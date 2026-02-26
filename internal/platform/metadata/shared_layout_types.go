package metadata

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SharedLayout represents a reusable layout fragment (ADR-0027).
// Referenced via layout_ref in FieldConfig. Deletion is RESTRICT while referenced.
type SharedLayout struct {
	ID        uuid.UUID       `json:"id"`
	APIName   string          `json:"api_name"`
	Type      string          `json:"type"`
	Label     string          `json:"label"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// CreateSharedLayoutInput is the input for creating a new SharedLayout.
type CreateSharedLayoutInput struct {
	APIName string
	Type    string
	Label   string
	Config  json.RawMessage
}

// UpdateSharedLayoutInput is the input for updating an existing SharedLayout.
type UpdateSharedLayoutInput struct {
	Label  string
	Config json.RawMessage
}
