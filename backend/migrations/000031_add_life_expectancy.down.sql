DROP INDEX IF EXISTS idx_vehicle_parts_key;
ALTER TABLE vehicle_parts DROP COLUMN IF EXISTS part_key;
ALTER TABLE vehicle_service_events DROP COLUMN IF EXISTS life_km;
ALTER TABLE products DROP COLUMN IF EXISTS rated_life_km;
