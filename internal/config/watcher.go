package config

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

// 持有当前配置的原子指针
type Loader struct {
	cfg atomic.Pointer[Config]
}

func NewLoader(path string) (*Loader, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	l := &Loader{}
	l.cfg.Store(cfg)
	return l, nil
}

// 从内存中的Config创建Loader, 用于单元测试
// 避免测试依赖外部文件, 支持直接注入 Mock 配置
func NewLoaderFromCfg(cfg *Config) *Loader {
	if cfg == nil {
		// 防御性: 防止测试中传入 nil 导致 panic
		cfg = &Config{}
	}

	l := &Loader{}
	l.cfg.Store(cfg)
	return l
}

func (l *Loader) Get() *Config {
	return l.cfg.Load()
}

// Watch 监听文件变化并原子替换配置
func (l *Loader) Watch(ctx context.Context, path string, logger *slog.Logger) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				newCfg, err := Load(path)
				if err != nil {
					logger.Warn("reload config failed, keeping old", "err", err)
					continue
				}
				if err := newCfg.Validate(); err != nil {
					logger.Warn("reload config validation failed, keeping old", "err", err)
					continue
				}
				l.cfg.Store(newCfg)
				logger.Info("configuration reloaded",
					"summary", newCfg.Summary(),
					"routes", func() []map[string]any {
						var routes []map[string]any
						for _, p := range newCfg.Providers {
							routes = append(routes, map[string]any{
								"provider": p.Name,
								"models":   p.Models,
								"base_url": maskURL(p.BaseURL),
								"timeout":  p.Timeout,
							})
						}
						return routes
					}(),
				)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.Error("config watcher error", "err", err)
		}
	}
}
