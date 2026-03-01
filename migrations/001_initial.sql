-- ============================================================
-- Board Governance System - Database Migration
-- Version: 001_initial
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── ENUM TYPES ──────────────────────────────────────────────────────────────

CREATE TYPE meeting_type    AS ENUM ('BOARD', 'COMMITTEE', 'AGM', 'SPECIAL');
CREATE TYPE document_status AS ENUM ('DRAFT', 'SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'ARCHIVED');
CREATE TYPE workflow_mode   AS ENUM ('SEQUENTIAL', 'PARALLEL');
CREATE TYPE approval_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

-- ─── USERS ────────────────────────────────────────────────────────────────────

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    full_name     VARCHAR(255) NOT NULL,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

ALTER TABLE users ADD COLUMN last_login_at TIMESTAMPTZ;


CREATE INDEX idx_users_email      ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- ─── ROLES & PERMISSIONS ─────────────────────────────────────────────────────

CREATE TABLE roles (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE permissions (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    resource    VARCHAR(100) NOT NULL,
    action      VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (resource, action)
);

CREATE TABLE role_permissions (
    role_id       UUID        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID        NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    UUID        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_user_roles_user       ON user_roles(user_id);

-- ─── MEETINGS ─────────────────────────────────────────────────────────────────

CREATE TABLE meetings (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    title        VARCHAR(255) NOT NULL,
    meeting_type meeting_type NOT NULL,
    scheduled_at TIMESTAMPTZ  NOT NULL,
    location     VARCHAR(500),
    agenda       TEXT,
    created_by   UUID         NOT NULL REFERENCES users(id),
    is_locked    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_meetings_type       ON meetings(meeting_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_meetings_scheduled  ON meetings(scheduled_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_meetings_deleted_at ON meetings(deleted_at);

-- ─── WORKFLOW CONFIGS ─────────────────────────────────────────────────────────

CREATE TABLE workflow_configs (
    id            UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_type  meeting_type  NOT NULL,
    workflow_mode workflow_mode NOT NULL,
    review_days   INT           NOT NULL DEFAULT 3,
    approval_days INT           NOT NULL DEFAULT 5,
    is_default    BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_workflow_configs_default ON workflow_configs(meeting_type) WHERE is_default = TRUE;

-- ─── DOCUMENTS ────────────────────────────────────────────────────────────────

CREATE TABLE documents (
    id            UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id    UUID            NOT NULL REFERENCES meetings(id) ON DELETE RESTRICT,
    title         VARCHAR(255)    NOT NULL,
    description   TEXT,
    status        document_status NOT NULL DEFAULT 'DRAFT',
    workflow_mode workflow_mode   NOT NULL DEFAULT 'SEQUENTIAL',
    is_locked     BOOLEAN         NOT NULL DEFAULT FALSE,
    created_by    UUID            NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_documents_meeting_id ON documents(meeting_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_status     ON documents(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_created_by ON documents(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_updated_at ON documents(updated_at) WHERE deleted_at IS NULL;

-- ─── DOCUMENT VERSIONS ────────────────────────────────────────────────────────

CREATE TABLE document_versions (
    id             UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id    UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version_number INT         NOT NULL,
    file_path      TEXT        NOT NULL,
    file_name      VARCHAR(255) NOT NULL,
    file_size      BIGINT      NOT NULL,
    mime_type      VARCHAR(100) NOT NULL,
    sha256_hash    CHAR(64)    NOT NULL,
    uploaded_by    UUID        NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, version_number)
);

CREATE INDEX idx_doc_versions_document ON document_versions(document_id);
CREATE INDEX idx_doc_versions_number   ON document_versions(document_id, version_number DESC);

-- ─── DOCUMENT REVIEWERS (DYNAMIC) ────────────────────────────────────────────

CREATE TABLE document_reviewers (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id),
    order_index INT         NOT NULL DEFAULT 1,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    comments    TEXT,
    UNIQUE (document_id, user_id)
);

CREATE INDEX idx_doc_reviewers_document ON document_reviewers(document_id);
CREATE INDEX idx_doc_reviewers_user     ON document_reviewers(user_id);

-- ─── DOCUMENT APPROVERS (DYNAMIC) ────────────────────────────────────────────

CREATE TABLE document_approvers (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id),
    order_index INT         NOT NULL DEFAULT 1,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    UNIQUE (document_id, user_id)
);

CREATE INDEX idx_doc_approvers_document ON document_approvers(document_id);
CREATE INDEX idx_doc_approvers_user     ON document_approvers(user_id);

-- ─── APPROVALS ────────────────────────────────────────────────────────────────

CREATE TABLE approvals (
    id          UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID            NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    approver_id UUID            NOT NULL REFERENCES users(id),
    status      approval_status NOT NULL DEFAULT 'PENDING',
    comments    TEXT,
    actioned_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, approver_id)
);

CREATE INDEX idx_approvals_document  ON approvals(document_id);
CREATE INDEX idx_approvals_approver  ON approvals(approver_id);
CREATE INDEX idx_approvals_status    ON approvals(status);
CREATE INDEX idx_approvals_actioned  ON approvals(actioned_at) WHERE actioned_at IS NOT NULL;

-- ─── AUDIT LOGS (IMMUTABLE) ───────────────────────────────────────────────────

CREATE TABLE audit_logs (
    id             UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID            REFERENCES users(id),
    action         VARCHAR(100)    NOT NULL,
    meeting_id     UUID            REFERENCES meetings(id),
    document_id    UUID            REFERENCES documents(id),
    previous_state document_status,
    new_state      document_status,
    ip_address     INET,
    metadata       TEXT,
    hash           CHAR(64)        NOT NULL,
    previous_hash  CHAR(64)        NOT NULL DEFAULT '',
    timestamp      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Immutability enforced via rule - no UPDATE or DELETE
CREATE RULE no_update_audit AS ON UPDATE TO audit_logs DO INSTEAD NOTHING;
CREATE RULE no_delete_audit AS ON DELETE TO audit_logs DO INSTEAD NOTHING;

CREATE INDEX idx_audit_logs_user       ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_document   ON audit_logs(document_id);
CREATE INDEX idx_audit_logs_meeting    ON audit_logs(meeting_id);
CREATE INDEX idx_audit_logs_action     ON audit_logs(action);
CREATE INDEX idx_audit_logs_timestamp  ON audit_logs(timestamp DESC);


CREATE TABLE refresh_tokens (
    id           UUID PRIMARY KEY,           -- Or BIGSERIAL if using integers
    user_id      UUID NOT NULL,              -- References your users table
    token_hash   TEXT NOT NULL,              -- Or VARCHAR(255)
    expires_at   TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);


-- ─── SEED DATA ────────────────────────────────────────────────────────────────


-- Default roles
INSERT INTO roles (id, name, description) VALUES
    (uuid_generate_v4(), 'super_admin',    'Full system access'),
    (uuid_generate_v4(), 'board_secretary','Manages meetings and documents'),
    (uuid_generate_v4(), 'board_member',   'Reviews and approves documents'),
    (uuid_generate_v4(), 'reviewer',       'Reviews documents only'),
    (uuid_generate_v4(), 'viewer',         'Read-only access');

-- Default permissions
INSERT INTO permissions (id, name, description, resource, action) VALUES
    (uuid_generate_v4(), 'meetings:create',   'Create meetings',       'meetings',  'create'),
    (uuid_generate_v4(), 'meetings:read',     'Read meetings',         'meetings',  'read'),
    (uuid_generate_v4(), 'meetings:update',   'Update meetings',       'meetings',  'update'),
    (uuid_generate_v4(), 'meetings:delete',   'Delete meetings',       'meetings',  'delete'),
    (uuid_generate_v4(), 'documents:create',  'Create documents',      'documents', 'create'),
    (uuid_generate_v4(), 'documents:read',    'Read documents',        'documents', 'read'),
    (uuid_generate_v4(), 'documents:upload',  'Upload document files', 'documents', 'upload'),
    (uuid_generate_v4(), 'documents:assign',  'Assign reviewers/approvers', 'documents', 'assign'),
    (uuid_generate_v4(), 'documents:submit',  'Submit documents',      'documents', 'submit'),
    (uuid_generate_v4(), 'documents:review',  'Review documents',      'documents', 'review'),
    (uuid_generate_v4(), 'documents:approve', 'Approve documents',     'documents', 'approve'),
    (uuid_generate_v4(), 'documents:download','Download documents',    'documents', 'download'),
    (uuid_generate_v4(), 'documents:archive', 'Archive documents',     'documents', 'archive'),
    (uuid_generate_v4(), 'audit:read',        'Read audit logs',       'audit',     'read');

-- Assign all permissions to super_admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'super_admin';

-- Board secretary permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'board_secretary'
  AND p.name IN ('meetings:create','meetings:read','meetings:update',
                 'documents:create','documents:read','documents:upload',
                 'documents:assign','documents:submit','documents:download','documents:archive');

-- Board member permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'board_member'
  AND p.name IN ('meetings:read','documents:read','documents:approve','documents:download');

-- Reviewer permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'reviewer'
  AND p.name IN ('meetings:read','documents:read','documents:review','documents:download');

-- Viewer permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer'
  AND p.name IN ('meetings:read','documents:read','documents:download');

-- Default workflow configs
INSERT INTO workflow_configs (id, meeting_type, workflow_mode, review_days, approval_days, is_default) VALUES
    (uuid_generate_v4(), 'BOARD',     'SEQUENTIAL', 3, 5, TRUE),
    (uuid_generate_v4(), 'COMMITTEE', 'PARALLEL',   2, 3, TRUE),
    (uuid_generate_v4(), 'AGM',       'SEQUENTIAL', 5, 7, TRUE),
    (uuid_generate_v4(), 'SPECIAL',   'PARALLEL',   1, 2, TRUE);
