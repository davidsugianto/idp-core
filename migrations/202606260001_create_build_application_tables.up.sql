CREATE TABLE IF NOT EXISTS build_applications (
    id VARCHAR(36) PRIMARY KEY,
    team_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(32) NOT NULL,
    repository_url VARCHAR(512) NOT NULL,
    repository_provider VARCHAR(64),
    default_branch VARCHAR(128),
    application_descriptor_path VARCHAR(512) NOT NULL,
    runtime_family VARCHAR(32),
    runtime_detection_mode VARCHAR(32),
    builder_profile_id VARCHAR(36),
    registry_target_id VARCHAR(36),
    deployment_automation_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    gitops_target_id VARCHAR(36),
    created_by VARCHAR(36),
    updated_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(team_id, name)
);

CREATE INDEX IF NOT EXISTS idx_build_applications_team ON build_applications(team_id);
CREATE INDEX IF NOT EXISTS idx_build_applications_status ON build_applications(status);

CREATE TABLE IF NOT EXISTS builds (
    id VARCHAR(36) PRIMARY KEY,
    application_id VARCHAR(36) NOT NULL,
    team_id VARCHAR(36) NOT NULL,
    sequence_number BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL,
    trigger_type VARCHAR(32) NOT NULL,
    triggered_by VARCHAR(36),
    idempotency_key VARCHAR(255),
    source_revision_requested VARCHAR(255),
    source_revision_resolved VARCHAR(255),
    retry_of_build_id VARCHAR(36),
    cancel_requested_by VARCHAR(36),
    failure_reason TEXT,
    queued_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_builds_application FOREIGN KEY (application_id) REFERENCES build_applications(id),
    UNIQUE(application_id, sequence_number)
);

CREATE INDEX IF NOT EXISTS idx_builds_team ON builds(team_id);
CREATE INDEX IF NOT EXISTS idx_builds_application ON builds(application_id);
CREATE INDEX IF NOT EXISTS idx_builds_status ON builds(status);
CREATE UNIQUE INDEX IF NOT EXISTS uq_builds_idempotency ON builds(application_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS build_artifacts (
    id VARCHAR(36) PRIMARY KEY,
    build_id VARCHAR(36) NOT NULL,
    application_id VARCHAR(36) NOT NULL,
    registry_target_id VARCHAR(36),
    image_repository VARCHAR(512),
    image_tag VARCHAR(255),
    image_digest VARCHAR(255),
    published_image_reference VARCHAR(1024),
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_artifacts_build FOREIGN KEY (build_id) REFERENCES builds(id),
    UNIQUE(build_id)
);

CREATE TABLE IF NOT EXISTS security_verifications (
    id VARCHAR(36) PRIMARY KEY,
    build_id VARCHAR(36) NOT NULL,
    artifact_id VARCHAR(36),
    status VARCHAR(32) NOT NULL,
    sbom_status VARCHAR(32),
    sbom_reference VARCHAR(1024),
    scan_status VARCHAR(32),
    scan_reference VARCHAR(1024),
    scan_summary TEXT,
    signing_status VARCHAR(32),
    signature_reference VARCHAR(1024),
    policy_gate_status VARCHAR(32),
    policy_gate_reason TEXT,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_security_build FOREIGN KEY (build_id) REFERENCES builds(id),
    UNIQUE(build_id)
);

CREATE TABLE IF NOT EXISTS deployment_updates (
    id VARCHAR(36) PRIMARY KEY,
    build_id VARCHAR(36) NOT NULL,
    application_id VARCHAR(36) NOT NULL,
    status VARCHAR(32) NOT NULL,
    gitops_target_id VARCHAR(36),
    requested_image_reference VARCHAR(1024),
    requested_manifest_path VARCHAR(1024),
    resulting_revision VARCHAR(255),
    failure_reason TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_deployment_build FOREIGN KEY (build_id) REFERENCES builds(id),
    UNIQUE(build_id)
);

CREATE TABLE IF NOT EXISTS lifecycle_events (
    id VARCHAR(36) PRIMARY KEY,
    team_id VARCHAR(36) NOT NULL,
    application_id VARCHAR(36),
    build_id VARCHAR(36),
    event_type VARCHAR(64) NOT NULL,
    event_source VARCHAR(64),
    payload_summary TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_lifecycle_application FOREIGN KEY (application_id) REFERENCES build_applications(id),
    CONSTRAINT fk_lifecycle_build FOREIGN KEY (build_id) REFERENCES builds(id)
);

CREATE INDEX IF NOT EXISTS idx_lifecycle_team ON lifecycle_events(team_id);
CREATE INDEX IF NOT EXISTS idx_lifecycle_application ON lifecycle_events(application_id);
CREATE INDEX IF NOT EXISTS idx_lifecycle_build ON lifecycle_events(build_id);
CREATE INDEX IF NOT EXISTS idx_lifecycle_occurred ON lifecycle_events(occurred_at DESC);

CREATE TABLE IF NOT EXISTS build_logs (
    id VARCHAR(36) PRIMARY KEY,
    build_id VARCHAR(36) NOT NULL,
    sequence BIGINT NOT NULL,
    line TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_build_logs_build FOREIGN KEY (build_id) REFERENCES builds(id),
    UNIQUE(build_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_build_logs_build ON build_logs(build_id);
