CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE currencies (
    code           TEXT        PRIMARY KEY,
    name           TEXT        NOT NULL,
    kind           TEXT        NOT NULL CHECK (kind IN ('fiat', 'crypto')),
    decimal_places SMALLINT    NOT NULL CHECK (decimal_places >= 0),
    is_active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_currencies_updated
    BEFORE UPDATE ON currencies
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();