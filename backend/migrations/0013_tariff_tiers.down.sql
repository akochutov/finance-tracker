-- Back to one price per tariff: the open-ended tier's price is kept,
-- the other tiers are lost.
ALTER TABLE tariffs ADD COLUMN price NUMERIC(24,8);

UPDATE tariffs t
SET price = tt.price
FROM tariff_tiers tt
WHERE tt.tariff_id = t.id AND tt.up_to IS NULL;

ALTER TABLE tariffs
    ALTER COLUMN price SET NOT NULL,
    ADD CONSTRAINT tariffs_price_check CHECK (price > 0);

DROP TABLE IF EXISTS tariff_tiers;

ALTER TABLE tariffs DROP COLUMN IF EXISTS tier_mode;