-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,

    -- Scope: platform, team, or environment
    scope VARCHAR(20) NOT NULL DEFAULT 'team', -- platform, team, environment

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_scope ON roles(scope);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS roles;

-- +goose StatementEnd
