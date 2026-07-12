-- Payments keep their base (USD) amount for the ledger, plus how they were
-- actually tendered — dual-currency (dollars/riel) is normal in Cambodia.
ALTER TABLE payments
  ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD' CHECK (currency IN ('USD', 'KHR')),
  ADD COLUMN tendered_amount DECIMAL(14, 2),
  ADD COLUMN exchange_rate DECIMAL(10, 2);
