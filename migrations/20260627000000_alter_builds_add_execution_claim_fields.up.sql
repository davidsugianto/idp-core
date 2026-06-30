ALTER TABLE builds
    ADD COLUMN IF NOT EXISTS execution_worker_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS execution_claimed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS execution_lease_expires_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS execution_attempts INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_builds_queue_claim
    ON builds(status, queued_at, created_at);

CREATE INDEX IF NOT EXISTS idx_builds_execution_worker_id
    ON builds(execution_worker_id);
