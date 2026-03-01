package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/boardgov/internal/domain"
)

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

// ─── Role Repository ──────────────────────────────────────────────────────────

type roleRepo struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) domain.RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) Create(ctx context.Context, role *domain.Role) error {
	role.ID = uuid.New()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx,
		`INSERT INTO roles (id, name, description, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`,
		role.ID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt)
	return err
}

func (r *roleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role := &domain.Role{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM roles WHERE id=$1 AND deleted_at IS NULL`, id).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	return role, mapError(err)
}

func (r *roleRepo) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	role := &domain.Role{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM roles WHERE name=$1 AND deleted_at IS NULL`, name).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	return role, mapError(err)
}

func (r *roleRepo) List(ctx context.Context) ([]*domain.Role, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, description FROM roles WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *roleRepo) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		userID, roleID, time.Now())
	return err
}

func (r *roleRepo) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1 AND role_id=$2`, userID, roleID)
	return err
}

func (r *roleRepo) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.name, r.description FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id=$1 AND r.deleted_at IS NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *roleRepo) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []*domain.Permission
	for rows.Next() {
		p := &domain.Permission{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *roleRepo) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		roleID, permissionID, time.Now())
	return err
}

func (r *roleRepo) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id=$1 AND p.resource=$2 AND p.action=$3`, userID, resource, action).Scan(&count)
	return count > 0, err
}
