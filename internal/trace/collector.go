package trace

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

// Collector assembles LLM Spans from gateway request/response data and
// persists them without blocking the HTTP response path.
type Collector struct {
	store      storage.Storage
	logger     *slog.Logger
	dispatchFn func(func()) // swappable for synchronous dispatch in tests
}

// NewCollector returns a Collector backed by store.
// Returns nil if store is nil so callers can use nil-check guards.
func NewCollector(store storage.Storage, logger *slog.Logger) *Collector {
	if store == nil {
		return nil
	}
	return &Collector{
		store:      store,
		logger:     logger,
		dispatchFn: func(fn func()) { go fn() },
	}
}

// CollectLLMSpan builds a Span from the request/response pair and persists it
// via dispatchFn (asynchronous by default). Always returns nil; storage errors
// are logged but never propagated to the caller.
func (c *Collector) CollectLLMSpan(
	req *http.Request,
	resp *http.Response,
	requestBody, responseBody []byte,
	startTime, endTime time.Time,
) error {
	if c == nil {
		return nil
	}

	tCtx := ExtractContext(req.Header)
	tCtx.GenerateTraceID() // no-op if already set
	tCtx.GenerateSpanID()

	llmReq, err := ParseRequest(requestBody)
	if err != nil {
		c.logger.Debug("trace: could not parse request body", "err", err)
		llmReq = &OpenAIRequest{}
	}

	llmResp, err := ParseResponse(responseBody)
	if err != nil {
		llmResp = &OpenAIResponse{}
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	span := &storage.Span{
		SpanID:       tCtx.SpanID,
		TraceID:      tCtx.TraceID,
		ParentSpanID: tCtx.ParentSpanID,
		SessionID:    tCtx.SessionID,

		TenantID:     tCtx.TenantID,
		UserID:       tCtx.UserID,
		AgentName:    tCtx.AgentName,
		AgentVersion: tCtx.AgentVersion,
		Environment:  tCtx.Environment,

		SpanType:   "llm",
		Name:       "llm.chat.completion",
		StartTime:  startTime,
		EndTime:    endTime,
		DurationMs: endTime.Sub(startTime).Milliseconds(),

		Model:            llmReq.Model,
		PromptTokens:     int32(llmResp.Usage.PromptTokens),
		CompletionTokens: int32(llmResp.Usage.CompletionTokens),
		CostUSD: CalculateCost(
			llmReq.Model,
			llmResp.Usage.PromptTokens,
			llmResp.Usage.CompletionTokens,
		),

		Input:     string(requestBody),
		Output:    string(responseBody),
		Status:    spanStatus(statusCode),
		Tags:      tCtx.Tags,
		CreatedAt: time.Now(),
	}

	c.dispatchFn(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.store.Spans().Create(ctx, span); err != nil {
			c.logger.Error("trace: failed to store span",
				"span_id", span.SpanID,
				"trace_id", span.TraceID,
				"err", err,
			)
		} else {
			c.logger.Debug("trace: span stored",
				"span_id", span.SpanID,
				"trace_id", span.TraceID,
				"agent", span.AgentName,
				"model", span.Model,
				"cost_usd", span.CostUSD,
				"duration_ms", span.DurationMs,
			)
		}
	})

	return nil
}

func spanStatus(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return "success"
	}
	return "error"
}
