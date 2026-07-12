DROP TABLE IF EXISTS vehicle_service_events;
ALTER TABLE vehicles DROP COLUMN IF EXISTS tire_interval_months;
ALTER TABLE vehicles DROP COLUMN IF EXISTS tire_interval_km;
ALTER TABLE vehicles DROP COLUMN IF EXISTS oil_interval_months;
ALTER TABLE vehicles DROP COLUMN IF EXISTS oil_interval_km;
ALTER TABLE products DROP COLUMN IF EXISTS is_oil_product;
