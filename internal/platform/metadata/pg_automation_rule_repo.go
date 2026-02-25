package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgAutomationRuleRepository is a PostgreSQL implementation of AutomationRuleRepository.
type PgAutomationRuleRepository struct {
	pool *pgxpool.Pool
}

// NewPgAutomationRuleRepository creates a new PgAutomationRuleRepository.
func NewPgAutomationRuleRepository(pool *pgxpool.Pool) *PgAutomationRuleRepository {
	return &PgAutomationRuleRepository{pool: pool}
}

func (r *PgAutomationRuleRepository) Create(ctx context.Context, input CreateAutomationRuleInput) (*AutomationRule, error) {
	rule := &AutomationRule{}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO metadata.automation_rules
			(object_id, name, description, event_type, condition,
			 procedure_code, execution_mode, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, object_id, name, description, event_type, condition,
			procedure_code, execution_mode, sort_order, is_active,
			created_at, updated_at`,
		input.ObjectID, input.Name, input.Description,
		input.EventType, input.Condition,
		input.ProcedureCode,
		input.ExecutionMode, input.SortOrder, input.IsActive,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.Name, &rule.Description,
		&rule.EventType, &rule.Condition,
		&rule.ProcedureCode,
		&rule.ExecutionMode, &rule.SortOrder, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pgAutomationRuleRepo.Create: %w", err)
	}
	return rule, nil
}

func (r *PgAutomationRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*AutomationRule, error) {
	rule := &AutomationRule{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, object_id, name, description, event_type, condition,
			procedure_code, execution_mode, sort_order, is_active,
			created_at, updated_at
		FROM metadata.automation_rules
		WHERE id = $1`, id,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.Name, &rule.Description,
		&rule.EventType, &rule.Condition,
		&rule.ProcedureCode,
		&rule.ExecutionMode, &rule.SortOrder, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgAutomationRuleRepo.GetByID: %w", err)
	}
	return rule, nil
}

func (r *PgAutomationRuleRepository) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]AutomationRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, name, description, event_type, condition,
			procedure_code, execution_mode, sort_order, is_active,
			created_at, updated_at
		FROM metadata.automation_rules
		WHERE object_id = $1
		ORDER BY sort_order, name`, objectID)
	if err != nil {
		return nil, fmt.Errorf("pgAutomationRuleRepo.ListByObjectID: %w", err)
	}
	defer rows.Close()

	return scanAutomationRules(rows)
}

func (r *PgAutomationRuleRepository) ListAll(ctx context.Context) ([]AutomationRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, object_id, name, description, event_type, condition,
			procedure_code, execution_mode, sort_order, is_active,
			created_at, updated_at
		FROM metadata.automation_rules
		ORDER BY object_id, sort_order, name`)
	if err != nil {
		return nil, fmt.Errorf("pgAutomationRuleRepo.ListAll: %w", err)
	}
	defer rows.Close()

	return scanAutomationRules(rows)
}

func (r *PgAutomationRuleRepository) Update(ctx context.Context, id uuid.UUID, input UpdateAutomationRuleInput) (*AutomationRule, error) {
	rule := &AutomationRule{}
	err := r.pool.QueryRow(ctx, `
		UPDATE metadata.automation_rules SET
			name = $2, description = $3, event_type = $4, condition = $5,
			procedure_code = $6, execution_mode = $7,
			sort_order = $8, is_active = $9, updated_at = now()
		WHERE id = $1
		RETURNING id, object_id, name, description, event_type, condition,
			procedure_code, execution_mode, sort_order, is_active,
			created_at, updated_at`,
		id, input.Name, input.Description,
		input.EventType, input.Condition,
		input.ProcedureCode,
		input.ExecutionMode, input.SortOrder, input.IsActive,
	).Scan(
		&rule.ID, &rule.ObjectID, &rule.Name, &rule.Description,
		&rule.EventType, &rule.Condition,
		&rule.ProcedureCode,
		&rule.ExecutionMode, &rule.SortOrder, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("pgAutomationRuleRepo.Update: %w", err)
	}
	return rule, nil
}

func (r *PgAutomationRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM metadata.automation_rules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("pgAutomationRuleRepo.Delete: %w", err)
	}
	return nil
}

func (r *PgAutomationRuleRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM metadata.automation_rules`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("pgAutomationRuleRepo.Count: %w", err)
	}
	return count, nil
}

func scanAutomationRules(rows pgx.Rows) ([]AutomationRule, error) {
	var rules []AutomationRule
	for rows.Next() {
		var rule AutomationRule
		if err := rows.Scan(
			&rule.ID, &rule.ObjectID, &rule.Name, &rule.Description,
			&rule.EventType, &rule.Condition,
			&rule.ProcedureCode,
			&rule.ExecutionMode, &rule.SortOrder, &rule.IsActive,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanAutomationRules: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
