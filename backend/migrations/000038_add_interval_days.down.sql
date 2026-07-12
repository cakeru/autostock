-- 000038_add_interval_days.down.sql

ALTER TABLE vehicles ADD COLUMN oil_interval_months INT;
ALTER TABLE vehicles ADD COLUMN tire_interval_months INT;

ALTER TABLE vehicles DROP COLUMN IF EXISTS oil_interval_days;
ALTER TABLE vehicles DROP COLUMN IF EXISTS tire_interval_days;
