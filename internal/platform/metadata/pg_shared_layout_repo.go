package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgSharedLayoutRepository is a PostgreSQL implementation of SharedLayoutRepository.
type PgSharedLayoutRepository struct {
	pool *pgxpool.Pool
}

// NewPgSharedLayoutRepository creates a new PgSharedLayoutRepository.
func NewPgSharedLayoutRepository(pool *pgxpool.Pool) *PgSharedLayoutRepository {
	return &PgSharedLayoutRepository{pool: pool}
}

func (r *PgSharedLayoutRepository) Create(ctx context.Context, input CreateSharedLayoutInput) (*SharedLayout, error) {
	configJSON := input.Config
	if configJSON == nil {
		configJSON = json.RawMessage(`{}`)
	}

	sl := &SharedLayout{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO metadata.shared_layouts
			(api_name, type, label, config)
		VALUES ($1, $2, $3, $4)
		RETURNING id, api_name, type, label, config, created_at, updated_at`,
		input.APIName, input.Type, input.Label, configJSON,
	).Scan(
		&sl.ID, &sl.APIName, &sl.Type, &sl.Label,
		&sl.Config, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgSharedLayoutRepo.Create: %w", err)
	}
	return sl, nil
}

func (r *PgSharedLayoutRepository) GetByID(ctx context.Context, id uuid.UUID) (*SharedLayout, error) {
	sl := &SharedLayout{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, type, label, config, created_at, updated_at
		FROM metadata.shared_layouts
		WHERE id = $1`, id,
	).Scan(
		&sl.ID, &sl.APIName, &sl.Type, &sl.Label,
		&sl.Config, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgSharedLayoutRepo.GetByID: %w", err)
	}
	return sl, nil
}

func (r *PgSharedLayoutRepository) GetByAPIName(ctx context.Context, apiName string) (*SharedLayout, error) {
	sl := &SharedLayout{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, type, label, config, created_at, updated_at
		FROM metadata.shared_layouts
		WHERE api_name = $1`, apiName,
	).Scan(
		&sl.ID, &sl.APIName, &sl.Type, &sl.Label,
		&sl.Config, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgSharedLayoutRepo.GetByAPIName: %w", err)
	}
	return sl, nil
}

func (r *PgSharedLayoutRepository) ListAll(ctx context.Context) ([]SharedLayout, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, api_name, type, label, config, created_at, updated_at
		FROM metadata.shared_layouts
		ORDER BY api_name`)
	if err != nil {
		return nil, fmt.Errorf("pgSharedLayoutRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanSharedLayouts(rows)
}

func (r *PgSharedLayoutRepository) Update(ctx context.Context, id uuid.UUID, input UpdateSharedLayoutInput) (*SharedLayout, error) {
	configJSON := input.Config
	if configJSON == nil {
		configJSON = json.RawMessage(`{}`)
	}

	sl := &SharedLayout{}
	err := r.pool.QueryRow(ctx, `
		UPDATE metadata.shared_layouts SET
			label = $2, config = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, api_name, type, label, config, created_at, updated_at`,
		id, input.Label, configJSON,
	).Scan(
		&sl.ID, &sl.APIName, &sl.Type, &sl.Label,
		&sl.Config, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgSharedLayoutRepo.Update: %w", err)
	}
	return sl, nil
}

func (r *PgSharedLayoutRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.shared_layouts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgSharedLayoutRepo.Delete: %w", err)
	}
	return nil
}

// CountReferences counts how many layouts reference this shared layout by api_name.
// Scans layout config JSONB for layout_ref matching the given api_name.
func (r *PgSharedLayoutRepository) CountReferences(ctx context.Context, apiName string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM metadata.layouts
		WHERE config::text LIKE '%' || $1 || '%'
		  AND EXISTS (
			SELECT 1
			FROM jsonb_each(COALESCE(config->'field_config', '{}'::jsonb)) AS fc(key, val)
			WHERE val->>'layout_ref' = $1
		  )`, apiName,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgSharedLayoutRepo.CountReferences: %w", err)
	}
	return count, nil
}

func scanSharedLayouts(rows pgx.Rows) ([]SharedLayout, error) {
	var layouts []SharedLayout
	for rows.Next() {
		var sl SharedLayout
		if err := rows.Scan(
			&sl.ID, &sl.APIName, &sl.Type, &sl.Label,
			&sl.Config, &sl.CreatedAt, &sl.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanSharedLayouts: %w", err)
		}
		layouts = append(layouts, sl)
	}
	return layouts, rows.Err()
}
