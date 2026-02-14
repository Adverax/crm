package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// ValidationRuleService provides business logic for validation rules.
type ValidationRuleService interface {
	Create(ctx context.Context, input CreateValidationRuleInput) (*ValidationRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ValidationRule, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]ValidationRule, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateValidationRuleInput) (*ValidationRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type validationRuleService struct {
	pool  *pgxpool.Pool
	repo  ValidationRuleRepository
	cache *MetadataCache
}

// NewValidationRuleService creates a new ValidationRuleService.
func NewValidationRuleService(
	pool *pgxpool.Pool,
	repo ValidationRuleRepository,
	cache *MetadataCache,
) ValidationRuleService {
	return &validationRuleService{
		pool:  pool,
		repo:  repo,
		cache: cache,
	}
}

func (s *validationRuleService) Create(ctx context.Context, input CreateValidationRuleInput) (*ValidationRule, error) {
	// Validate object exists
	if _, ok := s.cache.GetObjectByID(input.ObjectID); !ok {
		return nil, fmt.Errorf("validationRuleService.Create: %w",
			apperror.NotFound("object", input.ObjectID.String()))
	}

	// Apply defaults
	if input.Severity == "" {
		input.Severity = "error"
	}
	if input.ErrorCode == "" {
		input.ErrorCode = "validation_failed"
	}
	if input.AppliesTo == "" {
		input.AppliesTo = "create,update"
	}
	if !input.IsActive {
		input.IsActive = true
	}

	rule, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("validationRuleService.Create: %w", err)
	}

	// Reload cache
	if err := s.cache.LoadValidationRules(ctx); err != nil {
		return nil, fmt.Errorf("validationRuleService.Create: cache reload: %w", err)
	}

	return rule, nil
}

func (s *validationRuleService) GetByID(ctx context.Context, id uuid.UUID) (*ValidationRule, error) {
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("validationRuleService.GetByID: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("validationRuleService.GetByID: %w",
			apperror.NotFound("validation_rule", id.String()))
	}
	return rule, nil
}

func (s *validationRuleService) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]ValidationRule, error) {
	rules, err := s.repo.ListByObjectID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("validationRuleService.ListByObjectID: %w", err)
	}
	return rules, nil
}

func (s *validationRuleService) Update(ctx context.Context, id uuid.UUID, input UpdateValidationRuleInput) (*ValidationRule, error) {
	rule, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("validationRuleService.Update: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("validationRuleService.Update: %w",
			apperror.NotFound("validation_rule", id.String()))
	}

	// Reload cache
	if err := s.cache.LoadValidationRules(ctx); err != nil {
		return nil, fmt.Errorf("validationRuleService.Update: cache reload: %w", err)
	}

	return rule, nil
}

func (s *validationRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("validationRuleService.Delete: %w", err)
	}

	// Reload cache
	if err := s.cache.LoadValidationRules(ctx); err != nil {
		return fmt.Errorf("validationRuleService.Delete: cache reload: %w", err)
	}

	return nil
}
