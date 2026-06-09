-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS template_parameters (
    id VARCHAR(36) PRIMARY KEY,
    template_id VARCHAR(36) NOT NULL,
    version_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    "default" TEXT,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    validation TEXT,
    "order" INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_template_parameters_template FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE,
    CONSTRAINT fk_template_parameters_version FOREIGN KEY (version_id) REFERENCES template_versions(id) ON DELETE CASCADE,
    CONSTRAINT unique_template_parameter_name UNIQUE (version_id, name)
);

CREATE INDEX IF NOT EXISTS idx_template_parameters_template_id ON template_parameters(template_id);
CREATE INDEX IF NOT EXISTS idx_template_parameters_version_id ON template_parameters(version_id);
CREATE INDEX IF NOT EXISTS idx_template_parameters_order ON template_parameters(version_id, "order");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS template_parameters;

-- +goose StatementEnd
