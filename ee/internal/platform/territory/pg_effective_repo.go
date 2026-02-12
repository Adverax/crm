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

// PgEffectiveRepository implements EffectiveRepository using pgx.
// Delegates data-intensive operations to PL/pgSQL stored functions (ADR-0015).
type PgEffectiveRepository struct {
	pool *pgxpool.Pool
}

// NewPgEffectiveRepository creates a new PgEffectiveRepository.
func NewPgEffectiveRepository(pool *pgxpool.Pool) *PgEffectiveRepository {
	return &PgEffectiveRepository{pool: pool}
}

func (r *PgEffectiveRepository) RebuildHierarchy(ctx context.Context, tx pgx.Tx, modelID uuid.UUID) error {
	_, err := tx.Exec(ctx, `SELECT ee.rebuild_territory_hierarchy($1)`, modelID)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.RebuildHierarchy: %w", err)
	}
	return nil
}

func (r *PgEffectiveRepository) GenerateRecordShareEntries(ctx context.Context, tx pgx.Tx, recordID, objectID, territoryID uuid.UUID, shareTable string) error {
	_, err := tx.Exec(ctx, `SELECT ee.generate_record_share_entries($1, $2, $3, $4)`,
		recordID, objectID, territoryID, shareTable)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.GenerateRecordShareEntries: %w", err)
	}
	return nil
}

func (r *PgEffectiveRepository) ActivateModel(ctx context.Context, tx pgx.Tx, modelID uuid.UUID) error {
	_, err := tx.Exec(ctx, `SELECT ee.activate_territory_model($1)`, modelID)
	if err != nil {
		return fmt.Errorf("pgEffectiveRepo.ActivateModel: %w", err)
	}
	return nil
}

func (r *PgEffectiveRepository) GetUserTerritories(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT territory_id FROM security.effective_user_territory WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pgEffectiveRepo.GetUserTerritories: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("pgEffectiveRepo.GetUserTerritories: scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PgEffectiveRepository) GetTerritoryGroupIDs(ctx context.Context, territoryIDs []uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id FROM iam.groups
		WHERE group_type = 'territory' AND related_territory_id = ANY($1)
	`, territoryIDs)
	if err != nil {
		return nil, fmt.Errorf("pgEffectiveRepo.GetTerritoryGroupIDs: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("pgEffectiveRepo.GetTerritoryGroupIDs: scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
