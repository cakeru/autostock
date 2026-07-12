-- Oil-change reminders can't reliably key off the free-text `category` field
-- (an owner can rename "Engine oil" to anything, silently breaking any string
-- match) so oil products get an explicit, owner-editable flag. Tires need no
-- equivalent flag: products.type = 'tire' is already a hard CHECK constraint.
ALTER TABLE products ADD COLUMN is_oil_product BOOLEAN NOT NULL DEFAULT false;
UPDATE products SET is_oil_product = true WHERE type <> 'tire' AND category ILIKE '%oil%';

-- Per-vehicle interval overrides (NULL = use the branch's default setting) —
-- e.g. a taxi that drives far more than average per month.
ALTER TABLE vehicles ADD COLUMN oil_interval_km INT;
ALTER TABLE vehicles ADD COLUMN oil_interval_months INT;
ALTER TABLE vehicles ADD COLUMN tire_interval_km INT;
ALTER TABLE vehicles ADD COLUMN tire_interval_months INT;

-- One row per detected (or manually logged) oil change / tire install, so the
-- "due for service" estimate has a real history to project from instead of
-- guessing off a single most-recent invoice.
CREATE TABLE vehicle_service_events (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    event_type VARCHAR(10) NOT NULL CHECK (event_type IN ('oil', 'tire')),
    mileage INT,
    occurred_at TIMESTAMPTZ NOT NULL,
    invoice_id BIGINT REFERENCES invoices(id) ON DELETE SET NULL,
    service_job_id BIGINT REFERENCES service_jobs(id) ON DELETE SET NULL,
    product_name TEXT,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_service_events_vehicle ON vehicle_service_events (vehicle_id, event_type, occurred_at DESC);
-- One auto-logged event per invoice line item's product per type, so voiding
-- and re-issuing (or a bug retry) can't double-log the same sale.
CREATE UNIQUE INDEX idx_vehicle_service_events_invoice_dedup ON vehicle_service_events (vehicle_id, event_type, invoice_id) WHERE invoice_id IS NOT NULL;
