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

type userAssignmentServiceImpl struct {
	txBeginner     TxBeginner
	assignmentRepo UserAssignmentRepository
	territoryRepo  TerritoryRepository
}

// NewUserAssignmentService creates a new UserAssignmentService.
func NewUserAssignmentService(
	txBeginner TxBeginner,
	assignmentRepo UserAssignmentRepository,
	territoryRepo TerritoryRepository,
) UserAssignmentService {
	return &userAssignmentServiceImpl{
		txBeginner:     txBeginner,
		assignmentRepo: assignmentRepo,
		territoryRepo:  territoryRepo,
	}
}

func (s *userAssignmentServiceImpl) Assign(ctx context.Context, input AssignUserInput) (*UserTerritoryAssignment, error) {
	t, err := s.territoryRepo.GetByID(ctx, input.TerritoryID)
	if err != nil {
		return nil, fmt.Errorf("userAssignmentService.Assign: %w", err)
	}
	if t == nil {
		return nil, fmt.Errorf("userAssignmentService.Assign: %w",
			apperror.NotFound("Territory", input.TerritoryID.String()))
	}

	existing, _ := s.assignmentRepo.GetByUserAndTerritory(ctx, input.UserID, input.TerritoryID)
	if existing != nil {
		return nil, fmt.Errorf("userAssignmentService.Assign: %w",
			apperror.Conflict("user is already assigned to this territory"))
	}

	var result *UserTerritoryAssignment
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.assignmentRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("userAssignmentService.Assign: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *userAssignmentServiceImpl) Unassign(ctx context.Context, userID, territoryID uuid.UUID) error {
	existing, err := s.assignmentRepo.GetByUserAndTerritory(ctx, userID, territoryID)
	if err != nil {
		return fmt.Errorf("userAssignmentService.Unassign: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("userAssignmentService.Unassign: %w",
			apperror.NotFound("UserTerritoryAssignment", fmt.Sprintf("user=%s territory=%s", userID, territoryID)))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.assignmentRepo.Delete(ctx, tx, userID, territoryID); err != nil {
			return fmt.Errorf("userAssignmentService.Unassign: %w", err)
		}
		return nil
	})
}

func (s *userAssignmentServiceImpl) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]UserTerritoryAssignment, error) {
	assignments, err := s.assignmentRepo.ListByTerritoryID(ctx, territoryID)
	if err != nil {
		return nil, fmt.Errorf("userAssignmentService.ListByTerritoryID: %w", err)
	}
	return assignments, nil
}

func (s *userAssignmentServiceImpl) ListByUserID(ctx context.Context, userID uuid.UUID) ([]UserTerritoryAssignment, error) {
	assignments, err := s.assignmentRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("userAssignmentService.ListByUserID: %w", err)
	}
	return assignments, nil
}
