package metadata

import (
	"context"

	"github.com/google/uuid"
)

// ObjectService defines the business logic interface for object definitions.
type ObjectService interface {
	Create(ctx context.Context, input CreateObjectInput) (*ObjectDefinition, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ObjectDefinition, error)
	List(ctx context.Context, filter ObjectFilter) ([]ObjectDefinition, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateObjectInput) (*ObjectDefinition, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// FieldService defines the business logic interface for field definitions.
type FieldService interface {
	Create(ctx context.Context, input CreateFieldInput) (*FieldDefinition, error)
	GetByID(ctx context.Context, id uuid.UUID) (*FieldDefinition, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]FieldDefinition, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateFieldInput) (*FieldDefinition, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// CacheInvalidator invalidates metadata cache.
type CacheInvalidator interface {
	Invalidate(ctx context.Context) error
}
