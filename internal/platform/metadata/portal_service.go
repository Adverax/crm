package metadata

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
)

var validPortalAPIName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// PortalService provides business logic for object views.
type PortalService interface {
	Create(ctx context.Context, input CreatePortalInput) (*Portal, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Portal, error)
	GetByAPIName(ctx context.Context, apiName string) (*Portal, error)
	ListAll(ctx context.Context) ([]Portal, error)
	Update(ctx context.Context, id uuid.UUID, input UpdatePortalInput) (*Portal, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type portalService struct {
	pool  *pgxpool.Pool
	repo  PortalRepository
	cache *MetadataCache
}

// NewPortalService creates a new PortalService.
func NewPortalService(
	pool *pgxpool.Pool,
	repo PortalRepository,
	cache *MetadataCache,
) PortalService {
	return &portalService{
		pool:  pool,
		repo:  repo,
		cache: cache,
	}
}

func (s *portalService) Create(ctx context.Context, input CreatePortalInput) (*Portal, error) {
	if err := s.validateCreate(input); err != nil {
		return nil, fmt.Errorf("portalService.Create: %w", err)
	}

	ov, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("portalService.Create: %w", err)
	}

	if err := s.cache.LoadPortals(ctx); err != nil {
		return nil, fmt.Errorf("portalService.Create: %w", err)
	}

	return ov, nil
}

func (s *portalService) GetByID(ctx context.Context, id uuid.UUID) (*Portal, error) {
	ov, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("portalService.GetByID: %w", err)
	}
	if ov == nil {
		return nil, fmt.Errorf("portalService.GetByID: %w",
			apperror.NotFound("portal", id.String()))
	}
	return ov, nil
}

func (s *portalService) GetByAPIName(ctx context.Context, apiName string) (*Portal, error) {
	ov, err := s.repo.GetByAPIName(ctx, apiName)
	if err != nil {
		return nil, fmt.Errorf("portalService.GetByAPIName: %w", err)
	}
	if ov == nil {
		return nil, fmt.Errorf("portalService.GetByAPIName: %w",
			apperror.NotFound("portal", apiName))
	}
	return ov, nil
}

func (s *portalService) ListAll(ctx context.Context) ([]Portal, error) {
	views, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("portalService.ListAll: %w", err)
	}
	return views, nil
}

func (s *portalService) Update(ctx context.Context, id uuid.UUID, input UpdatePortalInput) (*Portal, error) {
	if err := s.validateUpdate(input); err != nil {
		return nil, fmt.Errorf("portalService.Update: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("portalService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("portalService.Update: %w",
			apperror.NotFound("portal", id.String()))
	}

	ov, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("portalService.Update: %w", err)
	}
	if ov == nil {
		return nil, fmt.Errorf("portalService.Update: %w",
			apperror.NotFound("portal", id.String()))
	}

	if err := s.cache.LoadPortals(ctx); err != nil {
		return nil, fmt.Errorf("portalService.Update: %w", err)
	}

	return ov, nil
}

func (s *portalService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("portalService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("portalService.Delete: %w",
			apperror.NotFound("portal", id.String()))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("portalService.Delete: %w", err)
	}

	if err := s.cache.LoadPortals(ctx); err != nil {
		return fmt.Errorf("portalService.Delete: %w", err)
	}

	return nil
}

func (s *portalService) validateCreate(input CreatePortalInput) error {
	if !validPortalAPIName.MatchString(input.APIName) {
		return apperror.BadRequest("api_name must match ^[a-z][a-z0-9_]*$")
	}
	if len(input.APIName) > 100 {
		return apperror.BadRequest("api_name must be at most 100 characters")
	}
	if input.Label == "" {
		return apperror.BadRequest("label is required")
	}
	if len(input.Label) > 255 {
		return apperror.BadRequest("label must be at most 255 characters")
	}
	if err := validateViewConfig(input.Config); err != nil {
		return err
	}
	return nil
}

func (s *portalService) validateUpdate(input UpdatePortalInput) error {
	if input.Label == "" {
		return apperror.BadRequest("label is required")
	}
	if len(input.Label) > 255 {
		return apperror.BadRequest("label must be at most 255 characters")
	}
	if err := validateViewConfig(input.Config); err != nil {
		return err
	}
	return nil
}
