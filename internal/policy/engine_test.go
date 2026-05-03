package policy_test

import (
	"context"
	"testing"
	"time"

	"github.com/skylunna/luner/internal/policy"
	"github.com/skylunna/luner/internal/storage"
	"github.com/skylunna/luner/internal/storage/sqlite"
)

func newStore(t *testing.T) storage.Storage {
	t.Helper()
	s, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func mkPolicy(id, expr, action string, priority int, enabled bool) *storage.Policy {
	now := time.Now()
	return &storage.Policy{
		ID:         id,
		TenantID:   "t1",
		Name:       id,
		Expression: expr,
		Action:     action,
		Priority:   priority,
		Enabled:    enabled,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestEvaluate_NoPolicies(t *testing.T) {
	store := newStore(t)
	eng, err := policy.NewCELEngine(store, nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	result, err := eng.Evaluate(context.Background(), &policy.EvaluationContext{Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result with no policies, got %+v", result)
	}
}

func TestEvaluate_BlockOnModel(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	if err := store.Policies().Create(ctx, mkPolicy("p1", `model == "gpt-4o"`, "block", 0, true)); err != nil {
		t.Fatalf("create policy: %v", err)
	}

	eng, _ := policy.NewCELEngine(store, nil)

	result, err := eng.Evaluate(ctx, &policy.EvaluationContext{Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected block result, got nil")
	}
	if result.Action != policy.ActionBlock {
		t.Errorf("expected ActionBlock, got %s", result.Action)
	}
}

func TestEvaluate_NoMatchOnDifferentModel(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	_ = store.Policies().Create(ctx, mkPolicy("p1", `model == "gpt-4o"`, "block", 0, true))

	eng, _ := policy.NewCELEngine(store, nil)

	result, err := eng.Evaluate(ctx, &policy.EvaluationContext{Model: "gpt-3.5-turbo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for non-matching model, got %+v", result)
	}
}

func TestEvaluate_PriorityOrdering(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	_ = store.Policies().Create(ctx, mkPolicy("p-low", `model == "gpt-4o"`, "alert", 10, true))
	_ = store.Policies().Create(ctx, mkPolicy("p-high", `model == "gpt-4o"`, "block", 1, true))

	eng, _ := policy.NewCELEngine(store, nil)

	result, err := eng.Evaluate(ctx, &policy.EvaluationContext{Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.PolicyID != "p-high" {
		t.Errorf("expected high-priority policy (p-high) first, got %s", result.PolicyID)
	}
	if result.Action != policy.ActionBlock {
		t.Errorf("expected block action, got %s", result.Action)
	}
}

func TestEvaluate_InvalidExpressionFailOpen(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	_ = store.Policies().Create(ctx, mkPolicy("p-bad", `this is NOT valid CEL !!!`, "block", 0, true))

	eng, _ := policy.NewCELEngine(store, nil)

	result, err := eng.Evaluate(ctx, &policy.EvaluationContext{Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("expected fail-open, got error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for invalid expression, got %+v", result)
	}
}

func TestEvaluate_DisabledPolicyIgnored(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	_ = store.Policies().Create(ctx, mkPolicy("p-disabled", `model == "gpt-4o"`, "block", 0, false))

	eng, _ := policy.NewCELEngine(store, nil)

	result, err := eng.Evaluate(ctx, &policy.EvaluationContext{Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("disabled policy should not match, got %+v", result)
	}
}

func TestReload_ClearsCompiledCache(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	p := mkPolicy("p1", `user_id == "alice"`, "block", 0, true)
	_ = store.Policies().Create(ctx, p)

	eng, _ := policy.NewCELEngine(store, nil)

	result, _ := eng.Evaluate(ctx, &policy.EvaluationContext{UserID: "alice"})
	if result == nil || result.Action != policy.ActionBlock {
		t.Fatal("expected block on first eval")
	}

	// Disable the policy and reload the engine cache
	p.Enabled = false
	_ = store.Policies().Update(ctx, p)
	eng.Reload()

	result, _ = eng.Evaluate(ctx, &policy.EvaluationContext{UserID: "alice"})
	if result != nil {
		t.Errorf("after disable+reload expected nil, got %+v", result)
	}
}

func TestEvaluate_StatsVariables(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	_ = store.Policies().Create(ctx, mkPolicy("p1", `request_count > 100`, "block", 0, true))

	eng, _ := policy.NewCELEngine(store, nil)

	// No spans in db → request_count = 0 → no match
	result, err := eng.Evaluate(ctx, &policy.EvaluationContext{Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil when count is 0, got %+v", result)
	}
}
