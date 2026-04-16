package proxy

import (
	"io"
	"net/http"
	"time"
)

func isRetryable(status int) bool {
	return status >= 500 && status <= 599
}

func ExecuteWithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			// 网络错误可重试
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
				continue
			}
			return nil, err
		}
		if !isRetryable(resp.StatusCode) {
			return resp, nil // 非5xx状态码直接返回
		}
		// 5xx且还有重试次数
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		lastResp = resp
		lastErr = nil // 记录最后依次响应, 但最终会返回它

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Microsecond)
		}
	}

	// 所有重试失败，返回最后一次 5xx 响应
	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}
