package ddl

import (
	"strings"
	"testing"
)

func TestCreateObjectTable(t *testing.T) {
	t.Parallel()
	sql := CreateObjectTable("obj_invoice")
	tests := []struct {
		name     string
		contains string
	}{
		{"has CREATE TABLE", "CREATE TABLE"},
		{"has table name", `"obj_invoice"`},
		{"has id column", "id"},
		{"has owner_id", "owner_id"},
		{"has created_by", "created_by"},
		{"has created_at", "created_at"},
		{"has updated_by", "updated_by"},
		{"has updated_at", "updated_at"},
		{"has gen_random_uuid", "gen_random_uuid()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(sql, tt.contains) {
				t.Errorf("CreateObjectTable() missing %q in:\n%s", tt.contains, sql)
			}
		})
	}
}

func TestDropObjectTable(t *testing.T) {
	t.Parallel()
	sql := DropObjectTable("obj_invoice")
	if !strings.Contains(sql, "DROP TABLE") || !strings.Contains(sql, `"obj_invoice"`) {
		t.Errorf("DropObjectTable() = %q, want DROP TABLE with table name", sql)
	}
}

func TestAddColumn(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string { return &s }
	intPtr := func(n int) *int { return &n }

	tests := []struct {
		name          string
		tableName     string
		field         FieldInfo
		refTableName  string
		wantStmtCount int
		wantContains  []string
	}{
		{
			name:      "simple text column",
			tableName: "obj_contact",
			field: FieldInfo{
				APIName:      "first_name",
				FieldType:    "text",
				FieldSubtype: "plain",
				MaxLength:    intPtr(100),
			},
			wantStmtCount: 1,
			wantContains:  []string{"ALTER TABLE", "ADD COLUMN", "VARCHAR(100)"},
		},
		{
			name:      "required boolean with UNIQUE",
			tableName: "obj_contact",
			field: FieldInfo{
				APIName:    "is_active",
				FieldType:  "boolean",
				IsRequired: true,
				IsUnique:   true,
			},
			wantStmtCount: 2,
			wantContains:  []string{"NOT NULL", "UNIQUE"},
		},
		{
			name:      "association reference with FK and index",
			tableName: "obj_contact",
			field: FieldInfo{
				APIName:      "account_id",
				FieldType:    "reference",
				FieldSubtype: "association",
			},
			refTableName:  "obj_account",
			wantStmtCount: 2,
			wantContains:  []string{"REFERENCES", "SET NULL", "CREATE INDEX"},
		},
		{
			name:      "composition reference with CASCADE",
			tableName: "obj_line_item",
			field: FieldInfo{
				APIName:      "deal_id",
				FieldType:    "reference",
				FieldSubtype: "composition",
			},
			refTableName:  "obj_deal",
			wantStmtCount: 2,
			wantContains:  []string{"NOT NULL", "CASCADE", "CREATE INDEX"},
		},
		{
			name:      "polymorphic reference two columns + composite index",
			tableName: "obj_task",
			field: FieldInfo{
				APIName:      "what",
				FieldType:    "reference",
				FieldSubtype: "polymorphic",
			},
			wantStmtCount: 3,
			wantContains:  []string{"what_object_type", "what_record_id", "CREATE INDEX"},
		},
		{
			name:      "association with restrict on_delete",
			tableName: "obj_contact",
			field: FieldInfo{
				APIName:      "parent_id",
				FieldType:    "reference",
				FieldSubtype: "association",
				OnDelete:     strPtr("restrict"),
			},
			refTableName:  "obj_contact",
			wantStmtCount: 2,
			wantContains:  []string{"RESTRICT"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stmts, err := AddColumn(tt.tableName, tt.field, tt.refTableName)
			if err != nil {
				t.Fatalf("AddColumn() error = %v", err)
			}
			if len(stmts) != tt.wantStmtCount {
				t.Errorf("AddColumn() returned %d statements, want %d: %v", len(stmts), tt.wantStmtCount, stmts)
			}
			joined := strings.Join(stmts, " ")
			for _, c := range tt.wantContains {
				if !strings.Contains(joined, c) {
					t.Errorf("AddColumn() missing %q in: %s", c, joined)
				}
			}
		})
	}
}

func TestDropColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     FieldInfo
		wantCount int
	}{
		{
			name: "regular column drops 1",
			field: FieldInfo{
				APIName:   "first_name",
				FieldType: "text",
			},
			wantCount: 1,
		},
		{
			name: "polymorphic drops 2 columns",
			field: FieldInfo{
				APIName:      "what",
				FieldType:    "reference",
				FieldSubtype: "polymorphic",
			},
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stmts := DropColumn("obj_test", tt.field)
			if len(stmts) != tt.wantCount {
				t.Errorf("DropColumn() returned %d statements, want %d", len(stmts), tt.wantCount)
			}
		})
	}
}
