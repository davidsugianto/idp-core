-- +goose Up
CREATE TABLE IF NOT EXISTS service_dependencies (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL,
    depends_on_service_id VARCHAR(36) NOT NULL,
    dependency_type VARCHAR(20) NOT NULL DEFAULT 'runtime',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_service_dependencies_service ON service_dependencies(service_id);
CREATE INDEX IF NOT EXISTS idx_service_dependencies_depends_on ON service_dependencies(depends_on_service_id);
CREATE INDEX IF NOT EXISTS idx_service_dependencies_deleted_at ON service_dependencies(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_dependencies_unique ON service_dependencies(service_id, depends_on_service_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_service_dependencies_unique;
DROP INDEX IF EXISTS idx_service_dependencies_deleted_at;
DROP INDEX IF EXISTS idx_service_dependencies_depends_on;
DROP INDEX IF EXISTS idx_service_dependencies_service;
DROP TABLE IF EXISTS service_dependencies;
