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

	"github.com/adverax/crm/internal/pkg/apperror"
)

type assignmentRuleServiceImpl struct {
	txBeginner    TxBeginner
	ruleRepo      AssignmentRuleRepository
	territoryRepo TerritoryRepository
}

// NewAssignmentRuleService creates a new AssignmentRuleService.
func NewAssignmentRuleService(
	txBeginner TxBeginner,
	ruleRepo AssignmentRuleRepository,
	territoryRepo TerritoryRepository,
) AssignmentRuleService {
	return &assignmentRuleServiceImpl{
		txBeginner:    txBeginner,
		ruleRepo:      ruleRepo,
		territoryRepo: territoryRepo,
	}
}

func (s *assignmentRuleServiceImpl) Create(ctx context.Context, input CreateAssignmentRuleInput) (*AssignmentRule, error) {
	if err := ValidateCreateAssignmentRule(input); err != nil {
		return nil, fmt.Errorf("assignmentRuleService.Create: %w", err)
	}

	t, err := s.territoryRepo.GetByID(ctx, input.TerritoryID)
	if err != nil {
		return nil, fmt.Errorf("assignmentRuleService.Create: %w", err)
	}
	if t == nil {
		return nil, fmt.Errorf("assignmentRuleService.Create: %w",
			apperror.NotFound("Territory", input.TerritoryID.String()))
	}

	var result *AssignmentRule
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.ruleRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("assignmentRuleService.Create: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *assignmentRuleServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*AssignmentRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("assignmentRuleService.GetByID: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("assignmentRuleService.GetByID: %w",
			apperror.NotFound("AssignmentRule", id.String()))
	}
	return rule, nil
}

func (s *assignmentRuleServiceImpl) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]AssignmentRule, error) {
	rules, err := s.ruleRepo.ListByTerritoryID(ctx, territoryID)
	if err != nil {
		return nil, fmt.Errorf("assignmentRuleService.ListByTerritoryID: %w", err)
	}
	return rules, nil
}

func (s *assignmentRuleServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateAssignmentRuleInput) (*AssignmentRule, error) {
	if err := ValidateUpdateAssignmentRule(input); err != nil {
		return nil, fmt.Errorf("assignmentRuleService.Update: %w", err)
	}

	existing, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("assignmentRuleService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("assignmentRuleService.Update: %w",
			apperror.NotFound("AssignmentRule", id.String()))
	}

	var result *AssignmentRule
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.ruleRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("assignmentRuleService.Update: %w", err)
		}
		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *assignmentRuleServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("assignmentRuleService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("assignmentRuleService.Delete: %w",
			apperror.NotFound("AssignmentRule", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.ruleRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("assignmentRuleService.Delete: %w", err)
		}
		return nil
	})
}
