-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS template_resources (
    id VARCHAR(36) PRIMARY KEY,
    template_id VARCHAR(36) NOT NULL,
    version_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_template_resources_template FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE,
    CONSTRAINT fk_template_resources_version FOREIGN KEY (version_id) REFERENCES template_versions(id) ON DELETE CASCADE,
    CONSTRAINT unique_template_resource_name UNIQUE (version_id, name)
);

CREATE INDEX IF NOT EXISTS idx_template_resources_template_id ON template_resources(template_id);
CREATE INDEX IF NOT EXISTS idx_template_resources_version_id ON template_resources(version_id);
CREATE INDEX IF NOT EXISTS idx_template_resources_order ON template_resources(version_id, "order");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS template_resources;

-- +goose StatementEnd
