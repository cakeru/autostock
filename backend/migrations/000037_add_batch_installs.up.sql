-- 000037_add_batch_installs.up.sql
-- Optional batch traceability: a mechanic scans a batch QR before fitting the
-- part to a car, recording the ACTUAL batch that went on the vehicle (vs the
-- FIFO-assumed batch the sale deducts). Pure additive log — it never touches
-- stock_quantity, reserved_quantity, or the sale; the POS flow stays the single
-- source of truth for inventory and money.

CREATE TABLE batch_installs (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    batch_id BIGINT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    vehicle_id BIGINT REFERENCES vehicles(id) ON DELETE SET NULL,
    service_job_id BIGINT REFERENCES service_jobs(id) ON DELETE SET NULL,
    position VARCHAR(20),                                   -- e.g. FL / FR / front
    note TEXT,
    installed_by BIGINT REFERENCES users(id),               -- the device/user that scanned
    mechanic_employee_id BIGINT REFERENCES employees(id),   -- who physically fitted it
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batch_installs_batch ON batch_installs(batch_id);
CREATE INDEX idx_batch_installs_vehicle ON batch_installs(vehicle_id);
CREATE INDEX idx_batch_installs_job ON batch_installs(service_job_id);
CREATE INDEX idx_batch_installs_branch ON batch_installs(branch_id);
