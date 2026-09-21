CREATE TABLE rate_sources (
    kind                    text        NOT NULL CHECK (kind IN ('fiat', 'crypto')),
    source                  text        NOT NULL CHECK (source <> ''),
    url_template            text        NOT NULL CHECK (url_template <> ''),
    poll_interval_seconds   integer     NOT NULL CHECK (poll_interval_seconds > 0),
    request_timeout_seconds integer     NOT NULL DEFAULT 10 CHECK (request_timeout_seconds > 0),
    backfill_start          date,       -- history horizon; NULL until configured
    updated_at              timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (kind)
);

COMMENT ON TABLE  rate_sources                         IS 'Per-class (fiat/crypto) rate-fetch config: provider, URL template, poll interval, request timeout, backfill horizon.';
COMMENT ON COLUMN rate_sources.source                  IS 'Provider name written into exchange_rates.source (e.g. frankfurter). Free text — a new provider needs no migration.';
COMMENT ON COLUMN rate_sources.url_template            IS 'Request URL with placeholders {base}/{quote}/{date}. The adapter fills them and parses the response into the price-in-USD convention.';
COMMENT ON COLUMN rate_sources.poll_interval_seconds   IS 'Configured poll cadence. The 15-minute floor is enforced in the service, not here.';
COMMENT ON COLUMN rate_sources.request_timeout_seconds IS 'Max wait for a full provider response, per request. Guards against slow/hanging free APIs. Applied via context in the provider adapter.';
COMMENT ON COLUMN rate_sources.backfill_start          IS 'Earliest date to pull history from (per class). NULL until the user sets a horizon.';
