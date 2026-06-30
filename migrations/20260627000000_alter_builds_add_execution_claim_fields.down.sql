DROP INDEX IF EXISTS idx_builds_execution_worker_id;
DROP INDEX IF EXISTS idx_builds_queue_claim;

ALTER TABLE builds
    DROP COLUMN IF EXISTS execution_attempts,
    DROP COLUMN IF EXISTS execution_lease_expires_at,
    DROP COLUMN IF EXISTS execution_claimed_at,
    DROP COLUMN IF EXISTS execution_worker_id;
