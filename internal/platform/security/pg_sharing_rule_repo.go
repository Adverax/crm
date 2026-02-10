package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgSharingRuleRepository implements SharingRuleRepository using pgx.
type PgSharingRuleRepository struct {
	pool *pgxpool.Pool
}

// NewPgSharingRuleRepository creates a new PgSharingRuleRepository.
func NewPgSharingRuleRepository(pool *pgxpool.Pool) *PgSharingRuleRepository {
	return &PgSharingRuleRepository{pool: pool}
}

func (r *PgSharingRuleRepository) Create(ctx context.Context, tx pgx.Tx, input CreateSharingRuleInput) (*SharingRule, error) {
	var rule SharingRule
	err := tx.QueryRow(ctx, `
		INSERT INTO security.sharing_rules (
			object_id, rule_type, source_group_id, target_group_id, access_level,
			criteria_field, criteria_op, criteria_value
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, object_id, rule_type, source_group_id, target_group_id, access_level,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
	`,
		input.ObjectID, input.RuleType, input.SourceGroupID, input.TargetGroupID, input.AccessLevel,
		input.CriteriaField, input.CriteriaOp, input.CriteriaValue,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.RuleType, &rule.SourceGroupID, &rule.TargetGroupID,
		&rule.AccessLevel, &rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgSharingRuleRepo.Create: %w", err)
	}
	return &rule, nil
}

func (r *PgSharingRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*SharingRule, error) {
	var rule SharingRule
	err := r.pool.QueryRow(ctx, `
		SELECT id, object_id, rule_type, source_group_id, target_group_id, access_level,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
		FROM security.sharing_rules WHERE id = $1
	`, id).Scan(
		&rule.ID, &rule.ObjectID, &rule.RuleType, &rule.SourceGroupID, &rule.TargetGroupID,
		&rule.AccessLevel, &rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgSharingRuleRepo.GetByID: %w", err)
	}
	return &rule, nil
}

func (r *PgSharingRuleRepository) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]SharingRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, rule_type, source_group_id, target_group_id, access_level,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
		FROM security.sharing_rules WHERE object_id = $1
		ORDER BY created_at
	`, objectID)
	if err != nil {
		return nil, fmt.Errorf("pgSharingRuleRepo.ListByObjectID: %w", err)
	}
	defer rows.Close()

	rules := make([]SharingRule, 0)
	for rows.Next() {
		var rule SharingRule
		if err := rows.Scan(
			&rule.ID, &rule.ObjectID, &rule.RuleType, &rule.SourceGroupID, &rule.TargetGroupID,
			&rule.AccessLevel, &rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pgSharingRuleRepo.ListByObjectID: scan: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *PgSharingRuleRepository) Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateSharingRuleInput) (*SharingRule, error) {
	var rule SharingRule
	err := tx.QueryRow(ctx, `
		UPDATE security.sharing_rules SET
			target_group_id = $2, access_level = $3,
			criteria_field = $4, criteria_op = $5, criteria_value = $6,
			updated_at = now()
		WHERE id = $1
		RETURNING id, object_id, rule_type, source_group_id, target_group_id, access_level,
			criteria_field, criteria_op, criteria_value, created_at, updated_at
	`,
		id, input.TargetGroupID, input.AccessLevel,
		input.CriteriaField, input.CriteriaOp, input.CriteriaValue,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.RuleType, &rule.SourceGroupID, &rule.TargetGroupID,
		&rule.AccessLevel, &rule.CriteriaField, &rule.CriteriaOp, &rule.CriteriaValue,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgSharingRuleRepo.Update: %w", err)
	}
	return &rule, nil
}

func (r *PgSharingRuleRepository) Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM security.sharing_rules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgSharingRuleRepo.Delete: %w", err)
	}
	return nil
}
