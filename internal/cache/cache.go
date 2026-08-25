package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// Cache 非流式响应的 LRU + TTL 缓存。
// 命中率统计用于监控缓存效果。
type Cache struct {
	mu       sync.Mutex
	items    map[string]*entry
	order    []string // 简易 LRU：按插入顺序，超容量时淘汰最旧
	capacity int
	ttl      time.Duration

	hits, misses uint64
}

type entry struct {
	value     json.RawMessage
	expiresAt time.Time
}

func New(capacity int, ttl time.Duration) *Cache {
	if capacity <= 0 {
		capacity = 1024
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Cache{
		items:    make(map[string]*entry),
		capacity: capacity,
		ttl:      ttl,
	}
}

// Key 根据模型与请求体生成缓存键
func Key(model string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(model))
	h.Write([]byte{0})
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Cache) Get(k string) (json.RawMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[k]
	if !ok || time.Now().After(e.expiresAt) {
		if ok {
			delete(c.items, k)
			c.removeOrder(k)
		}
		c.misses++
		return nil, false
	}
	c.hits++
	return e.value, true
}

func (c *Cache) Set(k string, v json.RawMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.items[k]; !exists && len(c.items) >= c.capacity {
		// 淘汰最旧
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.items, oldest)
	}
	if _, exists := c.items[k]; !exists {
		c.order = append(c.order, k)
	}
	c.items[k] = &entry{value: v, expiresAt: time.Now().Add(c.ttl)}
}

func (c *Cache) removeOrder(k string) {
	for i, o := range c.order {
		if o == k {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

// Stats 返回命中率（0-100）
func (c *Cache) Stats() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total) * 100
}
