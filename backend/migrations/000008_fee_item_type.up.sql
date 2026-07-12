-- Allow a dedicated 'fee' line type on invoice items (disposal/environmental,
-- shop supplies) so fee revenue is reportable separately from labor.
ALTER TABLE invoice_items DROP CONSTRAINT IF EXISTS invoice_items_item_type_check;
ALTER TABLE invoice_items ADD CONSTRAINT invoice_items_item_type_check
    CHECK (item_type IN ('product', 'labor', 'custom', 'fee'));
