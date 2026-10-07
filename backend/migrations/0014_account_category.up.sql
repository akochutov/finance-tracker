-- The expense category of an account's payments: expense lines of this
-- category, with a period, are what was actually paid for a billing month.
ALTER TABLE accounts
    ADD COLUMN expense_category_id UUID
        REFERENCES expense_categories (id) ON DELETE RESTRICT;