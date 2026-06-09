-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS delivery_targets (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    purpose VARCHAR(50),
    team_id VARCHAR(36),
    cluster_name VARCHAR(255) NOT NULL,
    cluster_server VARCHAR(512),
    availability_state VARCHAR(20) NOT NULL DEFAULT 'available',
    health_state VARCHAR(20) NOT NULL DEFAULT 'unknown',
    capacity_summary TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_delivery_targets_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL,
    CONSTRAINT unique_delivery_targets_slug UNIQUE (slug),
    CONSTRAINT chk_delivery_targets_availability CHECK (availability_state IN ('available', 'maintenance', 'disabled')),
    CONSTRAINT chk_delivery_targets_health CHECK (health_state IN ('healthy', 'degraded', 'unhealthy', 'unknown'))
);

CREATE INDEX IF NOT EXISTS idx_delivery_targets_name ON delivery_targets(name);
CREATE INDEX IF NOT EXISTS idx_delivery_targets_team_id ON delivery_targets(team_id);
CREATE INDEX IF NOT EXISTS idx_delivery_targets_purpose ON delivery_targets(purpose);
CREATE INDEX IF NOT EXISTS idx_delivery_targets_availability ON delivery_targets(availability_state);
CREATE INDEX IF NOT EXISTS idx_delivery_targets_health ON delivery_targets(health_state);
CREATE INDEX IF NOT EXISTS idx_delivery_targets_deleted_at ON delivery_targets(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS delivery_targets;

-- +goose StatementEnd
