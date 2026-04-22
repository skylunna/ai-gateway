package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode 统一错误码枚举
type ErrorCode string

const (
	ErrInvalidRequest ErrorCode = "INVALID_REQUEST"
	ErrModelNotFound  ErrorCode = "MODEL_NOT_FOUND"
	ErrProviderDown   ErrorCode = "PROVIDER_DOWN"
	ErrRateLimited    ErrorCode = "RATE_LIMITED"
	ErrCacheMiss      ErrorCode = "CACHE_MISS" // 仅日志，不返回客户端
	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrTimeout        ErrorCode = "UPSTREAM_TIMEOUT"
)

// ErrorDetail 标准化错误响应体（兼容 OpenAI 格式）
type ErrorDetail struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`            // 固定 "luner_error"
	Param     *string   `json:"param,omitempty"` // 可选：哪个参数出错
	RequestID string    `json:"request_id"`      // 关联 TraceID，便于排查
}

// APIError 实现 error 接口，便于内部传递
type APIError struct {
	Detail ErrorDetail
	Status int // HTTP 状态码
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Detail.Code, e.Detail.Message)
}

// WriteJSON 将错误写入 http.ResponseWriter
func (e *APIError) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Request-Id", e.Detail.RequestID) // 注入 TraceID
	w.WriteHeader(e.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": e.Detail,
	})
}

// NewError 快速创建标准化错误
func NewError(status int, code ErrorCode, message string, requestID string, param ...string) *APIError {
	var p *string
	if len(param) > 0 && param[0] != "" {
		p = &param[0]
	}
	return &APIError{
		Status: status,
		Detail: ErrorDetail{
			Code:      code,
			Message:   message,
			Type:      "luner_error",
			Param:     p,
			RequestID: requestID,
		},
	}
}

// FromUpstream 将上游错误转换为标准化错误
func FromUpstream(status int, body []byte, requestID string) *APIError {
	// 尝试解析上游 OpenAI 风格错误
	var upstream struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &upstream); err == nil && upstream.Error.Message != "" {
		return NewError(status, ErrProviderDown, upstream.Error.Message, requestID)
	}
	// 降级：返回原始响应片段
	msg := fmt.Sprintf("upstream returned %d: %s", status, string(body)[:min(100, len(body))])
	return NewError(status, ErrProviderDown, msg, requestID)
}
