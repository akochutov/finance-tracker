-- Meters show one value; the provider splits a month into zones.
-- No real readings exist yet, so readings are rebuilt from scratch.

DROP TABLE readings;
DROP TABLE meter_registers;

-- Tariff zones are a term of the account, chosen when it is created.
ALTER TABLE accounts
    ADD COLUMN zones TEXT NOT NULL DEFAULT 'single'
        CHECK (zones IN ('single', 'day_night'));

-- ---------------------------------------------------------------------------
-- readings: one value per meter per day; the initial one is the starting
-- point of the meter (e.g. the value found when moving in), never consumption
-- ---------------------------------------------------------------------------
CREATE TABLE readings (
    id         UUID          PRIMARY KEY,
    meter_id   UUID          NOT NULL REFERENCES meters (id) ON DELETE CASCADE,
    taken_on   DATE          NOT NULL,
    value      NUMERIC(14,3) NOT NULL CHECK (value >= 0),
    is_initial BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT uq_readings_meter_day UNIQUE (meter_id, taken_on)
);

-- At most one initial reading per meter.
CREATE UNIQUE INDEX uq_readings_initial ON readings (meter_id) WHERE is_initial;

CREATE TRIGGER trg_readings_updated
    BEFORE UPDATE ON readings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- zone_usage: the provider's split of a billing month into zones,
-- for accounts with day/night tariffs
-- ---------------------------------------------------------------------------
CREATE TABLE zone_usage (
    account_id UUID          NOT NULL REFERENCES accounts (id) ON DELETE RESTRICT,
    month      DATE          NOT NULL CHECK (EXTRACT(DAY FROM month) = 1),
    zone       TEXT          NOT NULL CHECK (zone IN ('day', 'night')),
    quantity   NUMERIC(14,3) NOT NULL CHECK (quantity >= 0),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, month, zone)
);

CREATE TRIGGER trg_zone_usage_updated
    BEFORE UPDATE ON zone_usage
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();