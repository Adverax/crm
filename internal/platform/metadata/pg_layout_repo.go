package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgLayoutRepository is a PostgreSQL implementation of LayoutRepository.
type PgLayoutRepository struct {
	pool *pgxpool.Pool
}

// NewPgLayoutRepository creates a new PgLayoutRepository.
func NewPgLayoutRepository(pool *pgxpool.Pool) *PgLayoutRepository {
	return &PgLayoutRepository{pool: pool}
}

func (r *PgLayoutRepository) Create(ctx context.Context, input CreateLayoutInput) (*Layout, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.Create: marshal config: %w", err)
	}

	layout := &Layout{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO metadata.layouts
			(object_view_id, form_factor, mode, config)
		VALUES ($1, $2, $3, $4)
		RETURNING id, object_view_id, form_factor, mode, config, created_at, updated_at`,
		input.ObjectViewID, input.FormFactor, input.Mode, configJSON,
	).Scan(
		&layout.ID, &layout.ObjectViewID, &layout.FormFactor, &layout.Mode,
		&configRaw, &layout.CreatedAt, &layout.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.Create: %w", err)
	}

	if err := json.Unmarshal(configRaw, &layout.Config); err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.Create: unmarshal config: %w", err)
	}
	return layout, nil
}

func (r *PgLayoutRepository) GetByID(ctx context.Context, id uuid.UUID) (*Layout, error) {
	layout := &Layout{}
	var configRaw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, object_view_id, form_factor, mode, config, created_at, updated_at
		FROM metadata.layouts
		WHERE id = $1`, id,
	).Scan(
		&layout.ID, &layout.ObjectViewID, &layout.FormFactor, &layout.Mode,
		&configRaw, &layout.CreatedAt, &layout.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgLayoutRepo.GetByID: %w", err)
	}

	if err := json.Unmarshal(configRaw, &layout.Config); err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.GetByID: unmarshal config: %w", err)
	}
	return layout, nil
}

func (r *PgLayoutRepository) ListByObjectViewID(ctx context.Context, ovID uuid.UUID) ([]Layout, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_view_id, form_factor, mode, config, created_at, updated_at
		FROM metadata.layouts
		WHERE object_view_id = $1
		ORDER BY form_factor, mode`, ovID)
	if err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.ListByObjectViewID: %w", err)
	}
	defer rows.Close()

	return scanLayouts(rows)
}

func (r *PgLayoutRepository) ListAll(ctx context.Context) ([]Layout, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_view_id, form_factor, mode, config, created_at, updated_at
		FROM metadata.layouts
		ORDER BY object_view_id, form_factor, mode`)
	if err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanLayouts(rows)
}

func (r *PgLayoutRepository) Update(ctx context.Context, id uuid.UUID, input UpdateLayoutInput) (*Layout, error) {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.Update: marshal config: %w", err)
	}

	layout := &Layout{}
	var configRaw []byte
	err = r.pool.QueryRow(ctx, `
		UPDATE metadata.layouts SET
			config = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, object_view_id, form_factor, mode, config, created_at, updated_at`,
		id, configJSON,
	).Scan(
		&layout.ID, &layout.ObjectViewID, &layout.FormFactor, &layout.Mode,
		&configRaw, &layout.CreatedAt, &layout.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgLayoutRepo.Update: %w", err)
	}

	if err := json.Unmarshal(configRaw, &layout.Config); err != nil {
		return nil, fmt.Errorf("pgLayoutRepo.Update: unmarshal config: %w", err)
	}
	return layout, nil
}

func (r *PgLayoutRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.layouts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgLayoutRepo.Delete: %w", err)
	}
	return nil
}

func scanLayouts(rows pgx.Rows) ([]Layout, error) {
	var layouts []Layout
	for rows.Next() {
		var layout Layout
		var configRaw []byte
		if err := rows.Scan(
			&layout.ID, &layout.ObjectViewID, &layout.FormFactor, &layout.Mode,
			&configRaw, &layout.CreatedAt, &layout.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanLayouts: %w", err)
		}
		if err := json.Unmarshal(configRaw, &layout.Config); err != nil {
			return nil, fmt.Errorf("scanLayouts: unmarshal config: %w", err)
		}
		layouts = append(layouts, layout)
	}
	return layouts, rows.Err()
}
