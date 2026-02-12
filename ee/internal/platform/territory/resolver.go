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

	"github.com/adverax/crm/internal/platform/security/rls"
)

// enterpriseTerritoryResolver implements rls.TerritoryResolver for enterprise edition.
// It resolves territory group IDs for a user by looking up the effective_user_territory
// cache and then finding the corresponding territory groups.
type enterpriseTerritoryResolver struct {
	effectiveRepo EffectiveRepository
}

// NewTerritoryResolver creates an enterprise TerritoryResolver.
func NewTerritoryResolver(effectiveRepo EffectiveRepository) rls.TerritoryResolver {
	return &enterpriseTerritoryResolver{effectiveRepo: effectiveRepo}
}

func (r *enterpriseTerritoryResolver) ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	territoryIDs, err := r.effectiveRepo.GetUserTerritories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("territoryResolver.ResolveTerritoryGroups: %w", err)
	}

	if len(territoryIDs) == 0 {
		return nil, nil
	}

	groupIDs, err := r.effectiveRepo.GetTerritoryGroupIDs(ctx, territoryIDs)
	if err != nil {
		return nil, fmt.Errorf("territoryResolver.ResolveTerritoryGroups: %w", err)
	}

	return groupIDs, nil
}
