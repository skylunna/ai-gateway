package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/skylunna/luner/internal/metrics"
)

// StreamUsageParser 拦截 SSE 流，透传给客户端并提取 token usage
// flush=true 时强制每次 chunk 立即发送，保持低延迟体验
func StreamUsageParser(w http.ResponseWriter, resp *http.Response, model, provider string, flush bool) error {
	flusher, canFlush := w.(http.Flusher)
	reader := bufio.NewReader(resp.Body)

	var usageFound atomic.Bool
	var totalPrompt, totalCompletion int64

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read stream chunk: %w", err)
		}

		// 1. 透传 chunk 给客户端
		if _, err := w.Write(line); err != nil {
			return fmt.Errorf("write to client: %w", err)
		}

		// 2. 强制刷新（保持流式低延迟）
		if canFlush && flush {
			flusher.Flush()
		}

		// 3. 解析 data 行提取 usage
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "data: ") {
			data := strings.TrimSpace(strings.TrimPrefix(lineStr, "data: "))
			if data == "[DONE]" {
				break
			}

			// 优化：仅当未找到 usage 时解析 JSON，降低 CPU 开销
			if !usageFound.Load() {
				var chunk struct {
					Usage struct {
						PromptTokens     int `json:"prompt_tokens"`
						CompletionTokens int `json:"completion_tokens"`
						TotalTokens      int `json:"total_tokens"`
					} `json:"usage"`
				}
				if err := json.Unmarshal([]byte(data), &chunk); err == nil && chunk.Usage.TotalTokens > 0 {
					totalPrompt = int64(chunk.Usage.PromptTokens)
					totalCompletion = int64(chunk.Usage.CompletionTokens)

					metrics.TokensUsed.WithLabelValues(model, "prompt").Add(float64(totalPrompt))
					metrics.TokensUsed.WithLabelValues(model, "completion").Add(float64(totalCompletion))
					metrics.TokensUsed.WithLabelValues(model, "total").Add(float64(chunk.Usage.TotalTokens))

					usageFound.Store(true)
				}
			}
		}
	}

	// 4. 若流结束仍未获取 usage（上游不支持或配置缺失），记录兜底指标
	if !usageFound.Load() {
		metrics.TokensUsed.WithLabelValues(model, "unknown").Add(1)
	}
	return nil
}
