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
// Split into View (presentation) and Edit (data contract) sub-objects.
type OVConfig struct {
	View OVViewConfig  `json:"view"`
	Edit *OVEditConfig `json:"edit,omitempty"`
}

// UnmarshalJSON handles three formats:
// 1. New format: "view"/"edit" keys
// 2. Legacy nested: "read"/"write" keys (mapped to View/Edit)
// 3. Legacy flat: no "view"/"read" key — convert from flat fields
func (c *OVConfig) UnmarshalJSON(data []byte) error {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}

	if _, hasView := probe["view"]; hasView {
		// New format with "view"/"edit" keys
		type Alias OVConfig
		var alias Alias
		if err := json.Unmarshal(data, &alias); err != nil {
			return err
		}
		*c = OVConfig(alias)
		return nil
	}

	if readRaw, hasRead := probe["read"]; hasRead {
		// Legacy nested format: "read"/"write" → View/Edit
		if err := json.Unmarshal(readRaw, &c.View); err != nil {
			return err
		}
		if writeRaw, hasWrite := probe["write"]; hasWrite {
			var edit OVEditConfig
			if err := json.Unmarshal(writeRaw, &edit); err != nil {
				return err
			}
			c.Edit = &edit
		}
		return nil
	}

	// Legacy flat format — convert to nested
	var legacy legacyOVConfig
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}

	c.View = OVViewConfig{
		Fields:   legacy.Fields,
		Actions:  legacy.Actions,
		Queries:  legacy.Queries,
		Computed: convertVirtualFieldsToViewComputed(legacy.VirtualFields),
	}

	// Only create Edit if there's any edit-related data
	if len(legacy.Validation) > 0 || len(legacy.Defaults) > 0 ||
		len(legacy.Computed) > 0 || len(legacy.Mutations) > 0 {
		c.Edit = &OVEditConfig{
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

func convertVirtualFieldsToViewComputed(vfs []OVVirtualField) []OVViewComputed {
	if len(vfs) == 0 {
		return nil
	}
	result := make([]OVViewComputed, len(vfs))
	for i, vf := range vfs {
		result[i] = OVViewComputed(vf)
	}
	return result
}

// OVViewConfig holds view-time (presentation) configuration.
type OVViewConfig struct {
	Fields   []string         `json:"fields"`
	Actions  []OVAction       `json:"actions"`
	Queries  []OVQuery        `json:"queries,omitempty"`
	Computed []OVViewComputed `json:"computed,omitempty"`
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

// OVViewComputed describes a computed virtual field for view context.
// Renamed from OVVirtualField.
type OVViewComputed struct {
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
// New code should use OVViewComputed instead.
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
