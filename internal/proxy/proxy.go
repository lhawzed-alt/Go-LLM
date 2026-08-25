package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go_llm/internal/cache"
	"go_llm/internal/optimizer"
	"go_llm/internal/router"
)

// Proxy 将 OpenAI 兼容请求转发到上游渠道
type Proxy struct {
	Router      *router.Router
	Client      *http.Client
	RespCache   *cache.Cache // 非流式响应缓存
}

func New(rt *router.Router) *Proxy {
	return &Proxy{
		Router:    rt,
		Client:    &http.Client{Timeout: 5 * time.Minute},
		RespCache: cache.New(2048, 30*time.Minute),
	}
}

// chatRequest 用于提取 model / stream / messages 字段
type chatRequest struct {
	Model    string             `json:"model"`
	Stream   bool               `json:"stream"`
	Messages []optimizer.Message `json:"messages"`
}

// Handler 处理 /v1/chat/completions
func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	var req chatRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Model == "" {
		httpError(w, http.StatusBadRequest, "无效的请求体或缺少 model 字段")
		return
	}

	// 1) 优化消息：清理多余空白 + 注入简洁指令，减少输入/输出 token
	if len(req.Messages) > 0 {
		req.Messages = optimizer.OptimizeMessages(req.Messages)
		var err2 error
		body, err2 = json.Marshal(req)
		if err2 != nil {
			httpError(w, http.StatusInternalServerError, "序列化优化后的请求失败")
			return
		}
	}

	// 2) 非流式请求先查缓存，命中直接返回（提升缓存命中率）
	cacheKey := cache.Key(req.Model, body)
	if !req.Stream {
		if v, ok := p.RespCache.Get(cacheKey); ok {
			log.Printf("[cache] HIT key=%s rate=%.1f%%", cacheKey[:12], p.RespCache.Stats())
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.Write(v)
			return
		}
	}

	ch, ok := p.Router.Pick(req.Model)
	if !ok {
		httpError(w, http.StatusNotFound, fmt.Sprintf("没有可用渠道支持模型: %s", req.Model))
		return
	}

	upstreamURL := strings.TrimRight(ch.BaseURL, "/") + "/v1/chat/completions"
	// 兼容 base_url 已带 /v1 的情况，避免出现 /v1/v1/... 导致 404
	if strings.HasSuffix(strings.TrimRight(ch.BaseURL, "/"), "/v1") {
		upstreamURL = strings.TrimRight(ch.BaseURL, "/") + "/chat/completions"
	}
	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(body))
	if err != nil {
		httpError(w, http.StatusInternalServerError, "构造上游请求失败")
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Authorization", "Bearer "+ch.APIKey)
	upReq.Header.Set("Accept", r.Header.Get("Accept"))

	resp, err := p.Client.Do(upReq)
	if err != nil {
		httpError(w, http.StatusBadGateway, "上游请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 复制响应头（含流式相关）
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	if req.Stream && resp.StatusCode == http.StatusOK {
		streamCopy(w, resp.Body)
	} else {
		respBody, _ := io.ReadAll(resp.Body)
		w.Write(respBody)
		// 仅缓存成功的非流式响应
		if !req.Stream && resp.StatusCode == http.StatusOK {
			p.RespCache.Set(cacheKey, json.RawMessage(respBody))
			log.Printf("[cache] MISS key=%s rate=%.1f%%", cacheKey[:12], p.RespCache.Stats())
		}
	}
	log.Printf("[proxy] model=%s channel=%s status=%d stream=%v", req.Model, ch.Name, resp.StatusCode, req.Stream)
}

// streamCopy 边读边写，支持 SSE 流式转发并立即刷新
func streamCopy(w http.ResponseWriter, src io.Reader) {
	flusher, canFlush := w.(http.Flusher)
	buf := make([]byte, 4096)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return
			}
			if canFlush {
				flusher.Flush()
			}
		}
		if err != nil {
			return
		}
	}
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"message": msg, "type": "relay_error"},
	})
}
