-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS workload_statuses (
    id VARCHAR(36) PRIMARY KEY,
    environment_id VARCHAR(36) NOT NULL,
    namespace VARCHAR(63) NOT NULL,

    -- Workload identification
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(32) NOT NULL,
    api_version VARCHAR(64),

    -- Replicas
    desired_replicas INTEGER DEFAULT 0,
    current_replicas INTEGER DEFAULT 0,
    ready_replicas INTEGER DEFAULT 0,
    updated_replicas INTEGER DEFAULT 0,
    available_replicas INTEGER DEFAULT 0,

    -- Status
    status VARCHAR(32),
    status_reason TEXT,

    -- Image info
    image VARCHAR(512),
    latest_image VARCHAR(512),

    -- Metadata
    labels TEXT,
    annotations TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_workload_statuses_environment_id ON workload_statuses(environment_id);
CREATE INDEX idx_workload_statuses_namespace ON workload_statuses(namespace);
CREATE INDEX idx_workload_statuses_deleted_at ON workload_statuses(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS workload_statuses;

-- +goose StatementEnd
