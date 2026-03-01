package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateViewConfig(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		config     PortalConfig
		wantErr    bool
		errContain string
	}{
		{
			name: "valid: empty config",
			config: PortalConfig{
				Read: PortalReadConfig{},
			},
		},
		// --- Args validation ---
		{
			name: "valid: single required arg",
			config: PortalConfig{
				Args: []PortalArg{{Name: "account_id", Type: "string"}},
				Read: PortalReadConfig{
					Queries: []PortalQuery{{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :account_id"}},
				},
			},
		},
		{
			name: "valid: optional arg with default",
			config: PortalConfig{
				Args: []PortalArg{{Name: "limit", Type: "int", Default: strPtr("10")}},
				Read: PortalReadConfig{},
			},
		},
		{
			name: "valid: mixed args",
			config: PortalConfig{
				Args: []PortalArg{
					{Name: "account_id", Type: "string"},
					{Name: "active", Type: "bool", Default: strPtr("true")},
					{Name: "threshold", Type: "float", Default: strPtr("1.5")},
				},
				Read: PortalReadConfig{
					Queries: []PortalQuery{{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :account_id"}},
				},
			},
		},
		{
			name: "valid: no args (empty slice)",
			config: PortalConfig{
				Args: []PortalArg{},
				Read: PortalReadConfig{},
			},
		},
		{
			name: "invalid: duplicate arg name",
			config: PortalConfig{
				Args: []PortalArg{{Name: "id", Type: "string"}, {Name: "id", Type: "int"}},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "duplicate arg name: id",
		},
		{
			name: "invalid: empty arg name",
			config: PortalConfig{
				Args: []PortalArg{{Name: "", Type: "string"}},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "arg name is required",
		},
		{
			name: "invalid: bad arg name format",
			config: PortalConfig{
				Args: []PortalArg{{Name: "MyParam", Type: "string"}},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "must match",
		},
		{
			name: "invalid: unknown arg type",
			config: PortalConfig{
				Args: []PortalArg{{Name: "val", Type: "date"}},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "type must be one of",
		},
		{
			name: "invalid: int default with non-numeric value",
			config: PortalConfig{
				Args: []PortalArg{{Name: "count", Type: "int", Default: strPtr("abc")}},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "not a valid int",
		},
		{
			name: "invalid: float default with non-numeric value",
			config: PortalConfig{
				Args: []PortalArg{{Name: "rate", Type: "float", Default: strPtr("xyz")}},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "not a valid float",
		},
		{
			name: "invalid: bool default with bad value",
			config: PortalConfig{
				Args: []PortalArg{{Name: "active", Type: "bool", Default: strPtr("yes")}},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "not a valid bool",
		},
		{
			name: "invalid: undeclared param in query",
			config: PortalConfig{
				Args: []PortalArg{{Name: "account_id", Type: "string"}},
				Read: PortalReadConfig{
					Queries: []PortalQuery{{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :unknown_id"}},
				},
			},
			wantErr:    true,
			errContain: "param :unknown_id is not declared",
		},
		{
			name: "valid: all params declared in args",
			config: PortalConfig{
				Args: []PortalArg{
					{Name: "account_id", Type: "string"},
					{Name: "status", Type: "string", Default: strPtr("active")},
				},
				Read: PortalReadConfig{
					Queries: []PortalQuery{{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :account_id AND Status = :status"}},
				},
			},
		},
		{
			name: "valid: query without params when args declared",
			config: PortalConfig{
				Args: []PortalArg{{Name: "page", Type: "int", Default: strPtr("1")}},
				Read: PortalReadConfig{
					Queries: []PortalQuery{{Name: "main", SOQL: "SELECT ROW Id FROM Account LIMIT 10"}},
				},
			},
		},
		{
			name: "valid: params in SOQL without args (backward compat, no validation)",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :id"}},
				},
			},
		},
		// --- Arg validation expressions ---
		{
			name: "valid: arg with validation and error_message",
			config: PortalConfig{
				Args: []PortalArg{
					{Name: "limit", Type: "int", Default: strPtr("10"), Validation: "args.limit > 0 && args.limit <= 100", ErrorMessage: "Limit must be between 1 and 100"},
				},
				Read: PortalReadConfig{},
			},
		},
		{
			name: "invalid: arg validation without error_message",
			config: PortalConfig{
				Args: []PortalArg{
					{Name: "limit", Type: "int", Default: strPtr("10"), Validation: "args.limit > 0"},
				},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "error_message is required when validation is set",
		},
		{
			name: "invalid: arg validation with bad CEL expression",
			config: PortalConfig{
				Args: []PortalArg{
					{Name: "limit", Type: "int", Default: strPtr("10"), Validation: "args.limit > }", ErrorMessage: "bad"},
				},
				Read: PortalReadConfig{},
			},
			wantErr:    true,
			errContain: "invalid validation expression",
		},
		{
			name: "valid: arg without validation (no error_message needed)",
			config: PortalConfig{
				Args: []PortalArg{
					{Name: "name", Type: "string"},
				},
				Read: PortalReadConfig{},
			},
		},
		{
			name: "valid: simple fields without queries",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{{Name: "name"}, {Name: "email"}},
				},
			},
		},
		{
			name: "valid: fields with queries",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id FROM Account"},
						{Name: "contacts", SOQL: "SELECT Id FROM Contact"},
					},
					Fields: []PortalViewField{
						{Name: "name"},
						{Name: "contact_count", Expr: "size(contacts)"},
					},
				},
			},
		},
		{
			name: "valid: DAG fields A -> B -> C",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{
						{Name: "a"},
						{Name: "b", Expr: "a + 1"},
						{Name: "c", Expr: "b + 1"},
					},
				},
			},
		},
		{
			name: "invalid: duplicate query name",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id FROM X"},
						{Name: "main", SOQL: "SELECT Id FROM Y"},
					},
				},
			},
			wantErr:    true,
			errContain: "duplicate query name: main",
		},
		{
			name: "invalid: empty query name",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "", SOQL: "SELECT ROW Id FROM X"},
					},
				},
			},
			wantErr:    true,
			errContain: "query name is required",
		},
		{
			name: "valid: multiple scalar queries (first is implicit default)",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "q1", SOQL: "SELECT ROW Id FROM X"},
						{Name: "q2", SOQL: "SELECT ROW Id FROM Y"},
					},
				},
			},
		},
		{
			name: "invalid: duplicate field name",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{
						{Name: "name"},
						{Name: "name"},
					},
				},
			},
			wantErr:    true,
			errContain: "duplicate field name: name",
		},
		{
			name: "invalid: empty field name",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{
						{Name: ""},
					},
				},
			},
			wantErr:    true,
			errContain: "field name is required",
		},
		{
			name: "invalid: field references non-existent query",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id FROM X"},
					},
					Fields: []PortalViewField{
						{Name: "total", Expr: "other.Amount * 1.2"},
					},
				},
			},
			wantErr:    true,
			errContain: "references unknown query",
		},
		{
			name: "invalid: direct cycle A -> B -> A",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{
						{Name: "a", Expr: "b + 1"},
						{Name: "b", Expr: "a + 1"},
					},
				},
			},
			wantErr:    true,
			errContain: "circular dependency",
		},
		{
			name: "invalid: transitive cycle A -> B -> C -> A",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{
						{Name: "a", Expr: "c + 1"},
						{Name: "b", Expr: "a + 1"},
						{Name: "c", Expr: "b + 1"},
					},
				},
			},
			wantErr:    true,
			errContain: "circular dependency",
		},
		{
			name: "invalid: self-reference",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{
						{Name: "a", Expr: "a + 1"},
					},
				},
			},
			wantErr:    true,
			errContain: "circular dependency",
		},
		{
			name: "valid: no default query (zero queries)",
			config: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{{Name: "name"}},
				},
			},
		},
		{
			name: "valid: query reference in expr",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id, Name FROM Account"},
					},
					Fields: []PortalViewField{
						{Name: "display", Expr: "main.Name"},
					},
				},
			},
		},
		{
			name: "valid: scalar query reference in computed field",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id FROM Account"},
						{Name: "stats", SOQL: "SELECT ROW COUNT(Id) AS total FROM Contact WHERE AccountId = :id"},
					},
					Fields: []PortalViewField{
						{Name: "name"},
						{Name: "contact_count", Expr: "stats.total"},
					},
				},
			},
		},
		{
			name: "invalid: field expr references list query",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id FROM Account"},
						{Name: "contacts", SOQL: "SELECT Id FROM Contact"},
					},
					Fields: []PortalViewField{
						{Name: "first_contact", Expr: "contacts.Name"},
					},
				},
			},
			wantErr:    true,
			errContain: "references list query",
		},
		{
			name: "invalid: field expr references list query with multiple fields",
			config: PortalConfig{
				Read: PortalReadConfig{
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id FROM Account"},
						{Name: "deals", SOQL: "SELECT Id, Amount FROM Deal"},
					},
					Fields: []PortalViewField{
						{Name: "name"},
						{Name: "deal_amount", Expr: "deals.Amount * 1.1"},
					},
				},
			},
			wantErr:    true,
			errContain: "references list query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateViewConfig(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
				return
			}

			require.NoError(t, err)
		})
	}
}
