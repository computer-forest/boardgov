package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrDocumentLocked     = errors.New("document is locked")
	ErrFileSizeExceeded   = errors.New("file size exceeded")
	ErrFileTypeNotAllowed = errors.New("file type not allowed")
	ErrInvalidTransition  = errors.New("invalid state transition")
)

// ─── User ────────────────────────────────────────────────────────────────────

type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	FullName     string     `json:"full_name"`
	PasswordHash string     `json:"-"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

// ─── Role / Permission ────────────────────────────────────────────────────────

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
}

// ─── Meeting ──────────────────────────────────────────────────────────────────

type MeetingType string

const (
	MeetingTypeBoard     MeetingType = "BOARD"
	MeetingTypeCommittee MeetingType = "COMMITTEE"
	MeetingTypeAGM       MeetingType = "AGM"
	MeetingTypeSpecial   MeetingType = "SPECIAL"
)

type MeetingStatus string

const (
	MeetingScheduled  MeetingStatus = "scheduled"
	MeetingInProgress MeetingStatus = "in_progress"
	MeetingCompleted  MeetingStatus = "completed"
	MeetingCancelled  MeetingStatus = "cancelled"
)

type Meeting struct {
	ID          uuid.UUID     `json:"id"`
	Title       string        `json:"title"`
	MeetingType string        `json:"meeting_type"`
	ScheduledAt time.Time     `json:"scheduled_at"`
	Location    string        `json:"location"`
	Agenda      string        `json:"agenda,omitempty"`
	Status      MeetingStatus `json:"status"`
	CreatedBy   uuid.UUID     `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   *time.Time    `json:"-"`
}

// ─── Document ─────────────────────────────────────────────────────────────────

type DocumentStatus string

const (
	DocDraft       DocumentStatus = "draft"
	DocSubmitted   DocumentStatus = "submitted"
	DocUnderReview DocumentStatus = "under_review"
	DocApproved    DocumentStatus = "approved"
	DocRejected    DocumentStatus = "rejected"
	DocArchived    DocumentStatus = "archived"
)

type WorkflowMode string

const (
	WorkflowSequential WorkflowMode = "sequential"
	WorkflowParallel   WorkflowMode = "parallel"
)

type Document struct {
	ID           uuid.UUID      `json:"id"`
	MeetingID    uuid.UUID      `json:"meeting_id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Status       DocumentStatus `json:"status"`
	WorkflowMode WorkflowMode   `json:"workflow_mode"`
	SubmittedAt  *time.Time     `json:"submitted_at,omitempty"`
	LockedAt     *time.Time     `json:"locked_at,omitempty"`
	CreatedBy    uuid.UUID      `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    *time.Time     `json:"-"`
}

// ─── DocumentVersion ──────────────────────────────────────────────────────────

type DocumentVersion struct {
	ID            uuid.UUID `json:"id"`
	DocumentID    uuid.UUID `json:"document_id"`
	VersionNumber int       `json:"version_number"`
	FileName      string    `json:"file_name"`
	StoragePath   string    `json:"-"`
	FileSize      int64     `json:"file_size"`
	MimeType      string    `json:"mime_type"`
	SHA256Hash    string    `json:"sha256_hash"`
	UploadedBy    uuid.UUID `json:"uploaded_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// ─── Reviewer / Approver ─────────────────────────────────────────────────────

type ReviewerStatus string

const (
	ReviewPending  ReviewerStatus = "pending"
	ReviewApproved ReviewerStatus = "approved"
	ReviewRejected ReviewerStatus = "rejected"
)

type DocumentReviewer struct {
	ID         uuid.UUID      `json:"id"`
	DocumentID uuid.UUID      `json:"document_id"`
	UserID     uuid.UUID      `json:"user_id"`
	OrderIndex int            `json:"order_index"`
	Status     ReviewerStatus `json:"status"`
	Comments   string         `json:"comments"`
	ReviewedAt *time.Time     `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ApproverStatus string

const (
	ApproverPending  ApproverStatus = "pending"
	ApproverApproved ApproverStatus = "approved"
	ApproverRejected ApproverStatus = "rejected"
)

type DocumentApprover struct {
	ID         uuid.UUID      `json:"id"`
	DocumentID uuid.UUID      `json:"document_id"`
	UserID     uuid.UUID      `json:"user_id"`
	OrderIndex int            `json:"order_index"`
	Status     ApproverStatus `json:"status"`
	Comments   string         `json:"comments"`
	ApprovedAt *time.Time     `json:"approved_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ─── Approval ─────────────────────────────────────────────────────────────────

type Approval struct {
	ID         uuid.UUID `json:"id"`
	DocumentID uuid.UUID `json:"document_id"`
	UserID     uuid.UUID `json:"user_id"`
	Role       string    `json:"role"`
	Action     string    `json:"action"`
	Comments   string    `json:"comments"`
	CreatedAt  time.Time `json:"created_at"`
}

// ─── WorkflowConfig ───────────────────────────────────────────────────────────

type WorkflowConfig struct {
	ID              uuid.UUID    `json:"id"`
	MeetingType     string       `json:"meeting_type"`
	WorkflowMode    WorkflowMode `json:"workflow_mode"`
	ReviewRequired  bool         `json:"review_required"`
	ApproveRequired bool         `json:"approve_required"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// ─── AuditLog ─────────────────────────────────────────────────────────────────

type AuditLog struct {
	ID            int64                  `json:"id"`
	UserID        *uuid.UUID             `json:"user_id,omitempty"`
	Action        string                 `json:"action"`
	MeetingID     *uuid.UUID             `json:"meeting_id,omitempty"`
	DocumentID    *uuid.UUID             `json:"document_id,omitempty"`
	PreviousState string                 `json:"previous_state,omitempty"`
	NewState      string                 `json:"new_state,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Hash          string                 `json:"hash"`
	PreviousHash  string                 `json:"previous_hash,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	User      *User      `json:"user,omitempty"`
}
