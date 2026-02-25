package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

const (
	maxDashboardWidgets = 12
)

// ProfileDashboardService provides business logic for profile dashboard configs.
type ProfileDashboardService interface {
	Create(ctx context.Context, input CreateProfileDashboardInput) (*ProfileDashboard, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ProfileDashboard, error)
	ListAll(ctx context.Context) ([]ProfileDashboard, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateProfileDashboardInput) (*ProfileDashboard, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ResolveForProfile(ctx context.Context, profileID uuid.UUID) (*ProfileDashboard, error)
}

type profileDashboardService struct {
	repo DashboardRepository
}

// NewProfileDashboardService creates a new ProfileDashboardService.
func NewProfileDashboardService(repo DashboardRepository) ProfileDashboardService {
	return &profileDashboardService{repo: repo}
}

func (s *profileDashboardService) Create(ctx context.Context, input CreateProfileDashboardInput) (*ProfileDashboard, error) {
	if input.ProfileID == uuid.Nil {
		return nil, fmt.Errorf("dashService.Create: %w",
			apperror.BadRequest("profile_id is required"))
	}

	if err := validateDashboardConfig(input.Config); err != nil {
		return nil, fmt.Errorf("dashService.Create: %w", err)
	}

	dash, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("dashService.Create: %w", err)
	}
	return dash, nil
}

func (s *profileDashboardService) GetByID(ctx context.Context, id uuid.UUID) (*ProfileDashboard, error) {
	dash, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("dashService.GetByID: %w", err)
	}
	if dash == nil {
		return nil, fmt.Errorf("dashService.GetByID: %w",
			apperror.NotFound("profile_dashboard", id.String()))
	}
	return dash, nil
}

func (s *profileDashboardService) ListAll(ctx context.Context) ([]ProfileDashboard, error) {
	dashes, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("dashService.ListAll: %w", err)
	}
	return dashes, nil
}

func (s *profileDashboardService) Update(ctx context.Context, id uuid.UUID, input UpdateProfileDashboardInput) (*ProfileDashboard, error) {
	if err := validateDashboardConfig(input.Config); err != nil {
		return nil, fmt.Errorf("dashService.Update: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("dashService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("dashService.Update: %w",
			apperror.NotFound("profile_dashboard", id.String()))
	}

	dash, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("dashService.Update: %w", err)
	}
	if dash == nil {
		return nil, fmt.Errorf("dashService.Update: %w",
			apperror.NotFound("profile_dashboard", id.String()))
	}
	return dash, nil
}

func (s *profileDashboardService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("dashService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashService.Delete: %w",
			apperror.NotFound("profile_dashboard", id.String()))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("dashService.Delete: %w", err)
	}
	return nil
}

// ResolveForProfile returns the dashboard config for a given profile, or nil if none exists.
func (s *profileDashboardService) ResolveForProfile(ctx context.Context, profileID uuid.UUID) (*ProfileDashboard, error) {
	dash, err := s.repo.GetByProfileID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("dashService.ResolveForProfile: %w", err)
	}
	return dash, nil
}

func validateDashboardConfig(cfg DashboardConfig) error {
	if len(cfg.Widgets) > maxDashboardWidgets {
		return apperror.BadRequest(fmt.Sprintf("maximum %d widgets allowed", maxDashboardWidgets))
	}

	keys := make(map[string]bool, len(cfg.Widgets))
	for i, w := range cfg.Widgets {
		if w.Key == "" {
			return apperror.BadRequest(fmt.Sprintf("widget %d: key is required", i))
		}
		if keys[w.Key] {
			return apperror.BadRequest(fmt.Sprintf("duplicate widget key: %s", w.Key))
		}
		keys[w.Key] = true

		if w.Label == "" {
			return apperror.BadRequest(fmt.Sprintf("widget %q: label is required", w.Key))
		}

		if err := validateWidgetByType(w); err != nil {
			return err
		}
	}

	return nil
}

func validateWidgetByType(w DashboardWidget) error {
	switch w.Type {
	case "list":
		if w.Query == "" {
			return apperror.BadRequest(fmt.Sprintf("widget %q: query is required for type 'list'", w.Key))
		}
		if len(w.Columns) == 0 {
			return apperror.BadRequest(fmt.Sprintf("widget %q: columns are required for type 'list'", w.Key))
		}
		if w.ObjectAPIName == "" {
			return apperror.BadRequest(fmt.Sprintf("widget %q: object_api_name is required for type 'list'", w.Key))
		}
	case "metric":
		if w.Query == "" {
			return apperror.BadRequest(fmt.Sprintf("widget %q: query is required for type 'metric'", w.Key))
		}
		if w.Format != "" && w.Format != "number" && w.Format != "currency" && w.Format != "percent" {
			return apperror.BadRequest(fmt.Sprintf("widget %q: format must be 'number', 'currency', or 'percent'", w.Key))
		}
	case "link_list":
		if len(w.Links) == 0 {
			return apperror.BadRequest(fmt.Sprintf("widget %q: links are required for type 'link_list'", w.Key))
		}
		for j, link := range w.Links {
			if link.Label == "" {
				return apperror.BadRequest(fmt.Sprintf("widget %q link %d: label is required", w.Key, j))
			}
			if link.URL == "" {
				return apperror.BadRequest(fmt.Sprintf("widget %q link %d: url is required", w.Key, j))
			}
		}
	default:
		return apperror.BadRequest(fmt.Sprintf("widget %q: type must be 'list', 'metric', or 'link_list'", w.Key))
	}

	if w.Size != "" && w.Size != "full" && w.Size != "half" && w.Size != "third" {
		return apperror.BadRequest(fmt.Sprintf("widget %q: size must be 'full', 'half', or 'third'", w.Key))
	}

	return nil
}
