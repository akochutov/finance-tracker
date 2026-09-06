ALTER TABLE incomes           DROP COLUMN IF EXISTS transaction_ref;
ALTER TABLE crypto_requisites DROP COLUMN IF EXISTS note;
ALTER TABLE bank_requisites   DROP COLUMN IF EXISTS note;