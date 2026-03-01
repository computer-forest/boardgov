package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ─── User Repository ──────────────────────────────────────────────────────────

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int) ([]*User, int, error)
	ListPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// ─── Role Repository ──────────────────────────────────────────────────────────

type RoleRepository interface {
	Create(ctx context.Context, r *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

// ─── RefreshToken Repository ──────────────────────────────────────────────────

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}

// ─── Meeting Repository ───────────────────────────────────────────────────────

type MeetingRepository interface {
	Create(ctx context.Context, m *Meeting) error
	GetByID(ctx context.Context, id uuid.UUID) (*Meeting, error)
	Update(ctx context.Context, m *Meeting) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter MeetingFilter) ([]*Meeting, int, error)
}

// ─── Document Repository ──────────────────────────────────────────────────────

type DocumentRepository interface {
	Create(ctx context.Context, d *Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	Update(ctx context.Context, d *Document) error
	List(ctx context.Context, filter DocumentFilter) ([]*Document, int, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Version control
	CreateVersion(ctx context.Context, v *DocumentVersion) error
	GetLatestVersion(ctx context.Context, documentID uuid.UUID) (*DocumentVersion, error)
	GetVersionByNumber(ctx context.Context, documentID uuid.UUID, version int) (*DocumentVersion, error)
	ListVersions(ctx context.Context, documentID uuid.UUID) ([]*DocumentVersion, error)

	// Reviewers
	GetReviewers(ctx context.Context, documentID uuid.UUID) ([]*DocumentReviewer, error)
	SetReviewers(ctx context.Context, documentID uuid.UUID, reviewers []*DocumentReviewer) error
	UpdateReviewer(ctx context.Context, r *DocumentReviewer) error

	// Approvers
	GetApprovers(ctx context.Context, documentID uuid.UUID) ([]*DocumentApprover, error)
	SetApprovers(ctx context.Context, documentID uuid.UUID, approvers []*DocumentApprover) error
	UpdateApprover(ctx context.Context, a *DocumentApprover) error

	// Approvals
	CreateApproval(ctx context.Context, a *Approval) error
}

// ─── Audit Repository ─────────────────────────────────────────────────────────

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetLatest(ctx context.Context) (*AuditLog, error)
	List(ctx context.Context, filter AuditFilter) ([]*AuditLog, int, error)
	ValidateChain(ctx context.Context, from, to int64) (bool, error)
}

// ─── Dashboard Repository ─────────────────────────────────────────────────────

type DashboardRepository interface {
	GetPendingApprovals(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*PendingItem, int64, error)
	GetPendingReviews(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*PendingItem, int64, error)
	GetOverdueDocuments(ctx context.Context, offset, limit int) ([]*OverdueDocument, int64, error)
	GetMeetingStatusSummary(ctx context.Context) ([]*MeetingStatusSummary, error)
	GetDocumentAgingAnalysis(ctx context.Context) ([]*AgingBucket, error)
	GetApprovalTurnaroundMetrics(ctx context.Context) (*TurnaroundMetrics, error)
	GetRejectionRatioAnalytics(ctx context.Context) (*RejectionAnalytics, error)
	GetActivityTimeline(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*AuditLog, int64, error)
}

// ─── Filter Types ─────────────────────────────────────────────────────────────

type AuditFilter struct {
	UserID     *uuid.UUID
	DocumentID *uuid.UUID
	MeetingID  *uuid.UUID
	Action     string
	From       *time.Time
	To         *time.Time
	Page       int
	PageSize   int
}

type DocumentFilter struct {
	MeetingID *uuid.UUID
	Status    DocumentStatus
	CreatedBy *uuid.UUID
	Page      int
	PageSize  int
}

type MeetingFilter struct {
	Status   MeetingStatus
	FromDate *time.Time
	ToDate   *time.Time
	Page     int
	PageSize int
}

// ─── Dashboard DTOs ───────────────────────────────────────────────────────────

type PendingItem struct {
	DocumentID    uuid.UUID      `db:"document_id" json:"document_id"`
	DocumentTitle string         `db:"document_title" json:"document_title"`
	MeetingID     uuid.UUID      `db:"meeting_id" json:"meeting_id"`
	MeetingTitle  string         `db:"meeting_title" json:"meeting_title"`
	Status        DocumentStatus `db:"status" json:"status"`
	DueDate       *time.Time     `db:"due_date" json:"due_date,omitempty"`
	AssignedAt    time.Time      `db:"assigned_at" json:"assigned_at"`
}

type OverdueDocument struct {
	DocumentID    uuid.UUID      `db:"document_id" json:"document_id"`
	DocumentTitle string         `db:"document_title" json:"document_title"`
	MeetingTitle  string         `db:"meeting_title" json:"meeting_title"`
	Status        DocumentStatus `db:"status" json:"status"`
	DaysOverdue   int            `db:"days_overdue" json:"days_overdue"`
}

type MeetingStatusSummary struct {
	MeetingID    uuid.UUID   `db:"meeting_id" json:"meeting_id"`
	MeetingTitle string      `db:"meeting_title" json:"meeting_title"`
	MeetingType  MeetingType `db:"meeting_type" json:"meeting_type"`
	Total        int         `db:"total" json:"total"`
	Draft        int         `db:"draft" json:"draft"`
	Submitted    int         `db:"submitted" json:"submitted"`
	UnderReview  int         `db:"under_review" json:"under_review"`
	Approved     int         `db:"approved" json:"approved"`
	Rejected     int         `db:"rejected" json:"rejected"`
	Archived     int         `db:"archived" json:"archived"`
}

type AgingBucket struct {
	Bucket   string `db:"bucket" json:"bucket"`
	Count    int    `db:"count" json:"count"`
	Statuses string `db:"statuses" json:"statuses"`
}

type TurnaroundMetrics struct {
	AvgReviewHours   float64 `db:"avg_review_hours" json:"avg_review_hours"`
	AvgApprovalHours float64 `db:"avg_approval_hours" json:"avg_approval_hours"`
	P50ReviewHours   float64 `db:"p50_review_hours" json:"p50_review_hours"`
	P90ReviewHours   float64 `db:"p90_review_hours" json:"p90_review_hours"`
	P50ApprovalHours float64 `db:"p50_approval_hours" json:"p50_approval_hours"`
	P90ApprovalHours float64 `db:"p90_approval_hours" json:"p90_approval_hours"`
}

type RejectionAnalytics struct {
	TotalDocuments int     `db:"total_documents" json:"total_documents"`
	TotalRejected  int     `db:"total_rejected" json:"total_rejected"`
	RejectionRate  float64 `db:"rejection_rate" json:"rejection_rate"`
	ByMeetingType  []struct {
		MeetingType   MeetingType `json:"meeting_type"`
		RejectionRate float64     `json:"rejection_rate"`
	} `json:"by_meeting_type"`
}
