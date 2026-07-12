-- Odometer reading (kilometers) at the time of service. Captured per visit —
-- it changes every time the car comes in — so it lives on the invoice and the
-- service job rather than as a static property of the vehicle. Printed on the
-- customer invoice next to the plate/vehicle details.
ALTER TABLE invoices ADD COLUMN mileage INT;
ALTER TABLE service_jobs ADD COLUMN mileage INT;
