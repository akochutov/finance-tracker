ALTER TABLE expense_categories
    ADD COLUMN include_in_dashboard BOOLEAN NOT NULL DEFAULT TRUE;

COMMENT ON COLUMN expense_categories.include_in_dashboard IS
    'Whether lines of this category count on the expenses dashboard. FALSE for transfers such as currency purchases.';

ALTER TABLE settings
    ADD COLUMN expense_base_currency TEXT REFERENCES currencies (code);

COMMENT ON COLUMN settings.base_currency IS
    'Display currency for the income dashboard. Must be fiat — enforced in the service layer. Defaults to USD when no row exists.';
COMMENT ON COLUMN settings.expense_base_currency IS
    'Display currency for the expenses dashboard. Must be fiat — enforced in the service layer. NULL falls back to base_currency.';