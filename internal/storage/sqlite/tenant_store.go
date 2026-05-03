package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

type tenantStore struct{ db *sql.DB }

func (s *tenantStore) Create(ctx context.Context, t *storage.Tenant) error {
	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, email, monthly_budget_usd, current_spend_usd, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.Email,
		t.MonthlyBudgetUSD, t.CurrentSpendUSD,
		boolToInt(t.Enabled), now, now,
	)
	if err != nil {
		return &storage.StorageError{Op: "CreateTenant", Err: err}
	}
	return nil
}

func (s *tenantStore) Get(ctx context.Context, tenantID string) (*storage.Tenant, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, email, monthly_budget_usd, current_spend_usd, enabled, created_at, updated_at
		FROM tenants WHERE id = ?`, tenantID)
	return scanTenant(row)
}

func (s *tenantStore) Update(ctx context.Context, t *storage.Tenant) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE tenants
		SET name = ?, email = ?, monthly_budget_usd = ?, current_spend_usd = ?, enabled = ?, updated_at = ?
		WHERE id = ?`,
		t.Name, t.Email, t.MonthlyBudgetUSD, t.CurrentSpendUSD,
		boolToInt(t.Enabled), time.Now().UnixMilli(), t.ID,
	)
	if err != nil {
		return &storage.StorageError{Op: "UpdateTenant", Err: err}
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *tenantStore) Delete(ctx context.Context, tenantID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM tenants WHERE id = ?", tenantID)
	if err != nil {
		return &storage.StorageError{Op: "DeleteTenant", Err: err}
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *tenantStore) List(ctx context.Context) ([]*storage.Tenant, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, email, monthly_budget_usd, current_spend_usd, enabled, created_at, updated_at
		FROM tenants ORDER BY created_at DESC`)
	if err != nil {
		return nil, &storage.StorageError{Op: "ListTenants", Err: err}
	}
	defer rows.Close()

	var tenants []*storage.Tenant
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			return nil, &storage.StorageError{Op: "scanTenant", Err: err}
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func scanTenant(row rowScanner) (*storage.Tenant, error) {
	var t storage.Tenant
	var enabledInt int
	var createdMs, updatedMs int64

	err := row.Scan(
		&t.ID, &t.Name, &t.Email,
		&t.MonthlyBudgetUSD, &t.CurrentSpendUSD,
		&enabledInt, &createdMs, &updatedMs,
	)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, &storage.StorageError{Op: "scanTenant", Err: err}
	}
	t.Enabled = enabledInt != 0
	t.CreatedAt = time.UnixMilli(createdMs).UTC()
	t.UpdatedAt = time.UnixMilli(updatedMs).UTC()
	return &t, nil
}
