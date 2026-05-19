-- +goose Up
CREATE TABLE IF NOT EXISTS service_endpoints (
    id VARCHAR(36) PRIMARY KEY,
    service_version_id VARCHAR(36) NOT NULL,
    url VARCHAR(2048) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'http',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_endpoints_version ON service_endpoints(service_version_id);
CREATE INDEX IF NOT EXISTS idx_service_endpoints_type ON service_endpoints(type);
CREATE INDEX IF NOT EXISTS idx_service_endpoints_status ON service_endpoints(status);

-- +goose Down
DROP INDEX IF EXISTS idx_service_endpoints_status;
DROP INDEX IF EXISTS idx_service_endpoints_type;
DROP INDEX IF EXISTS idx_service_endpoints_version;
DROP TABLE IF EXISTS service_endpoints;
