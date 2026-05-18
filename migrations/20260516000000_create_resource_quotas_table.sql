-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS resource_quotas (
    id VARCHAR(36) PRIMARY KEY,
    namespace VARCHAR(255) NOT NULL UNIQUE,
    team_id VARCHAR(36) NOT NULL,
    environment_id VARCHAR(36),

    -- CPU limits (in cores, e.g., "2" means 2 cores, "500m" means 0.5 cores)
    cpu_request_limit VARCHAR(50),
    cpu_limit_limit VARCHAR(50),

    -- Memory limits (in bytes, e.g., "2Gi" means 2 gibibytes)
    memory_request_limit VARCHAR(50),
    memory_limit_limit VARCHAR(50),

    -- Storage limits
    storage_request_limit VARCHAR(50),

    -- Pod count limit
    pod_count_limit INTEGER,

    -- Object count limits (optional)
    configmap_count_limit INTEGER,
    secret_count_limit INTEGER,
    pvc_count_limit INTEGER,

    -- Current usage (cached, updated periodically)
    current_cpu_request VARCHAR(50),
    current_cpu_limit VARCHAR(50),
    current_memory_request VARCHAR(50),
    current_memory_limit VARCHAR(50),
    current_storage_request VARCHAR(50),
    current_pod_count INTEGER,

    -- Enforcement settings
    enforce BOOLEAN NOT NULL DEFAULT true,
    grace_period_hours INTEGER DEFAULT 0,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    -- Metadata
    description TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resource_quotas_namespace ON resource_quotas(namespace);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_team ON resource_quotas(team_id);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_environment ON resource_quotas(environment_id);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_status ON resource_quotas(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS resource_quotas;

-- +goose StatementEnd
