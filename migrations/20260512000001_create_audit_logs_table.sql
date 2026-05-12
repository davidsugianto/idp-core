-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,

    -- Actor (who performed the action)
    user_id VARCHAR(36),
    user_email VARCHAR(255),
    actor_type VARCHAR(20) NOT NULL,  -- user, api_key, system

    -- Action details
    action VARCHAR(100) NOT NULL,     -- create, update, delete, login, etc.
    resource_type VARCHAR(50) NOT NULL,  -- environment, team, user, api_key, etc.
    resource_id VARCHAR(36),

    -- Context
    team_id VARCHAR(36),
    environment_id VARCHAR(36),

    -- Request details
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_method VARCHAR(10),
    request_path TEXT,
    request_id VARCHAR(36),

    -- Changes (optional, for tracking before/after)
    old_values JSONB,
    new_values JSONB,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'success',  -- success, failure, denied
    error_message TEXT,

    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_audit_logs_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL
);

-- Indexes for common queries
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_team_id ON audit_logs(team_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit_logs;

-- +goose StatementEnd
