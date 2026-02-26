package metadata

import (
	"encoding/json"
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
// Split into Read (presentation) and Write (data contract) sub-objects.
type OVConfig struct {
	Read  OVReadConfig   `json:"read"`
	Write *OVWriteConfig `json:"write,omitempty"`
}

// UnmarshalJSON handles both the new nested format and legacy flat format.
// Legacy flat format (pre-split) is detected by absence of the "read" key.
func (c *OVConfig) UnmarshalJSON(data []byte) error {
	// Try to detect format by checking for "read" key
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}

	if _, hasRead := probe["read"]; hasRead {
		// New nested format
		type Alias OVConfig
		var alias Alias
		if err := json.Unmarshal(data, &alias); err != nil {
			return err
		}
		*c = OVConfig(alias)
		return nil
	}

	// Legacy flat format — convert to nested
	var legacy legacyOVConfig
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}

	c.Read = OVReadConfig{
		Fields:   legacy.Fields,
		Actions:  legacy.Actions,
		Queries:  legacy.Queries,
		Computed: convertVirtualFieldsToReadComputed(legacy.VirtualFields),
	}

	// Only create Write if there's any write-related data
	if len(legacy.Validation) > 0 || len(legacy.Defaults) > 0 ||
		len(legacy.Computed) > 0 || len(legacy.Mutations) > 0 {
		c.Write = &OVWriteConfig{
			Validation: legacy.Validation,
			Defaults:   legacy.Defaults,
			Computed:   legacy.Computed,
			Mutations:  legacy.Mutations,
		}
	}

	return nil
}

// legacyOVConfig represents the pre-split flat config format.
type legacyOVConfig struct {
	Fields        []string         `json:"fields"`
	Actions       []OVAction       `json:"actions"`
	Queries       []OVQuery        `json:"queries,omitempty"`
	VirtualFields []OVVirtualField `json:"virtual_fields,omitempty"`
	Mutations     []OVMutation     `json:"mutations,omitempty"`
	Validation    []OVValidation   `json:"validation,omitempty"`
	Defaults      []OVDefault      `json:"defaults,omitempty"`
	Computed      []OVComputed     `json:"computed,omitempty"`
}

func convertVirtualFieldsToReadComputed(vfs []OVVirtualField) []OVReadComputed {
	if len(vfs) == 0 {
		return nil
	}
	result := make([]OVReadComputed, len(vfs))
	for i, vf := range vfs {
		result[i] = OVReadComputed(vf)
	}
	return result
}

// OVReadConfig holds read-time (presentation) configuration.
type OVReadConfig struct {
	Fields   []string         `json:"fields"`
	Actions  []OVAction       `json:"actions"`
	Queries  []OVQuery        `json:"queries,omitempty"`
	Computed []OVReadComputed `json:"computed,omitempty"`
}

// OVWriteConfig holds write-time (data contract) configuration.
// Optional — only present when create/update operations make sense.
type OVWriteConfig struct {
	Fields     []string       `json:"fields,omitempty"`
	Validation []OVValidation `json:"validation,omitempty"`
	Defaults   []OVDefault    `json:"defaults,omitempty"`
	Computed   []OVComputed   `json:"computed,omitempty"`
	Mutations  []OVMutation   `json:"mutations,omitempty"`
}

// OVReadComputed describes a computed virtual field for read context.
// Renamed from OVVirtualField.
type OVReadComputed struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Expr string `json:"expr"`
	When string `json:"when,omitempty"`
}

// OVAction describes a button action on the record detail page.
type OVAction struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Type           string `json:"type"`
	Icon           string `json:"icon"`
	VisibilityExpr string `json:"visibility_expr"`
}

// OVQuery describes a named SOQL query scoped to this Object View.
type OVQuery struct {
	Name string `json:"name"`
	SOQL string `json:"soql"`
	When string `json:"when,omitempty"`
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

// OVVirtualField is kept for backward compatibility during deserialization.
// New code should use OVReadComputed instead.
type OVVirtualField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Expr string `json:"expr"`
	When string `json:"when,omitempty"`
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
