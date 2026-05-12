-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS api_keys (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    team_id VARCHAR(36),
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(256) NOT NULL,
    key_prefix VARCHAR(10) NOT NULL,  -- First 8 chars for identification

    -- Permissions
    scopes JSONB,  -- Array of scopes/permissions

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, revoked, expired

    -- Expiry
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Usage tracking
    last_used_at TIMESTAMP WITH TIME ZONE,
    usage_count INTEGER NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT fk_api_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_api_keys_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_team_id ON api_keys(team_id);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_status ON api_keys(status);
CREATE INDEX idx_api_keys_deleted_at ON api_keys(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS api_keys;

-- +goose StatementEnd
