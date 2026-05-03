-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS pod_statuses (
    id VARCHAR(36) PRIMARY KEY,
    environment_id VARCHAR(36) NOT NULL,
    namespace VARCHAR(63) NOT NULL,

    -- Pod identification
    name VARCHAR(255) NOT NULL,
    owner_name VARCHAR(255),
    owner_kind VARCHAR(32),

    -- Status
    phase VARCHAR(32) NOT NULL,
    pod_ip VARCHAR(64),
    node_name VARCHAR(255),

    -- Conditions
    ready BOOLEAN DEFAULT FALSE,
    initialized BOOLEAN DEFAULT FALSE,
    containers_ready BOOLEAN DEFAULT FALSE,
    scheduled BOOLEAN DEFAULT FALSE,

    -- Container info
    container_count INTEGER DEFAULT 0,
    init_container_count INTEGER DEFAULT 0,

    -- Restart count
    restart_count INTEGER DEFAULT 0,

    -- Resource requests/limits
    cpu_request VARCHAR(32),
    cpu_limit VARCHAR(32),
    memory_request VARCHAR(32),
    memory_limit VARCHAR(32),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_pod_statuses_environment_id ON pod_statuses(environment_id);
CREATE INDEX idx_pod_statuses_namespace ON pod_statuses(namespace);
CREATE INDEX idx_pod_statuses_deleted_at ON pod_statuses(deleted_at);
CREATE INDEX idx_pod_statuses_phase ON pod_statuses(phase);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pod_statuses;

-- +goose StatementEnd
