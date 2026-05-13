CREATE TABLE IF NOT EXISTS cost_records (
    id VARCHAR(36) PRIMARY KEY,
    team_id VARCHAR(36),
    environment_id VARCHAR(36),
    namespace VARCHAR(255) NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    cpu_cost NUMERIC(12,4) NOT NULL DEFAULT 0,
    ram_cost NUMERIC(12,4) NOT NULL DEFAULT 0,
    pv_cost NUMERIC(12,4) NOT NULL DEFAULT 0,
    network_cost NUMERIC(12,4) NOT NULL DEFAULT 0,
    total_cost NUMERIC(12,4) NOT NULL DEFAULT 0,
    raw_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cost_records_team_period ON cost_records (team_id, period_start);
CREATE INDEX IF NOT EXISTS idx_cost_records_namespace_period ON cost_records (namespace, period_start);
CREATE INDEX IF NOT EXISTS idx_cost_records_period_start ON cost_records (period_start);