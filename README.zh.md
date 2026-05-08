![Product Logo](./assets/logo/luner-logo-fat.png)

# luner

<p align="center">
  <a href="README.md">English</a> | <strong>中文</strong>
</p>

[![Release](https://img.shields.io/github/v/release/skylunna/luner?label=Release&color=blue)](https://github.com/skylunna/luner/releases)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://docs.docker.com/compose/)
[![License](https://img.shields.io/github/license/skylunna/luner?color=green)](https://github.com/skylunna/luner/blob/main/LICENSE)


**具有实时治理功能的AI网关** —— 在错误的LLM请求花费你的钱之前阻止它们。代理、缓存、速率限制，并通过OpenAI兼容的接口观察您的AI工作负载。
内置的CEL策略引擎在请求到达LLM提供商之前强制执行预算和模型分配。

---

![architecture](./assets/architecture/architecture-v0.5.0.png)

---

## ✨ 特性

### CEL 策略引擎 — 花钱前先拦截

基于 Google CEL 表达式的实时治理：

```json
// 超出预算直接拦截
{ "expression": "cost_usd > 10.0", "action": "block" }

// 高频请求自动降级到便宜模型
{ "expression": "request_count > 100 && model == 'gpt-4o'", "action": "downgrade" }

// 可疑用量触发告警
{ "expression": "tokens_used > 50000", "action": "alert" }
```

策略存储于 SQLite，热重载无需重启。可实现模型白名单、用户级消费上限或自定义路由逻辑。

---
### 兼容 OpenAI
零侵入替换 `base_url`，兼容任何 OpenAI 兼容 SDK。

### LRU 缓存
 — 零依赖内存缓存，支持 TTL 配置。仅缓存非流式请求；缓存键包含 `model + messages + temperature`。

### 令牌桶限流
按 Provider 配置 QPS + Burst，超限立即返回 429。

### 全面可观测性

OpenTelemetry 链路追踪（OTLP）+ Prometheus 指标，Span 级别的成本归因存储于 SQLite。

### 内置 Web 控制台

暗色主题 React SPA，与网关同端口提供服务。包含 Dashboard、Traces 浏览器、Policies 管理、Settings 查看器，无需额外部署。

### 配置热重载

`fsnotify` + `atomic.Pointer[Config]` 原子切换路由表，零停机。

### 云原生就绪

多架构二进制、多阶段 Dockerfile、开箱即用的 `docker-compose`。

---

## 🚀 快速开始

[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/skylunna/luner/releases)

### 方案 1：Demo 模式 — 一条命令，立即看到 Dashboard（推荐用于体验）

```bash
git clone https://github.com/skylunna/luner.git
cd luner

# 构建镜像并启动全部服务（mock LLM + luner + 示例数据）
docker compose up -d --build

# 等待约 30 秒完成镜像构建和数据初始化
docker compose logs -f seed-data   # 看到 "Demo data ready" 即可
```

打开 **http://localhost:8080**，即可看到预加载的 Trace 时间线、成本图表和策略列表。

```bash
# 验证
curl http://localhost:8080/api/health
curl http://localhost:8080/api/dashboard/summary

# 停止并清理
docker compose down            # 保留数据卷
docker compose down -v         # 同时删除数据库
```

### 方案 2：生产模式 — 接入真实 LLM Provider

```bash
cd deployments/production
export OPENAI_API_KEY=sk-...
docker compose -f docker-compose.prod.yml up -d --build
```

### 方案 3：源码编译

```bash
make build                          # 构建前端 + Go 二进制
./bin/luner --config config/config.example.yaml
```

### 方案 4：完整监控栈（Prometheus + Grafana + Tempo）

```bash
docker compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d
# Grafana:    http://localhost:3000  (admin / admin)
# Prometheus: http://localhost:9091
```

### 常见问题

| 现象 | 解决方法 |
|---|---|
| `seed-data` 立即退出 | 检查 `docker compose logs luner`，luner 可能仍在启动中 |
| 8080 端口被占用 | `lsof -ti:8080 \| xargs kill`，或修改端口映射 |
| 镜像构建失败（Go 代理） | `GOPROXY=https://goproxy.io,direct docker compose build` |
| Dashboard 无数据 | 执行 `docker compose run --rm seed-data` 重新写入示例数据 |

---

## 配置说明

`luner` 实现了路由逻辑与敏感信息的分离。随时修改 `config/config.yaml`，变更原子生效，无需重启。

```yaml
# config/config.yaml
providers:
  - name: openai-prod
    base_url: "https://api.openai.com/v1"
    api_key: "${OPENAI_API_KEY}"   # 从环境变量展开
    models: ["gpt-4o", "gpt-4o-mini"]
    timeout: "30s"

cache:
  enabled: true
  max_items: 5000
  default_ttl: "2h"

rate_limit:
  enabled: true
  providers:
    - name: openai-prod
      qps: 50.0
      burst: 10

storage:
  backend: sqlite
  sqlite:
    path: "data/luner.db"
```

> **热重载**：保存 `config.yaml` 后，网关原子切换路由表，不会断开任何活跃连接。  
> **例外**：`server.listen`、`read_timeout`、`write_timeout` 变更需重启进程才能生效。

---

## Web 控制台

Web 控制台是一个暗色主题的 React SPA，通过 **`http://localhost:8080/`** 访问。构建时嵌入 Go 二进制，无需独立部署。

| 页面 | 说明 |
|---|---|
| **Dashboard** | 汇总指标（Traces 数、Spans 数、成本、延迟、错误率）+ 4 张实时图表：延迟 P50/P99、请求数/小时、Token 消耗、Agent 成本 |
| **Traces** | 分页 Trace 列表，支持 Agent/User 筛选和状态 Tab 切换。点击任意 Trace 查看 Span 时间线 |
| **Trace Detail** | 完整 Span 树，含每个 Span 的 Token 数、成本、耗时及比例时间线 |
| **Policies** | 列出、创建、切换 CEL 策略。每条策略展示表达式、执行动作、优先级和启用状态 |
| **Settings** | 网关配置查看器（只读，提示热重载路径） |


### Demo
![demo](./assets/demo/demo-web-v0.5.0.png)

---

## CEL 策略引擎

策略是对每条入站请求执行求值的 CEL 表达式。匹配成功时触发三种动作之一：`block`（返回 403 拒绝）、`alert`（记录日志后放行）、`downgrade`（替换为更便宜的模型）。

**可用变量：**

| 变量 | 类型 | 说明 |
|---|---|---|
| `model` | string | 请求中的模型名 |
| `user_id` | string | `X-User-ID` 请求头 |
| `tenant_id` | string | `X-Tenant-ID` 请求头 |
| `request_count` | int | 该用户最近一分钟的请求次数 |
| `cost_usd` | double | 该用户最近一分钟的累计成本（美元） |
| `tokens_used` | int | 该用户最近一分钟消耗的 Token 数 |

**策略示例：**

```json
// 拦截不在白名单的模型
{ "name": "model-allowlist", "expression": "!(model in ['gpt-4o-mini', 'claude-haiku-4-5'])", "action": "block" }

// 单用户一分钟内消费超过 $0.10 时告警
{ "name": "spend-alert", "expression": "cost_usd > 0.10", "action": "alert" }

// 高频用户自动降级到更便宜的模型
{ "name": "auto-downgrade", "expression": "request_count > 100", "action": "downgrade" }
```

策略存储于 SQLite，可通过 REST API 或 Web 控制台管理，变更即时生效无需重启。

---

## API 接口

所有 REST 接口与代理、Web 控制台共用 `:8080` 端口。

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/api/health` | 健康检查（K8s 存活探针） |
| `GET` | `/api/dashboard/summary` | 汇总统计：Traces、Spans、成本、延迟、错误率 |
| `GET` | `/api/dashboard/cost` | 按 Agent 分类的成本明细 |
| `GET` | `/api/traces` | 分页 Trace 列表（`?page=1&page_size=20&agent_name=&user_id=`） |
| `GET` | `/api/traces/{trace_id}` | Trace 详情：汇总 + Span 树 + 时间线 |
| `GET` | `/api/policies` | 获取所有策略 |
| `POST` | `/api/policies` | 创建策略 |
| `GET` | `/api/policies/{id}` | 获取单条策略 |
| `PUT` | `/api/policies/{id}` | 更新策略 |
| `DELETE` | `/api/policies/{id}` | 删除策略 |
| `POST` | `/api/policies/reload` | 强制重新编译 CEL 程序 |
| `POST` | `/v1/chat/completions` | 代理接口（兼容 OpenAI） |
| `GET` | `/metrics` | Prometheus 指标 |

---

## 可观测性

### Prometheus 指标（`:9090/metrics`）

| 指标 | 标签 | 说明 |
|---|---|---|
| `luner_requests_total` | `provider`, `model`, `status` | 请求计数 |
| `luner_request_duration_seconds` | `provider`, `model` | 延迟直方图 |
| `luner_tokens_used_total` | `provider`, `model`, `type` | Token 统计（`prompt`/`completion`/`total`） |

### OpenTelemetry 链路追踪

设置 `OTEL_EXPORTER_OTLP_ENDPOINT` 即可将 Span 导出到任意 OTLP 兼容后端（Jaeger、Grafana Tempo、Honeycomb 等）。未设置时静默跳过，开发环境不会报错。

---

## 客户端集成

兼容任何支持 OpenAI 接口的客户端，只需更新 `base_url`：

```python
from openai import OpenAI

client = OpenAI(
    api_key="any-value",           # 转发到上游；真实 Key 在 config.yaml 中配置
    base_url="http://localhost:8080/v1"
)
response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "你好"}]
)
```

---

## 📈 性能基准测试

测试环境：**Ubuntu 22.04 / 8 核 vCPU / 16 GB 内存**  
测试工具：`hey -c 50 -n 1000` | [复现脚本](scripts/bench.sh)

| 场景 | QPS | P50 | P99 | 缓存命中 | 内存 |
|---|---|---|---|---|---|
| 缓存命中（`temp=0`，重复请求） | **32 082** | **1.3 ms** | **6.9 ms** | 100% | ~42 MB |
| 冷启动（首次请求） | ~95 | ~380 ms | ~1.1 s | 0% | ~45 MB |
| 限流（`qps=10, burst=2`） | ~10 | ~45 ms | ~180 ms | — | ~43 MB |

> 缓存命中从内存 LRU 直接返回，无任何网络开销。  
> 结果因操作系统调度器和 Docker 运行时不同而有所差异，请使用 `scripts/bench.sh` 在自己的环境测试。

---

## 参与贡献

欢迎提交 PR、Issue 和反馈。请查阅 [CONTRIBUTING.md](CONTRIBUTING.md) 了解环境搭建指引、提交规范以及 `good first issue` 标签的入门任务。
