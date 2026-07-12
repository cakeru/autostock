-- The removed tire's tread at a replacement, so the visit summary can show the
-- before→after story a customer cares about ("your fronts were down to 2.5mm,
-- we fitted new at 8mm"). NULL = not a replacement / not measured; tread_mm
-- stays the current (post-service) reading.
ALTER TABLE wheel_service_corners ADD COLUMN tread_before_mm NUMERIC(4, 1);
