-- Job photos the admin takes and uploads at the counter: a captioned gallery
-- per vehicle, with an optional before/after tag. Internal only (never on the
-- customer report). A "before" and an "after" that share a caption render as a
-- side-by-side pair. Distinct from vehicle_records (note-centric) — this is a
-- pure photo gallery.
CREATE TABLE vehicle_photos (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    caption TEXT,
    phase VARCHAR(6) CHECK (phase IN ('before', 'after')),
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_photos_vehicle ON vehicle_photos (vehicle_id, created_at);
