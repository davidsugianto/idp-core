-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS teams (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(63) NOT NULL UNIQUE,
    description TEXT,

    -- Settings
    settings TEXT, -- JSON encoded settings

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, disabled

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_teams_name ON teams(name);
CREATE INDEX idx_teams_slug ON teams(slug);
CREATE INDEX idx_teams_status ON teams(status);
CREATE INDEX idx_teams_deleted_at ON teams(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS teams;

-- +goose StatementEnd
