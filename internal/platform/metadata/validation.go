package metadata

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/adverax/crm/internal/pkg/apperror"
)

var (
	apiNamePattern   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,98}[a-zA-Z0-9]$`)
	customAPISuffix  = "__c"
	tableNamePrefix  = "obj_"
	maxAPINameLength = 100
	reservedWords    = map[string]bool{
		"id": true, "created_at": true, "updated_at": true,
		"owner_id": true, "created_by": true, "updated_by": true,
		"select": true, "from": true, "where": true, "insert": true,
		"update": true, "delete": true, "table": true, "index": true,
		"create": true, "drop": true, "alter": true, "grant": true,
		"order": true, "group": true, "having": true, "limit": true,
		"offset": true, "join": true, "union": true, "null": true,
		"true": true, "false": true, "and": true, "or": true, "not": true,
	}
)

// ValidateObjectDefinition validates an ObjectDefinition for creation.
func ValidateObjectDefinition(obj *ObjectDefinition) error {
	if obj.APIName == "" {
		return apperror.Validation("api_name is required")
	}
	if len(obj.APIName) > maxAPINameLength {
		return apperror.Validation(fmt.Sprintf("api_name must be at most %d characters", maxAPINameLength))
	}
	if !apiNamePattern.MatchString(obj.APIName) {
		return apperror.Validation("api_name must start with a letter, contain only letters, digits, and underscores, and end with a letter or digit")
	}
	if reservedWords[strings.ToLower(obj.APIName)] {
		return apperror.Validation(fmt.Sprintf("api_name '%s' is a reserved word", obj.APIName))
	}

	if obj.ObjectType == ObjectTypeCustom && !strings.HasSuffix(obj.APIName, customAPISuffix) {
		return apperror.Validation("custom object api_name must end with '__c'")
	}
	if obj.ObjectType == ObjectTypeStandard && strings.HasSuffix(obj.APIName, customAPISuffix) {
		return apperror.Validation("standard object api_name must not end with '__c'")
	}

	if obj.ObjectType != ObjectTypeStandard && obj.ObjectType != ObjectTypeCustom {
		return apperror.Validation(fmt.Sprintf("object_type must be '%s' or '%s'", ObjectTypeStandard, ObjectTypeCustom))
	}

	if obj.Label == "" {
		return apperror.Validation("label is required")
	}
	if obj.PluralLabel == "" {
		return apperror.Validation("plural_label is required")
	}

	return nil
}

// ValidateFieldDefinition validates a FieldDefinition for creation.
func ValidateFieldDefinition(field *FieldDefinition) error {
	if field.APIName == "" {
		return apperror.Validation("api_name is required")
	}
	if len(field.APIName) > maxAPINameLength {
		return apperror.Validation(fmt.Sprintf("api_name must be at most %d characters", maxAPINameLength))
	}
	if !apiNamePattern.MatchString(field.APIName) {
		return apperror.Validation("api_name must start with a letter, contain only letters, digits, and underscores, and end with a letter or digit")
	}
	if reservedWords[strings.ToLower(field.APIName)] {
		return apperror.Validation(fmt.Sprintf("api_name '%s' is a reserved word", field.APIName))
	}

	if field.IsCustom && !strings.HasSuffix(field.APIName, customAPISuffix) {
		return apperror.Validation("custom field api_name must end with '__c'")
	}

	if field.Label == "" {
		return apperror.Validation("label is required")
	}

	spec, ok := LookupTypeSpec(field.FieldType, field.FieldSubtype)
	if !ok {
		sub := "null"
		if field.FieldSubtype != nil {
			sub = string(*field.FieldSubtype)
		}
		return apperror.Validation(fmt.Sprintf("invalid type/subtype combination: %s/%s", field.FieldType, sub))
	}

	if err := validateFieldConfig(field, spec); err != nil {
		return err
	}

	if err := validateReferenceField(field); err != nil {
		return err
	}

	return nil
}

func validateFieldConfig(field *FieldDefinition, spec TypeSpec) error {
	configBytes, err := json.Marshal(field.Config)
	if err != nil {
		return apperror.Validation("invalid config")
	}
	var configMap map[string]interface{}
	if err := json.Unmarshal(configBytes, &configMap); err != nil {
		return apperror.Validation("invalid config format")
	}

	for _, key := range spec.RequiredConfig {
		if _, ok := configMap[key]; !ok {
			return apperror.Validation(fmt.Sprintf("config.%s is required for %s/%s",
				key, field.FieldType, subtypeStr(field.FieldSubtype)))
		}
	}

	if field.FieldType == FieldTypeText && field.FieldSubtype != nil && *field.FieldSubtype == SubtypePlain {
		if field.Config.MaxLength != nil {
			if *field.Config.MaxLength < 1 || *field.Config.MaxLength > 255 {
				return apperror.Validation("config.max_length must be between 1 and 255 for text/plain")
			}
		}
	}

	if field.FieldType == FieldTypeNumber {
		if field.Config.Precision != nil && (*field.Config.Precision < 1 || *field.Config.Precision > 18) {
			return apperror.Validation("config.precision must be between 1 and 18")
		}
		if field.Config.Scale != nil && (*field.Config.Scale < 0 || *field.Config.Scale > 10) {
			return apperror.Validation("config.scale must be between 0 and 10")
		}
	}

	return nil
}

func validateReferenceField(field *FieldDefinition) error {
	if field.FieldType != FieldTypeReference {
		if field.ReferencedObjectID != nil {
			return apperror.Validation("referenced_object_id is only valid for reference fields")
		}
		return nil
	}

	if field.FieldSubtype == nil {
		return apperror.Validation("reference fields must have a subtype")
	}

	sub := *field.FieldSubtype

	switch sub {
	case SubtypeAssociation, SubtypeComposition:
		if field.ReferencedObjectID == nil {
			return apperror.Validation("referenced_object_id is required for association/composition fields")
		}
	case SubtypePolymorphic:
		if field.ReferencedObjectID != nil {
			return apperror.Validation("referenced_object_id must be null for polymorphic fields")
		}
	}

	if sub == SubtypeAssociation {
		onDelete := "set_null"
		if field.Config.OnDelete != nil {
			onDelete = *field.Config.OnDelete
		}
		if onDelete != "set_null" && onDelete != "restrict" {
			return apperror.Validation("on_delete must be 'set_null' or 'restrict' for association")
		}
	}

	if sub == SubtypeComposition {
		onDelete := "cascade"
		if field.Config.OnDelete != nil {
			onDelete = *field.Config.OnDelete
		}
		if onDelete != "cascade" && onDelete != "restrict" {
			return apperror.Validation("on_delete must be 'cascade' or 'restrict' for composition")
		}
	}

	return nil
}

// GenerateTableName generates the table name for an object from its api_name.
func GenerateTableName(apiName string) string {
	name := strings.ToLower(apiName)
	name = strings.TrimSuffix(name, customAPISuffix)
	return tableNamePrefix + name
}

// ValidateVisibility validates the visibility (OWD) value.
func ValidateVisibility(v Visibility) error {
	switch v {
	case VisibilityPrivate, VisibilityPublicRead, VisibilityPublicReadWrite, VisibilityControlledByParent:
		return nil
	case "":
		return nil // empty means default (private)
	default:
		return apperror.Validation(fmt.Sprintf(
			"visibility must be one of: private, public_read, public_read_write, controlled_by_parent; got '%s'", v))
	}
}

func subtypeStr(s *FieldSubtype) string {
	if s == nil {
		return "null"
	}
	return string(*s)
}
