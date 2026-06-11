-- +goose Up
-- +goose StatementBegin

ALTER TABLE delivery_targets
    ADD COLUMN IF NOT EXISTS control_plane_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS control_plane_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS kubeconfig_context VARCHAR(255),
    ADD COLUMN IF NOT EXISTS argocd_namespace VARCHAR(255),
    ADD COLUMN IF NOT EXISTS argocd_server VARCHAR(512),
    ADD COLUMN IF NOT EXISTS credential_reference VARCHAR(255);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE delivery_targets
    DROP COLUMN IF EXISTS credential_reference,
    DROP COLUMN IF EXISTS argocd_server,
    DROP COLUMN IF EXISTS argocd_namespace,
    DROP COLUMN IF EXISTS kubeconfig_context,
    DROP COLUMN IF EXISTS control_plane_type,
    DROP COLUMN IF EXISTS control_plane_name;

-- +goose StatementEnd
