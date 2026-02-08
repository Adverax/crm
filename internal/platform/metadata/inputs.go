package metadata

import "github.com/google/uuid"

// CreateObjectInput contains the input data for creating an object definition.
type CreateObjectInput struct {
	APIName     string     `json:"api_name"`
	Label       string     `json:"label"`
	PluralLabel string     `json:"plural_label"`
	Description string     `json:"description"`
	ObjectType  ObjectType `json:"object_type"`

	IsVisibleInSetup      bool `json:"is_visible_in_setup"`
	IsCustomFieldsAllowed bool `json:"is_custom_fields_allowed"`
	IsDeleteableObject    bool `json:"is_deleteable_object"`

	IsCreateable bool `json:"is_createable"`
	IsUpdateable bool `json:"is_updateable"`
	IsDeleteable bool `json:"is_deleteable"`
	IsQueryable  bool `json:"is_queryable"`
	IsSearchable bool `json:"is_searchable"`

	HasActivities      bool `json:"has_activities"`
	HasNotes           bool `json:"has_notes"`
	HasHistoryTracking bool `json:"has_history_tracking"`
	HasSharingRules    bool `json:"has_sharing_rules"`
}

// UpdateObjectInput contains the input data for updating an object definition.
type UpdateObjectInput struct {
	Label       string `json:"label"`
	PluralLabel string `json:"plural_label"`
	Description string `json:"description"`

	IsVisibleInSetup      bool `json:"is_visible_in_setup"`
	IsCustomFieldsAllowed bool `json:"is_custom_fields_allowed"`
	IsDeleteableObject    bool `json:"is_deleteable_object"`

	IsCreateable bool `json:"is_createable"`
	IsUpdateable bool `json:"is_updateable"`
	IsDeleteable bool `json:"is_deleteable"`
	IsQueryable  bool `json:"is_queryable"`
	IsSearchable bool `json:"is_searchable"`

	HasActivities      bool `json:"has_activities"`
	HasNotes           bool `json:"has_notes"`
	HasHistoryTracking bool `json:"has_history_tracking"`
	HasSharingRules    bool `json:"has_sharing_rules"`
}

// CreateFieldInput contains the input data for creating a field definition.
type CreateFieldInput struct {
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
	IsCustom           bool          `json:"is_custom"`
	SortOrder          int           `json:"sort_order"`
}

// UpdateFieldInput contains the input data for updating a field definition.
type UpdateFieldInput struct {
	Label       string      `json:"label"`
	Description string      `json:"description"`
	HelpText    string      `json:"help_text"`
	IsRequired  bool        `json:"is_required"`
	IsUnique    bool        `json:"is_unique"`
	Config      FieldConfig `json:"config"`
	SortOrder   int         `json:"sort_order"`
}

// ObjectFilter contains filtering criteria for listing objects.
type ObjectFilter struct {
	ObjectType *ObjectType `json:"object_type"`
	Page       int32       `json:"page"`
	PerPage    int32       `json:"per_page"`
}
