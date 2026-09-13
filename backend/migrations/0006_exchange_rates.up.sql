-- Foundation for the exchange-rates domain (Slice 1).
--   exchange_rates — historical rates stored against the USD pivot
--   settings       — global app settings (currently: display/base currency)

-- ---------------------------------------------------------------------------
-- exchange_rates
-- ---------------------------------------------------------------------------
-- One row = the price of 1 unit of `currency` expressed in USD, observed at
-- `rate_at`, delivered by `source`. USD is the pivot: its price in USD is
-- always 1, so it is NOT stored here — the conversion service short-circuits
-- the pivot to 1.
CREATE TABLE exchange_rates (
    currency   text            NOT NULL REFERENCES currencies (code),
    source     text            NOT NULL CHECK (source <> ''),
    rate_at    timestamptz     NOT NULL,                    -- market moment (as-of key)
    rate       numeric(36, 18) NOT NULL CHECK (rate > 0),   -- price of 1 unit in USD
    fetched_at timestamptz     NOT NULL DEFAULT now(),      -- when we recorded/pulled it

    -- Natural key: one rate per (currency, source, moment).
    -- Doubles as the ON CONFLICT target for idempotent upserts.
    PRIMARY KEY (currency, source, rate_at)
);

-- As-of lookup: latest rate for a currency at or before a target moment.
--   WHERE currency = $1 AND rate_at <= $2 ORDER BY rate_at DESC LIMIT 1
CREATE INDEX exchange_rates_currency_rate_at_idx
    ON exchange_rates (currency, rate_at DESC);

COMMENT ON TABLE  exchange_rates            IS 'Historical FX/crypto rates, each the price of 1 unit of the currency in USD (the pivot).';
COMMENT ON COLUMN exchange_rates.rate       IS 'Price of 1 unit of currency in USD. NUMERIC(36,18) for crypto-grade precision.';
COMMENT ON COLUMN exchange_rates.rate_at    IS 'Market moment the rate applies to; used for as-of lookups.';
COMMENT ON COLUMN exchange_rates.fetched_at IS 'When this row was recorded (manual) or pulled (provider).';
COMMENT ON COLUMN exchange_rates.source     IS 'Provenance: manual, frankfurter, ... Free text so new providers need no migration.';

-- ---------------------------------------------------------------------------
-- settings (singleton)
-- ---------------------------------------------------------------------------
-- Boolean PK pinned to TRUE guarantees at most one row ever exists.
CREATE TABLE settings (
    id            boolean     PRIMARY KEY DEFAULT true CHECK (id),
    base_currency text        NOT NULL REFERENCES currencies (code),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE  settings               IS 'Global app settings, exactly one row (singleton via boolean PK).';
COMMENT ON COLUMN settings.base_currency IS 'Display currency for dashboards. Must be fiat — enforced in the service layer. Defaults to USD when no row exists.';

-- No row seeded on purpose: currencies is user-managed reference data and may
-- be empty on a fresh DB, which would break the FK. The settings service
-- returns 'USD' when the table is empty and inserts the row on first write.