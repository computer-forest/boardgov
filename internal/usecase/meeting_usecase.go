package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/boardgov/internal/domain"
)

type MeetingUsecase interface {
	Create(ctx context.Context, m *domain.Meeting) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Meeting, error)
	List(ctx context.Context, filter domain.MeetingFilter) ([]*domain.Meeting, int, error)
	Update(ctx context.Context, m *domain.Meeting) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type meetingUsecase struct {
	repo      domain.MeetingRepository
	auditRepo domain.AuditLogRepository
}

func NewMeetingUsecase(repo domain.MeetingRepository, auditRepo domain.AuditLogRepository) MeetingUsecase {
	return &meetingUsecase{repo: repo, auditRepo: auditRepo}
}

func (u *meetingUsecase) Create(ctx context.Context, m *domain.Meeting) error {
	m.ID = uuid.New()
	m.Status = domain.MeetingScheduled
	return u.repo.Create(ctx, m)
}

func (u *meetingUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *meetingUsecase) List(ctx context.Context, filter domain.MeetingFilter) ([]*domain.Meeting, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	return u.repo.List(ctx, filter)
}

func (u *meetingUsecase) Update(ctx context.Context, m *domain.Meeting) error {
	existing, err := u.repo.GetByID(ctx, m.ID)
	if err != nil {
		return err
	}
	existing.Title = m.Title
	existing.MeetingType = m.MeetingType
	existing.ScheduledAt = m.ScheduledAt
	existing.Location = m.Location
	existing.Status = m.Status
	return u.repo.Update(ctx, existing)
}

func (u *meetingUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.SoftDelete(ctx, id)
}
