-- An outbox for Telegram notifications: business services just insert a row
-- (in the same transaction as the source event, so a notification only
-- exists if the underlying change actually committed); a background loop
-- delivers pending rows independent of the request that created them, so a
-- slow/unreachable Telegram API never adds latency or risk to the app itself.
CREATE TABLE telegram_events (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    topic VARCHAR(30) NOT NULL,
    event_type VARCHAR(40) NOT NULL,
    reference_type VARCHAR(30),
    reference_id BIGINT,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'skipped', 'failed')),
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ
);

CREATE INDEX idx_telegram_events_pending ON telegram_events (id) WHERE status = 'pending';

-- A job's own posted message, so a status change can edit that message in
-- place instead of spamming a new one per update.
ALTER TABLE service_jobs ADD COLUMN telegram_chat_id VARCHAR(64);
ALTER TABLE service_jobs ADD COLUMN telegram_message_id BIGINT;
