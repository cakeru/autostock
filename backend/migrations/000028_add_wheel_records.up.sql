-- Per-corner wheel data — the structured layer the vehicle profile was missing.
-- A "wheel service" is one snapshot of the car's four corners (plus optional
-- spare) taken during a visit: which tire is where, its tread/age, and the
-- alignment numbers read off the shop's alignment monitor. Grouping the corners
-- under one snapshot keeps the top-down car diagram coherent and lets each
-- visit be compared against the last.
CREATE TABLE vehicle_wheel_services (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    performed_at TIMESTAMPTZ NOT NULL,
    mileage INT,
    invoice_id BIGINT REFERENCES invoices(id) ON DELETE SET NULL,
    service_job_id BIGINT REFERENCES service_jobs(id) ON DELETE SET NULL,
    notes TEXT,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_wheel_services_vehicle ON vehicle_wheel_services (vehicle_id, performed_at DESC);

-- One row per wheel position. Every field is optional: an alignment-only job
-- fills the angle columns, a tire swap fills the tire columns, and a plain
-- rotation might only set tread + wear_note. Alignment readings are stored as
-- short text (e.g. '4°29'', '0.40', '-1°01''), so a tech types exactly what the
-- monitor shows — no unit conversion at the counter — while each
-- corner × metric × (before/after) stays its own discrete cell for the diagram.
CREATE TABLE wheel_service_corners (
    id BIGSERIAL PRIMARY KEY,
    wheel_service_id BIGINT NOT NULL REFERENCES vehicle_wheel_services(id) ON DELETE CASCADE,
    position VARCHAR(6) NOT NULL CHECK (position IN ('FL', 'FR', 'RL', 'RR', 'SPARE')),
    tire_product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    tire_brand TEXT,
    tire_size TEXT,
    tire_dot VARCHAR(10),          -- sidewall DOT week/year, e.g. '3823' = wk38/2023 (true tire age)
    tread_mm NUMERIC(4, 1),
    pressure NUMERIC(4, 1),
    camber_before VARCHAR(20),
    camber_after VARCHAR(20),
    caster_before VARCHAR(20),
    caster_after VARCHAR(20),
    toe_before VARCHAR(20),
    toe_after VARCHAR(20),
    wear_note TEXT
);
CREATE UNIQUE INDEX idx_wheel_service_corners_pos ON wheel_service_corners (wheel_service_id, position);

CREATE TABLE wheel_service_photos (
    id BIGSERIAL PRIMARY KEY,
    wheel_service_id BIGINT NOT NULL REFERENCES vehicle_wheel_services(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wheel_service_photos_service ON wheel_service_photos (wheel_service_id);

-- General parts-replaced log for non-tire items (brake pads, filters, wipers,
-- battery...). Deliberately flat and loose — part name + when + optional
-- position/inventory/invoice link — so the whole service history lives in one
-- place without forcing a rigid parts catalogue.
CREATE TABLE vehicle_parts (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    part_name TEXT NOT NULL,
    position TEXT,                 -- free text, e.g. 'front', 'rear-left'
    replaced_at TIMESTAMPTZ NOT NULL,
    mileage INT,
    product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    invoice_id BIGINT REFERENCES invoices(id) ON DELETE SET NULL,
    service_job_id BIGINT REFERENCES service_jobs(id) ON DELETE SET NULL,
    notes TEXT,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_parts_vehicle ON vehicle_parts (vehicle_id, replaced_at DESC);
