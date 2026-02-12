package engine

import (
	"context"
	"strings"
	"testing"
)

func TestKeysetFieldsInSelect(t *testing.T) {
	t.Skip("TODO: Requires full SOQL parser and compiler implementation")

	metadata := setupTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	compiler := NewCompiler(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		query      string
		wantId     bool // Should Id be in SELECT
		duplicates bool // Should NOT have duplicate Id
	}{
		{"without Id", "SELECT Name FROM Account LIMIT 5", true, false},
		{"with explicit Id", "SELECT Id, Name FROM Account LIMIT 5", true, false},
		{"with ORDER BY Name", "SELECT Name FROM Account ORDER BY Name LIMIT 5", true, false},
		{"with function", "SELECT UPPER(Name) FROM Account LIMIT 5", true, false},
		{"with expression", "SELECT Name || Industry FROM Account LIMIT 5", true, false},
		{"multiple ORDER BY fields", "SELECT Name FROM Account ORDER BY Name, Industry LIMIT 5", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			validated, err := validator.Validate(ctx, ast)
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			compiled, err := compiler.Compile(validated)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}

			// Check for Id in SQL
			hasId := strings.Contains(compiled.SQL, "id")
			if tt.wantId && !hasId {
				t.Errorf("SQL should contain id: %s", compiled.SQL)
			}

			// Check for duplicates in SELECT clause only
			// id can appear multiple times (SELECT and ORDER BY), but not twice in SELECT
			selectEnd := strings.Index(compiled.SQL, "FROM")
			if selectEnd > 0 {
				selectPart := compiled.SQL[:selectEnd]
				count := strings.Count(selectPart, "id")
				if count > 1 && !tt.duplicates {
					t.Errorf("SELECT should not have duplicate id (found %d): %s", count, selectPart)
				}
			}
		})
	}
}
