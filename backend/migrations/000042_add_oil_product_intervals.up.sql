-- Oil products carry their own change interval (distance + months) so an
-- oil change logged from an invoice drives that vehicle's due-for-service
-- estimate with the actual product's rating instead of only the branch default.
BEGIN;

ALTER TABLE products ADD COLUMN oil_interval_km INT;
ALTER TABLE products ADD COLUMN oil_interval_months INT;

-- Events record the interval that applied at the time (mirrors life_km for tires).
ALTER TABLE vehicle_service_events ADD COLUMN life_months INT;

COMMIT;
