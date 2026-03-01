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

type meetingRepo struct{ db *pgxpool.Pool }

func NewMeetingRepository(db *pgxpool.Pool) domain.MeetingRepository {
	return &meetingRepo{db: db}
}

func (r *meetingRepo) Create(ctx context.Context, m *domain.Meeting) error {
	q := `INSERT INTO meetings (id,title,meeting_type,scheduled_at,location,agenda,status,created_by,created_at,updated_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW()) RETURNING created_at,updated_at`
	return r.db.QueryRow(ctx, q,
		m.ID, m.Title, m.MeetingType, m.ScheduledAt, m.Location, m.Agenda, m.Status, m.CreatedBy,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
}

func (r *meetingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Meeting, error) {
	q := `SELECT id,title,meeting_type,scheduled_at,location,agenda,status,created_by,created_at,updated_at
	      FROM meetings WHERE id=$1 AND deleted_at IS NULL`
	m := &domain.Meeting{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&m.ID, &m.Title, &m.MeetingType, &m.ScheduledAt,
		&m.Location, &m.Agenda, &m.Status, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return m, err
}

func (r *meetingRepo) List(ctx context.Context, filter domain.MeetingFilter) ([]*domain.Meeting, int, error) {
	args := []interface{}{}
	conds := []string{"deleted_at IS NULL"}
	argIdx := 1

	if filter.Status != "" {
		conds = append(conds, fmt.Sprintf("status=$%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.FromDate != nil {
		conds = append(conds, fmt.Sprintf("scheduled_at >= $%d", argIdx))
		args = append(args, filter.FromDate)
		argIdx++
	}
	if filter.ToDate != nil {
		conds = append(conds, fmt.Sprintf("scheduled_at <= $%d", argIdx))
		args = append(args, filter.ToDate)
		argIdx++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	countQ := "SELECT COUNT(*) FROM meetings " + where
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	q := fmt.Sprintf(`SELECT id,title,meeting_type,scheduled_at,location,agenda,status,created_by,created_at,updated_at
	      FROM meetings %s ORDER BY scheduled_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var meetings []*domain.Meeting
	for rows.Next() {
		m := &domain.Meeting{}
		if err := rows.Scan(&m.ID, &m.Title, &m.MeetingType, &m.ScheduledAt,
			&m.Location, &m.Agenda, &m.Status, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		meetings = append(meetings, m)
	}
	return meetings, total, rows.Err()
}

func (r *meetingRepo) Update(ctx context.Context, m *domain.Meeting) error {
	q := `UPDATE meetings SET title=$1,meeting_type=$2,scheduled_at=$3,location=$4,agenda=$5,status=$6,updated_at=NOW()
	      WHERE id=$7 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, m.Title, m.MeetingType, m.ScheduledAt, m.Location, m.Agenda, m.Status, m.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *meetingRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE meetings SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
