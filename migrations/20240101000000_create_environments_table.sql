-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS environments (
    id VARCHAR(36) PRIMARY KEY,
    team_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    namespace VARCHAR(63) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'creating',

    -- GitOps configuration
    git_repo_url VARCHAR(512),
    git_revision VARCHAR(64) DEFAULT 'main',
    manifest_path VARCHAR(512),
    argo_app_name VARCHAR(63),

    -- Cluster information
    cluster_name VARCHAR(255),
    cluster_server VARCHAR(512),

    -- Resource quotas
    resource_quota_cpu VARCHAR(32),
    resource_quota_memory VARCHAR(32),

    -- Metadata
    labels TEXT,
    annotations TEXT,

    -- Ownership and lifecycle
    owner_id VARCHAR(36),
    expires_at TIMESTAMP WITH TIME ZONE,
    last_sync_at TIMESTAMP WITH TIME ZONE,

    -- Error tracking
    last_error TEXT,
    error_count INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_environments_team_id ON environments(team_id);
CREATE INDEX idx_environments_status ON environments(status);
CREATE INDEX idx_environments_expires_at ON environments(expires_at);
CREATE INDEX idx_environments_deleted_at ON environments(deleted_at);
CREATE INDEX idx_environments_namespace ON environments(namespace);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS environments;

-- +goose StatementEnd
