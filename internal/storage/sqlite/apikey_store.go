package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/skylunna/luner/internal/storage"
)

type apiKeyStore struct{ db *sql.DB }

func (s *apiKeyStore) Create(ctx context.Context, k *storage.APIKey) error {
	tagsJSON, _ := json.Marshal(k.Tags)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO api_keys (id, tenant_id, name, key_hash, agent_name, user_id, tags, rate_limit_rpm, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		k.ID, k.TenantID, k.Name, k.KeyHash,
		k.AgentName, k.UserID, tagsJSON,
		k.RateLimitRPM, boolToInt(k.Enabled),
		time.Now().UnixMilli(),
	)
	if err != nil {
		return &storage.StorageError{Op: "CreateAPIKey", Err: err}
	}
	return nil
}

func (s *apiKeyStore) Get(ctx context.Context, keyID string) (*storage.APIKey, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, tenant_id, name, key_hash, agent_name, user_id, tags, rate_limit_rpm, enabled, last_used, created_at FROM api_keys WHERE id = ?",
		keyID)
	return scanAPIKey(row)
}

func (s *apiKeyStore) GetByHash(ctx context.Context, hash string) (*storage.APIKey, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, tenant_id, name, key_hash, agent_name, user_id, tags, rate_limit_rpm, enabled, last_used, created_at FROM api_keys WHERE key_hash = ?",
		hash)
	return scanAPIKey(row)
}

func (s *apiKeyStore) Update(ctx context.Context, k *storage.APIKey) error {
	tagsJSON, _ := json.Marshal(k.Tags)
	result, err := s.db.ExecContext(ctx, `
		UPDATE api_keys
		SET name = ?, agent_name = ?, user_id = ?, tags = ?, rate_limit_rpm = ?, enabled = ?
		WHERE id = ?`,
		k.Name, k.AgentName, k.UserID, tagsJSON,
		k.RateLimitRPM, boolToInt(k.Enabled), k.ID,
	)
	if err != nil {
		return &storage.StorageError{Op: "UpdateAPIKey", Err: err}
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *apiKeyStore) Delete(ctx context.Context, keyID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM api_keys WHERE id = ?", keyID)
	if err != nil {
		return &storage.StorageError{Op: "DeleteAPIKey", Err: err}
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *apiKeyStore) ListByTenant(ctx context.Context, tenantID string) ([]*storage.APIKey, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, key_hash, agent_name, user_id, tags, rate_limit_rpm, enabled, last_used, created_at
		FROM api_keys WHERE tenant_id = ? ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, &storage.StorageError{Op: "ListAPIKeysByTenant", Err: err}
	}
	defer rows.Close()

	var keys []*storage.APIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func scanAPIKey(row rowScanner) (*storage.APIKey, error) {
	var k storage.APIKey
	var tagsJSON []byte
	var enabledInt int
	var lastUsedMs sql.NullInt64
	var createdMs int64

	err := row.Scan(
		&k.ID, &k.TenantID, &k.Name, &k.KeyHash,
		&k.AgentName, &k.UserID, &tagsJSON,
		&k.RateLimitRPM, &enabledInt, &lastUsedMs, &createdMs,
	)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, &storage.StorageError{Op: "scanAPIKey", Err: err}
	}
	k.Enabled = enabledInt != 0
	if lastUsedMs.Valid {
		k.LastUsed = time.UnixMilli(lastUsedMs.Int64).UTC()
	}
	k.CreatedAt = time.UnixMilli(createdMs).UTC()
	if len(tagsJSON) > 0 {
		_ = json.Unmarshal(tagsJSON, &k.Tags)
	}
	return &k, nil
}
