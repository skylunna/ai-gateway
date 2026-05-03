package storage

import "time"

// Span is the core data model for Agent Trace.
type Span struct {
	SpanID       string
	TraceID      string
	ParentSpanID string
	SessionID    string

	TenantID     string
	UserID       string
	AgentName    string
	AgentVersion string
	Environment  string

	SpanType   string // "llm", "tool", "retrieval", "custom"
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	DurationMs int64

	Model            string
	PromptTokens     int32
	CompletionTokens int32
	CostUSD          float64

	Input        string
	Output       string
	Status       string // "success", "error", "timeout"
	ErrorMessage string
	Tags         map[string]string

	CreatedAt time.Time
}

// Tenant holds multi-tenant configuration and budget state.
type Tenant struct {
	ID               string
	Name             string
	Email            string
	MonthlyBudgetUSD float64
	CurrentSpendUSD  float64
	Enabled          bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// APIKey is a virtual key used for cost attribution and routing.
type APIKey struct {
	ID           string
	TenantID     string
	Name         string
	KeyHash      string
	AgentName    string
	UserID       string
	Tags         map[string]string
	RateLimitRPM int
	Enabled      bool
	LastUsed     time.Time
	CreatedAt    time.Time
}

// Policy is a governance rule evaluated against each request.
type Policy struct {
	ID          string
	TenantID    string
	Name        string
	Expression  string // CEL expression, e.g. "cost_usd > 0.1"
	Action      string // "block", "alert", "downgrade"
	Priority    int    // lower = evaluated first
	Description string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SpanStats holds aggregate usage statistics for a user/tenant window.
type SpanStats struct {
	RequestCount int64
	CostUSD      float64
	TokensUsed   int64
}
