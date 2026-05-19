-- +goose Up
CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    team_id VARCHAR(36) NOT NULL,
    visibility VARCHAR(20) NOT NULL DEFAULT 'team',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);
CREATE INDEX IF NOT EXISTS idx_services_team ON services(team_id);
CREATE INDEX IF NOT EXISTS idx_services_visibility ON services(visibility);
CREATE INDEX IF NOT EXISTS idx_services_status ON services(status);
CREATE INDEX IF NOT EXISTS idx_services_deleted_at ON services(deleted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_services_deleted_at;
DROP INDEX IF EXISTS idx_services_status;
DROP INDEX IF EXISTS idx_services_visibility;
DROP INDEX IF EXISTS idx_services_team;
DROP INDEX IF EXISTS idx_services_name;
DROP TABLE IF EXISTS services;
