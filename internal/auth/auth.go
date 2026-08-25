package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"go_llm/internal/config"
)

// Middleware 校验客户端 Bearer Key；未配置任何 Key 时放行
func Middleware(cfg config.AuthConfig) func(http.Handler) http.Handler {
	keys := cfg.APIKeys
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(keys) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			ok := false
			for _, k := range keys {
				if subtle.ConstantTimeCompare([]byte(token), []byte(k)) == 1 {
					ok = true
					break
				}
			}
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"message":"Invalid API key","type":"auth_error"}}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
