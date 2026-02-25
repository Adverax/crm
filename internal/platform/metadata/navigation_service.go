package metadata

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

const (
	maxNavGroups        = 20
	maxNavItemsPerGroup = 50
)

// ProfileNavigationService provides business logic for profile navigation configs.
type ProfileNavigationService interface {
	Create(ctx context.Context, input CreateProfileNavigationInput) (*ProfileNavigation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ProfileNavigation, error)
	ListAll(ctx context.Context) ([]ProfileNavigation, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateProfileNavigationInput) (*ProfileNavigation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ResolveForProfile(ctx context.Context, profileID uuid.UUID) (*ProfileNavigation, error)
}

type profileNavigationService struct {
	repo NavigationRepository
}

// NewProfileNavigationService creates a new ProfileNavigationService.
func NewProfileNavigationService(repo NavigationRepository) ProfileNavigationService {
	return &profileNavigationService{repo: repo}
}

func (s *profileNavigationService) Create(ctx context.Context, input CreateProfileNavigationInput) (*ProfileNavigation, error) {
	if input.ProfileID == uuid.Nil {
		return nil, fmt.Errorf("navService.Create: %w",
			apperror.BadRequest("profile_id is required"))
	}

	if err := validateNavConfig(input.Config); err != nil {
		return nil, fmt.Errorf("navService.Create: %w", err)
	}

	nav, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("navService.Create: %w", err)
	}
	return nav, nil
}

func (s *profileNavigationService) GetByID(ctx context.Context, id uuid.UUID) (*ProfileNavigation, error) {
	nav, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("navService.GetByID: %w", err)
	}
	if nav == nil {
		return nil, fmt.Errorf("navService.GetByID: %w",
			apperror.NotFound("profile_navigation", id.String()))
	}
	return nav, nil
}

func (s *profileNavigationService) ListAll(ctx context.Context) ([]ProfileNavigation, error) {
	navs, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("navService.ListAll: %w", err)
	}
	return navs, nil
}

func (s *profileNavigationService) Update(ctx context.Context, id uuid.UUID, input UpdateProfileNavigationInput) (*ProfileNavigation, error) {
	if err := validateNavConfig(input.Config); err != nil {
		return nil, fmt.Errorf("navService.Update: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("navService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("navService.Update: %w",
			apperror.NotFound("profile_navigation", id.String()))
	}

	nav, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("navService.Update: %w", err)
	}
	if nav == nil {
		return nil, fmt.Errorf("navService.Update: %w",
			apperror.NotFound("profile_navigation", id.String()))
	}
	return nav, nil
}

func (s *profileNavigationService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("navService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("navService.Delete: %w",
			apperror.NotFound("profile_navigation", id.String()))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("navService.Delete: %w", err)
	}
	return nil
}

// ResolveForProfile returns the navigation config for a given profile, or nil if none exists.
func (s *profileNavigationService) ResolveForProfile(ctx context.Context, profileID uuid.UUID) (*ProfileNavigation, error) {
	nav, err := s.repo.GetByProfileID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("navService.ResolveForProfile: %w", err)
	}
	return nav, nil
}

func validateNavConfig(cfg NavConfig) error {
	if len(cfg.Groups) > maxNavGroups {
		return apperror.BadRequest(fmt.Sprintf("maximum %d groups allowed", maxNavGroups))
	}

	keys := make(map[string]bool, len(cfg.Groups))
	for _, g := range cfg.Groups {
		if g.Key == "" {
			return apperror.BadRequest("group key is required")
		}
		if keys[g.Key] {
			return apperror.BadRequest(fmt.Sprintf("duplicate group key: %s", g.Key))
		}
		keys[g.Key] = true

		if g.Label == "" {
			return apperror.BadRequest(fmt.Sprintf("group %q: label is required", g.Key))
		}

		if len(g.Items) > maxNavItemsPerGroup {
			return apperror.BadRequest(fmt.Sprintf("group %q: maximum %d items allowed", g.Key, maxNavItemsPerGroup))
		}

		for i, item := range g.Items {
			if err := validateNavItem(g.Key, i, item); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateNavItem(groupKey string, idx int, item NavItem) error {
	switch item.Type {
	case "object":
		if item.ObjectAPIName == "" {
			return apperror.BadRequest(fmt.Sprintf("group %q item %d: object_api_name is required for type 'object'", groupKey, idx))
		}
	case "link":
		if item.Label == "" {
			return apperror.BadRequest(fmt.Sprintf("group %q item %d: label is required for type 'link'", groupKey, idx))
		}
		if item.URL == "" {
			return apperror.BadRequest(fmt.Sprintf("group %q item %d: url is required for type 'link'", groupKey, idx))
		}
		if !strings.HasPrefix(item.URL, "/") && !strings.HasPrefix(item.URL, "https://") {
			return apperror.BadRequest(fmt.Sprintf("group %q item %d: url must start with '/' or 'https://'", groupKey, idx))
		}
		if strings.HasPrefix(strings.ToLower(item.URL), "javascript:") {
			return apperror.BadRequest(fmt.Sprintf("group %q item %d: javascript: URLs are not allowed", groupKey, idx))
		}
	case "divider":
		// no additional validation
	default:
		return apperror.BadRequest(fmt.Sprintf("group %q item %d: type must be 'object', 'link', or 'divider'", groupKey, idx))
	}
	return nil
}
