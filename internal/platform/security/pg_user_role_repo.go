package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgUserRoleRepository implements UserRoleRepository using pgx.
type PgUserRoleRepository struct {
	pool *pgxpool.Pool
}

// NewPgUserRoleRepository creates a new PgUserRoleRepository.
func NewPgUserRoleRepository(pool *pgxpool.Pool) *PgUserRoleRepository {
	return &PgUserRoleRepository{pool: pool}
}

func (r *PgUserRoleRepository) Create(ctx context.Context, tx pgx.Tx, input CreateUserRoleInput) (*UserRole, error) {
	var role UserRole
	err := tx.QueryRow(ctx, `
		INSERT INTO iam.user_roles (api_name, label, description, parent_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, api_name, label, description, parent_id, created_at, updated_at
	`, input.APIName, input.Label, input.Description, input.ParentID).Scan(
		&role.ID, &role.APIName, &role.Label, &role.Description,
		&role.ParentID, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgUserRoleRepo.Create: %w", err)
	}
	return &role, nil
}

func (r *PgUserRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*UserRole, error) {
	var role UserRole
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, parent_id, created_at, updated_at
		FROM iam.user_roles WHERE id = $1
	`, id).Scan(
		&role.ID, &role.APIName, &role.Label, &role.Description,
		&role.ParentID, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgUserRoleRepo.GetByID: %w", err)
	}
	return &role, nil
}

func (r *PgUserRoleRepository) GetByAPIName(ctx context.Context, apiName string) (*UserRole, error) {
	var role UserRole
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, parent_id, created_at, updated_at
		FROM iam.user_roles WHERE api_name = $1
	`, apiName).Scan(
		&role.ID, &role.APIName, &role.Label, &role.Description,
		&role.ParentID, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgUserRoleRepo.GetByAPIName: %w", err)
	}
	return &role, nil
}

func (r *PgUserRoleRepository) List(ctx context.Context, limit, offset int32) ([]UserRole, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, api_name, label, description, parent_id, created_at, updated_at
		FROM iam.user_roles
		ORDER BY created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pgUserRoleRepo.List: %w", err)
	}
	defer rows.Close()

	var roles []UserRole
	for rows.Next() {
		var role UserRole
		if err := rows.Scan(
			&role.ID, &role.APIName, &role.Label, &role.Description,
			&role.ParentID, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgUserRoleRepo.List: scan: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *PgUserRoleRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateUserRoleInput) (*UserRole, error) {
	var role UserRole
	err := tx.QueryRow(ctx, `
		UPDATE iam.user_roles SET
			label = $2, description = $3, parent_id = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, api_name, label, description, parent_id, created_at, updated_at
	`, id, input.Label, input.Description, input.ParentID).Scan(
		&role.ID, &role.APIName, &role.Label, &role.Description,
		&role.ParentID, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgUserRoleRepo.Update: %w", err)
	}
	return &role, nil
}

func (r *PgUserRoleRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM iam.user_roles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgUserRoleRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgUserRoleRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM iam.user_roles`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgUserRoleRepo.Count: %w", err)
	}
	return count, nil
}
