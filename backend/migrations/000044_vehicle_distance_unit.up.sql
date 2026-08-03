-- Per-vehicle distance unit: most cars are km, but imported cars may be
-- miles-only. Each vehicle states its unit (default: the shop's global
-- setting); odometer readings, reminders, and invoices are then consistent
-- per vehicle. The unit is frozen on invoices/jobs at creation (like the
-- exchange rate) so history stays correct even if the vehicle is later edited.
BEGIN;

ALTER TABLE vehicles
  ADD COLUMN distance_unit VARCHAR(3) NOT NULL DEFAULT 'km'
  CHECK (distance_unit IN ('km', 'mi'));

ALTER TABLE invoices
  ADD COLUMN mileage_unit VARCHAR(3) NOT NULL DEFAULT 'km'
  CHECK (mileage_unit IN ('km', 'mi'));

ALTER TABLE service_jobs
  ADD COLUMN mileage_unit VARCHAR(3) NOT NULL DEFAULT 'km'
  CHECK (mileage_unit IN ('km', 'mi'));

COMMIT;
