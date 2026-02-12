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

type recordAssignmentServiceImpl struct {
	txBeginner     TxBeginner
	assignmentRepo RecordAssignmentRepository
	territoryRepo  TerritoryRepository
	effectiveRepo  EffectiveRepository
	objLookup      ObjectDefinitionLookup
}

// NewRecordAssignmentService creates a new RecordAssignmentService.
func NewRecordAssignmentService(
	txBeginner TxBeginner,
	assignmentRepo RecordAssignmentRepository,
	territoryRepo TerritoryRepository,
	effectiveRepo EffectiveRepository,
	objLookup ObjectDefinitionLookup,
) RecordAssignmentService {
	return &recordAssignmentServiceImpl{
		txBeginner:     txBeginner,
		assignmentRepo: assignmentRepo,
		territoryRepo:  territoryRepo,
		effectiveRepo:  effectiveRepo,
		objLookup:      objLookup,
	}
}

func (s *recordAssignmentServiceImpl) Assign(ctx context.Context, input AssignRecordInput) (*RecordTerritoryAssignment, error) {
	if err := ValidateAssignRecord(input); err != nil {
		return nil, fmt.Errorf("recordAssignmentService.Assign: %w", err)
	}

	if input.Reason == "" {
		input.Reason = "manual"
	}

	t, err := s.territoryRepo.GetByID(ctx, input.TerritoryID)
	if err != nil {
		return nil, fmt.Errorf("recordAssignmentService.Assign: %w", err)
	}
	if t == nil {
		return nil, fmt.Errorf("recordAssignmentService.Assign: %w",
			apperror.NotFound("Territory", input.TerritoryID.String()))
	}

	existing, _ := s.assignmentRepo.GetByRecordAndTerritory(ctx, input.RecordID, input.ObjectID, input.TerritoryID)
	if existing != nil {
		return nil, fmt.Errorf("recordAssignmentService.Assign: %w",
			apperror.Conflict("record is already assigned to this territory"))
	}

	tableName, err := s.objLookup.GetTableName(ctx, input.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("recordAssignmentService.Assign: %w", err)
	}
	shareTable := tableName + "__share"

	var result *RecordTerritoryAssignment
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.assignmentRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("recordAssignmentService.Assign: %w", err)
		}

		if err := s.effectiveRepo.GenerateRecordShareEntries(ctx, tx, input.RecordID, input.ObjectID, input.TerritoryID, shareTable); err != nil {
			return fmt.Errorf("recordAssignmentService.Assign: generate shares: %w", err)
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *recordAssignmentServiceImpl) Unassign(ctx context.Context, recordID, objectID, territoryID uuid.UUID) error {
	existing, err := s.assignmentRepo.GetByRecordAndTerritory(ctx, recordID, objectID, territoryID)
	if err != nil {
		return fmt.Errorf("recordAssignmentService.Unassign: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("recordAssignmentService.Unassign: %w",
			apperror.NotFound("RecordTerritoryAssignment",
				fmt.Sprintf("record=%s object=%s territory=%s", recordID, objectID, territoryID)))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.assignmentRepo.Delete(ctx, tx, recordID, objectID, territoryID); err != nil {
			return fmt.Errorf("recordAssignmentService.Unassign: %w", err)
		}
		return nil
	})
}

func (s *recordAssignmentServiceImpl) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]RecordTerritoryAssignment, error) {
	assignments, err := s.assignmentRepo.ListByTerritoryID(ctx, territoryID)
	if err != nil {
		return nil, fmt.Errorf("recordAssignmentService.ListByTerritoryID: %w", err)
	}
	return assignments, nil
}

func (s *recordAssignmentServiceImpl) ListByRecordID(ctx context.Context, recordID, objectID uuid.UUID) ([]RecordTerritoryAssignment, error) {
	assignments, err := s.assignmentRepo.ListByRecordID(ctx, recordID, objectID)
	if err != nil {
		return nil, fmt.Errorf("recordAssignmentService.ListByRecordID: %w", err)
	}
	return assignments, nil
}
