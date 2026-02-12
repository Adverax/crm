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

// PgAssignmentRuleRepository implements AssignmentRuleRepository using pgx.
type PgAssignmentRuleRepository struct {
	pool *pgxpool.Pool
}

// NewPgAssignmentRuleRepository creates a new PgAssignmentRuleRepository.
func NewPgAssignmentRuleRepository(pool *pgxpool.Pool) *PgAssignmentRuleRepository {
	return &PgAssignmentRuleRepository{pool: pool}
}

func (r *PgAssignmentRuleRepository) Create(ctx context.Context, tx pgx.Tx, input CreateAssignmentRuleInput) (*AssignmentRule, error) {
	var rule AssignmentRule
	err := tx.QueryRow(ctx, `
		INSERT INTO ee.territory_assignment_rules
			(territory_id, object_id, is_active, rule_order, criteria_field, criteria_op, criteria_value)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, territory_id, object_id, is_active, rule_order,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
	`,
		input.TerritoryID, input.ObjectID, input.IsActive, input.RuleOrder,
		input.CriteriaField, input.CriteriaOp, input.CriteriaValue,
	).Scan(
		&rule.ID, &rule.TerritoryID, &rule.ObjectID, &rule.IsActive, &rule.RuleOrder,
		&rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgAssignmentRuleRepo.Create: %w", err)
	}
	return &rule, nil
}

func (r *PgAssignmentRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*AssignmentRule, error) {
	var rule AssignmentRule
	err := r.pool.QueryRow(ctx, `
		SELECT id, territory_id, object_id, is_active, rule_order,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
		FROM ee.territory_assignment_rules WHERE id = $1
	`, id).Scan(
		&rule.ID, &rule.TerritoryID, &rule.ObjectID, &rule.IsActive, &rule.RuleOrder,
		&rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgAssignmentRuleRepo.GetByID: %w", err)
	}
	return &rule, nil
}

func (r *PgAssignmentRuleRepository) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]AssignmentRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, territory_id, object_id, is_active, rule_order,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
		FROM ee.territory_assignment_rules WHERE territory_id = $1
		ORDER BY rule_order, created_at
	`, territoryID)
	if err != nil {
		return nil, fmt.Errorf("pgAssignmentRuleRepo.ListByTerritoryID: %w", err)
	}
	defer rows.Close()

	rules := make([]AssignmentRule, 0)
	for rows.Next() {
		var rule AssignmentRule
		if err := rows.Scan(
			&rule.ID, &rule.TerritoryID, &rule.ObjectID, &rule.IsActive, &rule.RuleOrder,
			&rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgAssignmentRuleRepo.ListByTerritoryID: scan: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *PgAssignmentRuleRepository) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]AssignmentRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, territory_id, object_id, is_active, rule_order,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
		FROM ee.territory_assignment_rules WHERE object_id = $1
		ORDER BY rule_order, created_at
	`, objectID)
	if err != nil {
		return nil, fmt.Errorf("pgAssignmentRuleRepo.ListByObjectID: %w", err)
	}
	defer rows.Close()

	rules := make([]AssignmentRule, 0)
	for rows.Next() {
		var rule AssignmentRule
		if err := rows.Scan(
			&rule.ID, &rule.TerritoryID, &rule.ObjectID, &rule.IsActive, &rule.RuleOrder,
			&rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgAssignmentRuleRepo.ListByObjectID: scan: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *PgAssignmentRuleRepository) ListActiveByObjectID(ctx context.Context, objectID uuid.UUID) ([]AssignmentRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, territory_id, object_id, is_active, rule_order,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
		FROM ee.territory_assignment_rules WHERE object_id = $1 AND is_active = true
		ORDER BY rule_order, created_at
	`, objectID)
	if err != nil {
		return nil, fmt.Errorf("pgAssignmentRuleRepo.ListActiveByObjectID: %w", err)
	}
	defer rows.Close()

	rules := make([]AssignmentRule, 0)
	for rows.Next() {
		var rule AssignmentRule
		if err := rows.Scan(
			&rule.ID, &rule.TerritoryID, &rule.ObjectID, &rule.IsActive, &rule.RuleOrder,
			&rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgAssignmentRuleRepo.ListActiveByObjectID: scan: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *PgAssignmentRuleRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateAssignmentRuleInput) (*AssignmentRule, error) {
	var rule AssignmentRule
	err := tx.QueryRow(ctx, `
		UPDATE ee.territory_assignment_rules SET
			is_active = $2, rule_order = $3, criteria_field = $4,
			criteria_op = $5, criteria_value = $6, updated_at = now()
		WHERE id = $1
		RETURNING id, territory_id, object_id, is_active, rule_order,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
	`,
		id, input.IsActive, input.RuleOrder,
		input.CriteriaField, input.CriteriaOp, input.CriteriaValue,
	).Scan(
		&rule.ID, &rule.TerritoryID, &rule.ObjectID, &rule.IsActive, &rule.RuleOrder,
		&rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgAssignmentRuleRepo.Update: %w", err)
	}
	return &rule, nil
}

func (r *PgAssignmentRuleRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM ee.territory_assignment_rules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgAssignmentRuleRepo.Delete: %w", err)
	}
	return nil
}
