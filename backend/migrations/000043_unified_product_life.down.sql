BEGIN;

ALTER TABLE vehicle_service_events DROP COLUMN IF EXISTS life_days;

ALTER TABLE products
  ADD COLUMN rated_life_km INT,
  ADD COLUMN oil_interval_km INT,
  ADD COLUMN oil_interval_months INT;

UPDATE products SET rated_life_km = life_km WHERE life_km IS NOT NULL;
UPDATE products SET oil_interval_km = life_km WHERE life_km IS NOT NULL;
UPDATE products SET oil_interval_months = life_months WHERE life_months IS NOT NULL;

ALTER TABLE products
  DROP COLUMN life_km,
  DROP COLUMN life_days,
  DROP COLUMN life_months;

COMMIT;
