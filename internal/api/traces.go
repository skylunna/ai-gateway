package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

// TraceListResponse is the paginated trace list response.
type TraceListResponse struct {
	Traces     []TraceItem `json:"traces"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

// TraceItem is a single row in the trace list — aggregated from spans sharing a trace_id.
type TraceItem struct {
	TraceID      string    `json:"trace_id"`
	AgentName    string    `json:"agent_name,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	Model        string    `json:"model,omitempty"`
	SpanCount    int       `json:"span_count"`
	TotalCostUSD float64   `json:"total_cost_usd"`
	DurationMs   int64     `json:"duration_ms"`
	StartTime    time.Time `json:"start_time"`
	Status       string    `json:"status"`
}

// handleListTraces handles GET /api/traces
func (s *RestServer) handleListTraces(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := storage.SpanFilter{
		AgentName: r.URL.Query().Get("agent_name"),
		UserID:    r.URL.Query().Get("user_id"),
		Limit:     pageSize * 10, // fetch more spans to aggregate into traces
		Offset:    0,
	}

	spans, err := s.store.Spans().Query(r.Context(), filter)
	if err != nil {
		s.logger.Error("list traces: query failed", "err", err)
		s.errJSON(w, http.StatusInternalServerError, "failed to query traces")
		return
	}

	// Aggregate spans → traces, preserving insertion order via a slice.
	order := []string{}
	traceMap := map[string]*TraceItem{}
	for _, sp := range spans {
		item, ok := traceMap[sp.TraceID]
		if !ok {
			item = &TraceItem{
				TraceID:   sp.TraceID,
				AgentName: sp.AgentName,
				UserID:    sp.UserID,
				Model:     sp.Model,
				StartTime: sp.StartTime,
				Status:    "success",
			}
			traceMap[sp.TraceID] = item
			order = append(order, sp.TraceID)
		}
		item.SpanCount++
		item.TotalCostUSD += sp.CostUSD
		if sp.ParentSpanID == "" {
			item.DurationMs = sp.DurationMs
		}
		if sp.Status == "error" {
			item.Status = "error"
		}
	}

	// Paginate over aggregated traces.
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(order) {
		start = len(order)
	}
	if end > len(order) {
		end = len(order)
	}

	page_items := make([]TraceItem, 0, end-start)
	for _, id := range order[start:end] {
		page_items = append(page_items, *traceMap[id])
	}

	s.json(w, http.StatusOK, TraceListResponse{
		Traces:     page_items,
		TotalCount: len(order),
		Page:       page,
		PageSize:   pageSize,
	})
}

// SpanNode is a tree node in the trace detail response.
type SpanNode struct {
	SpanID       string    `json:"span_id"`
	ParentSpanID string    `json:"parent_span_id,omitempty"`
	Name         string    `json:"name"`
	SpanType     string    `json:"span_type"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	DurationMs   int64     `json:"duration_ms"`

	// Milliseconds relative to the trace start — drives the frontend timeline bars.
	RelativeStartMs int64 `json:"relative_start_ms"`
	RelativeEndMs   int64 `json:"relative_end_ms"`

	Model            string  `json:"model,omitempty"`
	PromptTokens     int32   `json:"prompt_tokens,omitempty"`
	CompletionTokens int32   `json:"completion_tokens,omitempty"`
	CostUSD          float64 `json:"cost_usd,omitempty"`

	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`

	Children []SpanNode `json:"children,omitempty"`
}

// Timeline carries the absolute start/end and total duration of a trace.
type Timeline struct {
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	DurationMs int64     `json:"duration_ms"`
}

// TraceSummary rolls up key metrics for a trace.
type TraceSummary struct {
	TotalSpans   int     `json:"total_spans"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	DurationMs   int64   `json:"duration_ms"`
	Status       string  `json:"status"`
}

// TraceDetailResponse is the full detail for a single trace.
type TraceDetailResponse struct {
	TraceID  string       `json:"trace_id"`
	Spans    []SpanNode   `json:"spans"`
	Summary  TraceSummary `json:"summary"`
	Timeline Timeline     `json:"timeline"`
}

// handleGetTrace handles GET /api/traces/{trace_id}
func (s *RestServer) handleGetTrace(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("trace_id")
	if traceID == "" {
		s.errJSON(w, http.StatusBadRequest, "trace_id is required")
		return
	}

	spans, err := s.store.Spans().ListByTrace(r.Context(), traceID)
	if err != nil {
		s.logger.Error("get trace: query failed", "trace_id", traceID, "err", err)
		s.errJSON(w, http.StatusInternalServerError, "failed to get trace")
		return
	}
	if len(spans) == 0 {
		s.errJSON(w, http.StatusNotFound, "trace not found")
		return
	}

	timeline := calculateTimeline(spans)
	tree := buildSpanTree(spans, timeline.StartTime)
	summary := calculateSummary(spans, timeline)

	s.json(w, http.StatusOK, TraceDetailResponse{
		TraceID:  traceID,
		Spans:    tree,
		Summary:  summary,
		Timeline: timeline,
	})
}

// buildSpanTree converts a flat span list into a tree rooted at spans with no
// known parent. Children are sorted by start time. O(n²) — fine for trace sizes.
func buildSpanTree(spans []*storage.Span, traceStart time.Time) []SpanNode {
	inTrace := make(map[string]bool, len(spans))
	for _, sp := range spans {
		inTrace[sp.SpanID] = true
	}

	var build func(sp *storage.Span) SpanNode
	build = func(sp *storage.Span) SpanNode {
		node := SpanNode{
			SpanID:           sp.SpanID,
			ParentSpanID:     sp.ParentSpanID,
			Name:             sp.Name,
			SpanType:         sp.SpanType,
			StartTime:        sp.StartTime,
			EndTime:          sp.EndTime,
			DurationMs:       sp.DurationMs,
			RelativeStartMs:  sp.StartTime.Sub(traceStart).Milliseconds(),
			RelativeEndMs:    sp.EndTime.Sub(traceStart).Milliseconds(),
			Model:            sp.Model,
			PromptTokens:     sp.PromptTokens,
			CompletionTokens: sp.CompletionTokens,
			CostUSD:          sp.CostUSD,
			Status:           sp.Status,
			ErrorMessage:     sp.ErrorMessage,
		}
		for _, child := range spans {
			if child.ParentSpanID == sp.SpanID {
				node.Children = append(node.Children, build(child))
			}
		}
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].StartTime.Before(node.Children[j].StartTime)
		})
		return node
	}

	var roots []SpanNode
	for _, sp := range spans {
		if sp.ParentSpanID == "" || !inTrace[sp.ParentSpanID] {
			roots = append(roots, build(sp))
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].StartTime.Before(roots[j].StartTime)
	})
	return roots
}

func calculateTimeline(spans []*storage.Span) Timeline {
	if len(spans) == 0 {
		return Timeline{}
	}
	start := spans[0].StartTime
	end := spans[0].EndTime
	for _, sp := range spans[1:] {
		if sp.StartTime.Before(start) {
			start = sp.StartTime
		}
		if sp.EndTime.After(end) {
			end = sp.EndTime
		}
	}
	return Timeline{
		StartTime:  start,
		EndTime:    end,
		DurationMs: end.Sub(start).Milliseconds(),
	}
}

func calculateSummary(spans []*storage.Span, tl Timeline) TraceSummary {
	s := TraceSummary{TotalSpans: len(spans), DurationMs: tl.DurationMs, Status: "success"}
	for _, sp := range spans {
		s.TotalCostUSD += sp.CostUSD
		if sp.Status == "error" {
			s.Status = "error"
		}
	}
	return s
}
