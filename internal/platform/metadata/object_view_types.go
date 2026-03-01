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
type OVConfig struct {
	Read OVReadConfig `json:"read"`
}

// OVReadConfig holds the read-time (presentation + actions) configuration (ADR-0036).
// Fields is a unified list of regular and computed fields (ADR-0035).
// Actions include CRUD and custom operations as first-class units.
type OVReadConfig struct {
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

// OVAction describes an operation available on the page (ADR-0036).
// When Apply is nil, the action is UI-only (e.g. a link or client-side toggle).
// When Apply is set, the action is executable server-side via POST /view/:ov/action/:key.
type OVAction struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Type           string `json:"type"`
	Icon           string `json:"icon"`
	VisibilityExpr string `json:"visibility_expr"`

	Apply *OVActionApply `json:"apply,omitempty"`
}

// OVActionApply describes the transactional execution model for an action.
type OVActionApply struct {
	Type     string         `json:"type"`               // "dml" | "scenario"
	DML      []string       `json:"dml,omitempty"`      // DML statements for type="dml"
	Scenario *OVScenarioRef `json:"scenario,omitempty"` // scenario reference for type="scenario"
}

// OVScenarioRef references a scenario to start when an action is executed.
type OVScenarioRef struct {
	APIName string            `json:"api_name"`
	Params  map[string]string `json:"params,omitempty"`
}

// OVQuery describes a named SOQL query scoped to this Object View (ADR-0035).
// The query type (scalar vs list) is determined by the SOQL syntax:
// SELECT ROW ... = scalar (single record), SELECT ... = list (multiple records).
// The first scalar query in the array is the implicit default (context record).
type OVQuery struct {
	Name string `json:"name"`
	SOQL string `json:"soql"`
	When string `json:"when,omitempty"`
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
