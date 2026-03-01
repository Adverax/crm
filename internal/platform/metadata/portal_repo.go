package metadata

import (
	"context"

	"github.com/google/uuid"
)

// PortalRepository provides CRUD operations for object views.
type PortalRepository interface {
	Create(ctx context.Context, input CreatePortalInput) (*Portal, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Portal, error)
	GetByAPIName(ctx context.Context, apiName string) (*Portal, error)
	ListAll(ctx context.Context) ([]Portal, error)
	Update(ctx context.Context, id uuid.UUID, input UpdatePortalInput) (*Portal, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
