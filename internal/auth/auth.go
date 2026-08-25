package auth

import (
	"net/http"
	"strings"
	"sync"
)

// Store 管理多个客户端 API Key，支持运行时动态增删（多人共用网关，每人一个 Key）
type Store struct {
	mu   sync.RWMutex
	keys map[string]struct{}
}

func NewStore(keys []string) *Store {
	s := &Store{keys: make(map[string]struct{}, len(keys))}
	for _, k := range keys {
		if k != "" {
			s.keys[k] = struct{}{}
		}
	}
	return s
}

// Validate 校验 token 是否有效；未配置任何 Key 时放行
func (s *Store) Validate(token string) bool {
	if s == nil || len(s.keys) == 0 {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.keys[token]
	return ok
}

// Add 添加一个新 Key
func (s *Store) Add(key string) {
	if key == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[key] = struct{}{}
}

// Remove 删除一个 Key（泄露/人员变动时吊销）
func (s *Store) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keys, key)
}

// List 返回所有 Key 的脱敏列表（用于管理接口展示）
func (s *Store) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.keys))
	for k := range s.keys {
		out = append(out, mask(k))
	}
	return out
}

func mask(k string) string {
	if len(k) <= 8 {
		return "****"
	}
	return k[:6] + "****" + k[len(k)-4:]
}

// Middleware 基于 Store 的鉴权中间件
func Middleware(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !store.Validate(token) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"message":"Invalid API key","type":"auth_error"}}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
