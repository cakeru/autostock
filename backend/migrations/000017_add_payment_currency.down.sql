ALTER TABLE payments
  DROP COLUMN IF EXISTS currency,
  DROP COLUMN IF EXISTS tendered_amount,
  DROP COLUMN IF EXISTS exchange_rate;
