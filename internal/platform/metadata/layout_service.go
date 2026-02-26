package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
)

var validFormFactors = map[string]bool{
	"desktop": true,
	"tablet":  true,
	"mobile":  true,
}

var validModes = map[string]bool{
	"edit": true,
	"view": true,
}

// LayoutService provides business logic for layouts.
type LayoutService interface {
	Create(ctx context.Context, input CreateLayoutInput) (*Layout, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Layout, error)
	ListByObjectViewID(ctx context.Context, ovID uuid.UUID) ([]Layout, error)
	ListAll(ctx context.Context) ([]Layout, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateLayoutInput) (*Layout, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type layoutService struct {
	pool  *pgxpool.Pool
	repo  LayoutRepository
	cache *MetadataCache
}

// NewLayoutService creates a new LayoutService.
func NewLayoutService(
	pool *pgxpool.Pool,
	repo LayoutRepository,
	cache *MetadataCache,
) LayoutService {
	return &layoutService{
		pool:  pool,
		repo:  repo,
		cache: cache,
	}
}

func (s *layoutService) Create(ctx context.Context, input CreateLayoutInput) (*Layout, error) {
	if err := s.validateCreate(ctx, input); err != nil {
		return nil, fmt.Errorf("layoutService.Create: %w", err)
	}

	layout, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("layoutService.Create: %w", err)
	}

	if err := s.cache.LoadLayouts(ctx); err != nil {
		return nil, fmt.Errorf("layoutService.Create: %w", err)
	}

	return layout, nil
}

func (s *layoutService) GetByID(ctx context.Context, id uuid.UUID) (*Layout, error) {
	layout, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("layoutService.GetByID: %w", err)
	}
	if layout == nil {
		return nil, fmt.Errorf("layoutService.GetByID: %w",
			apperror.NotFound("layout", id.String()))
	}
	return layout, nil
}

func (s *layoutService) ListByObjectViewID(ctx context.Context, ovID uuid.UUID) ([]Layout, error) {
	layouts, err := s.repo.ListByObjectViewID(ctx, ovID)
	if err != nil {
		return nil, fmt.Errorf("layoutService.ListByObjectViewID: %w", err)
	}
	return layouts, nil
}

func (s *layoutService) ListAll(ctx context.Context) ([]Layout, error) {
	layouts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("layoutService.ListAll: %w", err)
	}
	return layouts, nil
}

func (s *layoutService) Update(ctx context.Context, id uuid.UUID, input UpdateLayoutInput) (*Layout, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("layoutService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("layoutService.Update: %w",
			apperror.NotFound("layout", id.String()))
	}

	layout, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("layoutService.Update: %w", err)
	}
	if layout == nil {
		return nil, fmt.Errorf("layoutService.Update: %w",
			apperror.NotFound("layout", id.String()))
	}

	if err := s.cache.LoadLayouts(ctx); err != nil {
		return nil, fmt.Errorf("layoutService.Update: %w", err)
	}

	return layout, nil
}

func (s *layoutService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("layoutService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("layoutService.Delete: %w",
			apperror.NotFound("layout", id.String()))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("layoutService.Delete: %w", err)
	}

	if err := s.cache.LoadLayouts(ctx); err != nil {
		return fmt.Errorf("layoutService.Delete: %w", err)
	}

	return nil
}

func (s *layoutService) validateCreate(_ context.Context, input CreateLayoutInput) error {
	if !validFormFactors[input.FormFactor] {
		return apperror.BadRequest("form_factor must be one of: desktop, tablet, mobile")
	}
	if !validModes[input.Mode] {
		return apperror.BadRequest("mode must be one of: edit, view")
	}

	if input.ObjectViewID == (uuid.UUID{}) {
		return apperror.BadRequest("object_view_id is required")
	}

	return nil
}
