#!/bin/bash
# scripts/test-policy-demo.sh - 策略引擎演示

set -e

echo "🚀 luner v0.5.0 策略引擎演示"
echo ""

# 启动 luner（假设已运行）
if ! curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "❌ luner 未运行，请先运行："
    echo "   ./scripts/test-product-demo.sh"
    exit 1
fi

echo "=== 1. 创建预算策略 ==="
curl -s -X POST http://localhost:8080/api/policies \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Budget - $0.0001",
    "description": "Block requests when user spent >$0.0001 (demo threshold)",
    "type": "budget",
    "expression": "cost_usd > 0.0001",
    "action": "deny",
    "enabled": true,
    "priority": 100
  }' | python3 -m json.tool

echo ""
echo "=== 2. 重新加载策略 ==="
curl -s -X POST http://localhost:8080/api/policies/reload | python3 -m json.tool

echo ""
echo "=== 3. 发送第一个请求（应该成功）==="
RESP1=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-Luner-User: demo-user" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello"}]
  }')

HTTP_CODE1=$(echo "$RESP1" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$HTTP_CODE1" = "200" ]; then
    echo "✅ 第一个请求成功 (HTTP $HTTP_CODE1)"
else
    echo "❌ 第一个请求失败 (HTTP $HTTP_CODE1)"
fi

echo ""
echo "=== 4. 等待 3 秒（让统计缓存过期）==="
sleep 3

echo ""
echo "=== 5. 发送第二个请求（应该被拒绝）==="
RESP2=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-Luner-User: demo-user" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello again"}]
  }')

HTTP_CODE2=$(echo "$RESP2" | grep "HTTP_CODE:" | cut -d: -f2)
BODY2=$(echo "$RESP2" | grep -v "HTTP_CODE:")

if [ "$HTTP_CODE2" = "403" ]; then
    echo "✅ 第二个请求被策略拦截 (HTTP $HTTP_CODE2)"
    echo "响应: $BODY2"
else
    echo "⚠️  第二个请求未被拦截 (HTTP $HTTP_CODE2)"
    echo "可能需要调整策略阈值或等待统计更新"
fi

echo ""
echo "=== 6. 查询用户统计 ==="
sqlite3 luner.db <<EOF
SELECT 
  user_id,
  COUNT(*) as requests,
  SUM(cost_usd) as total_cost
FROM spans
WHERE user_id = 'demo-user'
GROUP BY user_id;
EOF

echo ""
echo "=== 7. 查看所有策略 ==="
curl -s http://localhost:8080/api/policies | python3 -m json.tool

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 策略引擎演示完成"
echo ""
echo "💡 你可以尝试："
echo "   - 调整策略表达式"
echo "   - 创建速率限制策略"
echo "   - 创建模型路由策略"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"