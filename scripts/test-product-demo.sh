#!/bin/bash
# scripts/test-product-demo.sh

set -e


export NO_PROXY="localhost,127.0.0.1,::1"
export no_proxy="localhost,127.0.0.1,::1"

echo "🚀 luner v0.5.0 完整产品演示"
echo ""

# 0. 清理旧进程
echo "=== 清理环境 ==="
for port in 8080 9090 19999; do
    PID=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$PID" ]; then
        echo "✓ 清理端口 $port (PID: $PID)"
        kill -9 $PID 2>/dev/null || true
        sleep 0.5
    fi
done

# 1. 清理旧文件
rm -f luner.db luner.log config-demo.yaml
rm -f luner.db-shm luner.db-wal
echo "✓ 文件清理完成"

# 2. 构建
echo ""
echo "=== 构建产品 ==="
make build
echo "✓ 产品构建完成"

# 3. 启动 Mock LLM
echo ""
echo "=== 启动 Mock LLM ==="
python3 - <<'PYEOF' &
import http.server, json, time, sys

class H(http.server.BaseHTTPRequestHandler):
    def log_message(self, *a): pass
    def do_POST(self):
        try:
            time.sleep(0.5)
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({
                "id": "mock", "model": "gpt-4o-mini",
                "choices": [{"message": {"role": "assistant", "content": "Mock"}, "finish_reason": "stop"}],
                "usage": {"prompt_tokens": 15, "completion_tokens": 10, "total_tokens": 25}
            }).encode())
        except: pass

try:
    httpd = http.server.HTTPServer(("127.0.0.1", 19999), H)
    print("Mock LLM ready on :19999", file=sys.stderr)
    httpd.serve_forever()
except Exception as e:
    print(f"Failed: {e}", file=sys.stderr)
    sys.exit(1)
PYEOF

MOCK_PID=$!
sleep 2

if ! kill -0 $MOCK_PID 2>/dev/null; then
    echo "❌ Mock LLM 启动失败"
    exit 1
fi
echo "✓ Mock LLM 启动 (PID: $MOCK_PID)"

# 4. 配置
cat > config-demo.yaml <<EOF
server:
  port: 8080
storage:
  backend: sqlite
  sqlite:
    path: luner.db
providers:
  - name: mock-llm
    base_url: http://127.0.0.1:19999
    api_key: dummy
    timeout: 10s
    models: [gpt-4o-mini]
logging:
  level: info
EOF

echo "✓ 配置文件已创建"

# 5. 启动 luner
echo ""
echo "=== 启动 luner ==="
./bin/luner --config config-demo.yaml > luner.log 2>&1 &
LUNER_PID=$!
echo "✓ luner 启动中 (PID: $LUNER_PID)"

for i in {1..10}; do
    sleep 1
    if ! kill -0 $LUNER_PID 2>/dev/null; then
        echo "❌ luner 启动失败："
        cat luner.log
        kill $MOCK_PID 2>/dev/null || true
        exit 1
    fi
    
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        echo "✓ luner 启动成功"
        break
    fi
    
    if [ $i -eq 10 ]; then
        echo "❌ 超时"
        tail -20 luner.log
        kill $LUNER_PID $MOCK_PID 2>/dev/null || true
        exit 1
    fi
done

# 6. 安装并测试 SDK
echo ""
echo "=== 安装 Python SDK ==="
cd sdk/python

# 检测 Python 环境
if command -v uv &> /dev/null; then
    echo "使用 uv"
    
    # 方法 1: 使用 uv 的项目环境
    if [ ! -d ".venv" ]; then
        uv venv --quiet
    fi
    
    # 激活环境并安装
    source .venv/bin/activate
    uv pip install -e . --quiet
    
    # 验证安装
    python3 -c "import luner; print(f'✓ SDK version: {luner.__version__}')" || {
        echo "❌ SDK 安装失败"
        deactivate
        cd ../..
        kill $LUNER_PID $MOCK_PID 2>/dev/null || true
        exit 1
    }
    
    PYTHON_CMD="$(pwd)/.venv/bin/python3"
    # uv editable .pth isn't processed by Python's site module in some setups;
    # set PYTHONPATH so the venv Python can always find the luner package.
    export PYTHONPATH="$(pwd):${PYTHONPATH:-}"

elif command -v pip3 &> /dev/null; then
    echo "使用 pip3"
    pip3 install -e . --quiet --user
    python3 -c "import luner; print(f'✓ SDK version: {luner.__version__}')" || {
        echo "❌ SDK 安装失败"
        cd ../..
        kill $LUNER_PID $MOCK_PID 2>/dev/null || true
        exit 1
    }
    PYTHON_CMD="python3"
else
    echo "❌ 未找到 uv 或 pip3"
    cd ../..
    kill $LUNER_PID $MOCK_PID 2>/dev/null || true
    exit 1
fi

cd ../..

# 7. 生成示例数据
echo ""
echo "=== 生成示例 Trace 数据 ==="

# Agent 1
$PYTHON_CMD <<'PYEOF'
import luner, time, sys
from openai import OpenAI

try:
    luner.init(gateway_url="http://localhost:8080", agent_name="code-reviewer", agent_version="v2.1")
    luner.set_user(user_id="alice", tenant_id="acme-corp")
    client = luner.wrap(OpenAI(base_url="http://localhost:8080/v1", api_key="x"))
    
    for i in range(3):
        client.chat.completions.create(model="gpt-4o-mini", messages=[{"role": "user", "content": f"Review PR #{i+1}"}])
        time.sleep(0.5)
    
    print("✓ code-reviewer (alice): 3 traces")
except Exception as e:
    print(f"❌ {e}", file=sys.stderr)
    sys.exit(1)
PYEOF

# Agent 2
$PYTHON_CMD <<'PYEOF'
import luner, time, sys
from openai import OpenAI

try:
    luner.init(gateway_url="http://localhost:8080", agent_name="doc-writer", agent_version="v1.5")
    luner.set_user(user_id="bob", tenant_id="acme-corp")
    client = luner.wrap(OpenAI(base_url="http://localhost:8080/v1", api_key="x"))
    
    for i in range(2):
        client.chat.completions.create(model="gpt-4o-mini", messages=[{"role": "user", "content": f"Write docs {i+1}"}])
        time.sleep(0.5)
    
    print("✓ doc-writer (bob): 2 traces")
except Exception as e:
    print(f"❌ {e}", file=sys.stderr)
    sys.exit(1)
PYEOF

# Agent 3
$PYTHON_CMD <<'PYEOF'
import luner, time, sys
from openai import OpenAI

try:
    luner.init(gateway_url="http://localhost:8080", agent_name="test-generator", agent_version="v3.0")
    luner.set_user(user_id="alice", tenant_id="acme-corp")
    client = luner.wrap(OpenAI(base_url="http://localhost:8080/v1", api_key="x"))
    
    for i in range(2):
        with luner.span("analyze", tags={"type": "tool"}):
            time.sleep(0.1)
        client.chat.completions.create(model="gpt-4o-mini", messages=[{"role": "user", "content": f"Gen tests {i+1}"}])
        time.sleep(0.5)
    
    print("✓ test-generator (alice): 2 traces with spans")
except Exception as e:
    print(f"❌ {e}", file=sys.stderr)
    sys.exit(1)
PYEOF

echo ""
echo "⏳ 等待数据持久化..."
sleep 3

# 8. 验证
TOTAL=$(sqlite3 luner.db "SELECT COUNT(DISTINCT trace_id) FROM spans;" 2>/dev/null || echo "0")
SPAN_COUNT=$(sqlite3 luner.db "SELECT COUNT(*) FROM spans;" 2>/dev/null || echo "0")
COST=$(sqlite3 luner.db "SELECT ROUND(SUM(cost_usd), 6) FROM spans;" 2>/dev/null || echo "0")

if [ "$TOTAL" -eq "0" ]; then
    echo "⚠️  无数据，查看日志："
    tail -30 luner.log
else
    echo "✅ 数据生成完成: $TOTAL traces, $SPAN_COUNT spans"
fi

# 9. API 验证
echo ""
echo "=== API 验证 ==="
curl -s http://localhost:8080/api/dashboard/summary | python3 -m json.tool | head -8

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ luner v0.5.0 演示环境就绪"
echo ""
echo "🌐 Web 控制台: http://localhost:8080"
echo "   - Dashboard:  /dashboard"
echo "   - Traces:     /traces"
echo ""
echo "📊 当前数据:"
echo "   - Traces: $TOTAL"
echo "   - Spans: $SPAN_COUNT"
echo "   - Cost: \$$COST"
echo ""
echo "按 Ctrl+C 停止"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cleanup() {
    echo ""
    echo "🛑 停止中..."
    kill $LUNER_PID $MOCK_PID 2>/dev/null || true
    
    # 如果用了 uv venv，退出环境
    if [ -n "$VIRTUAL_ENV" ]; then
        deactivate 2>/dev/null || true
    fi
    
    echo "✓ 已停止"
}

trap cleanup INT TERM

while kill -0 $LUNER_PID 2>/dev/null; do
    sleep 1
done

cleanup