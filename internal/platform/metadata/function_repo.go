package metadata

import (
	"context"

	"github.com/google/uuid"
)

// FunctionRepository provides CRUD operations for custom functions.
type FunctionRepository interface {
	Create(ctx context.Context, input CreateFunctionInput) (*Function, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Function, error)
	GetByName(ctx context.Context, name string) (*Function, error)
	ListAll(ctx context.Context) ([]Function, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateFunctionInput) (*Function, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
}
