-- User-configurable database backup schedules (instance-wide, not per-branch:
-- there is one database per server). The Settings page manages these; the
-- backend scheduler runs pg_dump per schedule into BACKUP_DIR.
CREATE TABLE backup_schedules (
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT NOT NULL,
    cron           TEXT NOT NULL,
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    retention_days INTEGER NOT NULL DEFAULT 14,
    last_run_at    TIMESTAMPTZ,
    last_status    TEXT NOT NULL DEFAULT 'never', -- never | success | error
    last_error     TEXT NOT NULL DEFAULT '',
    next_run_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Default schedule so a fresh install keeps the old "nightly dump" behaviour.
INSERT INTO backup_schedules (name, cron, enabled, retention_days)
VALUES ('Nightly backup', '0 2 * * *', TRUE, 14);
