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

	"go_llm/internal/router"
)

// Proxy 将 OpenAI 兼容请求转发到上游渠道
type Proxy struct {
	Router *router.Router
	Client *http.Client
}

func New(rt *router.Router) *Proxy {
	return &Proxy{
		Router: rt,
		Client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// chatRequest 仅用于提取 model 字段
type chatRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
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
		io.Copy(w, resp.Body)
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
