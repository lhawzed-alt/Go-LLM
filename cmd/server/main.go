package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go_llm/internal/auth"
	"go_llm/internal/config"
	"go_llm/internal/proxy"
	"go_llm/internal/router"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	rt := router.NewRouter(cfg.Channels)
	p := proxy.New(rt)
	keyStore := auth.NewStore(cfg.Auth.APIKeys)

	// 持有当前配置，支持运行时热更新
	var (
		mu   sync.RWMutex
		curC = cfg
	)

	// 热更新：重建 router 与 keyStore
	reload := func() {
		mu.Lock()
		defer mu.Unlock()
		rt.Rebuild(curC.Channels)
		keyStore.Replace(curC.Auth.APIKeys)
		log.Printf("[reload] channels=%d keys=%d", len(curC.Channels), len(curC.Auth.APIKeys))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", p.Handler)
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		type model struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		}
		var data []model
		for _, m := range rt.Models() {
			data = append(data, model{ID: m, Object: "model", Created: time.Now().Unix(), OwnedBy: "relay"})
		}
		json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"cache_hit_rate_percent": p.RespCache.Stats(),
		})
	})

	// 管理 API：动态增删客户端 Key
	mux.HandleFunc("/admin/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]any{"keys": keyStore.List()})
		case http.MethodPost:
			var body struct{ Key string `json:"key"` }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
				http.Error(w, `{"error":"需要 key 字段"}`, http.StatusBadRequest)
				return
			}
			keyStore.Add(body.Key)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"message": "已添加", "masked": keyStore.List()})
		case http.MethodDelete:
			var body struct{ Key string `json:"key"` }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
				http.Error(w, `{"error":"需要 key 字段"}`, http.StatusBadRequest)
				return
			}
			keyStore.Remove(body.Key)
			json.NewEncoder(w).Encode(map[string]any{"message": "已删除", "masked": keyStore.List()})
		default:
			http.Error(w, `{"error":"不支持的方法"}`, http.StatusMethodNotAllowed)
		}
	})

	// 管理 API：完整配置读写（仅内网使用）
	mux.HandleFunc("/admin/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			mu.RLock()
			defer mu.RUnlock()
			out := maskConfigForRead(curC)
			json.NewEncoder(w).Encode(out)
		case http.MethodPut:
			var incoming config.Config
			if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
				http.Error(w, `{"error":"JSON 解析失败"}`, http.StatusBadRequest)
				return
			}
			mu.Lock()
			// 合并：脱敏的 API Key（形如 "sk-xxx****xxxx"）保持原值
			merged := mergeConfig(curC, &incoming)
			// 写入磁盘（先备份）
			if err := backupFile(*cfgPath); err != nil {
				log.Printf("[warn] 备份配置文件失败: %v", err)
			}
			if err := config.Save(*cfgPath, merged); err != nil {
				mu.Unlock()
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
				return
			}
			// 展开环境变量引用
			for i := range merged.Channels {
				merged.Channels[i].APIKey = expandEnvLocal(merged.Channels[i].APIKey)
			}
			curC = merged
			mu.Unlock()
			reload()
			json.NewEncoder(w).Encode(map[string]any{
				"message": "已保存并热更新",
				"config":  maskConfigForRead(merged),
			})
		default:
			http.Error(w, `{"error":"不支持的方法"}`, http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("AI 中转站已启动，监听 %s", addr)
	// /admin/*、/healthz、/stats 不走客户端鉴权
	authedHandler := skipPaths(auth.Middleware(keyStore), []string{"/admin/", "/healthz", "/stats"})(mux)
	log.Fatal(http.ListenAndServe(addr, withCORS(authedHandler)))
}

// skipPaths 让指定前缀的路径直接放行（不经过鉴权）
func skipPaths(mw func(http.Handler) http.Handler, prefixes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		wrapped := mw(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range prefixes {
				if strings.HasPrefix(r.URL.Path, p) {
					next.ServeHTTP(w, r)
					return
				}
				// 同时支持不带结尾斜杠的精确匹配
				if r.URL.Path == strings.TrimSuffix(p, "/") {
					next.ServeHTTP(w, r)
					return
				}
			}
			wrapped.ServeHTTP(w, r)
		})
	}
}

// withCORS 允许控制台（任意 Origin）跨域访问管理接口
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
