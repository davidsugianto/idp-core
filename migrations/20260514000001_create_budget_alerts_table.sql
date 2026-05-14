-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS budget_alerts (
    id VARCHAR(36) PRIMARY KEY,
    budget_id VARCHAR(36) NOT NULL,
    "timestamp" TIMESTAMP WITH TIME ZONE NOT NULL,
    threshold INTEGER NOT NULL,
    current_spend NUMERIC(12,4) NOT NULL DEFAULT 0,
    "limit" NUMERIC(12,4) NOT NULL DEFAULT 0,
    percentage NUMERIC(6,2) NOT NULL DEFAULT 0,
    sent_to TEXT NOT NULL DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'sent',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_budget_alerts_budget FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_budget_alerts_budget_id ON budget_alerts(budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_alerts_timestamp ON budget_alerts("timestamp");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS budget_alerts;

-- +goose StatementEnd