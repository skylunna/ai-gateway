# luner Python SDK

Python SDK for [luner](https://github.com/skylunna/luner) — the AI Control Plane.

Add observability to any LLM agent in 3 lines of code, with zero changes to your existing OpenAI calls.

## Installation

```bash
pip install luner
```

## Quick Start

```python
import luner
from openai import OpenAI

# 1. Initialize (one-time setup)
luner.init(
    gateway_url="http://localhost:8080",
    agent_name="my-agent",
    agent_version="v1.0",
)

# 2. Wrap your OpenAI client
client = OpenAI(base_url="http://localhost:8080/v1")
client = luner.wrap(client)

# 3. Use as normal — traces are automatic
response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello"}]
)
```

That's it. The luner Gateway now sees:

- ✅ Agent name and version
- ✅ Complete W3C trace chain
- ✅ Cost and token usage
- ✅ User and tenant attribution

## API Reference

### `luner.init()`

Initialize the SDK with your agent configuration.

| Parameter | Type | Description |
|-----------|------|-------------|
| `gateway_url` | `str` | luner Gateway URL |
| `agent_name` | `str` | Your agent's name |
| `agent_version` | `str` | Version identifier (default: `"default"`) |
| `environment` | `str` | `production` / `staging` / `dev` (default: `"production"`) |
| `user_id` | `str` | Optional end-user identifier |
| `tenant_id` | `str` | Optional tenant identifier |
| `session_id` | `str` | Optional session identifier |
| `tags` | `dict` | Optional free-form key-value tags |

### `luner.wrap(client)` / `luner.wrap_async(client)`

Wrap an `OpenAI` or `AsyncOpenAI` client to inject trace headers automatically.

```python
client = luner.wrap(OpenAI(base_url="..."))
async_client = luner.wrap_async(AsyncOpenAI(base_url="..."))
```

### `luner.set_user(user_id, tenant_id=None)`

Set user context after initialization — useful when identity is resolved after an auth step.

### `luner.set_tag(key, value)` / `luner.set_tags(tags)`

Add custom tags to the current trace.

### `luner.span(name, tags=None)`

Context manager that creates a child span for a block of work (Mode 2).

```python
with luner.span("retrieval") as s:
    s.tags["query"] = query
    results = search(query)
```

### `@luner.traced(name=None)`

Decorator that wraps a function in a luner span (Mode 2).

```python
@luner.traced("search-step")
def search(query: str) -> list:
    ...
```

## HTTP Headers

luner injects the following headers into every request made through a wrapped client:

| Header | Value |
|--------|-------|
| `traceparent` | W3C trace context: `00-{trace_id}-{span_id}-01` |
| `X-Luner-Agent` | Agent name |
| `X-Luner-Agent-Version` | Agent version |
| `X-Luner-User` | User ID |
| `X-Luner-Tenant` | Tenant ID |
| `X-Luner-Env` | Environment |
| `X-Luner-Session` | Session ID |
| `X-Luner-Tags` | Tags as `k1=v1,k2=v2` |

## Examples

See the [examples/](examples/) directory:

- [`basic_usage.py`](examples/basic_usage.py) — Mode 1: minimal integration
- [`with_spans.py`](examples/with_spans.py) — Mode 2: explicit span instrumentation
- [`langchain_example.py`](examples/langchain_example.py) — LangChain integration

## Requirements

- Python 3.9+
- `openai >= 1.0.0`

## License

MIT
