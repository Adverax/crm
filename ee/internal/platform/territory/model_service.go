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

type modelServiceImpl struct {
	txBeginner    TxBeginner
	modelRepo     ModelRepository
	effectiveRepo EffectiveRepository
}

// NewModelService creates a new ModelService.
func NewModelService(
	txBeginner TxBeginner,
	modelRepo ModelRepository,
	effectiveRepo EffectiveRepository,
) ModelService {
	return &modelServiceImpl{
		txBeginner:    txBeginner,
		modelRepo:     modelRepo,
		effectiveRepo: effectiveRepo,
	}
}

func (s *modelServiceImpl) Create(ctx context.Context, input CreateModelInput) (*TerritoryModel, error) {
	if err := ValidateCreateModel(input); err != nil {
		return nil, fmt.Errorf("modelService.Create: %w", err)
	}

	existing, _ := s.modelRepo.GetByAPIName(ctx, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("modelService.Create: %w",
			apperror.Conflict(fmt.Sprintf("territory model with api_name '%s' already exists", input.APIName)))
	}

	var result *TerritoryModel
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.modelRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("modelService.Create: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *modelServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*TerritoryModel, error) {
	model, err := s.modelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("modelService.GetByID: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("modelService.GetByID: %w",
			apperror.NotFound("TerritoryModel", id.String()))
	}
	return model, nil
}

func (s *modelServiceImpl) GetActive(ctx context.Context) (*TerritoryModel, error) {
	model, err := s.modelRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("modelService.GetActive: %w", err)
	}
	return model, nil
}

func (s *modelServiceImpl) List(ctx context.Context, page, perPage int32) ([]TerritoryModel, int64, error) {
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	models, err := s.modelRepo.List(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("modelService.List: %w", err)
	}

	total, err := s.modelRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("modelService.List: count: %w", err)
	}

	return models, total, nil
}

func (s *modelServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateModelInput) (*TerritoryModel, error) {
	if err := ValidateUpdateModel(input); err != nil {
		return nil, fmt.Errorf("modelService.Update: %w", err)
	}

	existing, err := s.modelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("modelService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("modelService.Update: %w",
			apperror.NotFound("TerritoryModel", id.String()))
	}

	var result *TerritoryModel
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.modelRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("modelService.Update: %w", err)
		}
		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *modelServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.modelRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("modelService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("modelService.Delete: %w",
			apperror.NotFound("TerritoryModel", id.String()))
	}
	if existing.Status == ModelStatusActive {
		return fmt.Errorf("modelService.Delete: %w",
			apperror.Validation("cannot delete an active territory model; archive it first"))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.modelRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("modelService.Delete: %w", err)
		}
		return nil
	})
}

func (s *modelServiceImpl) Activate(ctx context.Context, id uuid.UUID) error {
	existing, err := s.modelRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("modelService.Activate: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("modelService.Activate: %w",
			apperror.NotFound("TerritoryModel", id.String()))
	}
	if existing.Status != ModelStatusPlanning {
		return fmt.Errorf("modelService.Activate: %w",
			apperror.Validation(fmt.Sprintf("only planning models can be activated; current status is '%s'", existing.Status)))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.effectiveRepo.ActivateModel(ctx, tx, id); err != nil {
			return fmt.Errorf("modelService.Activate: %w", err)
		}
		return nil
	})
}

func (s *modelServiceImpl) Archive(ctx context.Context, id uuid.UUID) error {
	existing, err := s.modelRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("modelService.Archive: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("modelService.Archive: %w",
			apperror.NotFound("TerritoryModel", id.String()))
	}
	if existing.Status != ModelStatusActive {
		return fmt.Errorf("modelService.Archive: %w",
			apperror.Validation(fmt.Sprintf("only active models can be archived; current status is '%s'", existing.Status)))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.modelRepo.UpdateStatus(ctx, tx, id, ModelStatusArchived); err != nil {
			return fmt.Errorf("modelService.Archive: %w", err)
		}
		return nil
	})
}
