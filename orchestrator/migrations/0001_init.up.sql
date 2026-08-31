CREATE TABLE bills (
    id BIGSERIAL PRIMARY KEY,
    source TEXT NOT NULL CHECK (source IN ('camara', 'senado')),
    external_id TEXT NOT NULL,
    type TEXT NOT NULL,
    number INTEGER NOT NULL,
    year INTEGER NOT NULL,
    ementa TEXT NOT NULL,
    author TEXT,
    presented_date DATE NOT NULL,
    url TEXT NOT NULL,
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source, external_id)
);

CREATE TABLE ingestion_runs (
    id BIGSERIAL PRIMARY KEY,
    run_date DATE NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('camara', 'senado')),
    status TEXT NOT NULL CHECK (status IN ('ok', 'partial', 'failed')),
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    items_fetched INTEGER NOT NULL DEFAULT 0,
    error_message TEXT
);

CREATE TABLE relevance_results (
    id BIGSERIAL PRIMARY KEY,
    bill_id BIGINT NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    method TEXT NOT NULL CHECK (method IN ('keyword', 'ml')),
    is_relevant BOOLEAN NOT NULL,
    matched_keywords JSONB NOT NULL DEFAULT '[]'::jsonb,
    confidence NUMERIC(4,3),
    model_version TEXT,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reports (
    id BIGSERIAL PRIMARY KEY,
    run_date DATE NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    output_ref TEXT NOT NULL,
    summary JSONB NOT NULL DEFAULT '{}'::jsonb
);
