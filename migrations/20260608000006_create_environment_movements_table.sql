-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS environment_movements (
    id VARCHAR(36) PRIMARY KEY,
    environment_id VARCHAR(36) NOT NULL,
    source_target_id VARCHAR(36),
    destination_target_id VARCHAR(36) NOT NULL,
    requested_by VARCHAR(36),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    progress_percent INTEGER NOT NULL DEFAULT 0,
    message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_environment_movements_environment FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE,
    CONSTRAINT fk_environment_movements_source_target FOREIGN KEY (source_target_id) REFERENCES delivery_targets(id) ON DELETE SET NULL,
    CONSTRAINT fk_environment_movements_destination_target FOREIGN KEY (destination_target_id) REFERENCES delivery_targets(id) ON DELETE RESTRICT,
    CONSTRAINT fk_environment_movements_requested_by FOREIGN KEY (requested_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_environment_movements_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    CONSTRAINT chk_environment_movements_progress CHECK (progress_percent >= 0 AND progress_percent <= 100)
);

CREATE INDEX IF NOT EXISTS idx_environment_movements_environment_id ON environment_movements(environment_id);
CREATE INDEX IF NOT EXISTS idx_environment_movements_destination_target_id ON environment_movements(destination_target_id);
CREATE INDEX IF NOT EXISTS idx_environment_movements_status ON environment_movements(status);
CREATE INDEX IF NOT EXISTS idx_environment_movements_created_at ON environment_movements(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS environment_movements;

-- +goose StatementEnd
