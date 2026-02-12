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

type territoryServiceImpl struct {
	txBeginner    TxBeginner
	territoryRepo TerritoryRepository
	modelRepo     ModelRepository
}

// NewTerritoryService creates a new TerritoryService.
func NewTerritoryService(
	txBeginner TxBeginner,
	territoryRepo TerritoryRepository,
	modelRepo ModelRepository,
) TerritoryService {
	return &territoryServiceImpl{
		txBeginner:    txBeginner,
		territoryRepo: territoryRepo,
		modelRepo:     modelRepo,
	}
}

func (s *territoryServiceImpl) Create(ctx context.Context, input CreateTerritoryInput) (*Territory, error) {
	if err := ValidateCreateTerritory(input); err != nil {
		return nil, fmt.Errorf("territoryService.Create: %w", err)
	}

	model, err := s.modelRepo.GetByID(ctx, input.ModelID)
	if err != nil {
		return nil, fmt.Errorf("territoryService.Create: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("territoryService.Create: %w",
			apperror.NotFound("TerritoryModel", input.ModelID.String()))
	}

	existing, _ := s.territoryRepo.GetByAPIName(ctx, input.ModelID, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("territoryService.Create: %w",
			apperror.Conflict(fmt.Sprintf("territory with api_name '%s' already exists in this model", input.APIName)))
	}

	if input.ParentID != nil {
		parent, err := s.territoryRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("territoryService.Create: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("territoryService.Create: %w",
				apperror.NotFound("Territory (parent)", input.ParentID.String()))
		}
		if parent.ModelID != input.ModelID {
			return nil, fmt.Errorf("territoryService.Create: %w",
				apperror.Validation("parent territory must belong to the same model"))
		}
	}

	var result *Territory
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.territoryRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("territoryService.Create: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *territoryServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*Territory, error) {
	t, err := s.territoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("territoryService.GetByID: %w", err)
	}
	if t == nil {
		return nil, fmt.Errorf("territoryService.GetByID: %w",
			apperror.NotFound("Territory", id.String()))
	}
	return t, nil
}

func (s *territoryServiceImpl) ListByModelID(ctx context.Context, modelID uuid.UUID) ([]Territory, error) {
	territories, err := s.territoryRepo.ListByModelID(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("territoryService.ListByModelID: %w", err)
	}
	return territories, nil
}

func (s *territoryServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateTerritoryInput) (*Territory, error) {
	if err := ValidateUpdateTerritory(input); err != nil {
		return nil, fmt.Errorf("territoryService.Update: %w", err)
	}

	existing, err := s.territoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("territoryService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("territoryService.Update: %w",
			apperror.NotFound("Territory", id.String()))
	}

	if input.ParentID != nil {
		if *input.ParentID == id {
			return nil, fmt.Errorf("territoryService.Update: %w",
				apperror.Validation("territory cannot be its own parent"))
		}
		parent, err := s.territoryRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("territoryService.Update: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("territoryService.Update: %w",
				apperror.NotFound("Territory (parent)", input.ParentID.String()))
		}
		if parent.ModelID != existing.ModelID {
			return nil, fmt.Errorf("territoryService.Update: %w",
				apperror.Validation("parent territory must belong to the same model"))
		}
	}

	var result *Territory
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.territoryRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("territoryService.Update: %w", err)
		}
		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *territoryServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.territoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("territoryService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("territoryService.Delete: %w",
			apperror.NotFound("Territory", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.territoryRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("territoryService.Delete: %w", err)
		}
		return nil
	})
}
