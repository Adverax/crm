package soql

import (
	"context"
	"testing"

	"github.com/adverax/crm/internal/platform/soql/engine"
)

func TestQueryService_Describe(t *testing.T) {
	t.Parallel()

	accountMeta := engine.NewObjectMeta("Account", "public", "obj_account").
		Field("Id", "id", engine.FieldTypeID).
		Field("Name", "name", engine.FieldTypeString).
		Field("Revenue", "revenue", engine.FieldTypeFloat).
		Field("IsActive", "is_active", engine.FieldTypeBoolean).
		Field("CreatedAt", "created_at", engine.FieldTypeDateTime).
		Build()

	metadata := engine.NewStaticMetadataProvider(map[string]*engine.ObjectMeta{
		"Account": accountMeta,
	})

	eng := engine.NewEngine(engine.WithMetadata(metadata))

	svc := NewQueryService(eng, nil)

	tests := []struct {
		name       string
		query      string
		wantObj    string
		wantFields []FieldInfo
		wantIsRow  bool
		wantErr    bool
	}{
		{
			name:    "returns field metadata for simple query",
			query:   "SELECT Id, Name FROM Account",
			wantObj: "Account",
			wantFields: []FieldInfo{
				{Name: "Id", Type: "id"},
				{Name: "Name", Type: "string"},
			},
		},
		{
			name:    "returns isRow true for SELECT ROW",
			query:   "SELECT ROW Id, Name FROM Account WHERE Id = '00000000-0000-0000-0000-000000000001'",
			wantObj: "Account",
			wantFields: []FieldInfo{
				{Name: "Id", Type: "id"},
				{Name: "Name", Type: "string"},
			},
			wantIsRow: true,
		},
		{
			name:    "returns error on invalid query",
			query:   "INVALID",
			wantErr: true,
		},
		{
			name:    "maps mixed field types",
			query:   "SELECT Id, Revenue, IsActive, CreatedAt FROM Account",
			wantObj: "Account",
			wantFields: []FieldInfo{
				{Name: "Id", Type: "id"},
				{Name: "Revenue", Type: "float"},
				{Name: "IsActive", Type: "boolean"},
				{Name: "CreatedAt", Type: "datetime"},
			},
		},
		{
			name:    "returns error for unknown object",
			query:   "SELECT Id FROM UnknownObject",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := svc.Describe(context.Background(), tt.query)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Describe() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Describe() unexpected error: %v", err)
			}

			if result.Object != tt.wantObj {
				t.Errorf("Object = %q, want %q", result.Object, tt.wantObj)
			}
			if result.IsRow != tt.wantIsRow {
				t.Errorf("IsRow = %v, want %v", result.IsRow, tt.wantIsRow)
			}
			if len(result.Fields) != len(tt.wantFields) {
				t.Fatalf("Fields count = %d, want %d", len(result.Fields), len(tt.wantFields))
			}
			for i, wf := range tt.wantFields {
				if result.Fields[i].Name != wf.Name {
					t.Errorf("Fields[%d].Name = %q, want %q", i, result.Fields[i].Name, wf.Name)
				}
				if result.Fields[i].Type != wf.Type {
					t.Errorf("Fields[%d].Type = %q, want %q", i, result.Fields[i].Type, wf.Type)
				}
			}
		})
	}
}
