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
// Validation is a CEL expression evaluated after type conversion; if it returns false, the request is rejected.
type PortalArg struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Default      *string `json:"default,omitempty"`
	Validation   string  `json:"validation,omitempty"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// PortalConfig holds the full Object View configuration stored as JSONB.
type PortalConfig struct {
	Args []PortalArg      `json:"args,omitempty"`
	Read PortalReadConfig `json:"read"`
}

// PortalReadConfig holds the read-time (presentation + actions) configuration (ADR-0036).
// Fields is a unified list of regular and computed fields (ADR-0035).
// Actions include CRUD and custom operations as first-class units.
type PortalReadConfig struct {
	Fields  []PortalViewField `json:"fields"`
	Actions []PortalAction    `json:"actions"`
	Queries []PortalQuery     `json:"queries,omitempty"`
}

// PortalViewField describes a field in the view configuration (ADR-0035).
// Fields without Expr are simple field references (resolved from the default query).
// Fields with Expr are computed from a CEL expression that can reference queries and other fields.
// The result type is inferred from the CEL expression at runtime — no explicit Type field needed.
type PortalViewField struct {
	Name string `json:"name"`
	Expr string `json:"expr,omitempty"`
	When string `json:"when,omitempty"`
}

// PortalAction describes an operation available on the page (ADR-0036).
// When Apply is nil, the action is UI-only (e.g. a link or client-side toggle).
// When Apply is set, the action is executable server-side via POST /view/:ov/action/:key.
type PortalAction struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Type           string `json:"type"`
	Icon           string `json:"icon"`
	VisibilityExpr string `json:"visibility_expr"`

	Apply *PortalActionApply `json:"apply,omitempty"`
}

// PortalActionApply describes the transactional execution model for an action.
type PortalActionApply struct {
	Type     string             `json:"type"`               // "dml" | "scenario"
	DML      []string           `json:"dml,omitempty"`      // DML statements for type="dml"
	Scenario *PortalScenarioRef `json:"scenario,omitempty"` // scenario reference for type="scenario"
}

// PortalScenarioRef references a scenario to start when an action is executed.
type PortalScenarioRef struct {
	APIName string            `json:"api_name"`
	Params  map[string]string `json:"params,omitempty"`
}

// PortalQuery describes a named SOQL query scoped to this Object View (ADR-0035).
// The query type (scalar vs list) is determined by the SOQL syntax:
// SELECT ROW ... = scalar (single record), SELECT ... = list (multiple records).
// The first scalar query in the array is the implicit default (context record).
type PortalQuery struct {
	Name string `json:"name"`
	SOQL string `json:"soql"`
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
