package metadata

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
)

var validSharedLayoutAPIName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

var validSharedLayoutTypes = map[string]bool{
	"field":   true,
	"section": true,
	"list":    true,
}

// SharedLayoutService provides business logic for shared layouts.
type SharedLayoutService interface {
	Create(ctx context.Context, input CreateSharedLayoutInput) (*SharedLayout, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SharedLayout, error)
	GetByAPIName(ctx context.Context, apiName string) (*SharedLayout, error)
	ListAll(ctx context.Context) ([]SharedLayout, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateSharedLayoutInput) (*SharedLayout, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type sharedLayoutService struct {
	pool  *pgxpool.Pool
	repo  SharedLayoutRepository
	cache *MetadataCache
}

// NewSharedLayoutService creates a new SharedLayoutService.
func NewSharedLayoutService(
	pool *pgxpool.Pool,
	repo SharedLayoutRepository,
	cache *MetadataCache,
) SharedLayoutService {
	return &sharedLayoutService{
		pool:  pool,
		repo:  repo,
		cache: cache,
	}
}

func (s *sharedLayoutService) Create(ctx context.Context, input CreateSharedLayoutInput) (*SharedLayout, error) {
	if err := s.validateCreate(input); err != nil {
		return nil, fmt.Errorf("sharedLayoutService.Create: %w", err)
	}

	sl, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("sharedLayoutService.Create: %w", err)
	}

	if err := s.cache.LoadSharedLayouts(ctx); err != nil {
		return nil, fmt.Errorf("sharedLayoutService.Create: %w", err)
	}

	return sl, nil
}

func (s *sharedLayoutService) GetByID(ctx context.Context, id uuid.UUID) (*SharedLayout, error) {
	sl, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sharedLayoutService.GetByID: %w", err)
	}
	if sl == nil {
		return nil, fmt.Errorf("sharedLayoutService.GetByID: %w",
			apperror.NotFound("shared_layout", id.String()))
	}
	return sl, nil
}

func (s *sharedLayoutService) GetByAPIName(ctx context.Context, apiName string) (*SharedLayout, error) {
	sl, err := s.repo.GetByAPIName(ctx, apiName)
	if err != nil {
		return nil, fmt.Errorf("sharedLayoutService.GetByAPIName: %w", err)
	}
	if sl == nil {
		return nil, fmt.Errorf("sharedLayoutService.GetByAPIName: %w",
			apperror.NotFound("shared_layout", apiName))
	}
	return sl, nil
}

func (s *sharedLayoutService) ListAll(ctx context.Context) ([]SharedLayout, error) {
	layouts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("sharedLayoutService.ListAll: %w", err)
	}
	return layouts, nil
}

func (s *sharedLayoutService) Update(ctx context.Context, id uuid.UUID, input UpdateSharedLayoutInput) (*SharedLayout, error) {
	if err := s.validateUpdate(input); err != nil {
		return nil, fmt.Errorf("sharedLayoutService.Update: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sharedLayoutService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("sharedLayoutService.Update: %w",
			apperror.NotFound("shared_layout", id.String()))
	}

	sl, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("sharedLayoutService.Update: %w", err)
	}
	if sl == nil {
		return nil, fmt.Errorf("sharedLayoutService.Update: %w",
			apperror.NotFound("shared_layout", id.String()))
	}

	if err := s.cache.LoadSharedLayouts(ctx); err != nil {
		return nil, fmt.Errorf("sharedLayoutService.Update: %w", err)
	}

	return sl, nil
}

func (s *sharedLayoutService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("sharedLayoutService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("sharedLayoutService.Delete: %w",
			apperror.NotFound("shared_layout", id.String()))
	}

	// RESTRICT: check references before deleting
	count, err := s.repo.CountReferences(ctx, existing.APIName)
	if err != nil {
		return fmt.Errorf("sharedLayoutService.Delete: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("sharedLayoutService.Delete: %w",
			apperror.Conflict(fmt.Sprintf("shared layout is referenced by %d layout(s)", count)))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("sharedLayoutService.Delete: %w", err)
	}

	if err := s.cache.LoadSharedLayouts(ctx); err != nil {
		return fmt.Errorf("sharedLayoutService.Delete: %w", err)
	}

	return nil
}

func (s *sharedLayoutService) validateCreate(input CreateSharedLayoutInput) error {
	if !validSharedLayoutAPIName.MatchString(input.APIName) {
		return apperror.BadRequest("api_name must match ^[a-z][a-z0-9_]*$")
	}
	if len(input.APIName) > 63 {
		return apperror.BadRequest("api_name must be at most 63 characters")
	}
	if !validSharedLayoutTypes[input.Type] {
		return apperror.BadRequest("type must be one of: field, section, list")
	}
	if len(input.Label) > 255 {
		return apperror.BadRequest("label must be at most 255 characters")
	}
	return nil
}

func (s *sharedLayoutService) validateUpdate(input UpdateSharedLayoutInput) error {
	if len(input.Label) > 255 {
		return apperror.BadRequest("label must be at most 255 characters")
	}
	return nil
}
