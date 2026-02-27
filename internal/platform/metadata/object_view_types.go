package metadata

import (
	"time"

	"github.com/google/uuid"
)

// ObjectView represents a UI screen configuration (ADR-0022).
// OV is not bound to a specific object — routing is done via Navigation config.
type ObjectView struct {
	ID          uuid.UUID  `json:"id"`
	ProfileID   *uuid.UUID `json:"profile_id"`
	APIName     string     `json:"api_name"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	Config      OVConfig   `json:"config"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// OVConfig holds the full Object View configuration stored as JSONB.
// Split into View (presentation) and Edit (data contract) sub-objects.
type OVConfig struct {
	View OVViewConfig  `json:"view"`
	Edit *OVEditConfig `json:"edit,omitempty"`
}

// OVViewConfig holds view-time (presentation) configuration.
// Fields is a unified list of regular and computed fields (ADR-0035).
type OVViewConfig struct {
	Fields  []OVViewField `json:"fields"`
	Actions []OVAction    `json:"actions"`
	Queries []OVQuery     `json:"queries,omitempty"`
}

// OVViewField describes a field in the view configuration (ADR-0035).
// Fields without Expr are simple field references (resolved from the default query).
// Fields with Expr are computed from a CEL expression that can reference queries.
type OVViewField struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	Expr string `json:"expr,omitempty"`
	When string `json:"when,omitempty"`
}

// OVEditConfig holds edit-time (data contract) configuration.
// Optional — only present when create/update operations make sense.
type OVEditConfig struct {
	Fields     []string       `json:"fields,omitempty"`
	Validation []OVValidation `json:"validation,omitempty"`
	Defaults   []OVDefault    `json:"defaults,omitempty"`
	Computed   []OVComputed   `json:"computed,omitempty"`
	Mutations  []OVMutation   `json:"mutations,omitempty"`
}

// OVAction describes a button action on the record detail page.
type OVAction struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Type           string `json:"type"`
	Icon           string `json:"icon"`
	VisibilityExpr string `json:"visibility_expr"`
}

// OVQuery describes a named SOQL query scoped to this Object View (ADR-0035).
type OVQuery struct {
	Name    string `json:"name"`
	SOQL    string `json:"soql"`
	Type    string `json:"type"`
	Default bool   `json:"default,omitempty"`
	When    string `json:"when,omitempty"`
}

// OVMutation describes a DML operation scoped to this Object View.
type OVMutation struct {
	DML     string     `json:"dml"`
	Foreach string     `json:"foreach,omitempty"`
	Sync    *OVMutSync `json:"sync,omitempty"`
	When    string     `json:"when,omitempty"`
}

// OVMutSync describes synchronization mapping for a mutation.
type OVMutSync struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// OVValidation describes a validation rule scoped to this Object View.
type OVValidation struct {
	Expr     string `json:"expr"`
	Message  string `json:"message"`
	Code     string `json:"code,omitempty"`
	Severity string `json:"severity"`
	When     string `json:"when,omitempty"`
}

// OVDefault describes a default value expression scoped to this Object View.
type OVDefault struct {
	Field string `json:"field"`
	Expr  string `json:"expr"`
	On    string `json:"on"`
	When  string `json:"when,omitempty"`
}

// OVComputed describes a computed field expression scoped to this Object View.
type OVComputed struct {
	Field string `json:"field"`
	Expr  string `json:"expr"`
}

// CreateObjectViewInput is the input for creating a new Object View.
type CreateObjectViewInput struct {
	ProfileID   *uuid.UUID
	APIName     string
	Label       string
	Description string
	Config      OVConfig
}

// UpdateObjectViewInput is the input for updating an existing Object View.
type UpdateObjectViewInput struct {
	Label       string
	Description string
	Config      OVConfig
}

// FieldNames extracts field API names from OVViewField slice.
func FieldNames(fields []OVViewField) []string {
	if len(fields) == 0 {
		return nil
	}
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names
}
