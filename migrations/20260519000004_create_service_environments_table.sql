-- +goose Up
CREATE TABLE IF NOT EXISTS service_environments (
    id VARCHAR(36) PRIMARY KEY,
    service_version_id VARCHAR(36) NOT NULL,
    environment_id VARCHAR(36) NOT NULL,
    deployed_by VARCHAR(36),
    status VARCHAR(20) NOT NULL DEFAULT 'deployed',
    deployment_metadata JSONB,
    deployed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_service_environments_version ON service_environments(service_version_id);
CREATE INDEX IF NOT EXISTS idx_service_environments_environment ON service_environments(environment_id);
CREATE INDEX IF NOT EXISTS idx_service_environments_status ON service_environments(status);
CREATE INDEX IF NOT EXISTS idx_service_environments_deleted_at ON service_environments(deleted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_service_environments_deleted_at;
DROP INDEX IF EXISTS idx_service_environments_status;
DROP INDEX IF EXISTS idx_service_environments_environment;
DROP INDEX IF EXISTS idx_service_environments_version;
DROP TABLE IF EXISTS service_environments;
