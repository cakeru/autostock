-- vehicles was missing branch_id entirely, which silently broke plate search
-- (the search query filters on v.branch_id, so it errored and was swallowed —
-- vehicle search has never actually returned a result). Backfill it from the
-- owning customer and make it a real scoping column going forward.
ALTER TABLE vehicles ADD COLUMN branch_id BIGINT;
UPDATE vehicles v SET branch_id = c.branch_id FROM customers c WHERE c.id = v.customer_id;
ALTER TABLE vehicles ALTER COLUMN branch_id SET NOT NULL;
ALTER TABLE vehicles ADD CONSTRAINT vehicles_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);
CREATE INDEX idx_vehicles_branch ON vehicles (branch_id);

-- A vehicle's permanent service record: the photo + note the mechanic takes
-- (alignment printout, damage, anything), optionally tied to the invoice/job
-- it happened during. Loose by design — no required structure — so it's fast
-- enough to actually get used at the counter.
CREATE TABLE vehicle_records (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    invoice_id BIGINT REFERENCES invoices(id) ON DELETE SET NULL,
    service_job_id BIGINT REFERENCES service_jobs(id) ON DELETE SET NULL,
    mileage INT,
    note TEXT,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_records_vehicle ON vehicle_records (vehicle_id, created_at DESC);

CREATE TABLE vehicle_record_photos (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES vehicle_records(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_record_photos_record ON vehicle_record_photos (record_id);
