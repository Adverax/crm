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

// PgModelRepository implements ModelRepository using pgx.
type PgModelRepository struct {
	pool *pgxpool.Pool
}

// NewPgModelRepository creates a new PgModelRepository.
func NewPgModelRepository(pool *pgxpool.Pool) *PgModelRepository {
	return &PgModelRepository{pool: pool}
}

func (r *PgModelRepository) Create(ctx context.Context, tx pgx.Tx, input CreateModelInput) (*TerritoryModel, error) {
	var m TerritoryModel
	err := tx.QueryRow(ctx, `
		INSERT INTO ee.territory_models (api_name, label, description)
		VALUES ($1, $2, $3)
		RETURNING id, api_name, label, description, status, activated_at, archived_at, created_at, updated_at
	`,
		input.APIName, input.Label, input.Description,
	).Scan(
		&m.ID, &m.APIName, &m.Label, &m.Description, &m.Status,
		&m.ActivatedAt, &m.ArchivedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgModelRepo.Create: %w", err)
	}
	return &m, nil
}

func (r *PgModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*TerritoryModel, error) {
	var m TerritoryModel
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, status, activated_at, archived_at, created_at, updated_at
		FROM ee.territory_models WHERE id = $1
	`, id).Scan(
		&m.ID, &m.APIName, &m.Label, &m.Description, &m.Status,
		&m.ActivatedAt, &m.ArchivedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgModelRepo.GetByID: %w", err)
	}
	return &m, nil
}

func (r *PgModelRepository) GetByAPIName(ctx context.Context, apiName string) (*TerritoryModel, error) {
	var m TerritoryModel
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, status, activated_at, archived_at, created_at, updated_at
		FROM ee.territory_models WHERE api_name = $1
	`, apiName).Scan(
		&m.ID, &m.APIName, &m.Label, &m.Description, &m.Status,
		&m.ActivatedAt, &m.ArchivedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgModelRepo.GetByAPIName: %w", err)
	}
	return &m, nil
}

func (r *PgModelRepository) GetActive(ctx context.Context) (*TerritoryModel, error) {
	var m TerritoryModel
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_name, label, description, status, activated_at, archived_at, created_at, updated_at
		FROM ee.territory_models WHERE status = 'active'
	`).Scan(
		&m.ID, &m.APIName, &m.Label, &m.Description, &m.Status,
		&m.ActivatedAt, &m.ArchivedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgModelRepo.GetActive: %w", err)
	}
	return &m, nil
}

func (r *PgModelRepository) List(ctx context.Context, limit, offset int32) ([]TerritoryModel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, api_name, label, description, status, activated_at, archived_at, created_at, updated_at
		FROM ee.territory_models
		ORDER BY created_at
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pgModelRepo.List: %w", err)
	}
	defer rows.Close()

	models := make([]TerritoryModel, 0)
	for rows.Next() {
		var m TerritoryModel
		if err := rows.Scan(
			&m.ID, &m.APIName, &m.Label, &m.Description, &m.Status,
			&m.ActivatedAt, &m.ArchivedAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgModelRepo.List: scan: %w", err)
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

func (r *PgModelRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateModelInput) (*TerritoryModel, error) {
	var m TerritoryModel
	err := tx.QueryRow(ctx, `
		UPDATE ee.territory_models SET
			label = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, api_name, label, description, status, activated_at, archived_at, created_at, updated_at
	`,
		id, input.Label, input.Description,
	).Scan(
		&m.ID, &m.APIName, &m.Label, &m.Description, &m.Status,
		&m.ActivatedAt, &m.ArchivedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgModelRepo.Update: %w", err)
	}
	return &m, nil
}

func (r *PgModelRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status ModelStatus) error {
	_, err := tx.Exec(ctx, `
		UPDATE ee.territory_models SET status = $2, updated_at = now()
		WHERE id = $1
	`, id, status)
	if err != nil {
		return fmt.Errorf("pgModelRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *PgModelRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM ee.territory_models WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgModelRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgModelRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ee.territory_models`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgModelRepo.Count: %w", err)
	}
	return count, nil
}
