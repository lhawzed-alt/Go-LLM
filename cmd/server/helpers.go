package main

import (
	"os"
	"regexp"
	"strings"

	"go_llm/internal/config"
)

// envRe 匹配 ${ENV_VAR} 形式的字符串
var envRe = regexp.MustCompile(`^\$\{([A-Z0-9_]+)\}$`)

// expandEnvLocal 与 config.expandEnv 相同；这里独立实现避免包内循环引用
func expandEnvLocal(s string) string {
	if m := envRe.FindStringSubmatch(s); m != nil {
		return os.Getenv(m[1])
	}
	return s
}

// backupFile 在覆盖前备份到 path.bak
func backupFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.WriteFile(path+".bak", data, 0o644)
}

// maskConfigForRead 返回配置的脱敏视图（API Key 仅显示前后几位）
func maskConfigForRead(c *config.Config) *config.Config {
	out := &config.Config{
		Server: c.Server,
		Auth: config.AuthConfig{
			APIKeys: make([]string, 0, len(c.Auth.APIKeys)),
		},
	}
	for _, k := range c.Auth.APIKeys {
		out.Auth.APIKeys = append(out.Auth.APIKeys, maskKey(k))
	}
	for _, ch := range c.Channels {
		nc := config.Channel{
			Name:    ch.Name,
			BaseURL: ch.BaseURL,
			APIKey:  maskKey(ch.APIKey),
			Models:  append([]string(nil), ch.Models...),
			Weight:  ch.Weight,
		}
		out.Channels = append(out.Channels, nc)
	}
	return out
}

// maskKey 与 auth.mask 规则一致
func maskKey(k string) string {
	if k == "" {
		return ""
	}
	if len(k) <= 8 {
		return "****"
	}
	return k[:6] + "****" + k[len(k)-4:]
}

// mergeConfig 把 incoming 合并到 current：API Key 若为脱敏形态（前缀 + "****" + 后缀长度匹配）
// 则保留 current 中的原值；否则采用 incoming 的新值。
// 通过 name 匹配 channel（顺序无关）。客户端 Key 列表按内容合并。
func mergeConfig(current, incoming *config.Config) *config.Config {
	out := &config.Config{
		Server: incoming.Server,
		Auth: config.AuthConfig{
			APIKeys: mergeKeyList(current.Auth.APIKeys, incoming.Auth.APIKeys),
		},
	}
	// 用 name 索引 current
	curByName := make(map[string]config.Channel, len(current.Channels))
	for _, ch := range current.Channels {
		curByName[ch.Name] = ch
	}
	for _, inc := range incoming.Channels {
		cur, ok := curByName[inc.Name]
		key := inc.APIKey
		if ok && isMaskedKey(inc.APIKey) {
			key = cur.APIKey
		}
		models := inc.Models
		if models == nil {
			models = []string{}
		}
		out.Channels = append(out.Channels, config.Channel{
			Name:    inc.Name,
			BaseURL: inc.BaseURL,
			APIKey:  key,
			Models:  models,
			Weight:  inc.Weight,
		})
	}
	return out
}

// isMaskedKey 判断字符串是否形如 "xxxxxx****yyyy"（脱敏形态）
func isMaskedKey(s string) bool {
	if !strings.Contains(s, "****") {
		return false
	}
	// 形如前缀(>=1) + **** + 后缀(>=1)，且前缀至少6位
	parts := strings.SplitN(s, "****", 2)
	return len(parts) == 2 && len(parts[0]) >= 1 && len(parts[1]) >= 1
}

// mergeKeyList 合并 Key 列表：incoming 的脱敏值（isMaskedKey 为 true）保留 current 对应原值；其他采用 incoming。
// 简单策略：以内容相等为同一 Key。
func mergeKeyList(current, incoming []string) []string {
	curSet := make(map[string]struct{}, len(current))
	for _, k := range current {
		curSet[k] = struct{}{}
	}
	out := make([]string, 0, len(incoming))
	seen := make(map[string]struct{}, len(incoming))
	for _, k := range incoming {
		if isMaskedKey(k) {
			// 脱敏：尝试在 current 中找一个尚未匹配且前缀后缀能对得上的
			matched := ""
			for ck := range curSet {
				if _, ok := seen[ck]; ok {
					continue
				}
				if maskKey(ck) == k {
					matched = ck
					break
				}
			}
			if matched != "" {
				out = append(out, matched)
				seen[matched] = struct{}{}
				continue
			}
			// 没找到：跳过
			continue
		}
		out = append(out, k)
		seen[k] = struct{}{}
	}
	return out
}
