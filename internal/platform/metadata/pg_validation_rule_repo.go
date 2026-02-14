package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgValidationRuleRepository is a PostgreSQL implementation of ValidationRuleRepository.
type PgValidationRuleRepository struct {
	pool *pgxpool.Pool
}

// NewPgValidationRuleRepository creates a new PgValidationRuleRepository.
func NewPgValidationRuleRepository(pool *pgxpool.Pool) *PgValidationRuleRepository {
	return &PgValidationRuleRepository{pool: pool}
}

func (r *PgValidationRuleRepository) Create(ctx context.Context, input CreateValidationRuleInput) (*ValidationRule, error) {
	rule := &ValidationRule{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO metadata.validation_rules
			(object_id, api_name, label, description, expression,
			 error_message, error_code, severity, when_expression,
			 applies_to, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, object_id, api_name, label, description, expression,
			error_message, error_code, severity, when_expression,
			applies_to, sort_order, is_active, created_at, updated_at`,
		input.ObjectID, input.APIName, input.Label, input.Description,
		input.Expression, input.ErrorMessage, input.ErrorCode, input.Severity,
		input.WhenExpression, input.AppliesTo, input.SortOrder, input.IsActive,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.APIName, &rule.Label, &rule.Description,
		&rule.Expression, &rule.ErrorMessage, &rule.ErrorCode, &rule.Severity,
		&rule.WhenExpression, &rule.AppliesTo, &rule.SortOrder, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgValidationRuleRepo.Create: %w", err)
	}
	return rule, nil
}

func (r *PgValidationRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*ValidationRule, error) {
	rule := &ValidationRule{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, object_id, api_name, label, description, expression,
			error_message, error_code, severity, when_expression,
			applies_to, sort_order, is_active, created_at, updated_at
		FROM metadata.validation_rules
		WHERE id = $1`, id,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.APIName, &rule.Label, &rule.Description,
		&rule.Expression, &rule.ErrorMessage, &rule.ErrorCode, &rule.Severity,
		&rule.WhenExpression, &rule.AppliesTo, &rule.SortOrder, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgValidationRuleRepo.GetByID: %w", err)
	}
	return rule, nil
}

func (r *PgValidationRuleRepository) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]ValidationRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, api_name, label, description, expression,
			error_message, error_code, severity, when_expression,
			applies_to, sort_order, is_active, created_at, updated_at
		FROM metadata.validation_rules
		WHERE object_id = $1
		ORDER BY sort_order, api_name`, objectID)
	if err != nil {
		return nil, fmt.Errorf("pgValidationRuleRepo.ListByObjectID: %w", err)
	}
	defer rows.Close()

	return scanValidationRules(rows)
}

func (r *PgValidationRuleRepository) Update(ctx context.Context, id uuid.UUID, input UpdateValidationRuleInput) (*ValidationRule, error) {
	rule := &ValidationRule{}
	err := r.pool.QueryRow(ctx, `
		UPDATE metadata.validation_rules SET
			label = $2, description = $3, expression = $4,
			error_message = $5, error_code = $6, severity = $7,
			when_expression = $8, applies_to = $9, sort_order = $10,
			is_active = $11, updated_at = now()
		WHERE id = $1
		RETURNING id, object_id, api_name, label, description, expression,
			error_message, error_code, severity, when_expression,
			applies_to, sort_order, is_active, created_at, updated_at`,
		id, input.Label, input.Description, input.Expression,
		input.ErrorMessage, input.ErrorCode, input.Severity,
		input.WhenExpression, input.AppliesTo, input.SortOrder, input.IsActive,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.APIName, &rule.Label, &rule.Description,
		&rule.Expression, &rule.ErrorMessage, &rule.ErrorCode, &rule.Severity,
		&rule.WhenExpression, &rule.AppliesTo, &rule.SortOrder, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgValidationRuleRepo.Update: %w", err)
	}
	return rule, nil
}

func (r *PgValidationRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.validation_rules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgValidationRuleRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgValidationRuleRepository) ListAll(ctx context.Context) ([]ValidationRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, api_name, label, description, expression,
			error_message, error_code, severity, when_expression,
			applies_to, sort_order, is_active, created_at, updated_at
		FROM metadata.validation_rules
		ORDER BY object_id, sort_order, api_name`)
	if err != nil {
		return nil, fmt.Errorf("pgValidationRuleRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanValidationRules(rows)
}

func scanValidationRules(rows pgx.Rows) ([]ValidationRule, error) {
	var rules []ValidationRule
	for rows.Next() {
		var rule ValidationRule
		if err := rows.Scan(
			&rule.ID, &rule.ObjectID, &rule.APIName, &rule.Label, &rule.Description,
			&rule.Expression, &rule.ErrorMessage, &rule.ErrorCode, &rule.Severity,
			&rule.WhenExpression, &rule.AppliesTo, &rule.SortOrder, &rule.IsActive,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanValidationRules: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
