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
				View: OVViewConfig{},
			},
		},
		{
			name: "valid: simple fields without queries",
			config: OVConfig{
				View: OVViewConfig{
					Fields: []OVViewField{{Name: "name"}, {Name: "email"}},
				},
			},
		},
		{
			name: "valid: fields with queries and default",
			config: OVConfig{
				View: OVViewConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar", Default: true},
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
				View: OVViewConfig{
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
				View: OVViewConfig{
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
				View: OVViewConfig{
					Queries: []OVQuery{
						{Name: "", SOQL: "SELECT Id FROM X", Type: "scalar"},
					},
				},
			},
			wantErr:    true,
			errContain: "query name is required",
		},
		{
			name: "invalid: more than one default query",
			config: OVConfig{
				View: OVViewConfig{
					Queries: []OVQuery{
						{Name: "q1", SOQL: "SELECT Id FROM X", Type: "scalar", Default: true},
						{Name: "q2", SOQL: "SELECT Id FROM Y", Type: "scalar", Default: true},
					},
				},
			},
			wantErr:    true,
			errContain: "at most one query can be marked as default",
		},
		{
			name: "invalid: bad query type",
			config: OVConfig{
				View: OVViewConfig{
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
				View: OVViewConfig{
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
				View: OVViewConfig{
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
				View: OVViewConfig{
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
				View: OVViewConfig{
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
				View: OVViewConfig{
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
				View: OVViewConfig{
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
				View: OVViewConfig{
					Fields: []OVViewField{{Name: "name"}},
				},
			},
		},
		{
			name: "valid: query reference in expr",
			config: OVConfig{
				View: OVViewConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id, Name FROM Account", Type: "scalar", Default: true},
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
				View: OVViewConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar", Default: true},
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
				View: OVViewConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar", Default: true},
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
				View: OVViewConfig{
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account", Type: "scalar", Default: true},
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
