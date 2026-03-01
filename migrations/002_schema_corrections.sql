-- ============================================================
-- Board Governance System - Schema Corrections
-- Version: 002_schema_corrections
-- Fixes mismatches between 001_initial.sql and Go repository layer
-- ============================================================

-- ─── 1. meetings: add missing status column ──────────────────────────────────

ALTER TABLE meetings ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'scheduled';

-- ─── 2. documents: convert ENUM columns to TEXT (Go uses lowercase values) ───
--    ENUM types defined in 001 are uppercase ('DRAFT', 'SEQUENTIAL', etc.)
--    Go constants are lowercase ("draft", "sequential", etc.)

ALTER TABLE documents ALTER COLUMN status       TYPE TEXT USING status::TEXT;
ALTER TABLE documents ALTER COLUMN workflow_mode TYPE TEXT USING workflow_mode::TEXT;

-- Lowercase any existing uppercase values from the old ENUMs
UPDATE documents SET status        = LOWER(status)        WHERE status        != LOWER(status);
UPDATE documents SET workflow_mode = LOWER(workflow_mode)  WHERE workflow_mode != LOWER(workflow_mode);

-- ─── 3. documents: add submitted_at / locked_at; drop is_locked ──────────────

ALTER TABLE documents ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS locked_at    TIMESTAMPTZ;

-- Populate locked_at from the old boolean flag so data is not lost
UPDATE documents SET locked_at = NOW() WHERE is_locked = TRUE AND locked_at IS NULL;

ALTER TABLE documents DROP COLUMN IF EXISTS is_locked;

-- ─── 4. document_versions: rename file_path → storage_path ──────────────────

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'document_versions' AND column_name = 'file_path'
    ) THEN
        ALTER TABLE document_versions RENAME COLUMN file_path TO storage_path;
    END IF;
END $$;

-- ─── 5. document_reviewers: add status and created_at; drop unused columns ───

ALTER TABLE document_reviewers ADD COLUMN IF NOT EXISTS status     VARCHAR(20)  NOT NULL DEFAULT 'pending';
ALTER TABLE document_reviewers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW();

-- Map old is_active=false to 'rejected' so existing data is preserved
UPDATE document_reviewers SET status = 'rejected' WHERE is_active = FALSE AND status = 'pending';

ALTER TABLE document_reviewers DROP COLUMN IF EXISTS is_active;
ALTER TABLE document_reviewers DROP COLUMN IF EXISTS assigned_at;

-- ─── 6. document_approvers: add status, comments, created_at ─────────────────

ALTER TABLE document_approvers ADD COLUMN IF NOT EXISTS status     VARCHAR(20)  NOT NULL DEFAULT 'pending';
ALTER TABLE document_approvers ADD COLUMN IF NOT EXISTS comments   TEXT;
ALTER TABLE document_approvers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW();

UPDATE document_approvers SET status = 'rejected' WHERE is_active = FALSE AND status = 'pending';

ALTER TABLE document_approvers DROP COLUMN IF EXISTS is_active;
ALTER TABLE document_approvers DROP COLUMN IF EXISTS assigned_at;

-- ─── 7. approvals: align with Go insert (user_id, role, action, comments) ────

-- Drop the unique constraint on (document_id, approver_id) — Go inserts
-- multiple rows per user (once as reviewer, once as approver).
ALTER TABLE approvals DROP CONSTRAINT IF EXISTS approvals_document_id_approver_id_key;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'approvals' AND column_name = 'approver_id'
    ) THEN
        ALTER TABLE approvals RENAME COLUMN approver_id TO user_id;
    END IF;
END $$;

ALTER TABLE approvals ADD COLUMN IF NOT EXISTS role   VARCHAR(20);
ALTER TABLE approvals ADD COLUMN IF NOT EXISTS action VARCHAR(20);

-- Drop columns that the Go layer no longer uses
ALTER TABLE approvals DROP COLUMN IF EXISTS actioned_at;
ALTER TABLE approvals DROP COLUMN IF EXISTS updated_at;
-- 'status' column from old schema is superseded by 'action'; drop it
-- (only safe if no business logic depends on the old status ENUM)
ALTER TABLE approvals DROP COLUMN IF EXISTS status;

-- ─── 8. audit_logs: id → BIGSERIAL, timestamp → created_at, states → TEXT ───

-- Rename the timestamp column first (simple rename, no data loss)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'audit_logs' AND column_name = 'timestamp'
    ) THEN
        ALTER TABLE audit_logs RENAME COLUMN "timestamp" TO created_at;
    END IF;
END $$;

-- Change previous_state / new_state from document_status ENUM to plain TEXT
-- so Go can store arbitrary transition strings (including empty string for
-- actions that have no meaningful before/after state).
ALTER TABLE audit_logs ALTER COLUMN previous_state TYPE TEXT USING previous_state::TEXT;
ALTER TABLE audit_logs ALTER COLUMN new_state      TYPE TEXT USING new_state::TEXT;

-- The Go AuditLog.ID field is int64 so we need a BIGSERIAL id.
-- Strategy: add a new bigserial sequence, backfill, swap primary key.
-- This is safe on an empty/append-only table. If the table has rows,
-- the sequence starts from max(existing sequential row number) + 1.

-- Add a bigint column driven by a sequence
CREATE SEQUENCE IF NOT EXISTS audit_logs_id_seq;

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS id_int BIGINT;

-- Backfill existing rows with sequential integers (row_number order by old UUID pk)
DO $$
DECLARE max_val BIGINT;
BEGIN
    UPDATE audit_logs al
    SET id_int = sub.rn
    FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY id) AS rn FROM audit_logs) sub
    WHERE al.id = sub.id;

    SELECT COALESCE(MAX(id_int), 0) INTO max_val FROM audit_logs;
    PERFORM setval('audit_logs_id_seq', GREATEST(max_val, 1));
END $$;

-- Set default for new rows
ALTER TABLE audit_logs ALTER COLUMN id_int SET DEFAULT nextval('audit_logs_id_seq');
ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id_int;

-- Swap primary key: drop old UUID pk, promote id_int
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_pkey;
ALTER TABLE audit_logs ALTER COLUMN id_int SET NOT NULL;
ALTER TABLE audit_logs ADD PRIMARY KEY (id_int);

-- Rename id → id_uuid (keep for any existing FK references), id_int → id
ALTER TABLE audit_logs RENAME COLUMN id     TO id_uuid;
ALTER TABLE audit_logs RENAME COLUMN id_int TO id;

-- ─── 9. Drop now-unused ENUM types (optional, non-breaking) ──────────────────
-- Cannot drop if still referenced by another column. Run after verifying.
-- DROP TYPE IF EXISTS document_status;
-- DROP TYPE IF EXISTS workflow_mode;
-- DROP TYPE IF EXISTS approval_status;
