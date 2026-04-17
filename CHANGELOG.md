# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2026-04-17

### ✨ Features
- OpenAI-compatible `/v1/chat/completions` proxy endpoint
- YAML configuration with environment variable expansion (`${ENV_VAR}`)
- Model-based routing to multiple LLM providers
- Prometheus metrics (`/metrics`): request count, latency, HTTP status codes
- Zero-buffer SSE streaming passthrough (`io.Copy` for O(1) memory usage)
- Automatic retry on 5xx or network errors (configurable, default: 1)
- Hot-reload configuration via `fsnotify` (atomic swap, zero downtime)
- Graceful shutdown with `context` timeout (SIGINT/SIGTERM handling)

### 🛠️ Engineering
- Go 1.24 ready, standard project layout (`cmd/`, `internal/`, `config/`)
- GitHub Actions CI: `golangci-lint`, race-tested tests, cross-compile check
- Table-driven unit tests for proxy handler
- `Makefile` for common dev tasks (`make run/test/lint/build`)

### 📝 Known Issues / Limitations
- Server listen address & read/write timeouts require restart to change
- Only supports OpenAI `/v1/chat/completions` (streaming & non-streaming)
- No advanced routing strategies (weight, fallback, A/B) yet
- Token usage & cost metrics not parsed yet (planned for v0.2.0)