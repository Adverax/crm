package metadata

import (
	"context"

	"github.com/google/uuid"
)

// ValidationRuleRepository provides CRUD operations for validation rules.
type ValidationRuleRepository interface {
	Create(ctx context.Context, input CreateValidationRuleInput) (*ValidationRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ValidationRule, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]ValidationRule, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateValidationRuleInput) (*ValidationRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context) ([]ValidationRule, error)
}
