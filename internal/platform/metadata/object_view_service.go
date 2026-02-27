package metadata

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
)

var validOVAPIName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ObjectViewService provides business logic for object views.
type ObjectViewService interface {
	Create(ctx context.Context, input CreateObjectViewInput) (*ObjectView, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ObjectView, error)
	GetByAPIName(ctx context.Context, apiName string) (*ObjectView, error)
	ListAll(ctx context.Context) ([]ObjectView, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateObjectViewInput) (*ObjectView, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type objectViewService struct {
	pool  *pgxpool.Pool
	repo  ObjectViewRepository
	cache *MetadataCache
}

// NewObjectViewService creates a new ObjectViewService.
func NewObjectViewService(
	pool *pgxpool.Pool,
	repo ObjectViewRepository,
	cache *MetadataCache,
) ObjectViewService {
	return &objectViewService{
		pool:  pool,
		repo:  repo,
		cache: cache,
	}
}

func (s *objectViewService) Create(ctx context.Context, input CreateObjectViewInput) (*ObjectView, error) {
	if err := s.validateCreate(input); err != nil {
		return nil, fmt.Errorf("objectViewService.Create: %w", err)
	}

	ov, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("objectViewService.Create: %w", err)
	}

	if err := s.cache.LoadObjectViews(ctx); err != nil {
		return nil, fmt.Errorf("objectViewService.Create: %w", err)
	}

	return ov, nil
}

func (s *objectViewService) GetByID(ctx context.Context, id uuid.UUID) (*ObjectView, error) {
	ov, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("objectViewService.GetByID: %w", err)
	}
	if ov == nil {
		return nil, fmt.Errorf("objectViewService.GetByID: %w",
			apperror.NotFound("object_view", id.String()))
	}
	return ov, nil
}

func (s *objectViewService) GetByAPIName(ctx context.Context, apiName string) (*ObjectView, error) {
	ov, err := s.repo.GetByAPIName(ctx, apiName)
	if err != nil {
		return nil, fmt.Errorf("objectViewService.GetByAPIName: %w", err)
	}
	if ov == nil {
		return nil, fmt.Errorf("objectViewService.GetByAPIName: %w",
			apperror.NotFound("object_view", apiName))
	}
	return ov, nil
}

func (s *objectViewService) ListAll(ctx context.Context) ([]ObjectView, error) {
	views, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("objectViewService.ListAll: %w", err)
	}
	return views, nil
}

func (s *objectViewService) Update(ctx context.Context, id uuid.UUID, input UpdateObjectViewInput) (*ObjectView, error) {
	if err := s.validateUpdate(input); err != nil {
		return nil, fmt.Errorf("objectViewService.Update: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("objectViewService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("objectViewService.Update: %w",
			apperror.NotFound("object_view", id.String()))
	}

	ov, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("objectViewService.Update: %w", err)
	}
	if ov == nil {
		return nil, fmt.Errorf("objectViewService.Update: %w",
			apperror.NotFound("object_view", id.String()))
	}

	if err := s.cache.LoadObjectViews(ctx); err != nil {
		return nil, fmt.Errorf("objectViewService.Update: %w", err)
	}

	return ov, nil
}

func (s *objectViewService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("objectViewService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("objectViewService.Delete: %w",
			apperror.NotFound("object_view", id.String()))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("objectViewService.Delete: %w", err)
	}

	if err := s.cache.LoadObjectViews(ctx); err != nil {
		return fmt.Errorf("objectViewService.Delete: %w", err)
	}

	return nil
}

func (s *objectViewService) validateCreate(input CreateObjectViewInput) error {
	if !validOVAPIName.MatchString(input.APIName) {
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

func (s *objectViewService) validateUpdate(input UpdateObjectViewInput) error {
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
