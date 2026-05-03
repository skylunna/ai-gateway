package config

/*
	配置解析(支持 YAML + 环境变量插值)
*/
import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

/*
version: 1.0.0
server:
listen: :8080
providers:
  - name: openai
    base_url: https://api.openai.com
*/
type Config struct {
	Version   string           `yaml:"version"`
	Server    ServerConfig     `yaml:"server"`
	Providers []ProviderConfig `yaml:"providers"`
	Cache     CacheConfig      `yaml:"cache"`
	RateLimit RateLimitConfig  `yaml:"rate_limit"`
	Storage   StorageConfig    `yaml:"storage"`
}

// StorageConfig selects the persistence backend.
// Leave Backend empty to disable persistent storage.
type StorageConfig struct {
	Backend string       `yaml:"backend"` // "sqlite" | "" (disabled)
	SQLite  SQLiteConfig `yaml:"sqlite"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"` // default: "luner.db"
}

type CacheConfig struct {
	Enabled  bool          `yaml:"enabled"`     // 是否启用缓存
	MaxItems int           `yaml:"max_items"`   // 最大缓存条数
	TTL      time.Duration `yaml:"default_ttl"` // 默认缓存过期时间
}

type RateLimitConfig struct {
	Enabled   bool                `yaml:"enabled"`   // 是否启用速率限制
	Providers []ProviderRateLimit `yaml:"providers"` // 每个厂商的速率限制配置
}

type ProviderRateLimit struct {
	Name  string  `yaml:"name"`  // 厂商名
	QPS   float64 `yaml:"qps"`   // 每秒请求数
	Burst int     `yaml:"burst"` // 突发请求数
}

// 网关 HTTP 服务的配置。
type ServerConfig struct {
	Listen       string        `yaml:"listen"`       // 监听地址 :8080
	ReadTimeout  time.Duration `yaml:"read_timeout"` // 读超时
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

/*
厂商配置 ProviderConfig
一个 Provider = 一个大模型厂商（通义千问、DeepSeek、OpenAI 等）
*/
type ProviderConfig struct {
	Name    string        `yaml:"name"`     // 厂商名 openai / qwen / doubao
	BaseURL string        `yaml:"base_url"` // 上游接口地址
	APIKey  string        `yaml:"api_key"`  // 密钥
	Models  []string      `yaml:"models"`   // 支持的模型列表
	Timeout time.Duration `yaml:"timeout"`  // 请求超时
}

/*
Load 读取 YAML 并展开环境变量

input: 配置文件路径
output: 配置结构体 / 错误
*/
func Load(path string) (*Config, error) {
	// 读取文件, 把config.yaml读取成字节
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// 展开 ${ENV_VAR}
	// 自动把 YAML 里的 ${VAR} 替换成系统环境变量
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// 设置默认值
	if cfg.Server.Listen == "" {
		cfg.Server.Listen = ":8080"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 120 * time.Second
	}
	if cfg.Cache.MaxItems == 0 {
		cfg.Cache.MaxItems = 5000
	}
	if cfg.Cache.TTL == 0 {
		cfg.Cache.TTL = 2 * time.Hour
	}
	if cfg.Storage.SQLite.Path == "" {
		cfg.Storage.SQLite.Path = "luner.db"
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Providers) == 0 {
		return fmt.Errorf("config: at least one provider is required")
	}
	for i, p := range c.Providers {
		if p.Name == "" {
			return fmt.Errorf("config: providers[%d].name is required", i)
		}
		if p.BaseURL == "" {
			return fmt.Errorf("config: providers[%d].base_url is required", i)
		}
		if p.APIKey == "" {
			return fmt.Errorf("config: providers[%d].api_key is required (check env vars)", i)
		}
		if len(p.Models) == 0 {
			return fmt.Errorf("config: providers[%d].models must contain at least one model", i)
		}
	}
	return nil
}

// String 返回 Provider 的精简描述（用于日志）
func (p *ProviderConfig) String() string {
	return fmt.Sprintf("%s(models=%v, base_url=%s, timeout=%s)",
		p.Name,
		p.Models,
		maskURL(p.BaseURL), // 脱敏：隐藏 API Key 参数
		p.Timeout,
	)
}

// MaskURL 脱敏 URL 中的敏感参数
func maskURL(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx] + "?***"
	}
	return url
}

func (c *Config) MaskURL(url string) string {
	return maskURL(url)
}

// Summary 返回配置摘要（用于启动日志）
func (c *Config) Summary() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("providers=%d, cache=%v, rate_limit=%v",
		len(c.Providers),
		c.Cache.Enabled,
		c.RateLimit.Enabled,
	))
	if len(c.Providers) > 0 {
		sb.WriteString(" [")
		for i, p := range c.Providers {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s:%v", p.Name, p.Models))
		}
		sb.WriteString("]")
	}
	return sb.String()
}
