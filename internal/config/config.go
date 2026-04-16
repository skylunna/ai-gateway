package config

/*
	配置解析(支持 YAML + 环境变量插值)
*/
import (
	"fmt"
	"os"
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

	return &cfg, nil
}
