package metadata

import (
	"time"

	"github.com/google/uuid"
)

// Portal represents a UI screen configuration (ADR-0022).
// OV is not bound to a specific object — routing is done via Navigation config.
type Portal struct {
	ID          uuid.UUID    `json:"id"`
	ProfileID   *uuid.UUID   `json:"profile_id"`
	APIName     string       `json:"api_name"`
	Label       string       `json:"label"`
	Description string       `json:"description"`
	Config      PortalConfig `json:"config"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// PortalArg declares a typed parameter for the portal.
// Args are resolved from URL query parameters at runtime and available in SOQL (:paramName) and CEL (args.*).
type PortalArg struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Default *string `json:"default,omitempty"`
}

// PortalArgRule is a validation rule for portal args.
// Condition is a CEL expression (env: args only) that must return true for the request to proceed.
type PortalArgRule struct {
	Name         string `json:"name"`
	Condition    string `json:"condition"`
	ErrorMessage string `json:"error_message"`
}

// PortalConfig holds the full Portal configuration stored as JSONB (ADR-0037).
// Portal is a graph of gates. EntryGate names the starting vertex.
type PortalConfig struct {
	Args      []PortalArg           `json:"args,omitempty"`
	ArgRules  []PortalArgRule       `json:"arg_rules,omitempty"`
	EntryGate string                `json:"entry_gate"`
	Gates     map[string]PortalGate `json:"gates"`
}

// PortalGate is a vertex in the portal graph (ADR-0037).
// Each gate is a stateless function: receives args, executes body, returns layout + datasets + outcomes.
type PortalGate struct {
	Label    string          `json:"label"`
	Args     []PortalArg     `json:"args,omitempty"`
	ArgRules []PortalArgRule `json:"arg_rules,omitempty"`
	Body     []GateBodyStep  `json:"body"`
	Layout   *GateLayout     `json:"layout,omitempty"`
	Outcomes []GateOutcome   `json:"outcomes,omitempty"`
}

// GateBodyStep is a single SOQL or DML operation in the gate body chain.
type GateBodyStep struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`                       // "soql" | "dml"
	SOQL            string   `json:"soql,omitempty"`             // for type="soql"
	DML             string   `json:"dml,omitempty"`              // for type="dml"
	When            string   `json:"when,omitempty"`             // CEL condition (skip if false)
	PageSize        int      `json:"page_size,omitempty"`        // enables pagination + dynamic filtering
	RestrictFilters []string `json:"restrict_filters,omitempty"` // opt-out: exclude fields from filtering
}

// GateLayout holds presentation config for a gate (ADR-0037).
// Merges what was previously split between Portal fields (ADR-0022) and Layout config (ADR-0027).
type GateLayout struct {
	Fields        []PortalViewField            `json:"fields,omitempty"`
	Root          *LayoutComponent             `json:"root,omitempty"`
	SectionConfig map[string]SectionConfig     `json:"section_config,omitempty"`
	FieldConfig   map[string]LayoutFieldConfig `json:"field_config,omitempty"`
	ListConfig    *ListConfig                  `json:"list_config,omitempty"`
}

// GateOutcome is an edge in the portal graph — a declared transition to another gate.
type GateOutcome struct {
	Name         string            `json:"name"`
	Gate         string            `json:"gate"`
	Label        string            `json:"label,omitempty"`
	Icon         string            `json:"icon,omitempty"`
	Type         string            `json:"type,omitempty"` // "primary"|"secondary"|"danger"
	ArgsTemplate map[string]string `json:"args_template,omitempty"`
}

// PortalViewField describes a field in the view configuration (ADR-0035).
// Fields without Expr are simple field references (resolved from the default query).
// Fields with Expr are computed from a CEL expression that can reference queries and other fields.
type PortalViewField struct {
	Name string `json:"name"`
	Expr string `json:"expr,omitempty"`
	When string `json:"when,omitempty"`
}

// CreatePortalInput is the input for creating a new Object View.
type CreatePortalInput struct {
	ProfileID   *uuid.UUID
	APIName     string
	Label       string
	Description string
	Config      PortalConfig
}

// UpdatePortalInput is the input for updating an existing Object View.
type UpdatePortalInput struct {
	Label       string
	Description string
	Config      PortalConfig
}

// FieldNames extracts field API names from PortalViewField slice.
func FieldNames(fields []PortalViewField) []string {
	if len(fields) == 0 {
		return nil
	}
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names
}
