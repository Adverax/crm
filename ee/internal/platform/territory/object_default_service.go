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

type objectDefaultServiceImpl struct {
	txBeginner     TxBeginner
	objDefaultRepo ObjectDefaultRepository
	territoryRepo  TerritoryRepository
}

// NewObjectDefaultService creates a new ObjectDefaultService.
func NewObjectDefaultService(
	txBeginner TxBeginner,
	objDefaultRepo ObjectDefaultRepository,
	territoryRepo TerritoryRepository,
) ObjectDefaultService {
	return &objectDefaultServiceImpl{
		txBeginner:     txBeginner,
		objDefaultRepo: objDefaultRepo,
		territoryRepo:  territoryRepo,
	}
}

func (s *objectDefaultServiceImpl) Set(ctx context.Context, input SetObjectDefaultInput) (*TerritoryObjectDefault, error) {
	if err := ValidateSetObjectDefault(input); err != nil {
		return nil, fmt.Errorf("objectDefaultService.Set: %w", err)
	}

	t, err := s.territoryRepo.GetByID(ctx, input.TerritoryID)
	if err != nil {
		return nil, fmt.Errorf("objectDefaultService.Set: %w", err)
	}
	if t == nil {
		return nil, fmt.Errorf("objectDefaultService.Set: %w",
			apperror.NotFound("Territory", input.TerritoryID.String()))
	}

	var result *TerritoryObjectDefault
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		upserted, err := s.objDefaultRepo.Upsert(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("objectDefaultService.Set: %w", err)
		}
		result = upserted
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *objectDefaultServiceImpl) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]TerritoryObjectDefault, error) {
	defaults, err := s.objDefaultRepo.ListByTerritoryID(ctx, territoryID)
	if err != nil {
		return nil, fmt.Errorf("objectDefaultService.ListByTerritoryID: %w", err)
	}
	return defaults, nil
}

func (s *objectDefaultServiceImpl) Remove(ctx context.Context, territoryID, objectID uuid.UUID) error {
	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.objDefaultRepo.Delete(ctx, tx, territoryID, objectID); err != nil {
			return fmt.Errorf("objectDefaultService.Remove: %w", err)
		}
		return nil
	})
}
