-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    team_id VARCHAR(36),
    environment_id VARCHAR(36),
    kind VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    payload TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_notifications_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL,
    CONSTRAINT fk_notifications_environment FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE SET NULL,
    CONSTRAINT chk_notifications_kind CHECK (kind IN ('environment', 'movement', 'target', 'template')),
    CONSTRAINT chk_notifications_severity CHECK (severity IN ('info', 'warning', 'error', 'success'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_team_id ON notifications(team_id);
CREATE INDEX IF NOT EXISTS idx_notifications_environment_id ON notifications(environment_id);
CREATE INDEX IF NOT EXISTS idx_notifications_kind ON notifications(kind);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS notifications;

-- +goose StatementEnd
