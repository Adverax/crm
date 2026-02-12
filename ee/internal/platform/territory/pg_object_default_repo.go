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

// PgObjectDefaultRepository implements ObjectDefaultRepository using pgx.
type PgObjectDefaultRepository struct {
	pool *pgxpool.Pool
}

// NewPgObjectDefaultRepository creates a new PgObjectDefaultRepository.
func NewPgObjectDefaultRepository(pool *pgxpool.Pool) *PgObjectDefaultRepository {
	return &PgObjectDefaultRepository{pool: pool}
}

func (r *PgObjectDefaultRepository) Upsert(ctx context.Context, tx pgx.Tx, input SetObjectDefaultInput) (*TerritoryObjectDefault, error) {
	var d TerritoryObjectDefault
	err := tx.QueryRow(ctx, `
		INSERT INTO ee.territory_object_defaults (territory_id, object_id, access_level)
		VALUES ($1, $2, $3)
		ON CONFLICT (territory_id, object_id) DO UPDATE SET
			access_level = EXCLUDED.access_level, updated_at = now()
		RETURNING id, territory_id, object_id, access_level, created_at, updated_at
	`,
		input.TerritoryID, input.ObjectID, input.AccessLevel,
	).Scan(
		&d.ID, &d.TerritoryID, &d.ObjectID, &d.AccessLevel, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgObjectDefaultRepo.Upsert: %w", err)
	}
	return &d, nil
}

func (r *PgObjectDefaultRepository) GetByTerritoryAndObject(ctx context.Context, territoryID, objectID uuid.UUID) (*TerritoryObjectDefault, error) {
	var d TerritoryObjectDefault
	err := r.pool.QueryRow(ctx, `
		SELECT id, territory_id, object_id, access_level, created_at, updated_at
		FROM ee.territory_object_defaults WHERE territory_id = $1 AND object_id = $2
	`, territoryID, objectID).Scan(
		&d.ID, &d.TerritoryID, &d.ObjectID, &d.AccessLevel, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgObjectDefaultRepo.GetByTerritoryAndObject: %w", err)
	}
	return &d, nil
}

func (r *PgObjectDefaultRepository) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]TerritoryObjectDefault, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, territory_id, object_id, access_level, created_at, updated_at
		FROM ee.territory_object_defaults WHERE territory_id = $1
		ORDER BY created_at
	`, territoryID)
	if err != nil {
		return nil, fmt.Errorf("pgObjectDefaultRepo.ListByTerritoryID: %w", err)
	}
	defer rows.Close()

	defaults := make([]TerritoryObjectDefault, 0)
	for rows.Next() {
		var d TerritoryObjectDefault
		if err := rows.Scan(
			&d.ID, &d.TerritoryID, &d.ObjectID, &d.AccessLevel, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgObjectDefaultRepo.ListByTerritoryID: scan: %w", err)
		}
		defaults = append(defaults, d)
	}
	return defaults, rows.Err()
}

func (r *PgObjectDefaultRepository) Delete(ctx context.Context, tx pgx.Tx, territoryID, objectID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM ee.territory_object_defaults WHERE territory_id = $1 AND object_id = $2
	`, territoryID, objectID)
	if err != nil {
		return fmt.Errorf("pgObjectDefaultRepo.Delete: %w", err)
	}
	return nil
}
