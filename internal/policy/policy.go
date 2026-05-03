package policy

import "context"

// Action defines what to do when a policy expression matches.
type Action string

const (
	ActionBlock     Action = "block"
	ActionAlert     Action = "alert"
	ActionDowngrade Action = "downgrade"
)

// EvaluationContext holds all variables available to CEL expressions.
type EvaluationContext struct {
	// Per-request identity
	Model    string
	UserID   string
	TenantID string

	// Aggregated stats (last hour, populated from the span store)
	RequestCount int64
	CostUSD      float64
	TokensUsed   int64
}

// EvaluationResult is returned when a policy matches.
type EvaluationResult struct {
	PolicyID   string
	PolicyName string
	Action     Action
	Message    string
}

// Engine evaluates policies against a request context.
type Engine interface {
	// Evaluate checks all active policies in priority order.
	// Returns the first matching result, or nil if all pass.
	// Errors during evaluation are logged and treated as non-matching (fail-open).
	Evaluate(ctx context.Context, evalCtx *EvaluationContext) (*EvaluationResult, error)

	// Reload clears the compiled program cache, forcing recompilation on the next call.
	Reload()
}
