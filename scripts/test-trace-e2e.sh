#!/bin/bash
# test-trace-e2e.sh - 端到端验证 Agent Trace 功能
set -e

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

echo "=== luner Agent Trace 端到端测试 ==="

LUNER_PID=""
MOCK_PID=""

cleanup() {
    [ -n "$LUNER_PID" ] && kill "$LUNER_PID" 2>/dev/null || true
    [ -n "$MOCK_PID" ]  && kill "$MOCK_PID"  2>/dev/null || true
    rm -f luner-test.yaml luner-test.db
}
trap cleanup EXIT

# 释放可能被上次异常中断占用的端口
lsof -ti tcp:18080 2>/dev/null | xargs kill -9 2>/dev/null || true
lsof -ti tcp:19999 2>/dev/null | xargs kill -9 2>/dev/null || true

# ── 1. 编译 ───────────────────────────────────────────────────────────────────
echo "⏳ 编译 luner..."
go build -o luner ./cmd/luner
echo "✓ 编译完成"

# ── 2. 清理旧数据 ─────────────────────────────────────────────────────────────
rm -f luner.db luner-test.db
echo "✓ 清理旧数据库"

# ── 3. 启动 Mock 上游服务 ──────────────────────────────────────────────────────
# 返回合法的 OpenAI 格式响应（含 usage），供 Span 收集使用
python3 - <<'PYEOF' &
import http.server, json, time

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *a): pass  # 静默
    def do_POST(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        resp = {
            "id": "mock-cmpl-001",
            "object": "chat.completion",
            "model": "gpt-4o-mini",
            "choices": [{"message": {"role": "assistant", "content": "mock response"}, "finish_reason": "stop"}],
            "usage": {"prompt_tokens": 12, "completion_tokens": 8, "total_tokens": 20}
        }
        self.wfile.write(json.dumps(resp).encode())

httpd = http.server.HTTPServer(("127.0.0.1", 19999), Handler)
httpd.serve_forever()
PYEOF

MOCK_PID=$!
sleep 0.5
echo "✓ Mock 上游启动 (PID: $MOCK_PID, :19999)"

# ── 4. 写入测试配置 ────────────────────────────────────────────────────────────
cat > luner-test.yaml <<'YAML'
version: v1
server:
  listen: ":18080"
  read_timeout: "10s"
  write_timeout: "30s"

providers:
  - name: mock-llm
    base_url: "http://127.0.0.1:19999"
    api_key: "test-key"
    models: ["gpt-4o-mini"]
    timeout: "5s"

storage:
  backend: sqlite
  sqlite:
    path: luner-test.db

cache:
  enabled: false

rate_limit:
  enabled: false
YAML
echo "✓ 测试配置已写入"

# ── 5. 启动 luner ─────────────────────────────────────────────────────────────
./luner -config luner-test.yaml &
LUNER_PID=$!

# 主动等待端口就绪（最多 5 秒），比固定 sleep 更可靠
READY=0
for i in $(seq 1 50); do
    if nc -z 127.0.0.1 18080 2>/dev/null; then
        READY=1; break
    fi
    sleep 0.1
done

if [ "$READY" -eq 0 ] || ! kill -0 "$LUNER_PID" 2>/dev/null; then
    echo "✗ luner 启动失败（端口 18080 未就绪）"
    exit 1
fi
echo "✓ luner 启动 (PID: $LUNER_PID, :18080)"

# ── 6. 测试 Mode 1（带 SDK Header）────────────────────────────────────────────
echo ""
echo "=== 测试 1: Mode 1（完整 Agent 上下文）==="
curl -s -X POST http://localhost:18080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-Luner-Agent: code-reviewer" \
  -H "X-Luner-Agent-Version: v2.1.3" \
  -H "X-Luner-User: alice" \
  -H "X-Luner-Tenant: acme-corp" \
  -H "X-Luner-Env: production" \
  -H "traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Review this code"}]}' \
  -o /dev/null

sleep 0.5
echo "✓ 请求已发送（Mode 1）"

# ── 7. 测试 Mode 0（零代码接入，无 Header）────────────────────────────────────
echo ""
echo "=== 测试 2: Mode 0（零代码接入）==="
curl -s -X POST http://localhost:18080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Hello"}]}' \
  -o /dev/null

sleep 0.5
echo "✓ 请求已发送（Mode 0）"

# ── 8. 数据库验证 ──────────────────────────────────────────────────────────────
echo ""
echo "=== 数据库验证 ==="
sqlite3 luner-test.db <<'EOF'
.headers on
.mode column

SELECT
  substr(span_id, 1, 8)   AS span_id,
  substr(trace_id, 1, 8)  AS trace_id,
  agent_name,
  user_id,
  model,
  prompt_tokens,
  completion_tokens,
  round(cost_usd, 6)      AS cost,
  status
FROM spans
ORDER BY created_at DESC;
EOF

# ── 9. 统计验证 ────────────────────────────────────────────────────────────────
echo ""
echo "=== 统计信息 ==="
SPAN_COUNT=$(sqlite3 luner-test.db "SELECT COUNT(*) FROM spans;")
echo "✓ Span 总数: $SPAN_COUNT"

MODE1_COUNT=$(sqlite3 luner-test.db "SELECT COUNT(*) FROM spans WHERE agent_name = 'code-reviewer';")
echo "✓ Mode 1 Span (agent=code-reviewer): $MODE1_COUNT"

MODE0_COUNT=$(sqlite3 luner-test.db "SELECT COUNT(*) FROM spans WHERE agent_name IS NULL OR agent_name = '';")
echo "✓ Mode 0 Span (无 agent): $MODE0_COUNT"

# ── 10. 断言 ──────────────────────────────────────────────────────────────────
echo ""
FAIL=0
if [ "$SPAN_COUNT" -lt 2 ]; then
    echo "✗ FAIL: 期望至少 2 个 Span，实际 $SPAN_COUNT"
    FAIL=1
fi
if [ "$MODE1_COUNT" -lt 1 ]; then
    echo "✗ FAIL: Mode 1 Span 数量为 0"
    FAIL=1
fi
if [ "$MODE0_COUNT" -lt 1 ]; then
    echo "✗ FAIL: Mode 0 Span 数量为 0"
    FAIL=1
fi
if [ "$FAIL" -eq 0 ]; then
    echo "✓ 所有断言通过"
fi

# ── 11. 清理 ──────────────────────────────────────────────────────────────────
kill "$LUNER_PID" 2>/dev/null
kill "$MOCK_PID" 2>/dev/null
rm -f luner-test.yaml luner-test.db
echo "✓ 测试完成，进程已停止"

exit "$FAIL"
