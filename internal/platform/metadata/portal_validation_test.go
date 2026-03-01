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
