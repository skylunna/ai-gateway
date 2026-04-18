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



## [v0.2.0] - 2026-04-18

### ✨ Features
- OpenTelemetry tracing integration: auto-inject `traceparent` to upstream LLM requests, support OTLP HTTP exporter (Jaeger/Tempo compatible)
- Token usage metrics: parse `usage` from non-streaming responses, expose Prometheus counters `aigw_tokens_used_total{model,type="prompt|completion|total"}`
- Enhanced graceful shutdown: flush OTel traces & wait for in-flight requests before exit
- Release engineering ready: `.goreleaser.yaml` for multi-arch binaries, `Dockerfile` + `docker-compose.yml` for one-click deployment
- Debug endpoint `/debug/config` (dev mode): view current effective configuration

### 🛠️ Engineering
- Refactor `config.Loader` with `atomic.Pointer[Config]` for lock-free hot-reload
- Add `config.NewLoaderFromCfg()` for pure-memory config injection in unit tests
- Structured logging migration: full adoption of `log/slog` with JSON handler support
- Error response normalization: align with OpenAI spec `{"error":{"message":"xxx","type":"api_error"}}`
- Add `CONTRIBUTING.md` + `ROADMAP.md` to lower contribution barrier
- CI enhancement: auto-tagged releases trigger `goreleaser` to upload Assets

### 🐛 Fixed
- Fix missing `Content-Length` header in streaming responses causing client-side buffering
- Fix race condition during config hot-reload (atomic pointer swap ensures zero-downtime)
- Fix potential nil-pointer panic in tests when injecting mock config

### 📝 Known Issues / Limitations
- Server listen address & read/write timeouts still require restart to change (network layer state)
- Token metrics only parsed for **non-streaming** responses (`stream: false`); streaming token counting planned for v0.3.0
- No advanced routing strategies yet (weight, fallback, least-latency); pure model-name matching only
- OTel exporter is optional: if `OTEL_EXPORTER_OTLP_ENDPOINT` not set, tracing silently skips (dev-friendly)
- No built-in auth middleware (API key validation for gateway itself); rely on reverse proxy (nginx/Ingress) for now