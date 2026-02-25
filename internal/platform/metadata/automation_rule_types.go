package metadata

import (
	"time"

	"github.com/google/uuid"
)

// AutomationRule represents an automation rule definition (ADR-0031).
type AutomationRule struct {
	ID            uuid.UUID `json:"id"`
	ObjectID      uuid.UUID `json:"object_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	EventType     string    `json:"event_type"`
	Condition     *string   `json:"condition"`
	ProcedureCode string    `json:"procedure_code"`
	ExecutionMode string    `json:"execution_mode"`
	SortOrder     int       `json:"sort_order"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateAutomationRuleInput is the input for creating an automation rule.
type CreateAutomationRuleInput struct {
	ObjectID      uuid.UUID
	Name          string
	Description   string
	EventType     string
	Condition     *string
	ProcedureCode string
	ExecutionMode string
	SortOrder     int
	IsActive      bool
}

// UpdateAutomationRuleInput is the input for updating an automation rule.
type UpdateAutomationRuleInput struct {
	Name          string
	Description   string
	EventType     string
	Condition     *string
	ProcedureCode string
	ExecutionMode string
	SortOrder     int
	IsActive      bool
}
