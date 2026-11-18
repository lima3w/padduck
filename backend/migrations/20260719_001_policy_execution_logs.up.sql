-- +migrate Up
CREATE TABLE policy_execution_logs (
    id              BIGSERIAL    PRIMARY KEY,
    organization_id BIGINT       REFERENCES organizations(id) ON DELETE SET NULL,
    policy_id       BIGINT       REFERENCES automation_policies(id) ON DELETE SET NULL,
    policy_name     TEXT         NOT NULL,
    resource_type   TEXT         NOT NULL,
    resource_id     BIGINT,
    resource_name   TEXT         NOT NULL DEFAULT '',
    workflow        TEXT         NOT NULL,
    action          TEXT         NOT NULL,
    decision        TEXT         NOT NULL CHECK (decision IN ('allow', 'deny', 'manual_review')),
    actions_taken   JSONB        NOT NULL DEFAULT '[]'::jsonb,
    triggered_by    TEXT         NOT NULL DEFAULT 'api',
    executed_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    rolled_back_at  TIMESTAMPTZ,
    rolled_back_by  BIGINT       REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_policy_execution_logs_policy ON policy_execution_logs(policy_id, executed_at DESC);
CREATE INDEX idx_policy_execution_logs_resource ON policy_execution_logs(resource_type, resource_id);
CREATE INDEX idx_policy_execution_logs_executed ON policy_execution_logs(executed_at DESC);
