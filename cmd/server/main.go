package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
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

	// 管理 API：动态增删客户端 Key（生产环境建议加管理员鉴权）
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

	addr := ":" + itoa(cfg.Server.Port)
	log.Printf("AI 中转站已启动，监听 %s", addr)
	log.Fatal(http.ListenAndServe(addr, auth.Middleware(keyStore)(mux)))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
