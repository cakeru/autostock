-- Customer-facing share link: a random, unguessable token per vehicle. The
-- token IS the authorization for the public read-only report (sequential row
-- ids stay internal-only behind JWT+branch scoping). NULL = no active link;
-- revoking a link just clears the column.
ALTER TABLE vehicles ADD COLUMN share_token TEXT;
CREATE UNIQUE INDEX idx_vehicles_share_token ON vehicles (share_token) WHERE share_token IS NOT NULL;
