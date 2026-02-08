package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
)

type profileServiceImpl struct {
	txBeginner TxBeginner
	profileRepo ProfileRepository
	psRepo      PermissionSetRepository
}

// NewProfileService creates a new ProfileService.
func NewProfileService(
	txBeginner TxBeginner,
	profileRepo ProfileRepository,
	psRepo PermissionSetRepository,
) ProfileService {
	return &profileServiceImpl{
		txBeginner:  txBeginner,
		profileRepo: profileRepo,
		psRepo:      psRepo,
	}
}

func (s *profileServiceImpl) Create(ctx context.Context, input CreateProfileInput) (*Profile, error) {
	if err := ValidateCreateProfile(input); err != nil {
		return nil, fmt.Errorf("profileService.Create: %w", err)
	}

	existing, _ := s.profileRepo.GetByAPIName(ctx, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("profileService.Create: %w",
			apperror.Conflict(fmt.Sprintf("profile with api_name '%s' already exists", input.APIName)))
	}

	var result *Profile
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		// Create base permission set atomically
		psInput := CreatePermissionSetInput{
			APIName:     input.APIName + "_base",
			Label:       input.Label + " Base",
			Description: fmt.Sprintf("Base permission set for profile %s", input.Label),
			PSType:      PSTypeGrant,
		}
		ps, err := s.psRepo.Create(ctx, tx, psInput)
		if err != nil {
			return fmt.Errorf("profileService.Create: create base PS: %w", err)
		}

		profile := &Profile{
			APIName:             input.APIName,
			Label:               input.Label,
			Description:         input.Description,
			BasePermissionSetID: ps.ID,
		}
		created, err := s.profileRepo.Create(ctx, tx, profile)
		if err != nil {
			return fmt.Errorf("profileService.Create: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *profileServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	profile, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("profileService.GetByID: %w", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("profileService.GetByID: %w",
			apperror.NotFound("Profile", id.String()))
	}
	return profile, nil
}

func (s *profileServiceImpl) List(ctx context.Context, page, perPage int32) ([]Profile, int64, error) {
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	profiles, err := s.profileRepo.List(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("profileService.List: %w", err)
	}

	total, err := s.profileRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("profileService.List: count: %w", err)
	}

	return profiles, total, nil
}

func (s *profileServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateProfileInput) (*Profile, error) {
	existing, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("profileService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("profileService.Update: %w",
			apperror.NotFound("Profile", id.String()))
	}

	if err := ValidateUpdateProfile(input); err != nil {
		return nil, fmt.Errorf("profileService.Update: %w", err)
	}

	var result *Profile
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.profileRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("profileService.Update: %w", err)
		}
		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *profileServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("profileService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("profileService.Delete: %w",
			apperror.NotFound("Profile", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.profileRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("profileService.Delete: %w", err)
		}
		// Base PS is cleaned up by CASCADE or manually
		return nil
	})
}
