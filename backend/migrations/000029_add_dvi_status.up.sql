-- Body type drives which top-down silhouette the vehicle profile draws
-- (sedan / suv / pickup / motorcycle...). Free-ish text with a soft default at
-- render time — an unknown/blank value just falls back to the sedan shape.
ALTER TABLE vehicles ADD COLUMN body_type VARCHAR(20);

-- DVI-style per-part condition, one current status per part per vehicle. The
-- tech taps a part on the car diagram and marks it green (good) / yellow
-- (watch) / red (needs attention); grey (not checked) is simply the absence of
-- a row. Tires and oil are auto-derived from tread/age/due-status elsewhere, so
-- this table holds the manually-assessed parts (brakes, battery, lights...).
CREATE TABLE vehicle_part_status (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    part_key VARCHAR(40) NOT NULL,
    status VARCHAR(10) NOT NULL CHECK (status IN ('green', 'yellow', 'red', 'grey')),
    note TEXT,
    updated_by BIGINT REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_vehicle_part_status_key ON vehicle_part_status (vehicle_id, part_key);
