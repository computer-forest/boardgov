package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/boardgov/config"
	"github.com/yourorg/boardgov/internal/domain"
	"github.com/yourorg/boardgov/pkg/crypto"
	"github.com/yourorg/boardgov/pkg/logger"
	"github.com/yourorg/boardgov/pkg/validator"
)

type DocumentUsecase interface {
	Create(ctx context.Context, doc *domain.Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error)
	List(ctx context.Context, filter domain.DocumentFilter) ([]*domain.Document, int, error)
	UploadVersion(ctx context.Context, docID, uploaderID uuid.UUID, fileName, mimeType string, data []byte, ip string) (*domain.DocumentVersion, error)
	DownloadVersion(ctx context.Context, docID uuid.UUID, versionNum int, requesterID uuid.UUID, ip string) ([]byte, *domain.DocumentVersion, error)
	SetReviewers(ctx context.Context, docID uuid.UUID, userIDs []uuid.UUID) error
	SetApprovers(ctx context.Context, docID uuid.UUID, userIDs []uuid.UUID) error
	GetReviewers(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentReviewer, error)
	GetApprovers(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentApprover, error)
	Submit(ctx context.Context, docID, userID uuid.UUID, ip string) error
	StartReview(ctx context.Context, docID, userID uuid.UUID, ip string) error
	Review(ctx context.Context, docID, userID uuid.UUID, action, comments, ip string) error
	Approve(ctx context.Context, docID, userID uuid.UUID, action, comments, ip string) error
	Archive(ctx context.Context, docID, userID uuid.UUID, ip string) error
	ListVersions(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentVersion, error)
}

type documentUsecase struct {
	docRepo     domain.DocumentRepository
	meetingRepo domain.MeetingRepository
	auditRepo   domain.AuditLogRepository
	storageCfg  config.StorageConfig
}

func NewDocumentUsecase(
	docRepo domain.DocumentRepository,
	meetingRepo domain.MeetingRepository,
	auditRepo domain.AuditLogRepository,
	storageCfg config.StorageConfig,
) DocumentUsecase {
	return &documentUsecase{
		docRepo:     docRepo,
		meetingRepo: meetingRepo,
		auditRepo:   auditRepo,
		storageCfg:  storageCfg,
	}
}

func (u *documentUsecase) Create(ctx context.Context, doc *domain.Document) error {
	// Verify meeting exists
	if _, err := u.meetingRepo.GetByID(ctx, doc.MeetingID); err != nil {
		return fmt.Errorf("meeting: %w", err)
	}
	doc.ID = uuid.New()
	doc.Status = domain.DocDraft
	if doc.WorkflowMode == "" {
		doc.WorkflowMode = domain.WorkflowSequential
	}
	return u.docRepo.Create(ctx, doc)
}

func (u *documentUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	return u.docRepo.GetByID(ctx, id)
}

func (u *documentUsecase) List(ctx context.Context, filter domain.DocumentFilter) ([]*domain.Document, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	return u.docRepo.List(ctx, filter)
}

func (u *documentUsecase) UploadVersion(ctx context.Context, docID, uploaderID uuid.UUID, fileName, mimeType string, data []byte, ip string) (*domain.DocumentVersion, error) {
	doc, err := u.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	if doc.LockedAt != nil {
		return nil, domain.ErrDocumentLocked
	}

	// Validate size
	if int64(len(data)) > u.storageCfg.MaxFileSize {
		return nil, domain.ErrFileSizeExceeded
	}

	// Validate MIME
	if err := validator.ValidateFileMIME(data[:min(512, len(data))], mimeType); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrFileTypeNotAllowed, err)
	}

	// Determine version number
	latest, err := u.docRepo.GetLatestVersion(ctx, docID)
	versionNum := 1
	if err == nil {
		versionNum = latest.VersionNumber + 1
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// Build storage path
	storageDir := filepath.Join(u.storageCfg.BasePath, doc.MeetingID.String(), docID.String(), fmt.Sprintf("v%d", versionNum))
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}

	fileID := uuid.New().String()
	storagePath := filepath.Join(storageDir, fileID)
	if err := os.WriteFile(storagePath, data, 0600); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	version := &domain.DocumentVersion{
		ID:            uuid.New(),
		DocumentID:    docID,
		VersionNumber: versionNum,
		FileName:      validator.SanitizeFilename(fileName),
		StoragePath:   storagePath,
		FileSize:      int64(len(data)),
		MimeType:      mimeType,
		SHA256Hash:    crypto.SHA256File(data),
		UploadedBy:    uploaderID,
	}

	if err := u.docRepo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	u.audit(ctx, uploaderID, "document.upload", ip, &doc.MeetingID, &docID, "", "")
	logger.Log.Info().Str("document_id", docID.String()).Int("version", versionNum).Msg("file uploaded")
	return version, nil
}

func (u *documentUsecase) DownloadVersion(ctx context.Context, docID uuid.UUID, versionNum int, requesterID uuid.UUID, ip string) ([]byte, *domain.DocumentVersion, error) {
	doc, err := u.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, nil, err
	}

	var version *domain.DocumentVersion
	if versionNum <= 0 {
		version, err = u.docRepo.GetLatestVersion(ctx, docID)
	} else {
		version, err = u.docRepo.GetVersionByNumber(ctx, docID, versionNum)
	}
	if err != nil {
		return nil, nil, err
	}

	f, err := os.Open(version.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}

	// Verify integrity
	if crypto.SHA256File(data) != version.SHA256Hash {
		return nil, nil, fmt.Errorf("file integrity check failed")
	}

	u.audit(ctx, requesterID, "document.download", ip, &doc.MeetingID, &docID, "", "")
	return data, version, nil
}

func (u *documentUsecase) SetReviewers(ctx context.Context, docID uuid.UUID, userIDs []uuid.UUID) error {
	reviewers := make([]*domain.DocumentReviewer, len(userIDs))
	for i, uid := range userIDs {
		reviewers[i] = &domain.DocumentReviewer{
			ID:         uuid.New(),
			DocumentID: docID,
			UserID:     uid,
			OrderIndex: i,
			Status:     domain.ReviewPending,
		}
	}
	return u.docRepo.SetReviewers(ctx, docID, reviewers)
}

func (u *documentUsecase) GetReviewers(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentReviewer, error) {
	return u.docRepo.GetReviewers(ctx, docID)
}

func (u *documentUsecase) GetApprovers(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentApprover, error) {
	return u.docRepo.GetApprovers(ctx, docID)
}

func (u *documentUsecase) SetApprovers(ctx context.Context, docID uuid.UUID, userIDs []uuid.UUID) error {
	approvers := make([]*domain.DocumentApprover, len(userIDs))
	for i, uid := range userIDs {
		approvers[i] = &domain.DocumentApprover{
			ID:         uuid.New(),
			DocumentID: docID,
			UserID:     uid,
			OrderIndex: i,
			Status:     domain.ApproverPending,
		}
	}
	return u.docRepo.SetApprovers(ctx, docID, approvers)
}

func (u *documentUsecase) Submit(ctx context.Context, docID, userID uuid.UUID, ip string) error {
	doc, err := u.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}

	if err := domain.ValidateTransition(doc.Status, domain.DocSubmitted); err != nil {
		return err
	}

	now := time.Now()
	doc.Status = domain.DocSubmitted
	doc.SubmittedAt = &now

	if err := u.docRepo.Update(ctx, doc); err != nil {
		return err
	}

	u.audit(ctx, userID, "document.submit", ip, &doc.MeetingID, &docID, string(domain.DocDraft), string(domain.DocSubmitted))
	return nil
}

func (u *documentUsecase) Review(ctx context.Context, docID, userID uuid.UUID, action, comments, ip string) error {
	doc, err := u.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}

	reviewers, err := u.docRepo.GetReviewers(ctx, docID)
	if err != nil {
		return err
	}

	// Find reviewer entry
	var reviewer *domain.DocumentReviewer
	for _, r := range reviewers {
		if r.UserID == userID {
			reviewer = r
			break
		}
	}
	if reviewer == nil {
		return domain.ErrForbidden
	}

	// Sequential: check order
	if doc.WorkflowMode == domain.WorkflowSequential {
		for _, r := range reviewers {
			if r.OrderIndex < reviewer.OrderIndex && r.Status == domain.ReviewPending {
				return fmt.Errorf("waiting for previous reviewer")
			}
		}
	}

	now := time.Now()
	reviewer.Status = domain.ReviewerStatus(action)
	reviewer.Comments = comments
	reviewer.ReviewedAt = &now

	if err := u.docRepo.UpdateReviewer(ctx, reviewer); err != nil {
		return err
	}

	// Record approval action
	_ = u.docRepo.CreateApproval(ctx, &domain.Approval{
		ID: uuid.New(), DocumentID: docID, UserID: userID,
		Role: "reviewer", Action: action, Comments: comments,
	})

	// Check if all reviewers done → move to under_review (or advance)
	if action == "rejected" {
		doc.Status = domain.DocRejected
		_ = u.docRepo.Update(ctx, doc)
		u.audit(ctx, userID, "document.review.rejected", ip, &doc.MeetingID, &docID, string(domain.DocSubmitted), string(domain.DocRejected))
		return nil
	}

	// Check if all approved
	allDone := true
	for _, r := range reviewers {
		if r.ID == reviewer.ID {
			continue
		}
		if r.Status == domain.ReviewPending {
			allDone = false
			break
		}
	}

	if allDone {
		doc.Status = domain.DocUnderReview
		_ = u.docRepo.Update(ctx, doc)
		u.audit(ctx, userID, "document.review.complete", ip, &doc.MeetingID, &docID, string(domain.DocSubmitted), string(domain.DocUnderReview))
	} else {
		u.audit(ctx, userID, "document.review.approved", ip, &doc.MeetingID, &docID, "", "")
	}

	return nil
}

func (u *documentUsecase) Approve(ctx context.Context, docID, userID uuid.UUID, action, comments, ip string) error {
	doc, err := u.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}

	approvers, err := u.docRepo.GetApprovers(ctx, docID)
	if err != nil {
		return err
	}

	var approver *domain.DocumentApprover
	for _, a := range approvers {
		if a.UserID == userID {
			approver = a
			break
		}
	}
	if approver == nil {
		return domain.ErrForbidden
	}

	// Sequential: check order
	if doc.WorkflowMode == domain.WorkflowSequential {
		for _, a := range approvers {
			if a.OrderIndex < approver.OrderIndex && a.Status == domain.ApproverPending {
				return fmt.Errorf("waiting for previous approver")
			}
		}
	}

	now := time.Now()
	approver.Status = domain.ApproverStatus(action)
	approver.Comments = comments
	approver.ApprovedAt = &now

	if err := u.docRepo.UpdateApprover(ctx, approver); err != nil {
		return err
	}

	_ = u.docRepo.CreateApproval(ctx, &domain.Approval{
		ID: uuid.New(), DocumentID: docID, UserID: userID,
		Role: "approver", Action: action, Comments: comments,
	})

	if action == "rejected" {
		doc.Status = domain.DocRejected
		_ = u.docRepo.Update(ctx, doc)
		u.audit(ctx, userID, "document.approve.rejected", ip, &doc.MeetingID, &docID, string(domain.DocUnderReview), string(domain.DocRejected))
		return nil
	}

	allApproved := true
	for _, a := range approvers {
		if a.ID == approver.ID {
			continue
		}
		if a.Status == domain.ApproverPending {
			allApproved = false
			break
		}
	}

	if allApproved {
		lockedAt := time.Now()
		doc.Status = domain.DocApproved
		doc.LockedAt = &lockedAt
		_ = u.docRepo.Update(ctx, doc)
		u.audit(ctx, userID, "document.approved", ip, &doc.MeetingID, &docID, string(domain.DocUnderReview), string(domain.DocApproved))
	} else {
		u.audit(ctx, userID, "document.approve.partial", ip, &doc.MeetingID, &docID, "", "")
	}

	return nil
}

func (u *documentUsecase) StartReview(ctx context.Context, docID, userID uuid.UUID, ip string) error {
	doc, err := u.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}
	if err := domain.ValidateTransition(doc.Status, domain.DocUnderReview); err != nil {
		return err
	}
	prev := string(doc.Status)
	doc.Status = domain.DocUnderReview
	if err := u.docRepo.Update(ctx, doc); err != nil {
		return err
	}
	u.audit(ctx, userID, "document.review.start", ip, &doc.MeetingID, &docID, prev, string(domain.DocUnderReview))
	return nil
}

func (u *documentUsecase) Archive(ctx context.Context, docID, userID uuid.UUID, ip string) error {
	doc, err := u.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}
	if err := domain.ValidateTransition(doc.Status, domain.DocArchived); err != nil {
		return err
	}
	prev := string(doc.Status)
	doc.Status = domain.DocArchived
	if err := u.docRepo.Update(ctx, doc); err != nil {
		return err
	}
	u.audit(ctx, userID, "document.archive", ip, &doc.MeetingID, &docID, prev, string(domain.DocArchived))
	return nil
}

func (u *documentUsecase) ListVersions(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentVersion, error) {
	return u.docRepo.ListVersions(ctx, docID)
}

func (u *documentUsecase) audit(ctx context.Context, userID uuid.UUID, action, ip string, meetingID, docID *uuid.UUID, prev, next string) {
	entry := &domain.AuditLog{
		UserID:        &userID,
		Action:        action,
		MeetingID:     meetingID,
		DocumentID:    docID,
		PreviousState: prev,
		NewState:      next,
		IPAddress:     ip,
	}
	if err := u.auditRepo.Create(ctx, entry); err != nil {
		logger.Log.Error().Err(err).Msg("audit log failed")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
