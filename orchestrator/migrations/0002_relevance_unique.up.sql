-- Prevents duplicate relevance rows when detect-relevance runs more than
-- once for the same bill on the same day (idempotency, matching the
-- UNIQUE(source, external_id) approach already used on bills).
--
-- A plain column (rather than an index on evaluated_at::date) is used
-- because casting timestamptz to date depends on the session's TimeZone
-- setting, so Postgres refuses to index that expression (not IMMUTABLE).
ALTER TABLE relevance_results ADD COLUMN evaluated_date DATE NOT NULL DEFAULT CURRENT_DATE;

CREATE UNIQUE INDEX relevance_results_bill_method_date_idx
    ON relevance_results (bill_id, method, evaluated_date);
