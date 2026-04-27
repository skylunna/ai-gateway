#!/bin/bash

GATEWAY_URL="http://localhost:8080"
MODEL="qwen-turbo"
CONCURRENCY=50
REQUESTS=1000

# ==== 请求体 ====
PAYLOAD_CACHE="{\"model\":\"$MODEL\",\"messages\":[{\"role\":\"user\",\"content\":\"Explain Go context in 1 sentence\"}],\"temperature\":0}"
PAYLOAD_COLD="{\"model\":\"$MODEL\",\"messages\":[{\"role\":\"user\",\"content\":\"What is the meaning of life?\"}],\"temperature\":0.7}"

# ==== 颜色输出 ====
Green='\033[0;32m'
Yellow='\033[1;33m'
Red='\033[0;31m'
NC='\033[0m' # No Color

info() { echo -e "${Green}[INFO] $*${NC}"; }
warn() { echo -e "${Yellow}[WARN] $*${NC}"; }
error() { echo -e "${Red}[ERROR] $*${NC}"; exit 1; }

# ==== 检查依赖 ====
check_deps() {
  if ! command -v hey &> /dev/null; then
    error "hey not found. Install: go install github.com/rakyll/hey@latest"
  fi
}

# ==== 等待网关健康 ====
wait_gateway() {
  info "Waiting for luner gateway at $GATEWAY_URL..."
  for ((i=0; i<30; i++)); do
    if curl -s "$GATEWAY_URL/health" -o /dev/null -w "%{http_code}" | grep -q 200; then
      info "Gateway is healthy"
      return
    fi
    sleep 1
  done
  error "Gateway not ready after 30s"
}

# ==== 预热缓存 ====
warm_cache() {
  info "Warming up cache..."
  for ((i=0; i<3; i++)); do
    curl -s -X POST "$GATEWAY_URL/v1/chat/completions" \
      -H "Content-Type: application/json" \
      -d "$PAYLOAD_CACHE" \
      -o /dev/null
  done
}

# ==== 执行压测 ====
run_bench() {
  local NAME="$1"
  local PAYLOAD="$2"
  local OUTPUT_FILE="bench_$NAME.txt"

  info "Running $NAME benchmark ($CONCURRENCY concurrent, $REQUESTS requests)..."

  # 无 BOM 写入临时文件
  local TMP=$(mktemp)
  echo -n "$PAYLOAD" > "$TMP"

  # 执行 hey
  hey -c "$CONCURRENCY" -n "$REQUESTS" -m POST \
    -H "Content-Type: application/json" \
    -D "$TMP" \
    "$GATEWAY_URL/v1/chat/completions" > "$OUTPUT_FILE" 2>&1

  rm -f "$TMP"

  # 解析结果
  local CONTENT=$(cat "$OUTPUT_FILE")

  # 提取 QPS
  local QPS=$(echo "$CONTENT" | grep -oP 'Requests/sec:\s*\K[\d,.]+' | head -1 | tr -d ',')
  QPS=${QPS:-N/A}

  local P50=$(echo "$CONTENT" | grep -oP '50%?\s+in\s+\K[\d.]+' | head -1)
  local P99=$(echo "$CONTENT" | grep -oP '99%?\s+in\s+\K[\d.]+' | head -1)
  
  if [[ -z "$P50" ]]; then
    P50=$(echo "$CONTENT" | grep -oP '50\s+[\d.]+' | head -1 | awk '{print $2}')
  fi
  if [[ -z "$P99" ]]; then
    P99=$(echo "$CONTENT" | grep -oP '99\s+[\d.]+' | head -1 | awk '{print $2}')
  fi
  
  P50=${P50:-N/A}
  P99=${P99:-N/A}

  echo -e "  QPS: $QPS | P50: ${P50}s | P99: ${P99}s"

  if [[ $P50 == "N/A" ]]; then
    echo -e "\n  [DEBUG] Raw output snippet:"
    head -30 "$OUTPUT_FILE"
  fi

  echo "$QPS $P50 $P99"
}

# ==== 收集 metrics ====
collect_metrics() {
  info "Collecting Prometheus metrics..."
  local METRICS=$(curl -s "$GATEWAY_URL/metrics" 2>/dev/null)

  echo -e "\n Key Metrics:"
  echo "$METRICS" | grep -E 'luner_requests_total|luner_tokens_used' | head -10
}

# ==== 输出总结 ====
print_summary() {
  echo -e "\n Benchmark Summary"
  echo "==================="
  printf "%-20s %-12s %-12s %-12s\n" "Scenario" "QPS" "P50" "P99"
  printf "%-20s %-12s %-12s %-12s\n" "--------" "---" "---" "---"

  local CACHE_RESULT=$(run_bench "cache_hit" "$PAYLOAD_CACHE")
  local COLD_RESULT=$(run_bench "cold_start" "$PAYLOAD_COLD")

  printf "%-20s %-12s %-12s %-12s\n" " Cache Hit" $CACHE_RESULT
  printf "%-20s %-12s %-12s %-12s\n" " Cold Start" $COLD_RESULT

  echo -e ""
  echo " Tip: Cache Hit QPS is theoretical max for identical requests."
  echo "   Real-world workloads will see lower QPS but still benefit from repeated prompts."
}

# ==================== 主流程 ====================
info "Starting luner benchmark suite"
check_deps
wait_gateway
warm_cache
print_summary
collect_metrics
info "Benchmark complete!"