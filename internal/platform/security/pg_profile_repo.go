package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgProfileRepository implements ProfileRepository using pgx.
type PgProfileRepository struct {
	pool *pgxpool.Pool
}

// NewPgProfileRepository creates a new PgProfileRepository.
func NewPgProfileRepository(pool *pgxpool.Pool) *PgProfileRepository {
	return &PgProfileRepository{pool: pool}
}

func (r *PgProfileRepository) Create(ctx context.Context, tx pgx.Tx, profile *Profile) (*Profile, error) {
	var p Profile
	err := tx.QueryRow(ctx, `
		INSERT INTO iam.profiles (api_name, label, description, base_permission_set_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, api_name, label, description, base_permission_set_id, created_at, updated_at
	`, profile.APIName, profile.Label, profile.Description, profile.BasePermissionSetID).Scan(
		&p.ID, &p.APIName, &p.Label, &p.Description,
		&p.BasePermissionSetID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgProfileRepo.Create: %w", err)
	}
	return &p, nil
}

func (r *PgProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	var p Profile
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, base_permission_set_id, created_at, updated_at
		FROM iam.profiles WHERE id = $1
	`, id).Scan(
		&p.ID, &p.APIName, &p.Label, &p.Description,
		&p.BasePermissionSetID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgProfileRepo.GetByID: %w", err)
	}
	return &p, nil
}

func (r *PgProfileRepository) GetByAPIName(ctx context.Context, apiName string) (*Profile, error) {
	var p Profile
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, base_permission_set_id, created_at, updated_at
		FROM iam.profiles WHERE api_name = $1
	`, apiName).Scan(
		&p.ID, &p.APIName, &p.Label, &p.Description,
		&p.BasePermissionSetID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgProfileRepo.GetByAPIName: %w", err)
	}
	return &p, nil
}

func (r *PgProfileRepository) List(ctx context.Context, limit, offset int32) ([]Profile, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, api_name, label, description, base_permission_set_id, created_at, updated_at
		FROM iam.profiles
		ORDER BY created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pgProfileRepo.List: %w", err)
	}
	defer rows.Close()

	profiles := make([]Profile, 0)
	for rows.Next() {
		var p Profile
		if err := rows.Scan(
			&p.ID, &p.APIName, &p.Label, &p.Description,
			&p.BasePermissionSetID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgProfileRepo.List: scan: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (r *PgProfileRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateProfileInput) (*Profile, error) {
	var p Profile
	err := tx.QueryRow(ctx, `
		UPDATE iam.profiles SET
			label = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, api_name, label, description, base_permission_set_id, created_at, updated_at
	`, id, input.Label, input.Description).Scan(
		&p.ID, &p.APIName, &p.Label, &p.Description,
		&p.BasePermissionSetID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgProfileRepo.Update: %w", err)
	}
	return &p, nil
}

func (r *PgProfileRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM iam.profiles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgProfileRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgProfileRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM iam.profiles`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgProfileRepo.Count: %w", err)
	}
	return count, nil
}
