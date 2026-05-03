package policy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/skylunna/luner/internal/storage"
)

// statsKey identifies a cached stats entry.
type statsKey struct {
	userID   string
	tenantID string
	bucket   int64 // Unix minute
}

type statsEntry struct {
	stats     storage.SpanStats
	expiresAt time.Time
}

// CELEngine is the production Engine implementation backed by Google CEL.
type CELEngine struct {
	store  storage.Storage
	logger *slog.Logger

	mu       sync.RWMutex
	programs map[string]cel.Program // keyed by policy ID

	statsCache sync.Map // statsKey → statsEntry

	celEnv *cel.Env
}

// NewCELEngine constructs a CELEngine and compiles the shared CEL environment.
// logger may be nil, in which case a no-op logger is used.
func NewCELEngine(store storage.Storage, logger *slog.Logger) (*CELEngine, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	env, err := cel.NewEnv(
		cel.Variable("model", cel.StringType),
		cel.Variable("user_id", cel.StringType),
		cel.Variable("tenant_id", cel.StringType),
		cel.Variable("request_count", cel.IntType),
		cel.Variable("cost_usd", cel.DoubleType),
		cel.Variable("tokens_used", cel.IntType),
	)
	if err != nil {
		return nil, fmt.Errorf("create CEL env: %w", err)
	}
	return &CELEngine{
		store:    store,
		logger:   logger,
		programs: make(map[string]cel.Program),
		celEnv:   env,
	}, nil
}

// Reload clears cached compiled programs so they recompile from the store next call.
func (e *CELEngine) Reload() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.programs = make(map[string]cel.Program)
}

// Evaluate runs all active policies against evalCtx and returns the first match.
// On any per-policy error the policy is skipped (fail-open).
func (e *CELEngine) Evaluate(ctx context.Context, evalCtx *EvaluationContext) (*EvaluationResult, error) {
	policies, err := e.store.Policies().ListActive(ctx)
	if err != nil {
		e.logger.Warn("policy: failed to load policies, failing open", "err", err)
		return nil, nil
	}
	if len(policies) == 0 {
		return nil, nil
	}

	if err := e.fillStats(ctx, evalCtx); err != nil {
		e.logger.Warn("policy: failed to fill stats, continuing without them", "err", err)
	}

	vars := map[string]any{
		"model":         evalCtx.Model,
		"user_id":       evalCtx.UserID,
		"tenant_id":     evalCtx.TenantID,
		"request_count": evalCtx.RequestCount,
		"cost_usd":      evalCtx.CostUSD,
		"tokens_used":   evalCtx.TokensUsed,
	}

	// ListActive already orders by priority ASC.
	for _, p := range policies {
		prog, err := e.getOrCompile(p.ID, p.Expression)
		if err != nil {
			e.logger.Warn("policy: compile error, skipping", "policy_id", p.ID, "err", err)
			continue
		}

		out, _, err := prog.Eval(vars)
		if err != nil {
			e.logger.Warn("policy: eval error, skipping", "policy_id", p.ID, "err", err)
			continue
		}

		if matched, _ := out.Value().(bool); matched {
			return &EvaluationResult{
				PolicyID:   p.ID,
				PolicyName: p.Name,
				Action:     Action(p.Action),
				Message:    fmt.Sprintf("policy %q matched", p.Name),
			}, nil
		}
	}
	return nil, nil
}

func (e *CELEngine) getOrCompile(id, expression string) (cel.Program, error) {
	e.mu.RLock()
	prog, ok := e.programs[id]
	e.mu.RUnlock()
	if ok {
		return prog, nil
	}

	ast, iss := e.celEnv.Compile(expression)
	if iss != nil && iss.Err() != nil {
		return nil, iss.Err()
	}
	prog, err := e.celEnv.Program(ast)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.programs[id] = prog
	e.mu.Unlock()
	return prog, nil
}

func (e *CELEngine) fillStats(ctx context.Context, evalCtx *EvaluationContext) error {
	now := time.Now()
	key := statsKey{
		userID:   evalCtx.UserID,
		tenantID: evalCtx.TenantID,
		bucket:   now.Unix() / 60,
	}

	if v, ok := e.statsCache.Load(key); ok {
		if entry := v.(statsEntry); now.Before(entry.expiresAt) {
			evalCtx.RequestCount = entry.stats.RequestCount
			evalCtx.CostUSD = entry.stats.CostUSD
			evalCtx.TokensUsed = entry.stats.TokensUsed
			return nil
		}
	}

	stats, err := e.store.Spans().QueryStats(ctx, evalCtx.UserID, evalCtx.TenantID, now.Add(-time.Hour))
	if err != nil {
		return err
	}
	evalCtx.RequestCount = stats.RequestCount
	evalCtx.CostUSD = stats.CostUSD
	evalCtx.TokensUsed = stats.TokensUsed

	e.statsCache.Store(key, statsEntry{stats: *stats, expiresAt: now.Add(time.Minute)})
	return nil
}
