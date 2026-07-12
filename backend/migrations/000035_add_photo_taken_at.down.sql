-- 000035_add_photo_taken_at.down.sql

ALTER TABLE vehicle_photos DROP COLUMN IF EXISTS taken_at;
