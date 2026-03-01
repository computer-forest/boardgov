-- ============================================================
-- Board Governance System - Migration 002
-- Adds meeting_status enum and status column to meetings table
-- ============================================================

CREATE TYPE meeting_status AS ENUM ('scheduled', 'in_progress', 'completed', 'cancelled');

ALTER TABLE meetings
    ADD COLUMN status meeting_status NOT NULL DEFAULT 'scheduled';

CREATE INDEX idx_meetings_status ON meetings(status) WHERE deleted_at IS NULL;
