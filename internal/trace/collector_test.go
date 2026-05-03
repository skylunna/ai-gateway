package trace

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/skylunna/luner/internal/storage"
	"github.com/skylunna/luner/internal/storage/sqlite"
)

// newSyncCollector creates a Collector backed by an in-memory SQLite store.
// dispatchFn is overridden to run synchronously so tests don't need sleeps.
func newSyncCollector(t *testing.T) *Collector {
	t.Helper()
	store, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("sqlite.New: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	c := NewCollector(store, slog.Default())
	c.dispatchFn = func(fn func()) { fn() }
	return c
}

func TestExtractContext_Full(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	req.Header.Set("X-Luner-Agent", "my-agent")
	req.Header.Set("X-Luner-Agent-Version", "1.2.3")
	req.Header.Set("X-Luner-Session", "sess-abc")
	req.Header.Set("X-Luner-User", "user-42")
	req.Header.Set("X-Luner-Tenant", "tenant-99")
	req.Header.Set("X-Luner-Env", "production")
	req.Header.Set("X-Luner-Tags", "team=infra,region=us-east")

	tCtx := ExtractContext(req.Header)

	if tCtx.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("TraceID = %q", tCtx.TraceID)
	}
	if tCtx.SpanID != "00f067aa0ba902b7" {
		t.Errorf("SpanID = %q", tCtx.SpanID)
	}
	if !tCtx.Sampled {
		t.Error("Sampled should be true")
	}
	if tCtx.AgentName != "my-agent" {
		t.Errorf("AgentName = %q", tCtx.AgentName)
	}
	if tCtx.AgentVersion != "1.2.3" {
		t.Errorf("AgentVersion = %q", tCtx.AgentVersion)
	}
	if tCtx.SessionID != "sess-abc" {
		t.Errorf("SessionID = %q", tCtx.SessionID)
	}
	if tCtx.UserID != "user-42" {
		t.Errorf("UserID = %q", tCtx.UserID)
	}
	if tCtx.TenantID != "tenant-99" {
		t.Errorf("TenantID = %q", tCtx.TenantID)
	}
	if tCtx.Environment != "production" {
		t.Errorf("Environment = %q", tCtx.Environment)
	}
	if tCtx.Tags["team"] != "infra" {
		t.Errorf("Tags[team] = %q", tCtx.Tags["team"])
	}
	if tCtx.Tags["region"] != "us-east" {
		t.Errorf("Tags[region] = %q", tCtx.Tags["region"])
	}
}

func TestExtractContext_NoHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	tCtx := ExtractContext(req.Header)

	if tCtx == nil {
		t.Fatal("ExtractContext returned nil")
	}
	if tCtx.TraceID != "" || tCtx.AgentName != "" || tCtx.UserID != "" {
		t.Error("expected empty context for request with no headers")
	}
	if !tCtx.IsEmpty() {
		t.Error("IsEmpty() should be true")
	}
}

func TestGenerateIDs(t *testing.T) {
	tCtx := &Context{}
	tCtx.GenerateTraceID()
	tCtx.GenerateSpanID()

	if len(tCtx.TraceID) != 32 {
		t.Errorf("TraceID length = %d, want 32", len(tCtx.TraceID))
	}
	if len(tCtx.SpanID) != 16 {
		t.Errorf("SpanID length = %d, want 16", len(tCtx.SpanID))
	}

	// Idempotent
	traceID := tCtx.TraceID
	spanID := tCtx.SpanID
	tCtx.GenerateTraceID()
	tCtx.GenerateSpanID()
	if tCtx.TraceID != traceID {
		t.Error("GenerateTraceID should be idempotent")
	}
	if tCtx.SpanID != spanID {
		t.Error("GenerateSpanID should be idempotent")
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		model            string
		promptTokens     int
		completionTokens int
		wantMin, wantMax float64
	}{
		// gpt-4o: $2.50/1M in + $10.00/1M out → 1000*2.5/1M + 500*10/1M = 0.0025 + 0.005 = 0.0075
		{"gpt-4o", 1000, 500, 0.0074, 0.0076},
		// gpt-4o-mini: $0.15/1M in + $0.60/1M out → 0.00015 + 0.0003 = 0.00045
		{"gpt-4o-mini", 1000, 500, 0.00044, 0.00046},
		// unknown model falls back to gpt-4o rates
		{"unknown-model", 1000, 0, 0.0024, 0.0026},
		// deepseek-chat: $0.14/1M in + $0.28/1M out → 0.14 + 0.28 = 0.42
		{"deepseek-chat", 1_000_000, 1_000_000, 0.41, 0.43},
	}
	for _, tt := range tests {
		cost := CalculateCost(tt.model, tt.promptTokens, tt.completionTokens)
		if cost < tt.wantMin || cost > tt.wantMax {
			t.Errorf("CalculateCost(%q, %d, %d) = %f, want [%f, %f]",
				tt.model, tt.promptTokens, tt.completionTokens, cost, tt.wantMin, tt.wantMax)
		}
	}
}

func TestCollector_CollectLLMSpan(t *testing.T) {
	c := newSyncCollector(t)

	reqBody, _ := json.Marshal(map[string]any{
		"model":    "gpt-4o",
		"messages": []map[string]string{{"role": "user", "content": "hello"}},
	})
	respBody, _ := json.Marshal(map[string]any{
		"id":    "chatcmpl-abc",
		"model": "gpt-4o",
		"choices": []map[string]any{
			{"message": map[string]string{"role": "assistant", "content": "world"}},
		},
		"usage": map[string]int{
			"prompt_tokens":     10,
			"completion_tokens": 5,
			"total_tokens":      15,
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(reqBody))
	req.Header.Set("X-Luner-Agent", "test-agent")
	req.Header.Set("X-Luner-Tenant", "tenant-1")

	upstream := &http.Response{StatusCode: http.StatusOK}
	start := time.Now().Add(-100 * time.Millisecond)
	end := time.Now()

	if err := c.CollectLLMSpan(req, upstream, reqBody, respBody, start, end); err != nil {
		t.Fatalf("CollectLLMSpan returned error: %v", err)
	}

	spans, err := c.store.Spans().Query(context.Background(), storage.SpanFilter{Limit: 10})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	s := spans[0]
	if s.AgentName != "test-agent" {
		t.Errorf("AgentName = %q", s.AgentName)
	}
	if s.TenantID != "tenant-1" {
		t.Errorf("TenantID = %q", s.TenantID)
	}
	if s.Model != "gpt-4o" {
		t.Errorf("Model = %q", s.Model)
	}
	if s.PromptTokens != 10 {
		t.Errorf("PromptTokens = %d", s.PromptTokens)
	}
	if s.CompletionTokens != 5 {
		t.Errorf("CompletionTokens = %d", s.CompletionTokens)
	}
	if s.Status != "success" {
		t.Errorf("Status = %q, want 'success'", s.Status)
	}
	if s.CostUSD <= 0 {
		t.Error("CostUSD should be > 0")
	}
	if s.SpanID == "" {
		t.Error("SpanID should be set")
	}
	if s.TraceID == "" {
		t.Error("TraceID should be set")
	}
}

func TestCollector_NilStore(t *testing.T) {
	c := NewCollector(nil, slog.Default())
	if c != nil {
		t.Error("NewCollector(nil) should return nil")
	}

	// nil receiver must be safe
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	if err := c.CollectLLMSpan(req, nil, nil, nil, time.Now(), time.Now()); err != nil {
		t.Errorf("nil Collector.CollectLLMSpan returned error: %v", err)
	}
}

func TestCollector_ErrorResponse(t *testing.T) {
	c := newSyncCollector(t)

	reqBody, _ := json.Marshal(map[string]any{"model": "gpt-4o", "messages": []any{}})

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(reqBody))
	upstream := &http.Response{StatusCode: http.StatusUnauthorized}

	if err := c.CollectLLMSpan(req, upstream, reqBody, []byte(`{"error":"unauthorized"}`), time.Now(), time.Now()); err != nil {
		t.Fatalf("CollectLLMSpan: %v", err)
	}

	spans, err := c.store.Spans().Query(context.Background(), storage.SpanFilter{Limit: 10})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(spans) == 0 {
		t.Fatal("expected span to be stored")
	}
	if spans[0].Status != "error" {
		t.Errorf("Status = %q, want 'error'", spans[0].Status)
	}
}
