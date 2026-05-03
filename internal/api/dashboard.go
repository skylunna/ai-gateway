package api

import (
	"net/http"
	"sort"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

// DashboardSummary holds 24-hour aggregate metrics.
type DashboardSummary struct {
	TotalTraces  int     `json:"total_traces"`
	TotalSpans   int     `json:"total_spans"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
}

func (s *RestServer) handleDashboardSummary(w http.ResponseWriter, r *http.Request) {
	spans, err := s.store.Spans().Query(r.Context(), storage.SpanFilter{
		StartTime: time.Now().Add(-24 * time.Hour),
		Limit:     10000,
	})
	if err != nil {
		s.errJSON(w, http.StatusInternalServerError, "failed to query spans")
		return
	}

	var (
		summary     DashboardSummary
		traceSet    = map[string]struct{}{}
		errorCount  int
		totalDurMs  int64
	)
	for _, sp := range spans {
		summary.TotalSpans++
		summary.TotalCostUSD += sp.CostUSD
		totalDurMs += sp.DurationMs
		traceSet[sp.TraceID] = struct{}{}
		if sp.Status == "error" {
			errorCount++
		}
	}
	summary.TotalTraces = len(traceSet)
	if summary.TotalSpans > 0 {
		summary.AvgLatencyMs = float64(totalDurMs) / float64(summary.TotalSpans)
		summary.ErrorRate = float64(errorCount) / float64(summary.TotalSpans)
	}

	s.json(w, http.StatusOK, summary)
}

// CostBreakdown groups cost by agent and by user.
type CostBreakdown struct {
	ByAgent []CostItem `json:"by_agent"`
	ByUser  []CostItem `json:"by_user"`
}

// CostItem is a single cost group.
type CostItem struct {
	Name  string  `json:"name"`
	Cost  float64 `json:"cost"`
	Count int     `json:"count"`
}

func (s *RestServer) handleCostBreakdown(w http.ResponseWriter, r *http.Request) {
	spans, err := s.store.Spans().Query(r.Context(), storage.SpanFilter{
		StartTime: time.Now().Add(-24 * time.Hour),
		Limit:     10000,
	})
	if err != nil {
		s.errJSON(w, http.StatusInternalServerError, "failed to query spans")
		return
	}

	byAgent := map[string]*CostItem{}
	byUser := map[string]*CostItem{}

	for _, sp := range spans {
		if sp.AgentName != "" {
			if byAgent[sp.AgentName] == nil {
				byAgent[sp.AgentName] = &CostItem{Name: sp.AgentName}
			}
			byAgent[sp.AgentName].Cost += sp.CostUSD
			byAgent[sp.AgentName].Count++
		}
		if sp.UserID != "" {
			if byUser[sp.UserID] == nil {
				byUser[sp.UserID] = &CostItem{Name: sp.UserID}
			}
			byUser[sp.UserID].Cost += sp.CostUSD
			byUser[sp.UserID].Count++
		}
	}

	toSlice := func(m map[string]*CostItem) []CostItem {
		out := make([]CostItem, 0, len(m))
		for _, v := range m {
			out = append(out, *v)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Cost > out[j].Cost })
		return out
	}

	s.json(w, http.StatusOK, CostBreakdown{
		ByAgent: toSlice(byAgent),
		ByUser:  toSlice(byUser),
	})
}

func (s *RestServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.json(w, http.StatusOK, map[string]string{"status": "ok", "version": "0.5.0"})
}
