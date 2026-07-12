-- Tire reminders move from a fixed shop-wide interval to a per-tire life:
-- the admin sets a rated life (km) once per tire product, every install
-- inherits it, and the reminder is anchored to the odometer at that install
-- (due at install_km + life_km). DOT/manufacture age no longer drives anything.
ALTER TABLE products ADD COLUMN rated_life_km INT;

-- The effective life captured on a specific tire install (from the product's
-- rated life, or a manual override). NULL falls back to the branch/vehicle
-- default tire life. Oil events ignore this column.
ALTER TABLE vehicle_service_events ADD COLUMN life_km INT;

-- A controlled key tying a logged part replacement to a reminder rule (the DVI
-- part keys: brakes_front, battery, air_filter, ...). NULL = a free-text part
-- with no reminder. Lets "last replaced" anchor a part's next-due estimate.
ALTER TABLE vehicle_parts ADD COLUMN part_key VARCHAR(40);
CREATE INDEX idx_vehicle_parts_key ON vehicle_parts (vehicle_id, part_key, replaced_at DESC) WHERE part_key IS NOT NULL;
