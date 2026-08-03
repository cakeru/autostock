BEGIN;

ALTER TABLE service_jobs DROP COLUMN IF EXISTS mileage_unit;
ALTER TABLE invoices DROP COLUMN IF EXISTS mileage_unit;
ALTER TABLE vehicles DROP COLUMN IF EXISTS distance_unit;

COMMIT;
