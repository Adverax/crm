package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgPermissionSetRepository implements PermissionSetRepository using pgx.
type PgPermissionSetRepository struct {
	pool *pgxpool.Pool
}

// NewPgPermissionSetRepository creates a new PgPermissionSetRepository.
func NewPgPermissionSetRepository(pool *pgxpool.Pool) *PgPermissionSetRepository {
	return &PgPermissionSetRepository{pool: pool}
}

func (r *PgPermissionSetRepository) Create(ctx context.Context, tx pgx.Tx, input CreatePermissionSetInput) (*PermissionSet, error) {
	var ps PermissionSet
	err := tx.QueryRow(ctx, `
		INSERT INTO iam.permission_sets (api_name, label, description, ps_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, api_name, label, description, ps_type, created_at, updated_at
	`, input.APIName, input.Label, input.Description, input.PSType).Scan(
		&ps.ID, &ps.APIName, &ps.Label, &ps.Description,
		&ps.PSType, &ps.CreatedAt, &ps.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgPermissionSetRepo.Create: %w", err)
	}
	return &ps, nil
}

func (r *PgPermissionSetRepository) GetByID(ctx context.Context, id uuid.UUID) (*PermissionSet, error) {
	var ps PermissionSet
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, ps_type, created_at, updated_at
		FROM iam.permission_sets WHERE id = $1
	`, id).Scan(
		&ps.ID, &ps.APIName, &ps.Label, &ps.Description,
		&ps.PSType, &ps.CreatedAt, &ps.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgPermissionSetRepo.GetByID: %w", err)
	}
	return &ps, nil
}

func (r *PgPermissionSetRepository) GetByAPIName(ctx context.Context, apiName string) (*PermissionSet, error) {
	var ps PermissionSet
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, ps_type, created_at, updated_at
		FROM iam.permission_sets WHERE api_name = $1
	`, apiName).Scan(
		&ps.ID, &ps.APIName, &ps.Label, &ps.Description,
		&ps.PSType, &ps.CreatedAt, &ps.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgPermissionSetRepo.GetByAPIName: %w", err)
	}
	return &ps, nil
}

func (r *PgPermissionSetRepository) List(ctx context.Context, limit, offset int32) ([]PermissionSet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, api_name, label, description, ps_type, created_at, updated_at
		FROM iam.permission_sets
		ORDER BY created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pgPermissionSetRepo.List: %w", err)
	}
	defer rows.Close()

	var sets []PermissionSet
	for rows.Next() {
		var ps PermissionSet
		if err := rows.Scan(
			&ps.ID, &ps.APIName, &ps.Label, &ps.Description,
			&ps.PSType, &ps.CreatedAt, &ps.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgPermissionSetRepo.List: scan: %w", err)
		}
		sets = append(sets, ps)
	}
	return sets, rows.Err()
}

func (r *PgPermissionSetRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdatePermissionSetInput) (*PermissionSet, error) {
	var ps PermissionSet
	err := tx.QueryRow(ctx, `
		UPDATE iam.permission_sets SET
			label = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, api_name, label, description, ps_type, created_at, updated_at
	`, id, input.Label, input.Description).Scan(
		&ps.ID, &ps.APIName, &ps.Label, &ps.Description,
		&ps.PSType, &ps.CreatedAt, &ps.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgPermissionSetRepo.Update: %w", err)
	}
	return &ps, nil
}

func (r *PgPermissionSetRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM iam.permission_sets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgPermissionSetRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgPermissionSetRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM iam.permission_sets`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgPermissionSetRepo.Count: %w", err)
	}
	return count, nil
}
