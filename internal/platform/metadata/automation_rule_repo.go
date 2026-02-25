package metadata

import (
	"context"

	"github.com/google/uuid"
)

// AutomationRuleRepository provides data access for automation rules (ADR-0031).
type AutomationRuleRepository interface {
	Create(ctx context.Context, input CreateAutomationRuleInput) (*AutomationRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AutomationRule, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]AutomationRule, error)
	ListAll(ctx context.Context) ([]AutomationRule, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateAutomationRuleInput) (*AutomationRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
}
