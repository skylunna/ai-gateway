package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/skylunna/luner/internal/storage"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type store struct {
	db       *sql.DB
	spans    *spanStore
	tenants  *tenantStore
	apiKeys  *apiKeyStore
	policies *policyStore
}

// New opens a SQLite database at dsn and runs any pending migrations.
// Use ":memory:" for an in-process ephemeral database (tests).
func New(dsn string) (storage.Storage, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	for _, p := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("sqlite pragma (%s): %w", p, err)
		}
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &store{
		db:       db,
		spans:    &spanStore{db: db},
		tenants:  &tenantStore{db: db},
		apiKeys:  &apiKeyStore{db: db},
		policies: &policyStore{db: db},
	}, nil
}

func (s *store) Spans() storage.SpanStore     { return s.spans }
func (s *store) Tenants() storage.TenantStore { return s.tenants }
func (s *store) APIKeys() storage.APIKeyStore { return s.apiKeys }
func (s *store) Policies() storage.PolicyStore { return s.policies }
func (s *store) Close() error                  { return s.db.Close() }

func (s *store) Health(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var version int
		fmt.Sscanf(entry.Name(), "%d_", &version)

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec(
			"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			version, time.Now().UnixMilli(),
		); err != nil {
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
