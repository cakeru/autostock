-- 000035_add_photo_taken_at.up.sql
-- Photo evidence is often uploaded at the counter days after the work was done.
-- taken_at lets the admin date a photo to the visit it documents, so the service
-- timeline clusters it onto the right day instead of the upload day. Falls back
-- to created_at when unset.

ALTER TABLE vehicle_photos ADD COLUMN taken_at TIMESTAMPTZ;
