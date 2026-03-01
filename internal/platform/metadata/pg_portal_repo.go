package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgPortalRepository is a PostgreSQL implementation of PortalRepository.
type PgPortalRepository struct {
	pool *pgxpool.Pool
}

// NewPgPortalRepository creates a new PgPortalRepository.
func NewPgPortalRepository(pool *pgxpool.Pool) *PgPortalRepository {
	return &PgPortalRepository{pool: pool}
}

func (r *PgPortalRepository) Create(ctx context.Context, input CreatePortalInput) (*Portal, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgPortalRepo.Create: marshal config: %w", err)
	}

	ov := &Portal{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO metadata.portals
			(profile_id, api_name, label, description, config)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, profile_id, api_name, label, description,
			config, created_at, updated_at`,
		input.ProfileID, input.APIName, input.Label,
		input.Description, configJSON,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgPortalRepo.Create: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgPortalRepo.Create: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgPortalRepository) GetByID(ctx context.Context, id uuid.UUID) (*Portal, error) {
	ov := &Portal{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, api_name, label, description,
			config, created_at, updated_at
		FROM metadata.portals
		WHERE id = $1`, id,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgPortalRepo.GetByID: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgPortalRepo.GetByID: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgPortalRepository) GetByAPIName(ctx context.Context, apiName string) (*Portal, error) {
	ov := &Portal{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, profile_id, api_name, label, description,
			config, created_at, updated_at
		FROM metadata.portals
		WHERE api_name = $1`, apiName,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgPortalRepo.GetByAPIName: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgPortalRepo.GetByAPIName: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgPortalRepository) ListAll(ctx context.Context) ([]Portal, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, profile_id, api_name, label, description,
			config, created_at, updated_at
		FROM metadata.portals
		ORDER BY api_name`)
	if err != nil {
		return nil, fmt.Errorf("pgPortalRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanPortals(rows)
}

func (r *PgPortalRepository) Update(ctx context.Context, id uuid.UUID, input UpdatePortalInput) (*Portal, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgPortalRepo.Update: marshal config: %w", err)
	}

	ov := &Portal{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		UPDATE metadata.portals SET
			label = $2, description = $3,
			config = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, profile_id, api_name, label, description,
			config, created_at, updated_at`,
		id, input.Label, input.Description, configJSON,
	).Scan(
		&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
		&ov.Description, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgPortalRepo.Update: %w", err)
	}

	if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
		return nil, fmt.Errorf("pgPortalRepo.Update: unmarshal config: %w", err)
	}
	return ov, nil
}

func (r *PgPortalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.portals WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgPortalRepo.Delete: %w", err)
	}
	return nil
}

func scanPortals(rows pgx.Rows) ([]Portal, error) {
	var views []Portal
	for rows.Next() {
		var ov Portal
		var configRaw []byte
		if err := rows.Scan(
			&ov.ID, &ov.ProfileID, &ov.APIName, &ov.Label,
			&ov.Description, &configRaw, &ov.CreatedAt, &ov.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanPortals: %w", err)
		}
		if err := json.Unmarshal(configRaw, &ov.Config); err != nil {
			return nil, fmt.Errorf("scanPortals: unmarshal config: %w", err)
		}
		views = append(views, ov)
	}
	return views, rows.Err()
}
