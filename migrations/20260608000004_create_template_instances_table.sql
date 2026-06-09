-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS template_instances (
    id VARCHAR(36) PRIMARY KEY,
    template_id VARCHAR(36) NOT NULL,
    version_id VARCHAR(36) NOT NULL,
    environment_id VARCHAR(36) NOT NULL,
    parameters TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_template_instances_template FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE RESTRICT,
    CONSTRAINT fk_template_instances_version FOREIGN KEY (version_id) REFERENCES template_versions(id) ON DELETE RESTRICT,
    CONSTRAINT fk_template_instances_environment FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE,
    CONSTRAINT unique_template_instance_environment UNIQUE (environment_id)
);

CREATE INDEX IF NOT EXISTS idx_template_instances_template_id ON template_instances(template_id);
CREATE INDEX IF NOT EXISTS idx_template_instances_version_id ON template_instances(version_id);
CREATE INDEX IF NOT EXISTS idx_template_instances_environment_id ON template_instances(environment_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS template_instances;

-- +goose StatementEnd
