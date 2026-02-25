package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

const maxAutomationRuleCount = 500

var validEventTypes = map[string]bool{
	"before_insert": true,
	"after_insert":  true,
	"before_update": true,
	"after_update":  true,
	"before_delete": true,
	"after_delete":  true,
}

var validExecutionModes = map[string]bool{
	"per_record": true,
	"per_batch":  true,
}

// AutomationRuleService provides business logic for automation rules (ADR-0031).
type AutomationRuleService interface {
	Create(ctx context.Context, input CreateAutomationRuleInput) (*AutomationRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AutomationRule, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]AutomationRule, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateAutomationRuleInput) (*AutomationRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// OnAutomationRulesChanged is a callback invoked after automation rules are modified.
type OnAutomationRulesChanged func(ctx context.Context) error

type automationRuleService struct {
	repo     AutomationRuleRepository
	cache    *MetadataCache
	onChange OnAutomationRulesChanged
}

// NewAutomationRuleService creates a new AutomationRuleService.
func NewAutomationRuleService(
	repo AutomationRuleRepository,
	cache *MetadataCache,
	onChange OnAutomationRulesChanged,
) AutomationRuleService {
	return &automationRuleService{
		repo:     repo,
		cache:    cache,
		onChange: onChange,
	}
}

func (s *automationRuleService) Create(ctx context.Context, input CreateAutomationRuleInput) (*AutomationRule, error) {
	if err := validateAutomationRuleInput(input.Name, input.EventType, input.ProcedureCode, input.ExecutionMode); err != nil {
		return nil, fmt.Errorf("automationRuleService.Create: %w", err)
	}

	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("automationRuleService.Create: %w", err)
	}
	if count >= maxAutomationRuleCount {
		return nil, fmt.Errorf("automationRuleService.Create: %w",
			apperror.BadRequest(fmt.Sprintf("max automation rule limit reached (%d)", maxAutomationRuleCount)))
	}

	rule, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("automationRuleService.Create: %w", err)
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("automationRuleService.Create: %w", err)
	}

	return rule, nil
}

func (s *automationRuleService) GetByID(ctx context.Context, id uuid.UUID) (*AutomationRule, error) {
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("automationRuleService.GetByID: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("automationRuleService.GetByID: %w",
			apperror.NotFound("automation_rule", id.String()))
	}
	return rule, nil
}

func (s *automationRuleService) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]AutomationRule, error) {
	rules, err := s.repo.ListByObjectID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("automationRuleService.ListByObjectID: %w", err)
	}
	return rules, nil
}

func (s *automationRuleService) Update(ctx context.Context, id uuid.UUID, input UpdateAutomationRuleInput) (*AutomationRule, error) {
	if err := validateAutomationRuleInput(input.Name, input.EventType, input.ProcedureCode, input.ExecutionMode); err != nil {
		return nil, fmt.Errorf("automationRuleService.Update: %w", err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("automationRuleService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("automationRuleService.Update: %w",
			apperror.NotFound("automation_rule", id.String()))
	}

	rule, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("automationRuleService.Update: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("automationRuleService.Update: %w",
			apperror.NotFound("automation_rule", id.String()))
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("automationRuleService.Update: %w", err)
	}

	return rule, nil
}

func (s *automationRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("automationRuleService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("automationRuleService.Delete: %w",
			apperror.NotFound("automation_rule", id.String()))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("automationRuleService.Delete: %w", err)
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return fmt.Errorf("automationRuleService.Delete: %w", err)
	}

	return nil
}

func (s *automationRuleService) reloadAndNotify(ctx context.Context) error {
	if err := s.cache.LoadAutomationRules(ctx); err != nil {
		return fmt.Errorf("cache reload: %w", err)
	}
	if s.onChange != nil {
		if err := s.onChange(ctx); err != nil {
			return fmt.Errorf("onChange callback: %w", err)
		}
	}
	return nil
}

// --- validation helpers ---

func validateAutomationRuleInput(name, eventType, procedureCode, executionMode string) error {
	if name == "" {
		return apperror.BadRequest("name is required")
	}
	if !validEventTypes[eventType] {
		return apperror.BadRequest(fmt.Sprintf("invalid event_type: %s", eventType))
	}
	if procedureCode == "" {
		return apperror.BadRequest("procedure_code is required")
	}
	if !validExecutionModes[executionMode] {
		return apperror.BadRequest(fmt.Sprintf("invalid execution_mode: %s", executionMode))
	}
	return nil
}
