package metrics

import (
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/client_golang/prometheus"
)

// CacheSnap holds a point-in-time view of cache counters.
type CacheSnap struct {
	Hits      float64            `json:"hits"`
	Misses    float64            `json:"misses"`
	HitRate   float64            `json:"hit_rate"`
	Size      float64            `json:"size"`
	Evictions map[string]float64 `json:"evictions"`
}

// LiveSnap is the full snapshot returned by /api/metrics/live.
type LiveSnap struct {
	Cache    CacheSnap          `json:"cache"`
	Requests map[string]float64 `json:"requests_by_status"`
	Tokens   map[string]float64 `json:"tokens_by_type"`
}

// Snapshot reads the current values from the default Prometheus registry.
func Snapshot() LiveSnap {
	mfs, _ := prometheus.DefaultGatherer.Gather()

	snap := LiveSnap{
		Cache:    CacheSnap{Evictions: map[string]float64{}},
		Requests: map[string]float64{},
		Tokens:   map[string]float64{},
	}

	for _, mf := range mfs {
		switch mf.GetName() {
		case "luner_cache_hits_total":
			snap.Cache.Hits = counterVal(mf)
		case "luner_cache_misses_total":
			snap.Cache.Misses = counterVal(mf)
		case "luner_cache_size":
			snap.Cache.Size = gaugeVal(mf)
		case "luner_cache_evictions_total":
			for _, m := range mf.GetMetric() {
				snap.Cache.Evictions[label(m, "reason")] = m.GetCounter().GetValue()
			}
		case "luner_requests_total":
			for _, m := range mf.GetMetric() {
				snap.Requests[label(m, "status")] += m.GetCounter().GetValue()
			}
		case "luner_tokens_used_total":
			for _, m := range mf.GetMetric() {
				snap.Tokens[label(m, "type")] += m.GetCounter().GetValue()
			}
		}
	}

	total := snap.Cache.Hits + snap.Cache.Misses
	if total > 0 {
		snap.Cache.HitRate = snap.Cache.Hits / total
	}
	return snap
}

func counterVal(mf *dto.MetricFamily) float64 {
	if m := mf.GetMetric(); len(m) > 0 {
		return m[0].GetCounter().GetValue()
	}
	return 0
}

func gaugeVal(mf *dto.MetricFamily) float64 {
	if m := mf.GetMetric(); len(m) > 0 {
		return m[0].GetGauge().GetValue()
	}
	return 0
}

func label(m *dto.Metric, name string) string {
	for _, lp := range m.GetLabel() {
		if lp.GetName() == name {
			return lp.GetValue()
		}
	}
	return ""
}
