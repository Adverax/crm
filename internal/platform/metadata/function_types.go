package metadata

import (
	"time"

	"github.com/google/uuid"
)

// Function represents a named reusable CEL expression with typed parameters (ADR-0026).
type Function struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Params      []FunctionParam `json:"params"`
	ReturnType  string          `json:"return_type"`
	Body        string          `json:"body"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// FunctionParam describes a single typed parameter of a custom function.
type FunctionParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// CreateFunctionInput is the input for creating a new custom function.
type CreateFunctionInput struct {
	Name        string
	Description string
	Params      []FunctionParam
	ReturnType  string
	Body        string
}

// UpdateFunctionInput is the input for updating an existing custom function.
type UpdateFunctionInput struct {
	Description string
	Params      []FunctionParam
	ReturnType  string
	Body        string
}
