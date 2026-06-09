-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS template_versions (
    id VARCHAR(36) PRIMARY KEY,
    template_id VARCHAR(36) NOT NULL,
    version VARCHAR(100) NOT NULL,
    description TEXT,
    changelog TEXT,
    is_latest BOOLEAN NOT NULL DEFAULT FALSE,
    is_stable BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_template_versions_template FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE,
    CONSTRAINT unique_template_version UNIQUE (template_id, version),
    CONSTRAINT chk_template_versions_status CHECK (status IN ('draft', 'stable', 'deprecated'))
);

CREATE INDEX IF NOT EXISTS idx_template_versions_template_id ON template_versions(template_id);
CREATE INDEX IF NOT EXISTS idx_template_versions_status ON template_versions(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_template_versions_latest_true ON template_versions(template_id) WHERE is_latest = TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS template_versions;

-- +goose StatementEnd
