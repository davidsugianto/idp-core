-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS budgets (
    id VARCHAR(36) PRIMARY KEY,
    team_id VARCHAR(36) NOT NULL,
    environment_id VARCHAR(36),
    name VARCHAR(255) NOT NULL,
    "limit" NUMERIC(12,4) NOT NULL,
    period VARCHAR(20) NOT NULL DEFAULT 'monthly',
    alert_thresholds TEXT NOT NULL DEFAULT '80,90,100',
    alert_channels TEXT NOT NULL DEFAULT '["slack"]',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_budgets_team_id ON budgets(team_id);
CREATE INDEX IF NOT EXISTS idx_budgets_status ON budgets(status);
CREATE INDEX IF NOT EXISTS idx_budgets_deleted_at ON budgets(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS budgets;

-- +goose StatementEnd