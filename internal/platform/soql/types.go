package soql

import (
	"github.com/adverax/crm/internal/platform/soql/engine"
)

// FieldInfo describes a single field in SOQL query output.
type FieldInfo struct {
	Name string `json:"name"` // SOQL field name or alias
	Type string `json:"type"` // "string", "integer", "float", "boolean", "date", "datetime", "id", "object", "array"
}

// DescribeResult contains metadata about a SOQL query without executing it.
type DescribeResult struct {
	Object string      `json:"object"` // Root SOQL object name
	Fields []FieldInfo `json:"fields"` // Output fields with types
	IsRow  bool        `json:"isRow"`  // SELECT ROW flag
}

// QueryResult represents the result of executing a SOQL query.
type QueryResult struct {
	// Fields describes the output fields with their types.
	Fields []FieldInfo `json:"fields"`

	// TotalSize is the total number of records matching the query.
	TotalSize int `json:"totalSize"`

	// Done indicates whether all records have been returned.
	Done bool `json:"done"`

	// Records contains the query results as maps.
	Records []map[string]any `json:"records"`

	// NextCursor is the cursor for fetching the next page.
	// Empty if Done is true.
	NextCursor string `json:"nextRecordsUrl,omitempty"`

	// IsRow indicates that this was a SELECT ROW query.
	// When true, Records contains at most one record.
	IsRow bool `json:"isRow,omitempty"`
}

// shapeToFieldInfo converts engine ResultShape fields to public FieldInfo slice.
func shapeToFieldInfo(shape *engine.ResultShape) []FieldInfo {
	if shape == nil || len(shape.Fields) == 0 {
		return nil
	}

	fields := make([]FieldInfo, len(shape.Fields))
	for i, f := range shape.Fields {
		fields[i] = FieldInfo{
			Name: f.Name,
			Type: fieldTypeToString(f.Type),
		}
	}
	return fields
}

// fieldTypeToString converts engine.FieldType to a JSON-friendly string.
func fieldTypeToString(ft engine.FieldType) string {
	if ft.IsArray() {
		return "array"
	}
	switch ft.Base() {
	case engine.FieldTypeString:
		return "string"
	case engine.FieldTypeInteger:
		return "integer"
	case engine.FieldTypeFloat:
		return "float"
	case engine.FieldTypeBoolean:
		return "boolean"
	case engine.FieldTypeDate:
		return "date"
	case engine.FieldTypeDateTime:
		return "datetime"
	case engine.FieldTypeID:
		return "id"
	case engine.FieldTypeObject:
		return "object"
	default:
		return "string"
	}
}

// QueryParams contains parameters for query execution.
type QueryParams struct {
	// PageSize overrides the LIMIT in the query.
	PageSize int
}

const (
	DefaultPageSize = 100
	MaxPageSize     = 2000
	MinPageSize     = 1
)
