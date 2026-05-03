package storage

import (
	"context"
	"time"
)

// Storage is the top-level interface for all persistence operations.
type Storage interface {
	Spans() SpanStore
	Tenants() TenantStore
	APIKeys() APIKeyStore
	Policies() PolicyStore

	Close() error
	Health(ctx context.Context) error
}

// SpanStore manages Agent Trace spans (high-frequency writes).
type SpanStore interface {
	Create(ctx context.Context, span *Span) error
	CreateBatch(ctx context.Context, spans []*Span) error
	Get(ctx context.Context, spanID string) (*Span, error)
	ListByTrace(ctx context.Context, traceID string) ([]*Span, error)
	Query(ctx context.Context, filter SpanFilter) ([]*Span, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
	// QueryStats returns aggregate usage for a user/tenant since the given time.
	// Empty userID or tenantID means "all".
	QueryStats(ctx context.Context, userID, tenantID string, since time.Time) (*SpanStats, error)
}

// TenantStore manages multi-tenant configuration.
type TenantStore interface {
	Create(ctx context.Context, tenant *Tenant) error
	Get(ctx context.Context, tenantID string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, tenantID string) error
	List(ctx context.Context) ([]*Tenant, error)
}

// APIKeyStore manages virtual API keys for cost attribution.
type APIKeyStore interface {
	Create(ctx context.Context, key *APIKey) error
	Get(ctx context.Context, keyID string) (*APIKey, error)
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	Update(ctx context.Context, key *APIKey) error
	Delete(ctx context.Context, keyID string) error
	ListByTenant(ctx context.Context, tenantID string) ([]*APIKey, error)
}

// PolicyStore manages governance policy rules.
type PolicyStore interface {
	Create(ctx context.Context, policy *Policy) error
	Get(ctx context.Context, policyID string) (*Policy, error)
	Update(ctx context.Context, policy *Policy) error
	Delete(ctx context.Context, policyID string) error
	ListByTenant(ctx context.Context, tenantID string) ([]*Policy, error)
	ListActive(ctx context.Context) ([]*Policy, error)
}

// SpanFilter specifies predicates for span queries.
type SpanFilter struct {
	TraceID     string
	AgentName   string
	UserID      string
	TenantID    string
	Environment string
	SpanType    string
	StartTime   time.Time
	EndTime     time.Time
	Limit       int
	Offset      int
}
