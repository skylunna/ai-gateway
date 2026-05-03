package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

func newTestStore(t *testing.T) storage.Storage {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// ── Health ───────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	s := newTestStore(t)
	if err := s.Health(context.Background()); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

// ── SpanStore ────────────────────────────────────────────────────────────────

func TestSpanStore_CreateAndGet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	span := &storage.Span{
		SpanID:           "span-001",
		TraceID:          "trace-001",
		TenantID:         "tenant-001",
		AgentName:        "test-agent",
		SpanType:         "llm",
		Name:             "chat",
		StartTime:        time.Now().Add(-time.Second).UTC().Truncate(time.Millisecond),
		EndTime:          time.Now().UTC().Truncate(time.Millisecond),
		DurationMs:       1000,
		Model:            "gpt-4o",
		PromptTokens:     100,
		CompletionTokens: 50,
		CostUSD:          0.001,
		Status:           "success",
		Tags:             map[string]string{"env": "test", "version": "v1"},
	}

	if err := s.Spans().Create(ctx, span); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.Spans().Get(ctx, "span-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.SpanID != span.SpanID {
		t.Errorf("SpanID: got %q want %q", got.SpanID, span.SpanID)
	}
	if got.Model != "gpt-4o" {
		t.Errorf("Model: got %q want %q", got.Model, "gpt-4o")
	}
	if got.PromptTokens != 100 {
		t.Errorf("PromptTokens: got %d want 100", got.PromptTokens)
	}
	if got.Tags["env"] != "test" {
		t.Errorf("Tags: got %v", got.Tags)
	}
	if !got.StartTime.Equal(span.StartTime) {
		t.Errorf("StartTime: got %v want %v", got.StartTime, span.StartTime)
	}
}

func TestSpanStore_GetNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Spans().Get(context.Background(), "nonexistent")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSpanStore_CreateBatch(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	spans := make([]*storage.Span, 5)
	for i := range spans {
		spans[i] = &storage.Span{
			SpanID:    "batch-span-" + string(rune('0'+i)),
			TraceID:   "trace-batch",
			TenantID:  "tenant-001",
			AgentName: "batch-agent",
			SpanType:  "llm",
			Name:      "batch",
			StartTime: time.Now().UTC().Truncate(time.Millisecond),
			EndTime:   time.Now().Add(time.Second).UTC().Truncate(time.Millisecond),
			Status:    "success",
		}
	}

	if err := s.Spans().CreateBatch(ctx, spans); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	got, err := s.Spans().ListByTrace(ctx, "trace-batch")
	if err != nil {
		t.Fatalf("ListByTrace: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("ListByTrace: got %d spans, want 5", len(got))
	}
}

func TestSpanStore_Query(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for _, agentName := range []string{"agent-a", "agent-a", "agent-b"} {
		if err := s.Spans().Create(ctx, &storage.Span{
			SpanID:    "q-" + agentName + "-" + time.Now().String(),
			TraceID:   "trace-q",
			TenantID:  "t1",
			AgentName: agentName,
			SpanType:  "llm",
			Name:      "op",
			StartTime: time.Now().UTC(),
			EndTime:   time.Now().Add(time.Second).UTC(),
			Status:    "success",
		}); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	got, err := s.Spans().Query(ctx, storage.SpanFilter{AgentName: "agent-a", Limit: 10})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Query agent-a: got %d, want 2", len(got))
	}
}

func TestSpanStore_DeleteOlderThan(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.Spans().Create(ctx, &storage.Span{
		SpanID:    "old-span",
		TraceID:   "trace-old",
		TenantID:  "t1",
		AgentName: "agent",
		SpanType:  "llm",
		Name:      "op",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().Add(time.Second).UTC(),
		Status:    "success",
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	n, err := s.Spans().DeleteOlderThan(ctx, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("DeleteOlderThan: %v", err)
	}
	if n != 1 {
		t.Errorf("DeleteOlderThan: deleted %d rows, want 1", n)
	}
}

// ── TenantStore ───────────────────────────────────────────────────────────────

func TestTenantStore_CRUD(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tenant := &storage.Tenant{
		ID:               "tenant-001",
		Name:             "Acme Corp",
		Email:            "admin@acme.com",
		MonthlyBudgetUSD: 100.0,
		Enabled:          true,
	}

	if err := s.Tenants().Create(ctx, tenant); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.Tenants().Get(ctx, "tenant-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Acme Corp" {
		t.Errorf("Name: got %q", got.Name)
	}
	if !got.Enabled {
		t.Error("expected Enabled=true")
	}

	got.Name = "Acme Corp Updated"
	if err := s.Tenants().Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := s.Tenants().Get(ctx, "tenant-001")
	if updated.Name != "Acme Corp Updated" {
		t.Errorf("after Update: got Name=%q", updated.Name)
	}

	list, err := s.Tenants().List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List: got %d tenants", len(list))
	}

	if err := s.Tenants().Delete(ctx, "tenant-001"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Tenants().Get(ctx, "tenant-001"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("after Delete: expected ErrNotFound, got %v", err)
	}
}

// ── APIKeyStore ───────────────────────────────────────────────────────────────

func TestAPIKeyStore_CRUD(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// API keys reference tenants via FK; create tenant first.
	if err := s.Tenants().Create(ctx, &storage.Tenant{
		ID: "t1", Name: "T1", Email: "t1@test.com", Enabled: true,
	}); err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	key := &storage.APIKey{
		ID:           "key-001",
		TenantID:     "t1",
		Name:         "default key",
		KeyHash:      "abc123hash",
		AgentName:    "my-agent",
		Tags:         map[string]string{"project": "demo"},
		RateLimitRPM: 60,
		Enabled:      true,
	}

	if err := s.APIKeys().Create(ctx, key); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.APIKeys().GetByHash(ctx, "abc123hash")
	if err != nil {
		t.Fatalf("GetByHash: %v", err)
	}
	if got.AgentName != "my-agent" {
		t.Errorf("AgentName: got %q", got.AgentName)
	}
	if got.Tags["project"] != "demo" {
		t.Errorf("Tags: got %v", got.Tags)
	}

	list, err := s.APIKeys().ListByTenant(ctx, "t1")
	if err != nil {
		t.Fatalf("ListByTenant: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListByTenant: got %d keys", len(list))
	}
}

// ── PolicyStore ───────────────────────────────────────────────────────────────

func TestPolicyStore_CRUD(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.Tenants().Create(ctx, &storage.Tenant{
		ID: "t1", Name: "T1", Email: "t1@test.com", Enabled: true,
	}); err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	policy := &storage.Policy{
		ID:        "pol-001",
		TenantID:  "t1",
		Name:      "block-expensive",
		Expression: "cost_usd > 0.1",
		Action:    "block",
		Enabled:   true,
	}

	if err := s.Policies().Create(ctx, policy); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.Policies().Get(ctx, "pol-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Action != "block" {
		t.Errorf("Action: got %q", got.Action)
	}

	active, err := s.Policies().ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(active) != 1 {
		t.Errorf("ListActive: got %d policies", len(active))
	}

	got.Enabled = false
	if err := s.Policies().Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	active, _ = s.Policies().ListActive(ctx)
	if len(active) != 0 {
		t.Errorf("after disable: ListActive should return 0, got %d", len(active))
	}
}
