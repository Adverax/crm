package metadata

import (
	"github.com/adverax/crm/internal/platform/metadata/ddl"
)

// ToFieldInfo converts a FieldDefinition to a ddl.FieldInfo for DDL generation.
func ToFieldInfo(f FieldDefinition) ddl.FieldInfo {
	subtype := ""
	if f.FieldSubtype != nil {
		subtype = string(*f.FieldSubtype)
	}
	return ddl.FieldInfo{
		APIName:      f.APIName,
		FieldType:    string(f.FieldType),
		FieldSubtype: subtype,
		IsRequired:   f.IsRequired,
		IsUnique:     f.IsUnique,
		MaxLength:    f.Config.MaxLength,
		Precision:    f.Config.Precision,
		Scale:        f.Config.Scale,
		OnDelete:     f.Config.OnDelete,
		DefaultValue: f.Config.DefaultValue,
	}
}
