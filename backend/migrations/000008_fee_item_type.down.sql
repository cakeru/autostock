UPDATE invoice_items SET item_type = 'custom' WHERE item_type = 'fee';
ALTER TABLE invoice_items DROP CONSTRAINT IF EXISTS invoice_items_item_type_check;
ALTER TABLE invoice_items ADD CONSTRAINT invoice_items_item_type_check
    CHECK (item_type IN ('product', 'labor', 'custom'));
