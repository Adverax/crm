package metadata

import (
	"context"

	"github.com/google/uuid"
)

// LayoutRepository provides CRUD operations for layouts.
type LayoutRepository interface {
	Create(ctx context.Context, input CreateLayoutInput) (*Layout, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Layout, error)
	ListByObjectViewID(ctx context.Context, ovID uuid.UUID) ([]Layout, error)
	ListAll(ctx context.Context) ([]Layout, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateLayoutInput) (*Layout, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
