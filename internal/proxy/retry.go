package proxy

import (
	"io"
	"log/slog"
	"net/http"
	"time"
)

func isRetryable(status int) bool {
	return status >= 500 && status <= 599
}

// ExecuteWithRetry executes the request with up to maxRetries retries on network errors or 5xx responses.
// logger may be nil.
func ExecuteWithRetry(client *http.Client, req *http.Request, maxRetries int, logger *slog.Logger) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				if logger != nil {
					logger.Debug("retrying upstream request", "attempt", attempt+1, "err", err)
				}
				time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
				continue
			}
			return nil, err
		}
		if !isRetryable(resp.StatusCode) {
			return resp, nil
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		lastResp = resp
		lastErr = nil

		if attempt < maxRetries {
			if logger != nil {
				logger.Debug("upstream returned 5xx, retrying", "attempt", attempt+1, "status", resp.StatusCode)
			}
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}

	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}
