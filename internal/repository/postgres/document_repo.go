package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/boardgov/internal/domain"
)

type documentRepo struct{ db *pgxpool.Pool }

func NewDocumentRepository(db *pgxpool.Pool) domain.DocumentRepository {
	return &documentRepo{db: db}
}

func (r *documentRepo) Create(ctx context.Context, d *domain.Document) error {
	q := `INSERT INTO documents (id,meeting_id,title,description,status,workflow_mode,created_by,created_at,updated_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW()) RETURNING created_at,updated_at`
	return r.db.QueryRow(ctx, q,
		d.ID, d.MeetingID, d.Title, d.Description, d.Status, d.WorkflowMode, d.CreatedBy,
	).Scan(&d.CreatedAt, &d.UpdatedAt)
}

func (r *documentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	q := `SELECT id,meeting_id,title,description,status,workflow_mode,submitted_at,locked_at,created_by,created_at,updated_at
	      FROM documents WHERE id=$1 AND deleted_at IS NULL`
	d := &domain.Document{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&d.ID, &d.MeetingID, &d.Title, &d.Description, &d.Status, &d.WorkflowMode,
		&d.SubmittedAt, &d.LockedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return d, err
}

func (r *documentRepo) List(ctx context.Context, filter domain.DocumentFilter) ([]*domain.Document, int, error) {
	args := []interface{}{}
	conds := []string{"deleted_at IS NULL"}
	argIdx := 1

	if filter.MeetingID != nil {
		conds = append(conds, fmt.Sprintf("meeting_id=$%d", argIdx))
		args = append(args, filter.MeetingID)
		argIdx++
	}
	if filter.Status != "" {
		conds = append(conds, fmt.Sprintf("status=$%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.CreatedBy != nil {
		conds = append(conds, fmt.Sprintf("created_by=$%d", argIdx))
		args = append(args, filter.CreatedBy)
		argIdx++
	}

	where := "WHERE " + strings.Join(conds, " AND ")
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM documents "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	q := fmt.Sprintf(`SELECT id,meeting_id,title,description,status,workflow_mode,submitted_at,locked_at,created_by,created_at,updated_at
	      FROM documents %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []*domain.Document
	for rows.Next() {
		d := &domain.Document{}
		if err := rows.Scan(&d.ID, &d.MeetingID, &d.Title, &d.Description, &d.Status, &d.WorkflowMode,
			&d.SubmittedAt, &d.LockedAt, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		docs = append(docs, d)
	}
	return docs, total, rows.Err()
}

func (r *documentRepo) Update(ctx context.Context, d *domain.Document) error {
	q := `UPDATE documents SET title=$1,description=$2,status=$3,workflow_mode=$4,submitted_at=$5,locked_at=$6,updated_at=NOW()
	      WHERE id=$7 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, d.Title, d.Description, d.Status, d.WorkflowMode, d.SubmittedAt, d.LockedAt, d.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *documentRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE documents SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *documentRepo) GetLatestVersion(ctx context.Context, docID uuid.UUID) (*domain.DocumentVersion, error) {
	q := `SELECT id,document_id,version_number,file_name,storage_path,file_size,mime_type,sha256_hash,uploaded_by,created_at
	      FROM document_versions WHERE document_id=$1 ORDER BY version_number DESC LIMIT 1`
	v := &domain.DocumentVersion{}
	err := r.db.QueryRow(ctx, q, docID).Scan(
		&v.ID, &v.DocumentID, &v.VersionNumber, &v.FileName, &v.StoragePath,
		&v.FileSize, &v.MimeType, &v.SHA256Hash, &v.UploadedBy, &v.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return v, err
}

func (r *documentRepo) CreateVersion(ctx context.Context, v *domain.DocumentVersion) error {
	q := `INSERT INTO document_versions (id,document_id,version_number,file_name,storage_path,file_size,mime_type,sha256_hash,uploaded_by,created_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW()) RETURNING created_at`
	return r.db.QueryRow(ctx, q,
		v.ID, v.DocumentID, v.VersionNumber, v.FileName, v.StoragePath,
		v.FileSize, v.MimeType, v.SHA256Hash, v.UploadedBy,
	).Scan(&v.CreatedAt)
}

func (r *documentRepo) ListVersions(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentVersion, error) {
	q := `SELECT id,document_id,version_number,file_name,storage_path,file_size,mime_type,sha256_hash,uploaded_by,created_at
	      FROM document_versions WHERE document_id=$1 ORDER BY version_number DESC`
	rows, err := r.db.Query(ctx, q, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var versions []*domain.DocumentVersion
	for rows.Next() {
		v := &domain.DocumentVersion{}
		if err := rows.Scan(&v.ID, &v.DocumentID, &v.VersionNumber, &v.FileName, &v.StoragePath,
			&v.FileSize, &v.MimeType, &v.SHA256Hash, &v.UploadedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *documentRepo) GetVersionByNumber(ctx context.Context, docID uuid.UUID, versionNum int) (*domain.DocumentVersion, error) {
	q := `SELECT id,document_id,version_number,file_name,storage_path,file_size,mime_type,sha256_hash,uploaded_by,created_at
	      FROM document_versions WHERE document_id=$1 AND version_number=$2`
	v := &domain.DocumentVersion{}
	err := r.db.QueryRow(ctx, q, docID, versionNum).Scan(
		&v.ID, &v.DocumentID, &v.VersionNumber, &v.FileName, &v.StoragePath,
		&v.FileSize, &v.MimeType, &v.SHA256Hash, &v.UploadedBy, &v.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return v, err
}

func (r *documentRepo) GetReviewers(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentReviewer, error) {
	q := `SELECT id,document_id,user_id,order_index,status,comments,reviewed_at,created_at
	      FROM document_reviewers WHERE document_id=$1 ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reviewers []*domain.DocumentReviewer
	for rows.Next() {
		rv := &domain.DocumentReviewer{}
		if err := rows.Scan(&rv.ID, &rv.DocumentID, &rv.UserID, &rv.OrderIndex,
			&rv.Status, &rv.Comments, &rv.ReviewedAt, &rv.CreatedAt); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, rv)
	}
	return reviewers, rows.Err()
}

func (r *documentRepo) SetReviewers(ctx context.Context, docID uuid.UUID, reviewers []*domain.DocumentReviewer) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM document_reviewers WHERE document_id=$1`, docID); err != nil {
		return err
	}
	for _, rv := range reviewers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO document_reviewers (id,document_id,user_id,order_index,status,created_at)
			 VALUES ($1,$2,$3,$4,'pending',NOW())`,
			rv.ID, rv.DocumentID, rv.UserID, rv.OrderIndex,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *documentRepo) UpdateReviewer(ctx context.Context, rv *domain.DocumentReviewer) error {
	q := `UPDATE document_reviewers SET status=$1,comments=$2,reviewed_at=$3 WHERE id=$4`
	_, err := r.db.Exec(ctx, q, rv.Status, rv.Comments, rv.ReviewedAt, rv.ID)
	return err
}

func (r *documentRepo) GetApprovers(ctx context.Context, docID uuid.UUID) ([]*domain.DocumentApprover, error) {
	q := `SELECT id,document_id,user_id,order_index,status,comments,approved_at,created_at
	      FROM document_approvers WHERE document_id=$1 ORDER BY order_index`
	rows, err := r.db.Query(ctx, q, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var approvers []*domain.DocumentApprover
	for rows.Next() {
		a := &domain.DocumentApprover{}
		if err := rows.Scan(&a.ID, &a.DocumentID, &a.UserID, &a.OrderIndex,
			&a.Status, &a.Comments, &a.ApprovedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		approvers = append(approvers, a)
	}
	return approvers, rows.Err()
}

func (r *documentRepo) SetApprovers(ctx context.Context, docID uuid.UUID, approvers []*domain.DocumentApprover) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM document_approvers WHERE document_id=$1`, docID); err != nil {
		return err
	}
	for _, a := range approvers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO document_approvers (id,document_id,user_id,order_index,status,created_at)
			 VALUES ($1,$2,$3,$4,'pending',NOW())`,
			a.ID, a.DocumentID, a.UserID, a.OrderIndex,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *documentRepo) UpdateApprover(ctx context.Context, a *domain.DocumentApprover) error {
	q := `UPDATE document_approvers SET status=$1,comments=$2,approved_at=$3 WHERE id=$4`
	_, err := r.db.Exec(ctx, q, a.Status, a.Comments, a.ApprovedAt, a.ID)
	return err
}

func (r *documentRepo) CreateApproval(ctx context.Context, a *domain.Approval) error {
	q := `INSERT INTO approvals (id,document_id,user_id,role,action,comments,created_at)
	      VALUES ($1,$2,$3,$4,$5,$6,NOW())`
	_, err := r.db.Exec(ctx, q, a.ID, a.DocumentID, a.UserID, a.Role, a.Action, a.Comments)
	return err
}
