-- Payment proofs for bank/wallet transfers (e.g. ABA): the customer's Trx ID
-- and an optional photo of the transfer confirmation.
BEGIN;

ALTER TABLE payments
  ADD COLUMN reference VARCHAR(120),
  ADD COLUMN proof_url TEXT;

COMMIT;
