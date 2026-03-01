package soql

import (
	"testing"

	"github.com/adverax/crm/internal/platform/soql/engine"
)

func TestShapeToFieldInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		shape *engine.ResultShape
		want  []FieldInfo
	}{
		{
			name:  "nil shape returns nil",
			shape: nil,
			want:  nil,
		},
		{
			name: "empty fields returns nil",
			shape: &engine.ResultShape{
				Object: "Account",
				Fields: []*engine.FieldShape{},
			},
			want: nil,
		},
		{
			name: "maps all basic field types",
			shape: &engine.ResultShape{
				Object: "Account",
				Fields: []*engine.FieldShape{
					{Name: "id", Column: "id", Type: engine.FieldTypeID},
					{Name: "name", Column: "name", Type: engine.FieldTypeString},
					{Name: "age", Column: "age", Type: engine.FieldTypeInteger},
					{Name: "revenue", Column: "revenue", Type: engine.FieldTypeFloat},
					{Name: "is_active", Column: "is_active", Type: engine.FieldTypeBoolean},
					{Name: "birth_date", Column: "birth_date", Type: engine.FieldTypeDate},
					{Name: "created_at", Column: "created_at", Type: engine.FieldTypeDateTime},
					{Name: "details", Column: "details", Type: engine.FieldTypeObject},
				},
			},
			want: []FieldInfo{
				{Name: "id", Type: "id"},
				{Name: "name", Type: "string"},
				{Name: "age", Type: "integer"},
				{Name: "revenue", Type: "float"},
				{Name: "is_active", Type: "boolean"},
				{Name: "birth_date", Type: "date"},
				{Name: "created_at", Type: "datetime"},
				{Name: "details", Type: "object"},
			},
		},
		{
			name: "array type maps to array",
			shape: &engine.ResultShape{
				Object: "Account",
				Fields: []*engine.FieldShape{
					{Name: "tags", Column: "tags", Type: engine.FieldTypeArray | engine.FieldTypeString},
				},
			},
			want: []FieldInfo{
				{Name: "tags", Type: "array"},
			},
		},
		{
			name: "unknown type defaults to string",
			shape: &engine.ResultShape{
				Object: "Account",
				Fields: []*engine.FieldShape{
					{Name: "unknown_field", Column: "unknown_field", Type: engine.FieldTypeUnknown},
				},
			},
			want: []FieldInfo{
				{Name: "unknown_field", Type: "string"},
			},
		},
		{
			name: "null type defaults to string",
			shape: &engine.ResultShape{
				Object: "Account",
				Fields: []*engine.FieldShape{
					{Name: "null_field", Column: "null_field", Type: engine.FieldTypeNull},
				},
			},
			want: []FieldInfo{
				{Name: "null_field", Type: "string"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := shapeToFieldInfo(tt.shape)

			if tt.want == nil {
				if got != nil {
					t.Errorf("shapeToFieldInfo() = %v, want nil", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("shapeToFieldInfo() returned %d fields, want %d", len(got), len(tt.want))
			}

			for i, wantField := range tt.want {
				if got[i].Name != wantField.Name {
					t.Errorf("field[%d].Name = %q, want %q", i, got[i].Name, wantField.Name)
				}
				if got[i].Type != wantField.Type {
					t.Errorf("field[%d].Type = %q, want %q", i, got[i].Type, wantField.Type)
				}
			}
		})
	}
}

func TestFieldTypeToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ft       engine.FieldType
		expected string
	}{
		{name: "string", ft: engine.FieldTypeString, expected: "string"},
		{name: "integer", ft: engine.FieldTypeInteger, expected: "integer"},
		{name: "float", ft: engine.FieldTypeFloat, expected: "float"},
		{name: "boolean", ft: engine.FieldTypeBoolean, expected: "boolean"},
		{name: "date", ft: engine.FieldTypeDate, expected: "date"},
		{name: "datetime", ft: engine.FieldTypeDateTime, expected: "datetime"},
		{name: "id", ft: engine.FieldTypeID, expected: "id"},
		{name: "object", ft: engine.FieldTypeObject, expected: "object"},
		{name: "unknown defaults to string", ft: engine.FieldTypeUnknown, expected: "string"},
		{name: "null defaults to string", ft: engine.FieldTypeNull, expected: "string"},
		{name: "array flag returns array", ft: engine.FieldTypeArray, expected: "array"},
		{name: "array|string returns array", ft: engine.FieldTypeArray | engine.FieldTypeString, expected: "array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := fieldTypeToString(tt.ft)
			if got != tt.expected {
				t.Errorf("fieldTypeToString(%v) = %q, want %q", tt.ft, got, tt.expected)
			}
		})
	}
}
