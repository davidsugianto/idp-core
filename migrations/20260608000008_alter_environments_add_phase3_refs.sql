-- +goose Up
-- +goose StatementBegin

ALTER TABLE environments
    ADD COLUMN IF NOT EXISTS delivery_target_id VARCHAR(36),
    ADD COLUMN IF NOT EXISTS template_instance_id VARCHAR(36);

CREATE INDEX IF NOT EXISTS idx_environments_delivery_target_id ON environments(delivery_target_id);
CREATE INDEX IF NOT EXISTS idx_environments_template_instance_id ON environments(template_instance_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_environments_delivery_target'
    ) THEN
        ALTER TABLE environments
            ADD CONSTRAINT fk_environments_delivery_target
            FOREIGN KEY (delivery_target_id) REFERENCES delivery_targets(id) ON DELETE SET NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_environments_template_instance'
    ) THEN
        ALTER TABLE environments
            ADD CONSTRAINT fk_environments_template_instance
            FOREIGN KEY (template_instance_id) REFERENCES template_instances(id) ON DELETE SET NULL;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE environments DROP CONSTRAINT IF EXISTS fk_environments_delivery_target;
ALTER TABLE environments DROP CONSTRAINT IF EXISTS fk_environments_template_instance;
DROP INDEX IF EXISTS idx_environments_delivery_target_id;
DROP INDEX IF EXISTS idx_environments_template_instance_id;
ALTER TABLE environments DROP COLUMN IF EXISTS delivery_target_id;
ALTER TABLE environments DROP COLUMN IF EXISTS template_instance_id;

-- +goose StatementEnd
