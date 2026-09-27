CREATE TABLE expenses (
    id           UUID        PRIMARY KEY,
    occurred_on  DATE        NOT NULL,
    currency     TEXT        NOT NULL REFERENCES currencies(code),
    payment_type TEXT        CHECK (payment_type IN ('bank', 'crypto', 'cash')),
    note         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_expenses_occurred_on ON expenses (occurred_on);

CREATE TRIGGER trg_expenses_updated
    BEFORE UPDATE ON expenses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE expense_items (
    id          UUID          PRIMARY KEY,
    expense_id  UUID          NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    line_no     INT           NOT NULL CHECK (line_no > 0),
    description TEXT          NOT NULL CHECK (btrim(description) <> ''),
    category_id UUID          NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    price       NUMERIC(24,8) NOT NULL CHECK (price >= 0),
    quantity    NUMERIC(12,3) NOT NULL DEFAULT 1 CHECK (quantity > 0),
    discount    NUMERIC(24,8) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    amount      NUMERIC(24,8) GENERATED ALWAYS AS (price * quantity - discount) STORED,
    period_from DATE,
    period_to   DATE,
    CONSTRAINT uq_expense_items_line UNIQUE (expense_id, line_no),
    CONSTRAINT chk_expense_items_amount CHECK (price * quantity - discount >= 0),
    CONSTRAINT chk_expense_items_period CHECK (
        (period_from IS NULL AND period_to IS NULL)
        OR (period_from IS NOT NULL AND period_to IS NOT NULL AND period_to >= period_from)
    )
);