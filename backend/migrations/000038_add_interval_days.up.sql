-- 000038_add_interval_days.up.sql
-- Per-vehicle day-based reminder overrides. The reminder engine works in days
-- (a fixed day interval yields a CERTAIN calendar date, unlike the drifting
-- mileage projection), so replace the never-read *_months columns with *_days.
-- Oil already has a shop-default day interval; tires gain one here so a car can
-- carry e.g. "change in 90 days" and land on a firm date in the calendar.

ALTER TABLE vehicles ADD COLUMN oil_interval_days INT;
ALTER TABLE vehicles ADD COLUMN tire_interval_days INT;

-- Dead columns: written by the intervals update but never read by the engine.
ALTER TABLE vehicles DROP COLUMN IF EXISTS oil_interval_months;
ALTER TABLE vehicles DROP COLUMN IF EXISTS tire_interval_months;
