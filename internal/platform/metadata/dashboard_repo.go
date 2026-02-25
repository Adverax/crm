package metadata

import (
	"context"

	"github.com/google/uuid"
)

// DashboardRepository provides CRUD operations for profile dashboard configs.
type DashboardRepository interface {
	Create(ctx context.Context, input CreateProfileDashboardInput) (*ProfileDashboard, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ProfileDashboard, error)
	GetByProfileID(ctx context.Context, profileID uuid.UUID) (*ProfileDashboard, error)
	ListAll(ctx context.Context) ([]ProfileDashboard, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateProfileDashboardInput) (*ProfileDashboard, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
