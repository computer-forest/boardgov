package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/boardgov/internal/domain"
)

type userRepo struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	q := `INSERT INTO users (id, email, full_name, password_hash, is_active, created_at, updated_at)
	      VALUES ($1,$2,$3,$4,$5,NOW(),NOW()) RETURNING created_at, updated_at`
	return r.db.QueryRow(ctx, q, u.ID, u.Email, u.FullName, u.PasswordHash, u.IsActive).
		Scan(&u.CreatedAt, &u.UpdatedAt)
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	q := `SELECT id,email,full_name,password_hash,is_active,last_login_at,created_at,updated_at
	      FROM users WHERE id=$1 AND deleted_at IS NULL`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.PasswordHash,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return u, err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := `SELECT id,email,full_name,password_hash,is_active,last_login_at,created_at,updated_at
	      FROM users WHERE email=$1 AND deleted_at IS NULL`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.FullName, &u.PasswordHash,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return u, err
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	q := `UPDATE users SET full_name=$1,is_active=$2,last_login_at=$3,updated_at=NOW()
	      WHERE id=$4 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, u.FullName, u.IsActive, u.LastLoginAt, u.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE users SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) List(ctx context.Context, page, pageSize int) ([]*domain.User, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id,email,full_name,is_active,last_login_at,created_at,updated_at
		 FROM users WHERE deleted_at IS NULL ORDER BY full_name ASC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *userRepo) ListPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	q := `SELECT DISTINCT p.name FROM permissions p
	      JOIN role_permissions rp ON rp.permission_id = p.id
	      JOIN user_roles ur ON ur.role_id = rp.role_id
	      WHERE ur.user_id = $1`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}
