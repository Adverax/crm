package templates

import "github.com/adverax/crm/internal/platform/metadata"

// Template defines a complete application template with objects and fields.
type Template struct {
	ID          string
	Label       string
	Description string
	Objects     []ObjectTemplate
	Fields      []FieldTemplate
}

// ObjectTemplate defines an object to be created when applying a template.
type ObjectTemplate struct {
	APIName     string
	Label       string
	PluralLabel string
	Description string
	Visibility  metadata.Visibility

	IsCreateable          bool
	IsUpdateable          bool
	IsDeleteable          bool
	IsQueryable           bool
	IsSearchable          bool
	IsCustomFieldsAllowed bool

	HasActivities      bool
	HasNotes           bool
	HasHistoryTracking bool
	HasSharingRules    bool
}

// FieldTemplate defines a field to be created when applying a template.
type FieldTemplate struct {
	ObjectAPIName           string
	APIName                 string
	Label                   string
	Description             string
	FieldType               metadata.FieldType
	FieldSubtype            *metadata.FieldSubtype
	ReferencedObjectAPIName string
	IsRequired              bool
	IsUnique                bool
	Config                  metadata.FieldConfig
	SortOrder               int
}

// TemplateInfo is the API response DTO for a template.
type TemplateInfo struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`
	Description string         `json:"description"`
	Status      TemplateStatus `json:"status"`
	Objects     int            `json:"objects"`
	Fields      int            `json:"fields"`
}

// TemplateStatus represents the availability of a template.
type TemplateStatus string

const (
	TemplateStatusAvailable TemplateStatus = "available"
	TemplateStatusApplied   TemplateStatus = "applied"
	TemplateStatusBlocked   TemplateStatus = "blocked"
)
