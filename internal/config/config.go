package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int `yaml:"port" json:"port"`
}

type Channel struct {
	Name    string   `yaml:"name" json:"name"`
	BaseURL string   `yaml:"base_url" json:"base_url"`
	APIKey  string   `yaml:"api_key" json:"api_key"`
	Models  []string `yaml:"models" json:"models"`
	Weight  int      `yaml:"weight" json:"weight"`
}

type AuthConfig struct {
	APIKeys []string `yaml:"api_keys" json:"api_keys"`
}

type Config struct {
	Server   ServerConfig `yaml:"server" json:"server"`
	Channels []Channel    `yaml:"channels" json:"channels"`
	Auth     AuthConfig   `yaml:"auth" json:"auth"`
}

var envRe = regexp.MustCompile(`^\$\{([A-Z0-9_]+)\}$`)

// expandEnv 支持 "${VAR}" 形式的环境变量引用
func expandEnv(s string) string {
	if m := envRe.FindStringSubmatch(s); m != nil {
		return os.Getenv(m[1])
	}
	return s
}

// Load 读取并解析 YAML 配置，展开 ${ENV} 引用
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	for i := range cfg.Channels {
		cfg.Channels[i].APIKey = expandEnv(cfg.Channels[i].APIKey)
	}
	return cfg, nil
}

// Save 把配置以原始 YAML 结构写回磁盘。
// 重要：${ENV} 引用会保留原样（如果用户传入的就是 ${...} 字符串），否则按传入值原样写入。
// 为了让 ${ENV} 形式可被保留，写入前由调用方在内存中维持原值；本函数只负责序列化。
func Save(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置为空")
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	header := "# AI 模型中转站配置（由控制台写入）\n"
	if err := os.WriteFile(path, []byte(header+string(data)), 0o644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}
