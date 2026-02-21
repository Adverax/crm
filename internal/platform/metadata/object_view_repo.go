package metadata

import (
	"context"

	"github.com/google/uuid"
)

// ObjectViewRepository provides CRUD operations for object views.
type ObjectViewRepository interface {
	Create(ctx context.Context, input CreateObjectViewInput) (*ObjectView, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ObjectView, error)
	ListAll(ctx context.Context) ([]ObjectView, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]ObjectView, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateObjectViewInput) (*ObjectView, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindForProfile(ctx context.Context, objectID uuid.UUID, profileID uuid.UUID) (*ObjectView, error)
	FindDefault(ctx context.Context, objectID uuid.UUID) (*ObjectView, error)
}
