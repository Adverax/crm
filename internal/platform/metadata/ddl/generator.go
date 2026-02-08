package ddl

import (
	"fmt"
	"strings"
)

// CreateObjectTable generates DDL to create a table for an object.
func CreateObjectTable(tableName string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID        NOT NULL,
    created_by  UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by  UUID        NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
)`, quoteIdent(tableName))
}

// DropObjectTable generates DDL to drop a table.
func DropObjectTable(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", quoteIdent(tableName))
}

// AddColumn generates DDL to add a column to a table.
func AddColumn(tableName string, field FieldInfo, referencedTableName string) ([]string, error) {
	columns, err := MapFieldToColumn(field)
	if err != nil {
		return nil, fmt.Errorf("ddl.AddColumn: %w", err)
	}

	var statements []string

	for _, col := range columns {
		stmt := buildAddColumnStatement(tableName, col, referencedTableName)
		statements = append(statements, stmt)
	}

	if field.FieldType == "reference" {
		if field.FieldSubtype == "polymorphic" {
			idxName := fmt.Sprintf("idx_%s_%s", sanitizeForIndex(tableName), field.APIName)
			statements = append(statements, fmt.Sprintf(
				"CREATE INDEX IF NOT EXISTS %s ON %s (%s, %s)",
				quoteIdent(idxName),
				quoteIdent(tableName),
				quoteIdent(field.APIName+"_object_type"),
				quoteIdent(field.APIName+"_record_id"),
			))
		} else {
			idxName := fmt.Sprintf("idx_%s_%s", sanitizeForIndex(tableName), field.APIName)
			statements = append(statements, fmt.Sprintf(
				"CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
				quoteIdent(idxName),
				quoteIdent(tableName),
				quoteIdent(field.APIName),
			))
		}
	}

	if field.IsUnique {
		cName := fmt.Sprintf("uq_%s_%s", sanitizeForIndex(tableName), field.APIName)
		statements = append(statements, fmt.Sprintf(
			"ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
			quoteIdent(tableName),
			quoteIdent(cName),
			quoteIdent(field.APIName),
		))
	}

	return statements, nil
}

func buildAddColumnStatement(tableName string, col ColumnDef, refTable string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ALTER TABLE %s ADD COLUMN %s %s",
		quoteIdent(tableName), quoteIdent(col.Name), col.DataType)

	if col.NotNull {
		b.WriteString(" NOT NULL")
	}

	if col.Default != "" {
		fmt.Fprintf(&b, " DEFAULT %s", col.Default)
	}

	if refTable != "" && col.OnDelete != "" {
		fmt.Fprintf(&b, " REFERENCES %s(id) ON DELETE %s",
			quoteIdent(refTable), col.OnDelete)
	}

	return b.String()
}

// DropColumn generates DDL to drop a column (or two columns for polymorphic).
func DropColumn(tableName string, field FieldInfo) []string {
	if field.FieldType == "reference" && field.FieldSubtype == "polymorphic" {
		return []string{
			fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s",
				quoteIdent(tableName), quoteIdent(field.APIName+"_object_type")),
			fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s",
				quoteIdent(tableName), quoteIdent(field.APIName+"_record_id")),
		}
	}
	return []string{
		fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s",
			quoteIdent(tableName), quoteIdent(field.APIName)),
	}
}

// AlterColumnSetNotNull generates DDL to set NOT NULL constraint.
func AlterColumnSetNotNull(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL",
		quoteIdent(tableName), quoteIdent(columnName))
}

// AlterColumnDropNotNull generates DDL to drop NOT NULL constraint.
func AlterColumnDropNotNull(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL",
		quoteIdent(tableName), quoteIdent(columnName))
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func sanitizeForIndex(tableName string) string {
	name := strings.ReplaceAll(tableName, `"`, "")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
