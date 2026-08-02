-- Unified service-life rating on products: distance (km/mi) + days + months,
-- for BOTH tire and oil products. Selling the product records the rating on
-- the vehicle's service event, which drives the due-for-service reminder.
BEGIN;

ALTER TABLE products
  ADD COLUMN life_km INT,
  ADD COLUMN life_days INT,
  ADD COLUMN life_months INT;

-- Carry over existing ratings (tire km life / oil km + months).
UPDATE products SET life_km = rated_life_km WHERE rated_life_km IS NOT NULL;
UPDATE products SET life_km = oil_interval_km WHERE oil_interval_km IS NOT NULL;
UPDATE products SET life_months = oil_interval_months WHERE oil_interval_months IS NOT NULL;

ALTER TABLE products
  DROP COLUMN rated_life_km,
  DROP COLUMN oil_interval_km,
  DROP COLUMN oil_interval_months;

-- Events record the full rating that applied at sale time (days in addition
-- to the existing km and months).
ALTER TABLE vehicle_service_events ADD COLUMN life_days INT;

COMMIT;
