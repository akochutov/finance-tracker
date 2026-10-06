DROP TABLE IF EXISTS zone_usage;
DROP TABLE IF EXISTS readings;

ALTER TABLE accounts DROP COLUMN IF EXISTS zones;

-- Back to the 0011 layout: registers per meter, readings per register.
CREATE TABLE meter_registers (
    id       UUID NOT NULL PRIMARY KEY,
    meter_id UUID NOT NULL REFERENCES meters (id) ON DELETE CASCADE,
    zone     TEXT NOT NULL CHECK (zone IN ('single', 'day', 'night')),
    CONSTRAINT uq_meter_registers_zone UNIQUE (meter_id, zone)
);

CREATE TABLE readings (
    id          UUID          PRIMARY KEY,
    register_id UUID          NOT NULL REFERENCES meter_registers (id) ON DELETE CASCADE,
    taken_on    DATE          NOT NULL,
    value       NUMERIC(14,3) NOT NULL CHECK (value >= 0),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT uq_readings_register_day UNIQUE (register_id, taken_on)
);

CREATE TRIGGER trg_readings_updated
    BEFORE UPDATE ON readings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();