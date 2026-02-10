package metadata

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FieldType represents the storage concern of a field.
type FieldType string

const (
	FieldTypeText      FieldType = "text"
	FieldTypeNumber    FieldType = "number"
	FieldTypeBoolean   FieldType = "boolean"
	FieldTypeDatetime  FieldType = "datetime"
	FieldTypePicklist  FieldType = "picklist"
	FieldTypeReference FieldType = "reference"
)

// FieldSubtype represents the semantic concern of a field.
type FieldSubtype string

const (
	// text subtypes
	SubtypePlain FieldSubtype = "plain"
	SubtypeArea  FieldSubtype = "area"
	SubtypeRich  FieldSubtype = "rich"
	SubtypeEmail FieldSubtype = "email"
	SubtypePhone FieldSubtype = "phone"
	SubtypeURL   FieldSubtype = "url"

	// number subtypes
	SubtypeInteger    FieldSubtype = "integer"
	SubtypeDecimal    FieldSubtype = "decimal"
	SubtypeCurrency   FieldSubtype = "currency"
	SubtypePercent    FieldSubtype = "percent"
	SubtypeAutoNumber FieldSubtype = "auto_number"

	// datetime subtypes
	SubtypeDate     FieldSubtype = "date"
	SubtypeDatetime FieldSubtype = "datetime"
	SubtypeTime     FieldSubtype = "time"

	// picklist subtypes
	SubtypeSingle FieldSubtype = "single"
	SubtypeMulti  FieldSubtype = "multi"

	// reference subtypes
	SubtypeAssociation FieldSubtype = "association"
	SubtypeComposition FieldSubtype = "composition"
	SubtypePolymorphic FieldSubtype = "polymorphic"
)

// ObjectType classifies whether an object is standard or custom.
type ObjectType string

const (
	ObjectTypeStandard ObjectType = "standard"
	ObjectTypeCustom   ObjectType = "custom"
)

// Visibility represents the Organization-Wide Default (OWD) for an object.
type Visibility string

const (
	VisibilityPrivate            Visibility = "private"
	VisibilityPublicRead         Visibility = "public_read"
	VisibilityPublicReadWrite    Visibility = "public_read_write"
	VisibilityControlledByParent Visibility = "controlled_by_parent"
)

// ObjectDefinition represents metadata of a CRM object.
type ObjectDefinition struct {
	ID          uuid.UUID  `json:"id"`
	APIName     string     `json:"api_name"`
	Label       string     `json:"label"`
	PluralLabel string     `json:"plural_label"`
	Description string     `json:"description"`
	TableName   string     `json:"table_name"`
	ObjectType  ObjectType `json:"object_type"`

	// Behavioral flags
	IsPlatformManaged     bool `json:"is_platform_managed"`
	IsVisibleInSetup      bool `json:"is_visible_in_setup"`
	IsCustomFieldsAllowed bool `json:"is_custom_fields_allowed"`
	IsDeleteableObject    bool `json:"is_deleteable_object"`

	// Record capabilities
	IsCreateable bool `json:"is_createable"`
	IsUpdateable bool `json:"is_updateable"`
	IsDeleteable bool `json:"is_deleteable"`
	IsQueryable  bool `json:"is_queryable"`
	IsSearchable bool `json:"is_searchable"`

	// Features
	HasActivities      bool `json:"has_activities"`
	HasNotes           bool `json:"has_notes"`
	HasHistoryTracking bool `json:"has_history_tracking"`
	HasSharingRules    bool `json:"has_sharing_rules"`

	// Security: Organization-Wide Default
	Visibility Visibility `json:"visibility"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FieldDefinition represents metadata of a field on a CRM object.
type FieldDefinition struct {
	ID                 uuid.UUID     `json:"id"`
	ObjectID           uuid.UUID     `json:"object_id"`
	APIName            string        `json:"api_name"`
	Label              string        `json:"label"`
	Description        string        `json:"description"`
	HelpText           string        `json:"help_text"`
	FieldType          FieldType     `json:"field_type"`
	FieldSubtype       *FieldSubtype `json:"field_subtype"`
	ReferencedObjectID *uuid.UUID    `json:"referenced_object_id"`
	IsRequired         bool          `json:"is_required"`
	IsUnique           bool          `json:"is_unique"`
	Config             FieldConfig   `json:"config"`
	IsSystemField      bool          `json:"is_system_field"`
	IsCustom           bool          `json:"is_custom"`
	IsPlatformManaged  bool          `json:"is_platform_managed"`
	SortOrder          int           `json:"sort_order"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// FieldConfig stores type-specific parameters as JSONB.
type FieldConfig struct {
	// text
	MaxLength *int `json:"max_length,omitempty"`

	// number
	Precision *int `json:"precision,omitempty"`
	Scale     *int `json:"scale,omitempty"`

	// auto_number
	Format     *string `json:"format,omitempty"`
	StartValue *int    `json:"start_value,omitempty"`

	// picklist
	PicklistID *uuid.UUID      `json:"picklist_id,omitempty"`
	Values     []PicklistValue `json:"values,omitempty"`

	// reference
	RelationshipName *string `json:"relationship_name,omitempty"`
	OnDelete         *string `json:"on_delete,omitempty"`
	IsReparentable   *bool   `json:"is_reparentable,omitempty"`

	// common
	DefaultValue *string `json:"default_value,omitempty"`
}

// Scan implements the sql.Scanner interface for JSONB.
func (fc *FieldConfig) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, fc)
	case string:
		return json.Unmarshal([]byte(v), fc)
	default:
		return json.Unmarshal(src.([]byte), fc)
	}
}

// PicklistValue represents a single value in a picklist.
type PicklistValue struct {
	ID        uuid.UUID `json:"id"`
	Value     string    `json:"value"`
	Label     string    `json:"label"`
	SortOrder int       `json:"sort_order"`
	IsDefault bool      `json:"is_default"`
	IsActive  bool      `json:"is_active"`
}

// PicklistDefinition represents a global picklist (Global Value Set).
type PicklistDefinition struct {
	ID          uuid.UUID `json:"id"`
	APIName     string    `json:"api_name"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PicklistValueRow represents a row in the picklist_values table.
type PicklistValueRow struct {
	ID                   uuid.UUID `json:"id"`
	PicklistDefinitionID uuid.UUID `json:"picklist_definition_id"`
	Value                string    `json:"value"`
	Label                string    `json:"label"`
	SortOrder            int       `json:"sort_order"`
	IsDefault            bool      `json:"is_default"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// PolymorphicTarget represents an allowed target object for a polymorphic field.
type PolymorphicTarget struct {
	ID        uuid.UUID `json:"id"`
	FieldID   uuid.UUID `json:"field_id"`
	ObjectID  uuid.UUID `json:"object_id"`
	CreatedAt time.Time `json:"created_at"`
}

// RelationshipInfo describes a relationship derived from field_definitions.
type RelationshipInfo struct {
	FieldID             uuid.UUID    `json:"field_id"`
	FieldAPIName        string       `json:"field_api_name"`
	RelationshipName    string       `json:"relationship_name"`
	ChildObjectID       uuid.UUID    `json:"child_object_id"`
	ChildObjectAPIName  string       `json:"child_object_api_name"`
	ParentObjectID      uuid.UUID    `json:"parent_object_id"`
	ParentObjectAPIName string       `json:"parent_object_api_name"`
	ReferenceSubtype    FieldSubtype `json:"reference_subtype"`
	OnDelete            string       `json:"on_delete"`
}
