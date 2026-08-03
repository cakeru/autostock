-- Services like wheel balancing/alignment are sold as labor lines; mark the
-- line so the vehicle record logs them automatically when the invoice has a
-- vehicle (value: 'service').
ALTER TABLE invoice_items ADD COLUMN vehicle_event_type TEXT;
ALTER TABLE service_job_items ADD COLUMN vehicle_event_type TEXT;

-- vehicle_service_events gains a generic 'service' type for labor-only work
-- (balancing, alignment, rotation, ...) alongside the product-driven oil/tire.
ALTER TABLE vehicle_service_events DROP CONSTRAINT IF EXISTS vehicle_service_events_event_type_check;
ALTER TABLE vehicle_service_events ADD CONSTRAINT vehicle_service_events_event_type_check
    CHECK (event_type IN ('oil', 'tire', 'service'));

-- Dedup per service name too, so one invoice can log several distinct services
-- (wheel balancing + alignment) instead of only the first per type.
DROP INDEX IF EXISTS idx_vehicle_service_events_invoice_dedup;
CREATE UNIQUE INDEX idx_vehicle_service_events_invoice_dedup
    ON vehicle_service_events (vehicle_id, event_type, invoice_id, product_name) NULLS NOT DISTINCT
    WHERE invoice_id IS NOT NULL;
