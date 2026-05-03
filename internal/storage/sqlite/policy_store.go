package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

type policyStore struct{ db *sql.DB }

func (s *policyStore) Create(ctx context.Context, p *storage.Policy) error {
	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO policies (id, tenant_id, name, expression, action, priority, description, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.TenantID, p.Name, p.Expression, p.Action,
		p.Priority, p.Description, boolToInt(p.Enabled), now, now,
	)
	if err != nil {
		return &storage.StorageError{Op: "CreatePolicy", Err: err}
	}
	return nil
}

func (s *policyStore) Get(ctx context.Context, policyID string) (*storage.Policy, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, expression, action, priority, description, enabled, created_at, updated_at
		FROM policies WHERE id = ?`, policyID)
	return scanPolicy(row)
}

func (s *policyStore) Update(ctx context.Context, p *storage.Policy) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE policies
		SET name = ?, expression = ?, action = ?, priority = ?, description = ?, enabled = ?, updated_at = ?
		WHERE id = ?`,
		p.Name, p.Expression, p.Action, p.Priority, p.Description,
		boolToInt(p.Enabled), time.Now().UnixMilli(), p.ID,
	)
	if err != nil {
		return &storage.StorageError{Op: "UpdatePolicy", Err: err}
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *policyStore) Delete(ctx context.Context, policyID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM policies WHERE id = ?", policyID)
	if err != nil {
		return &storage.StorageError{Op: "DeletePolicy", Err: err}
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *policyStore) ListByTenant(ctx context.Context, tenantID string) ([]*storage.Policy, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, expression, action, priority, description, enabled, created_at, updated_at
		FROM policies WHERE tenant_id = ? ORDER BY priority ASC, created_at DESC`, tenantID)
	if err != nil {
		return nil, &storage.StorageError{Op: "ListPoliciesByTenant", Err: err}
	}
	defer rows.Close()
	return collectPolicies(rows)
}

func (s *policyStore) ListActive(ctx context.Context) ([]*storage.Policy, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, expression, action, priority, description, enabled, created_at, updated_at
		FROM policies WHERE enabled = 1 ORDER BY priority ASC, created_at DESC`)
	if err != nil {
		return nil, &storage.StorageError{Op: "ListActivePolicies", Err: err}
	}
	defer rows.Close()
	return collectPolicies(rows)
}

func collectPolicies(rows *sql.Rows) ([]*storage.Policy, error) {
	var policies []*storage.Policy
	for rows.Next() {
		p, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func scanPolicy(row rowScanner) (*storage.Policy, error) {
	var p storage.Policy
	var enabledInt int
	var createdMs, updatedMs int64

	err := row.Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Expression, &p.Action,
		&p.Priority, &p.Description, &enabledInt, &createdMs, &updatedMs,
	)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, &storage.StorageError{Op: "scanPolicy", Err: err}
	}
	p.Enabled = enabledInt != 0
	p.CreatedAt = time.UnixMilli(createdMs).UTC()
	p.UpdatedAt = time.UnixMilli(updatedMs).UTC()
	return &p, nil
}
