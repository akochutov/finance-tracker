CREATE TABLE incomes (
    id                       UUID          PRIMARY KEY,
    payer_id                 UUID          NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    beneficiary_id           UUID          NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    amount                   NUMERIC(24,8) NOT NULL CHECK (amount > 0),
    currency                 TEXT          NOT NULL REFERENCES currencies(code) ON DELETE RESTRICT,
    occurred_at              TIMESTAMPTZ   NOT NULL,
    payment_type             TEXT          NOT NULL CHECK (payment_type IN ('bank', 'crypto')),
    payer_requisite_id       UUID          NOT NULL,
    beneficiary_requisite_id UUID          NOT NULL,
    note                     TEXT,
    is_active                BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_incomes_payer       ON incomes (payer_id);
CREATE INDEX idx_incomes_beneficiary ON incomes (beneficiary_id);
CREATE INDEX idx_incomes_occurred_at ON incomes (occurred_at DESC);
CREATE INDEX idx_incomes_currency    ON incomes (currency);

CREATE TRIGGER trg_incomes_updated
    BEFORE UPDATE ON incomes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();