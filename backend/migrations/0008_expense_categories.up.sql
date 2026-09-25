CREATE TABLE expense_groups (
    id         UUID        PRIMARY KEY,
    name       TEXT        NOT NULL CHECK (btrim(name) <> ''),
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_expense_groups_name
    ON expense_groups (lower(name));

CREATE TRIGGER trg_expense_groups_updated
    BEFORE UPDATE ON expense_groups
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE expense_categories (
    id         UUID        PRIMARY KEY,
    group_id   UUID        NOT NULL REFERENCES expense_groups(id) ON DELETE RESTRICT,
    name       TEXT        NOT NULL CHECK (btrim(name) <> ''),
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_expense_categories_group_name
    ON expense_categories (group_id, lower(name));

CREATE TRIGGER trg_expense_categories_updated
    BEFORE UPDATE ON expense_categories
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();