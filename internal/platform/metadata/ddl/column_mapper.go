package ddl

import "fmt"

// FieldInfo contains the field metadata needed for DDL generation.
type FieldInfo struct {
	APIName      string
	FieldType    string
	FieldSubtype string
	IsRequired   bool
	IsUnique     bool
	MaxLength    *int
	Precision    *int
	Scale        *int
	OnDelete     *string
	DefaultValue *string
}

// ColumnDef represents a DDL column definition.
type ColumnDef struct {
	Name     string
	DataType string
	NotNull  bool
	Default  string
	FKTable  string
	FKColumn string
	OnDelete string
}

// MapFieldToColumn maps a FieldInfo to one or more DDL column definitions.
func MapFieldToColumn(field FieldInfo) ([]ColumnDef, error) {
	if field.FieldType == "reference" && field.FieldSubtype == "polymorphic" {
		return mapPolymorphicColumns(field)
	}
	col, err := mapSingleColumn(field)
	if err != nil {
		return nil, err
	}
	return []ColumnDef{col}, nil
}

func mapSingleColumn(field FieldInfo) (ColumnDef, error) {
	col := ColumnDef{
		Name:    field.APIName,
		NotNull: field.IsRequired,
	}

	switch field.FieldType {
	case "text":
		col.DataType = mapTextType(field)
	case "number":
		dt, err := mapNumberType(field)
		if err != nil {
			return ColumnDef{}, err
		}
		col.DataType = dt
	case "boolean":
		col.DataType = "BOOLEAN"
		if field.DefaultValue != nil {
			col.Default = *field.DefaultValue
		}
	case "datetime":
		col.DataType = mapDatetimeType(field)
	case "picklist":
		col.DataType = mapPicklistType(field)
	case "reference":
		return mapReferenceColumn(field)
	default:
		return ColumnDef{}, fmt.Errorf("ddl.mapSingleColumn: unsupported field type: %s", field.FieldType)
	}

	return col, nil
}

func mapTextType(field FieldInfo) string {
	switch field.FieldSubtype {
	case "plain":
		maxLen := 255
		if field.MaxLength != nil {
			maxLen = *field.MaxLength
		}
		return fmt.Sprintf("VARCHAR(%d)", maxLen)
	case "email":
		return "VARCHAR(255)"
	case "phone":
		return "VARCHAR(40)"
	case "url":
		return "VARCHAR(2048)"
	default:
		return "TEXT"
	}
}

func mapNumberType(field FieldInfo) (string, error) {
	switch field.FieldSubtype {
	case "integer":
		precision := 18
		if field.Precision != nil {
			precision = *field.Precision
		}
		return fmt.Sprintf("NUMERIC(%d,0)", precision), nil
	case "decimal":
		precision := 18
		scale := 2
		if field.Precision != nil {
			precision = *field.Precision
		}
		if field.Scale != nil {
			scale = *field.Scale
		}
		return fmt.Sprintf("NUMERIC(%d,%d)", precision, scale), nil
	case "currency":
		return "NUMERIC(18,2)", nil
	case "percent":
		return "NUMERIC(5,2)", nil
	case "auto_number":
		return "INTEGER GENERATED ALWAYS AS IDENTITY", nil
	default:
		return "NUMERIC(18,0)", nil
	}
}

func mapDatetimeType(field FieldInfo) string {
	switch field.FieldSubtype {
	case "date":
		return "DATE"
	case "datetime":
		return "TIMESTAMPTZ"
	case "time":
		return "TIME"
	default:
		return "TIMESTAMPTZ"
	}
}

func mapPicklistType(field FieldInfo) string {
	if field.FieldSubtype == "multi" {
		return "TEXT[]"
	}
	return "VARCHAR(255)"
}

func mapReferenceColumn(field FieldInfo) (ColumnDef, error) {
	col := ColumnDef{
		Name:     field.APIName,
		DataType: "UUID",
	}

	switch field.FieldSubtype {
	case "association":
		onDelete := "SET NULL"
		if field.OnDelete != nil && *field.OnDelete == "restrict" {
			onDelete = "RESTRICT"
		}
		col.OnDelete = onDelete
	case "composition":
		col.NotNull = true
		onDelete := "CASCADE"
		if field.OnDelete != nil && *field.OnDelete == "restrict" {
			onDelete = "RESTRICT"
		}
		col.OnDelete = onDelete
	default:
		return ColumnDef{}, fmt.Errorf("ddl.mapReferenceColumn: unexpected subtype %s for non-polymorphic reference", field.FieldSubtype)
	}

	return col, nil
}

func mapPolymorphicColumns(field FieldInfo) ([]ColumnDef, error) {
	baseName := field.APIName
	return []ColumnDef{
		{
			Name:     baseName + "_object_type",
			DataType: "VARCHAR(100)",
			NotNull:  true,
		},
		{
			Name:     baseName + "_record_id",
			DataType: "UUID",
			NotNull:  true,
		},
	}, nil
}
