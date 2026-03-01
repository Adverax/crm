package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateViewConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     OVConfig
		wantErr    bool
		errContain string
	}{
		{
			name: "valid: empty config",
			config: OVConfig{
				Read: OVReadConfig{},
			},
		},
		{
			name: "valid: simple fields without queries",
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{{Name: "name"}, {Name: "email"}},
				},
			},
		},
		{
			name: "valid: fields with queries",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar"},
						{Name: "contacts", SOQL: "SELECT Id FROM Contact", Type: "list"},
					},
					Fields: []OVViewField{
						{Name: "name"},
						{Name: "contact_count", Type: "int", Expr: "size(contacts)"},
					},
				},
			},
		},
		{
			name: "valid: DAG fields A -> B -> C",
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{
						{Name: "a"},
						{Name: "b", Expr: "a + 1"},
						{Name: "c", Expr: "b + 1"},
					},
				},
			},
		},
		{
			name: "invalid: duplicate query name",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM X", Type: "scalar"},
						{Name: "main", SOQL: "SELECT Id FROM Y", Type: "list"},
					},
				},
			},
			wantErr:    true,
			errContain: "duplicate query name: main",
		},
		{
			name: "invalid: empty query name",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "", SOQL: "SELECT Id FROM X", Type: "scalar"},
					},
				},
			},
			wantErr:    true,
			errContain: "query name is required",
		},
		{
			name: "valid: multiple scalar queries (first is implicit default)",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "q1", SOQL: "SELECT Id FROM X", Type: "scalar"},
						{Name: "q2", SOQL: "SELECT Id FROM Y", Type: "scalar"},
					},
				},
			},
		},
		{
			name: "invalid: bad query type",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "q1", SOQL: "SELECT Id FROM X", Type: "batch"},
					},
				},
			},
			wantErr:    true,
			errContain: "type must be 'scalar' or 'list'",
		},
		{
			name: "invalid: duplicate field name",
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{
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
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{
						{Name: ""},
					},
				},
			},
			wantErr:    true,
			errContain: "field name is required",
		},
		{
			name: "invalid: field references non-existent query",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM X", Type: "scalar"},
					},
					Fields: []OVViewField{
						{Name: "total", Expr: "other.Amount * 1.2"},
					},
				},
			},
			wantErr:    true,
			errContain: "references unknown query",
		},
		{
			name: "invalid: direct cycle A -> B -> A",
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{
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
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{
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
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{
						{Name: "a", Expr: "a + 1"},
					},
				},
			},
			wantErr:    true,
			errContain: "circular dependency",
		},
		{
			name: "valid: no default query (zero queries)",
			config: OVConfig{
				Read: OVReadConfig{
					Fields: []OVViewField{{Name: "name"}},
				},
			},
		},
		{
			name: "valid: query reference in expr",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id, Name FROM Account", Type: "scalar"},
					},
					Fields: []OVViewField{
						{Name: "display", Type: "string", Expr: "main.Name"},
					},
				},
			},
		},
		{
			name: "valid: scalar query reference in computed field",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar"},
						{Name: "stats", SOQL: "SELECT COUNT(Id) AS total FROM Contact WHERE AccountId = :id", Type: "scalar"},
					},
					Fields: []OVViewField{
						{Name: "name"},
						{Name: "contact_count", Type: "int", Expr: "stats.total"},
					},
				},
			},
		},
		{
			name: "invalid: field expr references list query",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar"},
						{Name: "contacts", SOQL: "SELECT Id FROM Contact", Type: "list"},
					},
					Fields: []OVViewField{
						{Name: "first_contact", Expr: "contacts.Name"},
					},
				},
			},
			wantErr:    true,
			errContain: "references list query",
		},
		{
			name: "invalid: field expr references list query with multiple fields",
			config: OVConfig{
				Read: OVReadConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar"},
						{Name: "deals", SOQL: "SELECT Id, Amount FROM Deal", Type: "list"},
					},
					Fields: []OVViewField{
						{Name: "name"},
						{Name: "deal_amount", Type: "float", Expr: "deals.Amount * 1.1"},
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
