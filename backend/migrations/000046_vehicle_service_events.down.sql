ALTER TABLE invoice_items DROP COLUMN IF EXISTS vehicle_event_type;
ALTER TABLE service_job_items DROP COLUMN IF EXISTS vehicle_event_type;

ALTER TABLE vehicle_service_events DROP CONSTRAINT IF EXISTS vehicle_service_events_event_type_check;
ALTER TABLE vehicle_service_events ADD CONSTRAINT vehicle_service_events_event_type_check
    CHECK (event_type IN ('oil', 'tire'));

DROP INDEX IF EXISTS idx_vehicle_service_events_invoice_dedup;
CREATE UNIQUE INDEX idx_vehicle_service_events_invoice_dedup
    ON vehicle_service_events (vehicle_id, event_type, invoice_id)
    WHERE invoice_id IS NOT NULL;
