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

// PgUserAssignmentRepository implements UserAssignmentRepository using pgx.
type PgUserAssignmentRepository struct {
	pool *pgxpool.Pool
}

// NewPgUserAssignmentRepository creates a new PgUserAssignmentRepository.
func NewPgUserAssignmentRepository(pool *pgxpool.Pool) *PgUserAssignmentRepository {
	return &PgUserAssignmentRepository{pool: pool}
}

func (r *PgUserAssignmentRepository) Create(ctx context.Context, tx pgx.Tx, input AssignUserInput) (*UserTerritoryAssignment, error) {
	var a UserTerritoryAssignment
	err := tx.QueryRow(ctx, `
		INSERT INTO ee.user_territory_assignments (user_id, territory_id)
		VALUES ($1, $2)
		RETURNING id, user_id, territory_id, created_at
	`,
		input.UserID, input.TerritoryID,
	).Scan(
		&a.ID, &a.UserID, &a.TerritoryID, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgUserAssignmentRepo.Create: %w", err)
	}
	return &a, nil
}

func (r *PgUserAssignmentRepository) Delete(ctx context.Context, tx pgx.Tx, userID, territoryID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM ee.user_territory_assignments WHERE user_id = $1 AND territory_id = $2
	`, userID, territoryID)
	if err != nil {
		return fmt.Errorf("pgUserAssignmentRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgUserAssignmentRepository) GetByUserAndTerritory(ctx context.Context, userID, territoryID uuid.UUID) (*UserTerritoryAssignment, error) {
	var a UserTerritoryAssignment
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, territory_id, created_at
		FROM ee.user_territory_assignments WHERE user_id = $1 AND territory_id = $2
	`, userID, territoryID).Scan(
		&a.ID, &a.UserID, &a.TerritoryID, &a.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgUserAssignmentRepo.GetByUserAndTerritory: %w", err)
	}
	return &a, nil
}

func (r *PgUserAssignmentRepository) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]UserTerritoryAssignment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, territory_id, created_at
		FROM ee.user_territory_assignments WHERE territory_id = $1
		ORDER BY created_at
	`, territoryID)
	if err != nil {
		return nil, fmt.Errorf("pgUserAssignmentRepo.ListByTerritoryID: %w", err)
	}
	defer rows.Close()

	assignments := make([]UserTerritoryAssignment, 0)
	for rows.Next() {
		var a UserTerritoryAssignment
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.TerritoryID, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgUserAssignmentRepo.ListByTerritoryID: scan: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

func (r *PgUserAssignmentRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]UserTerritoryAssignment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, territory_id, created_at
		FROM ee.user_territory_assignments WHERE user_id = $1
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pgUserAssignmentRepo.ListByUserID: %w", err)
	}
	defer rows.Close()

	assignments := make([]UserTerritoryAssignment, 0)
	for rows.Next() {
		var a UserTerritoryAssignment
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.TerritoryID, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgUserAssignmentRepo.ListByUserID: scan: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}
