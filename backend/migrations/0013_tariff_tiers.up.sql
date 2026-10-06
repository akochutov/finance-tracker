-- ---------------------------------------------------------------------------
-- Tariff tiers: a price can depend on the monthly volume.
-- A flat price is a single tier without an upper bound.
-- ---------------------------------------------------------------------------

-- How tiers apply to a month's volume:
--   whole       — the whole volume at the price of the tier it falls into (ENA)
--   progressive — each part of the volume at its own tier's price
ALTER TABLE tariffs
    ADD COLUMN tier_mode TEXT NOT NULL DEFAULT 'whole'
        CHECK (tier_mode IN ('whole', 'progressive'));

CREATE TABLE tariff_tiers (
    tariff_id UUID          NOT NULL REFERENCES tariffs (id) ON DELETE CASCADE,
    position  SMALLINT      NOT NULL CHECK (position >= 1),
    up_to     NUMERIC(14,3) CHECK (up_to > 0),   -- NULL: no upper bound
    price     NUMERIC(24,8) NOT NULL CHECK (price > 0),
    PRIMARY KEY (tariff_id, position)
);

-- Only one open-ended tier per tariff: the last one.
CREATE UNIQUE INDEX uq_tariff_tiers_open ON tariff_tiers (tariff_id) WHERE up_to IS NULL;

-- Every existing tariff becomes one open-ended tier with its price.
INSERT INTO tariff_tiers (tariff_id, position, up_to, price)
SELECT id, 1, NULL, price FROM tariffs;

ALTER TABLE tariffs DROP COLUMN price;