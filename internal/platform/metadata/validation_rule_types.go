package metadata

import (
	"time"

	"github.com/google/uuid"
)

// ValidationRule represents a CEL-based validation rule attached to an object.
type ValidationRule struct {
	ID             uuid.UUID `json:"id"`
	ObjectID       uuid.UUID `json:"object_id"`
	APIName        string    `json:"api_name"`
	Label          string    `json:"label"`
	Description    string    `json:"description"`
	Expression     string    `json:"expression"`
	ErrorMessage   string    `json:"error_message"`
	ErrorCode      string    `json:"error_code"`
	Severity       string    `json:"severity"`
	WhenExpression *string   `json:"when_expression"`
	AppliesTo      string    `json:"applies_to"`
	SortOrder      int       `json:"sort_order"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateValidationRuleInput is the input for creating a new validation rule.
type CreateValidationRuleInput struct {
	ObjectID       uuid.UUID
	APIName        string
	Label          string
	Description    string
	Expression     string
	ErrorMessage   string
	ErrorCode      string
	Severity       string
	WhenExpression *string
	AppliesTo      string
	SortOrder      int
	IsActive       bool
}

// UpdateValidationRuleInput is the input for updating an existing validation rule.
type UpdateValidationRuleInput struct {
	Label          string
	Description    string
	Expression     string
	ErrorMessage   string
	ErrorCode      string
	Severity       string
	WhenExpression *string
	AppliesTo      string
	SortOrder      int
	IsActive       bool
}
