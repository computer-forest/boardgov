package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/boardgov/internal/domain"
)

type refreshTokenRepo struct{ db *pgxpool.Pool }

func NewRefreshTokenRepository(db *pgxpool.Pool) domain.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	q := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
	      VALUES ($1,$2,$3,$4,NOW()) RETURNING created_at`
	return r.db.QueryRow(ctx, q, token.ID, token.UserID, token.TokenHash, token.ExpiresAt).
		Scan(&token.CreatedAt)
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	q := `SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
	      FROM refresh_tokens WHERE token_hash=$1`
	t := &domain.RefreshToken{}
	err := r.db.QueryRow(ctx, q, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=NOW() WHERE id=$1`, id)
	return err
}
