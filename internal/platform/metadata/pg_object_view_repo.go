package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgObjectViewRepository is a PostgreSQL implementation of ObjectViewRepository.
type PgObjectViewRepository struct {
	pool *pgxpool.Pool
}

// NewPgObjectViewRepository creates a new PgObjectViewRepository.
func NewPgObjectViewRepository(pool *pgxpool.Pool) *PgObjectViewRepository {
	return &PgObjectViewRepository{pool: pool}
}

func (r *PgObjectViewRepository) Create(ctx context.Context, input CreateObjectViewInput) (*ObjectView, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.Create: marshal config: %w", err)
	}

	ov := &ObjectView{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO metadata.object_views
			(profile_id, api_name, label, description, is_default, config)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, profile_id, api_name, label, description,
			is_default, config, created_at, updated_at`,
		input.ProfileID, input.APIName, input.Label,
		input.Description, input.IsDefault, configJSON,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &ov.IsDefault, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.Create: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.Create: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgObjectViewRepository) GetByID(ctx context.Context, id uuid.UUID) (*ObjectView, error) {
	ov := &ObjectView{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, api_name, label, description,
			is_default, config, created_at, updated_at
		FROM metadata.object_views
		WHERE id = $1`, id,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &ov.IsDefault, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgObjectViewRepo.GetByID: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.GetByID: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgObjectViewRepository) GetByAPIName(ctx context.Context, apiName string) (*ObjectView, error) {
	ov := &ObjectView{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, api_name, label, description,
			is_default, config, created_at, updated_at
		FROM metadata.object_views
		WHERE api_name = $1`, apiName,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &ov.IsDefault, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgObjectViewRepo.GetByAPIName: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.GetByAPIName: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgObjectViewRepository) ListAll(ctx context.Context) ([]ObjectView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, profile_id, api_name, label, description,
			is_default, config, created_at, updated_at
		FROM metadata.object_views
		ORDER BY api_name`)
	if err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanObjectViews(rows)
}

func (r *PgObjectViewRepository) Update(ctx context.Context, id uuid.UUID, input UpdateObjectViewInput) (*ObjectView, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.Update: marshal config: %w", err)
	}

	ov := &ObjectView{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		UPDATE metadata.object_views SET
			label = $2, description = $3, is_default = $4,
			config = $5, updated_at = now()
		WHERE id = $1
		RETURNING id, profile_id, api_name, label, description,
			is_default, config, created_at, updated_at`,
		id, input.Label, input.Description, input.IsDefault, configJSON,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &ov.IsDefault, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgObjectViewRepo.Update: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgObjectViewRepo.Update: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgObjectViewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.object_views WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgObjectViewRepo.Delete: %w", err)
	}
	return nil
}

func scanObjectViews(rows pgx.Rows) ([]ObjectView, error) {
	var views []ObjectView
	for rows.Next() {
		var ov ObjectView
		var configRaw []byte
		if err := rows.Scan(
			&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
			&ov.Description, &ov.IsDefault, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanObjectViews: %w", err)
		}
		if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
			return nil, fmt.Errorf("scanObjectViews: unmarshal config: %w", err)
		}
		views = append(views, ov)
	}
	return views, rows.Err()
}
