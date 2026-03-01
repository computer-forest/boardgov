package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/boardgov/internal/domain"
)

type AuditIntegrityResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

type DashboardUsecase interface {
	PendingApprovals(ctx context.Context, userID uuid.UUID) ([]*domain.PendingItem, error)
	PendingReviews(ctx context.Context, userID uuid.UUID) ([]*domain.PendingItem, error)
	OverdueDocuments(ctx context.Context) ([]*domain.OverdueDocument, error)
	MeetingStatusSummary(ctx context.Context) ([]*domain.MeetingStatusSummary, error)
	DocumentAgingAnalysis(ctx context.Context) ([]*domain.AgingBucket, error)
	ApprovalTurnaround(ctx context.Context) (*domain.TurnaroundMetrics, error)
	RejectionRatios(ctx context.Context) (*domain.RejectionAnalytics, error)
	ActivityTimeline(ctx context.Context, page, pageSize int) ([]*domain.AuditLog, int64, error)
	AuditIntegrityStatus(ctx context.Context) (*AuditIntegrityResult, error)
}

type dashboardUsecase struct {
	dashRepo  domain.DashboardRepository
	auditRepo domain.AuditLogRepository
}

func NewDashboardUsecase(dashRepo domain.DashboardRepository, auditRepo domain.AuditLogRepository) DashboardUsecase {
	return &dashboardUsecase{dashRepo: dashRepo, auditRepo: auditRepo}
}

func (u *dashboardUsecase) PendingApprovals(ctx context.Context, userID uuid.UUID) ([]*domain.PendingItem, error) {
	items, _, err := u.dashRepo.GetPendingApprovals(ctx, userID, 0, 100)
	return items, err
}

func (u *dashboardUsecase) PendingReviews(ctx context.Context, userID uuid.UUID) ([]*domain.PendingItem, error) {
	items, _, err := u.dashRepo.GetPendingReviews(ctx, userID, 0, 100)
	return items, err
}

func (u *dashboardUsecase) OverdueDocuments(ctx context.Context) ([]*domain.OverdueDocument, error) {
	items, _, err := u.dashRepo.GetOverdueDocuments(ctx, 0, 100)
	return items, err
}

func (u *dashboardUsecase) MeetingStatusSummary(ctx context.Context) ([]*domain.MeetingStatusSummary, error) {
	return u.dashRepo.GetMeetingStatusSummary(ctx)
}

func (u *dashboardUsecase) DocumentAgingAnalysis(ctx context.Context) ([]*domain.AgingBucket, error) {
	return u.dashRepo.GetDocumentAgingAnalysis(ctx)
}

func (u *dashboardUsecase) ApprovalTurnaround(ctx context.Context) (*domain.TurnaroundMetrics, error) {
	return u.dashRepo.GetApprovalTurnaroundMetrics(ctx)
}

func (u *dashboardUsecase) RejectionRatios(ctx context.Context) (*domain.RejectionAnalytics, error) {
	return u.dashRepo.GetRejectionRatioAnalytics(ctx)
}

func (u *dashboardUsecase) ActivityTimeline(ctx context.Context, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize
	return u.dashRepo.GetActivityTimeline(ctx, nil, offset, pageSize)
}

func (u *dashboardUsecase) AuditIntegrityStatus(ctx context.Context) (*AuditIntegrityResult, error) {
	valid, err := u.auditRepo.ValidateChain(ctx, 1, 9_999_999_999)
	if err != nil {
		return &AuditIntegrityResult{Valid: false, Errors: []string{err.Error()}}, nil
	}
	return &AuditIntegrityResult{Valid: valid}, nil
}
