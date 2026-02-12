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

// PgRecordAssignmentRepository implements RecordAssignmentRepository using pgx.
type PgRecordAssignmentRepository struct {
	pool *pgxpool.Pool
}

// NewPgRecordAssignmentRepository creates a new PgRecordAssignmentRepository.
func NewPgRecordAssignmentRepository(pool *pgxpool.Pool) *PgRecordAssignmentRepository {
	return &PgRecordAssignmentRepository{pool: pool}
}

func (r *PgRecordAssignmentRepository) Create(ctx context.Context, tx pgx.Tx, input AssignRecordInput) (*RecordTerritoryAssignment, error) {
	var a RecordTerritoryAssignment
	err := tx.QueryRow(ctx, `
		INSERT INTO ee.record_territory_assignments (record_id, object_id, territory_id, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id, record_id, object_id, territory_id, reason, created_at
	`,
		input.RecordID, input.ObjectID, input.TerritoryID, input.Reason,
	).Scan(
		&a.ID, &a.RecordID, &a.ObjectID, &a.TerritoryID, &a.Reason, &a.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgRecordAssignmentRepo.Create: %w", err)
	}
	return &a, nil
}

func (r *PgRecordAssignmentRepository) Delete(ctx context.Context, tx pgx.Tx, recordID, objectID, territoryID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM ee.record_territory_assignments
		WHERE record_id = $1 AND object_id = $2 AND territory_id = $3
	`, recordID, objectID, territoryID)
	if err != nil {
		return fmt.Errorf("pgRecordAssignmentRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgRecordAssignmentRepository) GetByRecordAndTerritory(ctx context.Context, recordID, objectID, territoryID uuid.UUID) (*RecordTerritoryAssignment, error) {
	var a RecordTerritoryAssignment
	err := r.pool.QueryRow(ctx, `
		SELECT id, record_id, object_id, territory_id, reason, created_at
		FROM ee.record_territory_assignments
		WHERE record_id = $1 AND object_id = $2 AND territory_id = $3
	`, recordID, objectID, territoryID).Scan(
		&a.ID, &a.RecordID, &a.ObjectID, &a.TerritoryID, &a.Reason, &a.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgRecordAssignmentRepo.GetByRecordAndTerritory: %w", err)
	}
	return &a, nil
}

func (r *PgRecordAssignmentRepository) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]RecordTerritoryAssignment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, record_id, object_id, territory_id, reason, created_at
		FROM ee.record_territory_assignments WHERE territory_id = $1
		ORDER BY created_at
	`, territoryID)
	if err != nil {
		return nil, fmt.Errorf("pgRecordAssignmentRepo.ListByTerritoryID: %w", err)
	}
	defer rows.Close()

	assignments := make([]RecordTerritoryAssignment, 0)
	for rows.Next() {
		var a RecordTerritoryAssignment
		if err := rows.Scan(
			&a.ID, &a.RecordID, &a.ObjectID, &a.TerritoryID, &a.Reason, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgRecordAssignmentRepo.ListByTerritoryID: scan: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

func (r *PgRecordAssignmentRepository) ListByRecordID(ctx context.Context, recordID, objectID uuid.UUID) ([]RecordTerritoryAssignment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, record_id, object_id, territory_id, reason, created_at
		FROM ee.record_territory_assignments WHERE record_id = $1 AND object_id = $2
		ORDER BY created_at
	`, recordID, objectID)
	if err != nil {
		return nil, fmt.Errorf("pgRecordAssignmentRepo.ListByRecordID: %w", err)
	}
	defer rows.Close()

	assignments := make([]RecordTerritoryAssignment, 0)
	for rows.Next() {
		var a RecordTerritoryAssignment
		if err := rows.Scan(
			&a.ID, &a.RecordID, &a.ObjectID, &a.TerritoryID, &a.Reason, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgRecordAssignmentRepo.ListByRecordID: scan: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}
