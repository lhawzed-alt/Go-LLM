package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int `yaml:"port"`
}

type Channel struct {
	Name     string   `yaml:"name"`
	BaseURL  string   `yaml:"base_url"`
	APIKey   string   `yaml:"api_key"`
	Models   []string `yaml:"models"`
	Weight   int      `yaml:"weight"`
}

type AuthConfig struct {
	APIKeys []string `yaml:"api_keys"`
}

type Config struct {
	Server   ServerConfig `yaml:"server"`
	Channels []Channel    `yaml:"channels"`
	Auth     AuthConfig   `yaml:"auth"`
}

var envRe = regexp.MustCompile(`^\$\{([A-Z0-9_]+)\}$`)

// expandEnv 支持 "${VAR}" 形式的环境变量引用
func expandEnv(s string) string {
	if m := envRe.FindStringSubmatch(s); m != nil {
		return os.Getenv(m[1])
	}
	return s
}

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
