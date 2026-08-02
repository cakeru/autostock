BEGIN;

ALTER TABLE vehicle_service_events DROP COLUMN IF EXISTS life_months;
ALTER TABLE products DROP COLUMN IF EXISTS oil_interval_months;
ALTER TABLE products DROP COLUMN IF EXISTS oil_interval_km;

COMMIT;
