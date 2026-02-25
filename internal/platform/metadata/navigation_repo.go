package metadata

import (
	"context"

	"github.com/google/uuid"
)

// NavigationRepository provides CRUD operations for profile navigation configs.
type NavigationRepository interface {
	Create(ctx context.Context, input CreateProfileNavigationInput) (*ProfileNavigation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ProfileNavigation, error)
	GetByProfileID(ctx context.Context, profileID uuid.UUID) (*ProfileNavigation, error)
	ListAll(ctx context.Context) ([]ProfileNavigation, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateProfileNavigationInput) (*ProfileNavigation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
