//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgTerritoryRepository implements TerritoryRepository using pgx.
type PgTerritoryRepository struct {
	pool *pgxpool.Pool
}

// NewPgTerritoryRepository creates a new PgTerritoryRepository.
func NewPgTerritoryRepository(pool *pgxpool.Pool) *PgTerritoryRepository {
	return &PgTerritoryRepository{pool: pool}
}

func (r *PgTerritoryRepository) Create(ctx context.Context, tx pgx.Tx, input CreateTerritoryInput) (*Territory, error) {
	var t Territory
	err := tx.QueryRow(ctx, `
		INSERT INTO ee.territories (model_id, parent_id, api_name, label, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, model_id, parent_id, api_name, label, description, created_at, updated_at
	`,
		input.ModelID, input.ParentID, input.APIName, input.Label, input.Description,
	).Scan(
		&t.ID, &t.ModelID, &t.ParentID, &t.APIName, &t.Label, &t.Description,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgTerritoryRepo.Create: %w", err)
	}
	return &t, nil
}

func (r *PgTerritoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*Territory, error) {
	var t Territory
	err := r.pool.QueryRow(ctx, `
		SELECT id, model_id, parent_id, api_name, label, description, created_at, updated_at
		FROM ee.territories WHERE id = $1
	`, id).Scan(
		&t.ID, &t.ModelID, &t.ParentID, &t.APIName, &t.Label, &t.Description,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgTerritoryRepo.GetByID: %w", err)
	}
	return &t, nil
}

func (r *PgTerritoryRepository) GetByAPIName(ctx context.Context, modelID uuid.UUID, apiName string) (*Territory, error) {
	var t Territory
	err := r.pool.QueryRow(ctx, `
		SELECT id, model_id, parent_id, api_name, label, description, created_at, updated_at
		FROM ee.territories WHERE model_id = $1 AND api_name = $2
	`, modelID, apiName).Scan(
		&t.ID, &t.ModelID, &t.ParentID, &t.APIName, &t.Label, &t.Description,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgTerritoryRepo.GetByAPIName: %w", err)
	}
	return &t, nil
}

func (r *PgTerritoryRepository) ListByModelID(ctx context.Context, modelID uuid.UUID) ([]Territory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, model_id, parent_id, api_name, label, description, created_at, updated_at
		FROM ee.territories WHERE model_id = $1
		ORDER BY created_at
	`, modelID)
	if err != nil {
		return nil, fmt.Errorf("pgTerritoryRepo.ListByModelID: %w", err)
	}
	defer rows.Close()

	territories := make([]Territory, 0)
	for rows.Next() {
		var t Territory
		if err := rows.Scan(
			&t.ID, &t.ModelID, &t.ParentID, &t.APIName, &t.Label, &t.Description,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgTerritoryRepo.ListByModelID: scan: %w", err)
		}
		territories = append(territories, t)
	}
	return territories, rows.Err()
}

func (r *PgTerritoryRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateTerritoryInput) (*Territory, error) {
	var t Territory
	err := tx.QueryRow(ctx, `
		UPDATE ee.territories SET
			parent_id = $2, label = $3, description = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, model_id, parent_id, api_name, label, description, created_at, updated_at
	`,
		id, input.ParentID, input.Label, input.Description,
	).Scan(
		&t.ID, &t.ModelID, &t.ParentID, &t.APIName, &t.Label, &t.Description,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgTerritoryRepo.Update: %w", err)
	}
	return &t, nil
}

func (r *PgTerritoryRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM ee.territories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgTerritoryRepo.Delete: %w", err)
	}
	return nil
}
