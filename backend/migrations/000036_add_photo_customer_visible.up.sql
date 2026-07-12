-- 000036_add_photo_customer_visible.up.sql
-- Per-photo opt-in to the customer-facing condition report. Photos are internal
-- by default (the shop's own record); the admin flips the ones worth showing —
-- a clean before/after alignment, say — so the good evidence reaches the
-- customer while messy shop shots stay private.

ALTER TABLE vehicle_photos ADD COLUMN customer_visible BOOLEAN NOT NULL DEFAULT false;
