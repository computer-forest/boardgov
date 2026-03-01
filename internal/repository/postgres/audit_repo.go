package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/boardgov/internal/domain"
)

type auditRepo struct{ db *pgxpool.Pool }

func NewAuditLogRepository(db *pgxpool.Pool) domain.AuditLogRepository {
	return &auditRepo{db: db}
}

func (r *auditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	// Build hash chain
	latest, err := r.GetLatest(ctx)
	var prevHash string
	if err == nil && latest != nil {
		prevHash = latest.Hash
	}

	metaJSON, _ := json.Marshal(log.Metadata)

	// Compute this entry's hash
	raw := fmt.Sprintf("%v|%v|%s|%v|%v|%s|%s|%s|%s",
		log.UserID, log.Action, log.IPAddress,
		log.MeetingID, log.DocumentID,
		log.PreviousState, log.NewState,
		string(metaJSON), prevHash,
	)
	h := sha256.Sum256([]byte(raw))
	log.Hash = hex.EncodeToString(h[:])
	log.PreviousHash = prevHash

	q := `INSERT INTO audit_logs (user_id,action,meeting_id,document_id,previous_state,new_state,ip_address,metadata,hash,previous_hash,created_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW()) RETURNING id,created_at`
	return r.db.QueryRow(ctx, q,
		log.UserID, log.Action, log.MeetingID, log.DocumentID,
		log.PreviousState, log.NewState, log.IPAddress,
		metaJSON, log.Hash, log.PreviousHash,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *auditRepo) GetLatest(ctx context.Context) (*domain.AuditLog, error) {
	q := `SELECT id,hash,previous_hash FROM audit_logs ORDER BY id DESC LIMIT 1`
	l := &domain.AuditLog{}
	err := r.db.QueryRow(ctx, q).Scan(&l.ID, &l.Hash, &l.PreviousHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return l, err
}

func (r *auditRepo) List(ctx context.Context, filter domain.AuditFilter) ([]*domain.AuditLog, int, error) {
	args := []interface{}{}
	conds := []string{}
	argIdx := 1

	if filter.UserID != nil {
		conds = append(conds, fmt.Sprintf("user_id=$%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.DocumentID != nil {
		conds = append(conds, fmt.Sprintf("document_id=$%d", argIdx))
		args = append(args, filter.DocumentID)
		argIdx++
	}
	if filter.MeetingID != nil {
		conds = append(conds, fmt.Sprintf("meeting_id=$%d", argIdx))
		args = append(args, filter.MeetingID)
		argIdx++
	}
	if filter.Action != "" {
		conds = append(conds, fmt.Sprintf("action=$%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	q := fmt.Sprintf(`SELECT id,user_id,action,meeting_id,document_id,previous_state,new_state,ip_address,hash,previous_hash,created_at
	      FROM audit_logs %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, q, args...)
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
	return logs, total, rows.Err()
}

func (r *auditRepo) ValidateChain(ctx context.Context, from, to int64) (bool, error) {
	q := `SELECT id,hash,previous_hash FROM audit_logs WHERE id BETWEEN $1 AND $2 ORDER BY id ASC`
	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	type entry struct {
		id, hash, prevHash string
	}
	var entries []entry
	for rows.Next() {
		var e entry
		var id int64
		if err := rows.Scan(&id, &e.hash, &e.prevHash); err != nil {
			return false, err
		}
		e.id = fmt.Sprintf("%d", id)
		entries = append(entries, e)
	}

	for i := 1; i < len(entries); i++ {
		if entries[i].prevHash != entries[i-1].hash {
			return false, nil
		}
	}
	return true, nil
}
