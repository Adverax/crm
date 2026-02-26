package metadata

import (
	"context"

	"github.com/google/uuid"
)

// SharedLayoutRepository provides CRUD operations for shared layouts.
type SharedLayoutRepository interface {
	Create(ctx context.Context, input CreateSharedLayoutInput) (*SharedLayout, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SharedLayout, error)
	GetByAPIName(ctx context.Context, apiName string) (*SharedLayout, error)
	ListAll(ctx context.Context) ([]SharedLayout, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateSharedLayoutInput) (*SharedLayout, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountReferences(ctx context.Context, apiName string) (int, error)
}
