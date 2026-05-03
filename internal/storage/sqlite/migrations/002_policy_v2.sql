-- 002_policy_v2.sql: add priority/description, rename condition→expression, remove FK to tenants
CREATE TABLE policies_new (
    id          TEXT    PRIMARY KEY,
    tenant_id   TEXT    NOT NULL,
    name        TEXT    NOT NULL,
    expression  TEXT    NOT NULL,
    action      TEXT    NOT NULL,
    priority    INTEGER NOT NULL DEFAULT 0,
    description TEXT    NOT NULL DEFAULT '',
    enabled     INTEGER NOT NULL DEFAULT 1,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

INSERT OR IGNORE INTO policies_new
    (id, tenant_id, name, expression, action, priority, description, enabled, created_at, updated_at)
    SELECT id, tenant_id, name, condition, action, 0, '', enabled, created_at, updated_at
    FROM policies;

DROP TABLE policies;

ALTER TABLE policies_new RENAME TO policies;

CREATE INDEX IF NOT EXISTS idx_policies_tenant   ON policies (tenant_id);
CREATE INDEX IF NOT EXISTS idx_policies_enabled  ON policies (enabled) WHERE enabled = 1;
CREATE INDEX IF NOT EXISTS idx_policies_priority ON policies (enabled, priority);
