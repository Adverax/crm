//go:build !enterprise

package rls

import (
	"context"

	"github.com/google/uuid"
)

// noopTerritoryResolver is the community edition stub.
// Returns nil — no territory-based access in community edition.
type noopTerritoryResolver struct{}

// NewTerritoryResolver returns a noop resolver for community edition.
func NewTerritoryResolver() TerritoryResolver {
	return &noopTerritoryResolver{}
}

func (r *noopTerritoryResolver) ResolveTerritoryGroups(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
