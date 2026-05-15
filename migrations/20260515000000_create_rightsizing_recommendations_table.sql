-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS rightsizing_recommendations (
    id VARCHAR(36) PRIMARY KEY,
    namespace VARCHAR(255) NOT NULL,
    workload_name VARCHAR(255) NOT NULL,
    workload_type VARCHAR(20) NOT NULL,
    container_name VARCHAR(255) NOT NULL,

    -- Current resource configuration
    current_cpu_request VARCHAR(50),
    current_cpu_limit VARCHAR(50),
    current_memory_request VARCHAR(50),
    current_memory_limit VARCHAR(50),

    -- Recommended resource configuration
    recommended_cpu_request VARCHAR(50),
    recommended_cpu_limit VARCHAR(50),
    recommended_memory_request VARCHAR(50),
    recommended_memory_limit VARCHAR(50),

    -- Usage metrics (for transparency)
    cpu_usage_avg VARCHAR(50),
    cpu_usage_max VARCHAR(50),
    memory_usage_avg VARCHAR(50),
    memory_usage_max VARCHAR(50),

    -- Recommendation details
    recommendation_type VARCHAR(20) NOT NULL,
    savings_potential NUMERIC(12,4) DEFAULT 0,
    confidence_score NUMERIC(5,2) DEFAULT 0,

    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    applied_at TIMESTAMP WITH TIME ZONE,
    applied_by VARCHAR(36),
    previous_state JSONB,

    -- Analysis period
    analysis_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    analysis_period_end TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Metadata
    team_id VARCHAR(36),
    environment_id VARCHAR(36),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rightsizing_namespace ON rightsizing_recommendations(namespace);
CREATE INDEX IF NOT EXISTS idx_rightsizing_workload ON rightsizing_recommendations(workload_name);
CREATE INDEX IF NOT EXISTS idx_rightsizing_status ON rightsizing_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_rightsizing_team ON rightsizing_recommendations(team_id);
CREATE INDEX IF NOT EXISTS idx_rightsizing_created ON rightsizing_recommendations(created_at);
CREATE INDEX IF NOT EXISTS idx_rightsizing_analysis_period ON rightsizing_recommendations(analysis_period_start, analysis_period_end);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS rightsizing_recommendations;

-- +goose StatementEnd
