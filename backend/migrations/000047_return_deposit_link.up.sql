-- Lets undoing a store-credit return remove exactly the deposit it created.
ALTER TABLE deposits ADD COLUMN return_id BIGINT REFERENCES returns(id) ON DELETE SET NULL;
CREATE INDEX idx_deposits_return_id ON deposits (return_id);
