package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/boardgov/internal/domain"
)

type dashboardRepo struct{ db *pgxpool.Pool }

func NewDashboardRepository(db *pgxpool.Pool) domain.DashboardRepository {
	return &dashboardRepo{db: db}
}

func (r *dashboardRepo) GetPendingApprovals(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.PendingItem, int64, error) {
	var total int64
	_ = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM approvals a
		JOIN documents d ON d.id=a.document_id
		WHERE a.approver_id=$1 AND a.status='PENDING' AND d.deleted_at IS NULL`, userID).Scan(&total)

	rows, err := r.db.Query(ctx, `
		SELECT d.id,d.title,d.meeting_id,m.title,d.status,da.assigned_at
		FROM approvals a
		JOIN documents d ON d.id=a.document_id
		JOIN meetings m ON m.id=d.meeting_id
		JOIN document_approvers da ON da.document_id=d.id AND da.user_id=a.approver_id
		WHERE a.approver_id=$1 AND a.status='PENDING' AND d.deleted_at IS NULL
		ORDER BY da.assigned_at ASC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*domain.PendingItem
	for rows.Next() {
		it := &domain.PendingItem{}
		if err := rows.Scan(&it.DocumentID, &it.DocumentTitle, &it.MeetingID, &it.MeetingTitle, &it.Status, &it.AssignedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	return items, total, nil
}

func (r *dashboardRepo) GetPendingReviews(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.PendingItem, int64, error) {
	var total int64
	_ = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM document_reviewers dr
		JOIN documents d ON d.id=dr.document_id
		WHERE dr.user_id=$1 AND dr.reviewed_at IS NULL AND d.status='UNDER_REVIEW' AND d.deleted_at IS NULL`, userID).Scan(&total)

	rows, err := r.db.Query(ctx, `
		SELECT d.id,d.title,d.meeting_id,m.title,d.status,dr.assigned_at
		FROM document_reviewers dr
		JOIN documents d ON d.id=dr.document_id
		JOIN meetings m ON m.id=d.meeting_id
		WHERE dr.user_id=$1 AND dr.reviewed_at IS NULL AND d.status='UNDER_REVIEW' AND d.deleted_at IS NULL
		ORDER BY dr.assigned_at ASC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*domain.PendingItem
	for rows.Next() {
		it := &domain.PendingItem{}
		if err := rows.Scan(&it.DocumentID, &it.DocumentTitle, &it.MeetingID, &it.MeetingTitle, &it.Status, &it.AssignedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	return items, total, nil
}

func (r *dashboardRepo) GetOverdueDocuments(ctx context.Context, offset, limit int) ([]*domain.OverdueDocument, int64, error) {
	cutoff := time.Now().AddDate(0, 0, -7)
	var total int64
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM documents d WHERE d.status IN ('SUBMITTED','UNDER_REVIEW') AND d.updated_at<$1 AND d.deleted_at IS NULL`, cutoff).Scan(&total)
	rows, err := r.db.Query(ctx, `
		SELECT d.id,d.title,m.title,d.status,EXTRACT(DAY FROM NOW()-d.updated_at)::int AS days_overdue
		FROM documents d JOIN meetings m ON m.id=d.meeting_id
		WHERE d.status IN ('SUBMITTED','UNDER_REVIEW') AND d.updated_at<$1 AND d.deleted_at IS NULL
		ORDER BY days_overdue DESC LIMIT $2 OFFSET $3`, cutoff, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*domain.OverdueDocument
	for rows.Next() {
		it := &domain.OverdueDocument{}
		if err := rows.Scan(&it.DocumentID, &it.DocumentTitle, &it.MeetingTitle, &it.Status, &it.DaysOverdue); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	return items, total, nil
}

func (r *dashboardRepo) GetMeetingStatusSummary(ctx context.Context) ([]*domain.MeetingStatusSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.id,m.title,m.meeting_type,
		  COUNT(d.id),
		  COUNT(d.id) FILTER (WHERE d.status='DRAFT'),
		  COUNT(d.id) FILTER (WHERE d.status='SUBMITTED'),
		  COUNT(d.id) FILTER (WHERE d.status='UNDER_REVIEW'),
		  COUNT(d.id) FILTER (WHERE d.status='APPROVED'),
		  COUNT(d.id) FILTER (WHERE d.status='REJECTED'),
		  COUNT(d.id) FILTER (WHERE d.status='ARCHIVED')
		FROM meetings m
		LEFT JOIN documents d ON d.meeting_id=m.id AND d.deleted_at IS NULL
		WHERE m.deleted_at IS NULL GROUP BY m.id,m.title,m.meeting_type ORDER BY m.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.MeetingStatusSummary
	for rows.Next() {
		s := &domain.MeetingStatusSummary{}
		if err := rows.Scan(&s.MeetingID, &s.MeetingTitle, &s.MeetingType,
			&s.Total, &s.Draft, &s.Submitted, &s.UnderReview, &s.Approved, &s.Rejected, &s.Archived); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *dashboardRepo) GetDocumentAgingAnalysis(ctx context.Context) ([]*domain.AgingBucket, error) {
	rows, err := r.db.Query(ctx, `
		SELECT CASE
		  WHEN age_days<=1 THEN '0-1 day' WHEN age_days<=3 THEN '2-3 days'
		  WHEN age_days<=7 THEN '4-7 days' WHEN age_days<=14 THEN '8-14 days'
		  ELSE '15+ days' END AS bucket, COUNT(*) as count
		FROM (SELECT EXTRACT(DAY FROM NOW()-created_at)::int AS age_days FROM documents
		      WHERE status NOT IN ('APPROVED','ARCHIVED') AND deleted_at IS NULL) t
		GROUP BY bucket ORDER BY MIN(age_days)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.AgingBucket
	for rows.Next() {
		b := &domain.AgingBucket{}
		if err := rows.Scan(&b.Bucket, &b.Count); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *dashboardRepo) GetApprovalTurnaroundMetrics(ctx context.Context) (*domain.TurnaroundMetrics, error) {
	m := &domain.TurnaroundMetrics{}
	_ = r.db.QueryRow(ctx, `
		SELECT
		  COALESCE(AVG(EXTRACT(EPOCH FROM (actioned_at-created_at))/3600),0),
		  COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (actioned_at-created_at))/3600),0),
		  COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (actioned_at-created_at))/3600),0)
		FROM approvals WHERE actioned_at IS NOT NULL`).
		Scan(&m.AvgApprovalHours, &m.P50ApprovalHours, &m.P90ApprovalHours)
	return m, nil
}

func (r *dashboardRepo) GetRejectionRatioAnalytics(ctx context.Context) (*domain.RejectionAnalytics, error) {
	ra := &domain.RejectionAnalytics{}
	_ = r.db.QueryRow(ctx, `
		SELECT COUNT(*),
		  COUNT(*) FILTER (WHERE status='REJECTED'),
		  ROUND(COUNT(*) FILTER (WHERE status='REJECTED')::numeric/NULLIF(COUNT(*),0)*100,2)
		FROM documents WHERE deleted_at IS NULL`).
		Scan(&ra.TotalDocuments, &ra.TotalRejected, &ra.RejectionRate)
	return ra, nil
}

func (r *dashboardRepo) GetActivityTimeline(ctx context.Context, userID *uuid.UUID, offset, limit int) ([]*domain.AuditLog, int64, error) {
	var total int64
	var rows interface {
		Next() bool
		Scan(dest ...any) error
		Close()
	}
	var err error

	if userID != nil {
		_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE user_id=$1`, userID).Scan(&total)
		rows, err = r.db.Query(ctx, `
			SELECT id,user_id,action,meeting_id,document_id,previous_state,new_state,ip_address,hash,previous_hash,created_at
			FROM audit_logs WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	} else {
		_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&total)
		rows, err = r.db.Query(ctx, `
			SELECT id,user_id,action,meeting_id,document_id,previous_state,new_state,ip_address,hash,previous_hash,created_at
			FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		l := &domain.AuditLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.MeetingID, &l.DocumentID,
			&l.PreviousState, &l.NewState, &l.IPAddress, &l.Hash, &l.PreviousHash, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}
