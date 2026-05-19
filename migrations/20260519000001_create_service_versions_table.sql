-- +goose Up
CREATE TABLE IF NOT EXISTS service_versions (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL,
    version VARCHAR(100) NOT NULL,
    git_ref VARCHAR(255),
    changelog TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_versions_service ON service_versions(service_id);
CREATE INDEX IF NOT EXISTS idx_service_versions_status ON service_versions(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_versions_unique ON service_versions(service_id, version);

-- +goose Down
DROP INDEX IF EXISTS idx_service_versions_unique;
DROP INDEX IF EXISTS idx_service_versions_status;
DROP INDEX IF EXISTS idx_service_versions_service;
DROP TABLE IF EXISTS service_versions;
