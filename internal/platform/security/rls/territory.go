package rls

import (
	"context"

	"github.com/google/uuid"
)

// TerritoryResolver resolves territory-based group memberships for a user.
// Enterprise edition provides the full implementation; community edition uses a noop.
type TerritoryResolver interface {
	ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
