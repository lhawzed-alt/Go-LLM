package router

import (
	"math/rand"
	"sync"

	"go_llm/internal/config"
)

// Router 根据模型名选择上游渠道，支持加权随机负载均衡
type Router struct {
	mu       sync.RWMutex
	modelMap map[string][]config.Channel
}

func NewRouter(channels []config.Channel) *Router {
	m := make(map[string][]config.Channel)
	for _, ch := range channels {
		for _, model := range ch.Models {
			m[model] = append(m[model], ch)
		}
	}
	return &Router{modelMap: m}
}

// Pick 返回支持该模型的渠道；多个渠道时按权重随机
func (r *Router) Pick(model string) (config.Channel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cands := r.modelMap[model]
	if len(cands) == 0 {
		return config.Channel{}, false
	}
	total := 0
	for _, c := range cands {
		w := c.Weight
		if w <= 0 {
			w = 1
		}
		total += w
	}
	n := rand.Intn(total)
	for _, c := range cands {
		w := c.Weight
		if w <= 0 {
			w = 1
		}
		if n < w {
			return c, true
		}
		n -= w
	}
	return cands[0], true
}

// Models 返回所有可用模型名
func (r *Router) Models() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	models := make([]string, 0, len(r.modelMap))
	for k := range r.modelMap {
		models = append(models, k)
	}
	return models
}

// Rebuild 重建路由表（用于配置热更新）
func (r *Router) Rebuild(channels []config.Channel) {
	m := make(map[string][]config.Channel)
	for _, ch := range channels {
		for _, model := range ch.Models {
			m[model] = append(m[model], ch)
		}
	}
	r.mu.Lock()
	r.modelMap = m
	r.mu.Unlock()
}
