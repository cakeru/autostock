-- Customer type: garages (wholesale), retail walk-ins, and companies (fleet/credit).
BEGIN;

ALTER TABLE customers ADD COLUMN customer_type VARCHAR(20) NOT NULL DEFAULT 'retail'
  CHECK (customer_type IN ('garage', 'retail', 'company'));

COMMIT;
