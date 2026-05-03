-- 001_init.sql: initial schema
-- Times are stored as INTEGER (Unix milliseconds) for efficiency and unambiguous timezone handling.
-- Booleans are stored as INTEGER (0=false, 1=true).
-- JSON objects (tags) are stored as TEXT.

CREATE TABLE IF NOT EXISTS spans (
    span_id           TEXT    PRIMARY KEY,
    trace_id          TEXT    NOT NULL,
    parent_span_id    TEXT,
    session_id        TEXT,

    tenant_id         TEXT    NOT NULL,
    user_id           TEXT,
    agent_name        TEXT    NOT NULL,
    agent_version     TEXT,
    environment       TEXT,

    span_type         TEXT    NOT NULL, -- 'llm', 'tool', 'retrieval', 'custom'
    name              TEXT    NOT NULL,
    start_time        INTEGER NOT NULL, -- Unix milliseconds
    end_time          INTEGER NOT NULL, -- Unix milliseconds
    duration_ms       INTEGER NOT NULL,

    model             TEXT,
    prompt_tokens     INTEGER,
    completion_tokens INTEGER,
    cost_usd          REAL,

    input             TEXT,
    output            TEXT,
    status            TEXT    NOT NULL, -- 'success', 'error', 'timeout'
    error_message     TEXT,
    tags              TEXT,             -- JSON object

    created_at        INTEGER NOT NULL  -- Unix milliseconds
);

CREATE INDEX IF NOT EXISTS idx_spans_trace   ON spans (trace_id);
CREATE INDEX IF NOT EXISTS idx_spans_agent   ON spans (agent_name, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_spans_tenant  ON spans (tenant_id,  start_time DESC);
CREATE INDEX IF NOT EXISTS idx_spans_user    ON spans (user_id,    start_time DESC);
CREATE INDEX IF NOT EXISTS idx_spans_time    ON spans (start_time DESC);

CREATE TABLE IF NOT EXISTS tenants (
    id                  TEXT    PRIMARY KEY,
    name                TEXT    NOT NULL,
    email               TEXT    UNIQUE NOT NULL,
    monthly_budget_usd  REAL    NOT NULL DEFAULT 0,
    current_spend_usd   REAL    NOT NULL DEFAULT 0,
    enabled             INTEGER NOT NULL DEFAULT 1,
    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS api_keys (
    id              TEXT    PRIMARY KEY,
    tenant_id       TEXT    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            TEXT    NOT NULL,
    key_hash        TEXT    UNIQUE NOT NULL,
    agent_name      TEXT,
    user_id         TEXT,
    tags            TEXT,             -- JSON object
    rate_limit_rpm  INTEGER NOT NULL DEFAULT 0,
    enabled         INTEGER NOT NULL DEFAULT 1,
    last_used       INTEGER,          -- Unix milliseconds, nullable
    created_at      INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_apikeys_tenant ON api_keys (tenant_id);
CREATE INDEX IF NOT EXISTS idx_apikeys_hash   ON api_keys (key_hash);

CREATE TABLE IF NOT EXISTS policies (
    id         TEXT    PRIMARY KEY,
    tenant_id  TEXT    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name       TEXT    NOT NULL,
    condition  TEXT    NOT NULL,  -- CEL expression
    action     TEXT    NOT NULL,  -- 'block', 'alert', 'downgrade'
    enabled    INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_policies_tenant  ON policies (tenant_id);
CREATE INDEX IF NOT EXISTS idx_policies_enabled ON policies (enabled) WHERE enabled = 1;
