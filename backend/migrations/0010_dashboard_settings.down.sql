ALTER TABLE settings DROP COLUMN IF EXISTS expense_base_currency;
ALTER TABLE expense_categories DROP COLUMN IF EXISTS include_in_dashboard;

COMMENT ON COLUMN settings.base_currency IS
    'Display currency for dashboards. Must be fiat — enforced in the service layer. Defaults to USD when no row exists.';