-- 000036_add_photo_customer_visible.down.sql

ALTER TABLE vehicle_photos DROP COLUMN IF EXISTS customer_visible;
