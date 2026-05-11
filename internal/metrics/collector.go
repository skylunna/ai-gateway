package metrics

/*
	Prometheus 指标收集器
*/
import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var initOnce sync.Once

// 定义两个核心指标 (全局变量)
var (
	/*
		请求总数指标（Counter）
		计数器: 只增不减, 用来统计请求总量
		标签(维度):
			- model: 模型名 (qwen-turbo/deepseek)
			- provider: 厂商 (ali/openai/volc)
			- status: 状态码 (200/400/500/502)

		luner_requests_total{model="qwen-turbo",provider="ali",status="200"} 123
	*/
	RequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "luner_requests_total",                   // 指标名称
		Help: "Total number of LLM requests processed", // 说明
	}, []string{"model", "provider", "status"}) // 标签维度

	/*
		请求耗时指标（Histogram）
			直方图: 统计请求耗时分布 (慢请求、快请求)
			标签:
				- model
				- provider

		luner_request_duration_seconds_bucket{model="qwen-turbo",provider="ali",le="0.1"} 89
	*/
	RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "luner_request_duration_seconds",
		Help:    "Request latency in seconds",
		Buckets: prometheus.DefBuckets, // 默认耗时区间 [.005, .01, .025, .05, ... , 10]
	}, []string{"model", "provider", "type"})

	TokensUsed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "luner_tokens_used_total",
		Help: "Total tokens consumed by model",
	}, []string{"model", "type"})

	CacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "luner_cache_hits_total",
		Help: "Total number of cache hits",
	})

	CacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "luner_cache_misses_total",
		Help: "Total number of cache misses (key not found)",
	})

	// reason: "capacity" (LRU eviction) or "ttl" (TTL expiration)
	CacheEvictions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "luner_cache_evictions_total",
		Help: "Total number of cache entries removed",
	}, []string{"reason"})

	CacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "luner_cache_size",
		Help: "Current number of entries in the cache",
	})
)

// Init registers Prometheus metrics. Safe to call multiple times (e.g. in tests).
func Init() {
	initOnce.Do(func() {
		prometheus.MustRegister(
			RequestTotal, RequestDuration, TokensUsed,
			CacheHits, CacheMisses, CacheEvictions, CacheSize,
		)
	})
}

// CacheObserver implements cache.Observer and reports events to Prometheus.
// It carries no state — all writes go directly to package-level metric vars.
type CacheObserver struct{}

func NewCacheObserver() *CacheObserver { return &CacheObserver{} }

func (*CacheObserver) OnHit()    { CacheHits.Inc() }
func (*CacheObserver) OnMiss()   { CacheMisses.Inc() }
func (*CacheObserver) OnExpire() { CacheEvictions.WithLabelValues("ttl").Inc(); CacheSize.Dec() }
func (*CacheObserver) OnEvict()  { CacheEvictions.WithLabelValues("capacity").Inc(); CacheSize.Dec() }
func (*CacheObserver) OnAdd()    { CacheSize.Inc() }

// Handler 返回 Prometheus metrics HTTP handler
// 提供 /metrics 接口
// 创建一个HTTP处理器, 访问 /metrics 就能看到所有监控数据
func Handler() http.Handler {
	return promhttp.Handler()
}
