package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

type spanStore struct{ db *sql.DB }

const spanColumns = `span_id, trace_id, parent_span_id, session_id,
	tenant_id, user_id, agent_name, agent_version, environment,
	span_type, name, start_time, end_time, duration_ms,
	model, prompt_tokens, completion_tokens, cost_usd,
	input, output, status, error_message, tags`

const insertSpanSQL = `
	INSERT INTO spans (
		span_id, trace_id, parent_span_id, session_id,
		tenant_id, user_id, agent_name, agent_version, environment,
		span_type, name, start_time, end_time, duration_ms,
		model, prompt_tokens, completion_tokens, cost_usd,
		input, output, status, error_message, tags, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanSpan(row rowScanner) (*storage.Span, error) {
	var s storage.Span
	var tagsJSON []byte
	var startMs, endMs int64

	err := row.Scan(
		&s.SpanID, &s.TraceID, &s.ParentSpanID, &s.SessionID,
		&s.TenantID, &s.UserID, &s.AgentName, &s.AgentVersion, &s.Environment,
		&s.SpanType, &s.Name, &startMs, &endMs, &s.DurationMs,
		&s.Model, &s.PromptTokens, &s.CompletionTokens, &s.CostUSD,
		&s.Input, &s.Output, &s.Status, &s.ErrorMessage, &tagsJSON,
	)
	if err != nil {
		return nil, err
	}
	s.StartTime = time.UnixMilli(startMs).UTC()
	s.EndTime = time.UnixMilli(endMs).UTC()
	if len(tagsJSON) > 0 {
		_ = json.Unmarshal(tagsJSON, &s.Tags)
	}
	return &s, nil
}

func spanArgs(span *storage.Span) []any {
	tagsJSON, _ := json.Marshal(span.Tags)
	return []any{
		span.SpanID, span.TraceID, span.ParentSpanID, span.SessionID,
		span.TenantID, span.UserID, span.AgentName, span.AgentVersion, span.Environment,
		span.SpanType, span.Name,
		span.StartTime.UnixMilli(), span.EndTime.UnixMilli(), span.DurationMs,
		span.Model, span.PromptTokens, span.CompletionTokens, span.CostUSD,
		span.Input, span.Output, span.Status, span.ErrorMessage, tagsJSON,
		time.Now().UnixMilli(),
	}
}

func (s *spanStore) Create(ctx context.Context, span *storage.Span) error {
	if _, err := s.db.ExecContext(ctx, insertSpanSQL, spanArgs(span)...); err != nil {
		return &storage.StorageError{Op: "CreateSpan", Err: err}
	}
	return nil
}

func (s *spanStore) CreateBatch(ctx context.Context, spans []*storage.Span) error {
	if len(spans) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &storage.StorageError{Op: "CreateBatch", Err: err}
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, insertSpanSQL)
	if err != nil {
		return &storage.StorageError{Op: "CreateBatch.prepare", Err: err}
	}
	defer stmt.Close()

	for _, span := range spans {
		if _, err := stmt.ExecContext(ctx, spanArgs(span)...); err != nil {
			return &storage.StorageError{Op: "CreateBatch.exec", Err: err}
		}
	}
	if err := tx.Commit(); err != nil {
		return &storage.StorageError{Op: "CreateBatch.commit", Err: err}
	}
	return nil
}

func (s *spanStore) Get(ctx context.Context, spanID string) (*storage.Span, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT "+spanColumns+" FROM spans WHERE span_id = ?", spanID)
	span, err := scanSpan(row)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, &storage.StorageError{Op: "GetSpan", Err: err}
	}
	return span, nil
}

func (s *spanStore) ListByTrace(ctx context.Context, traceID string) ([]*storage.Span, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT "+spanColumns+" FROM spans WHERE trace_id = ? ORDER BY start_time ASC", traceID)
	if err != nil {
		return nil, &storage.StorageError{Op: "ListByTrace", Err: err}
	}
	defer rows.Close()
	return collectSpans(rows)
}

func (s *spanStore) Query(ctx context.Context, filter storage.SpanFilter) ([]*storage.Span, error) {
	q := "SELECT " + spanColumns + " FROM spans WHERE 1=1"
	var args []any

	if filter.TraceID != "" {
		q += " AND trace_id = ?"
		args = append(args, filter.TraceID)
	}
	if filter.AgentName != "" {
		q += " AND agent_name = ?"
		args = append(args, filter.AgentName)
	}
	if filter.UserID != "" {
		q += " AND user_id = ?"
		args = append(args, filter.UserID)
	}
	if filter.TenantID != "" {
		q += " AND tenant_id = ?"
		args = append(args, filter.TenantID)
	}
	if filter.Environment != "" {
		q += " AND environment = ?"
		args = append(args, filter.Environment)
	}
	if filter.SpanType != "" {
		q += " AND span_type = ?"
		args = append(args, filter.SpanType)
	}
	if !filter.StartTime.IsZero() {
		q += " AND start_time >= ?"
		args = append(args, filter.StartTime.UnixMilli())
	}
	if !filter.EndTime.IsZero() {
		q += " AND start_time <= ?"
		args = append(args, filter.EndTime.UnixMilli())
	}

	q += " ORDER BY start_time DESC"

	if filter.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, &storage.StorageError{Op: "QuerySpans", Err: err}
	}
	defer rows.Close()
	return collectSpans(rows)
}

func (s *spanStore) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM spans WHERE created_at < ?", before.UnixMilli())
	if err != nil {
		return 0, &storage.StorageError{Op: "DeleteOlderThan", Err: err}
	}
	n, _ := result.RowsAffected()
	return n, nil
}

func (s *spanStore) QueryStats(ctx context.Context, userID, tenantID string, since time.Time) (*storage.SpanStats, error) {
	q := `SELECT COUNT(*), COALESCE(SUM(cost_usd), 0), COALESCE(SUM(prompt_tokens + completion_tokens), 0)
	      FROM spans WHERE span_type = 'llm' AND start_time >= ?`
	args := []any{since.UnixMilli()}
	if userID != "" {
		q += " AND user_id = ?"
		args = append(args, userID)
	}
	if tenantID != "" {
		q += " AND tenant_id = ?"
		args = append(args, tenantID)
	}

	var stats storage.SpanStats
	if err := s.db.QueryRowContext(ctx, q, args...).Scan(
		&stats.RequestCount, &stats.CostUSD, &stats.TokensUsed,
	); err != nil {
		return nil, &storage.StorageError{Op: "QueryStats", Err: err}
	}
	return &stats, nil
}

func collectSpans(rows *sql.Rows) ([]*storage.Span, error) {
	var spans []*storage.Span
	for rows.Next() {
		s, err := scanSpan(rows)
		if err != nil {
			return nil, &storage.StorageError{Op: "scanSpan", Err: err}
		}
		spans = append(spans, s)
	}
	return spans, rows.Err()
}
