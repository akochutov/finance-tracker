-- Zones belong to the meter: a day/night meter keeps two registers and
-- every reading names its register. The account no longer carries zones,
-- and the provider's split is gone: day and night come from the readings.

DROP TABLE zone_usage;

ALTER TABLE accounts DROP COLUMN zones;

ALTER TABLE meters
    ADD COLUMN registers TEXT NOT NULL DEFAULT 'single'
        CHECK (registers IN ('single', 'day_night'));

-- Existing readings are of single-register meters.
ALTER TABLE readings
    ADD COLUMN zone TEXT NOT NULL DEFAULT 'single'
        CHECK (zone IN ('single', 'day', 'night'));

-- One reading per register per day; one initial reading per register.
ALTER TABLE readings DROP CONSTRAINT uq_readings_meter_day;
ALTER TABLE readings ADD CONSTRAINT uq_readings_meter_day_zone UNIQUE (meter_id, taken_on, zone);

DROP INDEX uq_readings_initial;
CREATE UNIQUE INDEX uq_readings_initial ON readings (meter_id, zone) WHERE is_initial;