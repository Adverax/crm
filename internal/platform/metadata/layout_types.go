package metadata

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Layout represents a presentation layer for an Object View (ADR-0027).
// Layout per (object_view_id, form_factor, mode) defines HOW data is displayed.
type Layout struct {
	ID           uuid.UUID    `json:"id"`
	ObjectViewID uuid.UUID    `json:"object_view_id"`
	FormFactor   string       `json:"form_factor"`
	Mode         string       `json:"mode"`
	Config       LayoutConfig `json:"config"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// LayoutConfig holds the full layout configuration stored as JSONB.
type LayoutConfig struct {
	Root          *LayoutComponent          `json:"root,omitempty"`
	SectionConfig map[string]SectionConfig  `json:"section_config,omitempty"`
	FieldConfig   map[string]LayoutFieldConfig    `json:"field_config,omitempty"`
	ListConfig    *ListConfig               `json:"list_config,omitempty"`
}

// LayoutComponent represents a node in the page component tree.
type LayoutComponent struct {
	Type     string            `json:"type"`               // grid | group | highlights | field_section | related_list | query_widget | actions_bar | activity_feed | tabs
	Key      string            `json:"key,omitempty"`      // for field_section, related_list, query_widget
	Columns  int               `json:"columns,omitempty"`  // for grid
	ColSpan  int               `json:"col_span,omitempty"` // for children in grid
	Limit    int               `json:"limit,omitempty"`    // for related_list
	Children []LayoutComponent `json:"children,omitempty"` // for grid, group, tabs
}

// SectionConfig holds layout overrides for a section.
type SectionConfig struct {
	Columns        int    `json:"columns,omitempty"`
	Collapsed      bool   `json:"collapsed,omitempty"`
	Collapsible    bool   `json:"collapsible,omitempty"`
	VisibilityExpr string `json:"visibility_expr,omitempty"`
}

// LayoutFieldConfig holds layout overrides for a field.
type LayoutFieldConfig struct {
	LayoutRef      string          `json:"layout_ref,omitempty"`
	ColSpan        int             `json:"col_span,omitempty"`
	UIKind         json.RawMessage `json:"ui_kind,omitempty"`
	RequiredExpr   string          `json:"required_expr,omitempty"`
	ReadonlyExpr   string          `json:"readonly_expr,omitempty"`
	VisibilityExpr string          `json:"visibility_expr,omitempty"`
	Reference      *RefConfig      `json:"reference,omitempty"`
}

// RefConfig holds reference field display configuration.
type RefConfig struct {
	DisplayFields []string         `json:"display_fields,omitempty"`
	SearchFields  []string         `json:"search_fields,omitempty"`
	Target        string           `json:"target,omitempty"`
	Hint          string           `json:"hint,omitempty"`
	Filter        *RefFilterConfig `json:"filter,omitempty"`
}

// RefFilterConfig holds filter configuration for reference lookups.
type RefFilterConfig struct {
	Items []RefFilterItem `json:"items"`
}

// RefFilterItem holds a single filter condition for reference lookups.
type RefFilterItem struct {
	Field    string          `json:"field"`
	Operator string          `json:"operator"`
	Value    json.RawMessage `json:"value"`
}

// ListConfig holds configuration for list/table views.
type ListConfig struct {
	View       string             `json:"view,omitempty"`
	Columns    []ListColumnConfig `json:"columns,omitempty"`
	SortBy     []ListSortConfig   `json:"sort_by,omitempty"`
	Search     *ListSearchConfig  `json:"search,omitempty"`
	RowActions json.RawMessage    `json:"row_actions,omitempty"`
}

// ListColumnConfig holds configuration for a single list column.
type ListColumnConfig struct {
	Field      string          `json:"field"`
	Label      string          `json:"label,omitempty"`
	Width      string          `json:"width,omitempty"`
	Align      string          `json:"align,omitempty"`
	Sortable   *bool           `json:"sortable,omitempty"`
	SortDir    string          `json:"sort_dir,omitempty"`
	Filterable *bool           `json:"filterable,omitempty"`
	UIKind     json.RawMessage `json:"ui_kind,omitempty"`
}

// ListSortConfig holds a sort direction for a field.
type ListSortConfig struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// ListSearchConfig holds search configuration for list views.
type ListSearchConfig struct {
	Fields      []string `json:"fields"`
	Placeholder string   `json:"placeholder,omitempty"`
}

// CreateLayoutInput is the input for creating a new Layout.
type CreateLayoutInput struct {
	ObjectViewID uuid.UUID
	FormFactor   string
	Mode         string
	Config       LayoutConfig
}

// UpdateLayoutInput is the input for updating an existing Layout.
type UpdateLayoutInput struct {
	Config LayoutConfig
}
