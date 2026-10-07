-- Back to single-value meters: readings of day and night registers have
-- no place in the old layout and are removed.
DELETE FROM readings WHERE zone <> 'single';

DROP INDEX uq_readings_initial;
CREATE UNIQUE INDEX uq_readings_initial ON readings (meter_id) WHERE is_initial;

ALTER TABLE readings DROP CONSTRAINT uq_readings_meter_day_zone;
ALTER TABLE readings ADD CONSTRAINT uq_readings_meter_day UNIQUE (meter_id, taken_on);

ALTER TABLE readings DROP COLUMN zone;
ALTER TABLE meters DROP COLUMN registers;

ALTER TABLE accounts
    ADD COLUMN zones TEXT NOT NULL DEFAULT 'single'
        CHECK (zones IN ('single', 'day_night'));

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