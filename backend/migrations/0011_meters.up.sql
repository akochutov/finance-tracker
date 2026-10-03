-- ---------------------------------------------------------------------------
-- services: what a meter measures. Reference data with stable text keys,
-- so it is seeded here (unlike categories, there are no per-database UUIDs).
-- ---------------------------------------------------------------------------
CREATE TABLE services (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    unit TEXT NOT NULL
);

INSERT INTO services (code, name, unit) VALUES
    ('electricity', 'Electricity', 'kWh'),
    ('gas',         'Gas',         'm³'),
    ('water',       'Water',       'm³');

-- ---------------------------------------------------------------------------
-- addresses
-- ---------------------------------------------------------------------------
CREATE TABLE addresses (
    id         UUID        PRIMARY KEY,
    address    TEXT        NOT NULL CHECK (btrim(address) <> ''),
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_addresses_address ON addresses (lower(address));

CREATE TRIGGER trg_addresses_updated
    BEFORE UPDATE ON addresses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- accounts: a utility account (лицевой счёт) for one service at one address
-- ---------------------------------------------------------------------------
CREATE TABLE accounts (
    id         UUID        PRIMARY KEY,
    address_id UUID        NOT NULL REFERENCES addresses (id) ON DELETE RESTRICT,
    service    TEXT        NOT NULL REFERENCES services (code) ON DELETE RESTRICT,
    number     TEXT        NOT NULL CHECK (btrim(number) <> ''),
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_accounts_address_service_number UNIQUE (address_id, service, number)
);

CREATE TRIGGER trg_accounts_updated
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- meters: a physical meter on an account; replaced meters keep their history
-- ---------------------------------------------------------------------------
CREATE TABLE meters (
    id           UUID        PRIMARY KEY,
    account_id   UUID        NOT NULL REFERENCES accounts (id) ON DELETE RESTRICT,
    serial       TEXT        NOT NULL CHECK (btrim(serial) <> ''),
    installed_on DATE        NOT NULL,
    removed_on   DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_meters_account_serial UNIQUE (account_id, serial),
    CONSTRAINT chk_meters_dates CHECK (removed_on IS NULL OR removed_on >= installed_on)
);

CREATE TRIGGER trg_meters_updated
    BEFORE UPDATE ON meters
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- meter_registers: the dials of a meter — one 'single', or 'day' + 'night'
-- ---------------------------------------------------------------------------
CREATE TABLE meter_registers (
    id       UUID NOT NULL PRIMARY KEY,
    meter_id UUID NOT NULL REFERENCES meters (id) ON DELETE CASCADE,
    zone     TEXT NOT NULL CHECK (zone IN ('single', 'day', 'night')),
    CONSTRAINT uq_meter_registers_zone UNIQUE (meter_id, zone)
);

-- ---------------------------------------------------------------------------
-- readings: the value shown on a register on a given day
-- ---------------------------------------------------------------------------
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

-- ---------------------------------------------------------------------------
-- tariffs: price per unit for a service and zone, valid until the next one
-- ---------------------------------------------------------------------------
CREATE TABLE tariffs (
    id         UUID          PRIMARY KEY,
    service    TEXT          NOT NULL REFERENCES services (code) ON DELETE RESTRICT,
    zone       TEXT          NOT NULL CHECK (zone IN ('single', 'day', 'night')),
    valid_from DATE          NOT NULL,
    price      NUMERIC(24,8) NOT NULL CHECK (price > 0),
    currency   TEXT          NOT NULL REFERENCES currencies (code) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT uq_tariffs_service_zone_from UNIQUE (service, zone, valid_from)
);

CREATE TRIGGER trg_tariffs_updated
    BEFORE UPDATE ON tariffs
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();