CREATE TABLE IF NOT EXISTS privacy_policy_versions (
    id BIGSERIAL PRIMARY KEY,
    version TEXT NOT NULL UNIQUE,
    effective_date DATE NOT NULL,
    summary TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO privacy_policy_versions (version, effective_date, summary)
VALUES ('1.0', CURRENT_DATE, 'Initial privacy policy')
ON CONFLICT (version) DO NOTHING;
