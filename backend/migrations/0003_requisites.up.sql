CREATE TABLE bank_requisites (
    id                        UUID        PRIMARY KEY,
    company_id                UUID        NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    beneficiary_name          TEXT        NOT NULL,
    account_number            TEXT        NOT NULL,
    bank_name                 TEXT        NOT NULL,
    bank_swift                TEXT        NOT NULL,
    bank_address              TEXT,
    correspondent_bank_name   TEXT,
    correspondent_bank_swift  TEXT,
    intermediary_bank_name    TEXT,
    intermediary_bank_swift   TEXT,
    valid_from                TIMESTAMPTZ NOT NULL DEFAULT now(),
    valid_to                  TIMESTAMPTZ,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bank_requisites_valid_period CHECK (valid_to IS NULL OR valid_to > valid_from)
);

CREATE INDEX idx_bank_requisites_company ON bank_requisites (company_id);

CREATE TRIGGER trg_bank_requisites_updated
    BEFORE UPDATE ON bank_requisites
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE crypto_requisites (
    id             UUID        PRIMARY KEY,
    company_id     UUID        NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    network        TEXT        NOT NULL,
    wallet_address TEXT        NOT NULL,
    valid_from     TIMESTAMPTZ NOT NULL DEFAULT now(),
    valid_to       TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT crypto_requisites_valid_period CHECK (valid_to IS NULL OR valid_to > valid_from)
);

CREATE INDEX idx_crypto_requisites_company ON crypto_requisites (company_id);

CREATE TRIGGER trg_crypto_requisites_updated
    BEFORE UPDATE ON crypto_requisites
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();