package metadata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOVConfig_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    OVConfig
		wantErr bool
	}{
		{
			name:  "new format: view only",
			input: `{"view":{"fields":["name","email"],"actions":[{"key":"send","label":"Send","type":"primary","icon":"mail","visibility_expr":""}]}}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"name", "email"},
					Actions: []OVAction{{Key: "send", Label: "Send", Type: "primary", Icon: "mail"}},
				},
			},
		},
		{
			name: "new format: view + edit",
			input: `{
				"view":{"fields":["name"],"actions":[],"queries":[{"name":"q1","soql":"SELECT Id FROM X"}],"computed":[{"name":"total","type":"float","expr":"a+b"}]},
				"edit":{"fields":["name"],"validation":[{"expr":"a>0","message":"positive","severity":"error"}],"defaults":[{"field":"status","expr":"'draft'","on":"create"}],"computed":[{"field":"total","expr":"a+b"}],"mutations":[{"dml":"INSERT INTO X"}]}
			}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:   []string{"name"},
					Actions:  []OVAction{},
					Queries:  []OVQuery{{Name: "q1", SOQL: "SELECT Id FROM X"}},
					Computed: []OVViewComputed{{Name: "total", Type: "float", Expr: "a+b"}},
				},
				Edit: &OVEditConfig{
					Fields:     []string{"name"},
					Validation: []OVValidation{{Expr: "a>0", Message: "positive", Severity: "error"}},
					Defaults:   []OVDefault{{Field: "status", Expr: "'draft'", On: "create"}},
					Computed:   []OVComputed{{Field: "total", Expr: "a+b"}},
					Mutations:  []OVMutation{{DML: "INSERT INTO X"}},
				},
			},
		},
		{
			name:  "new format: view with empty arrays",
			input: `{"view":{"fields":[],"actions":[]}}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{},
					Actions: []OVAction{},
				},
			},
		},
		{
			name:  "new format: edit is null",
			input: `{"view":{"fields":["a"],"actions":[]},"edit":null}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"a"},
					Actions: []OVAction{},
				},
			},
		},
		{
			name: "legacy nested format: read + write mapped to view + edit",
			input: `{
				"read":{"fields":["name"],"actions":[],"queries":[{"name":"q1","soql":"SELECT Id FROM X"}],"computed":[{"name":"total","type":"float","expr":"a+b"}]},
				"write":{"fields":["name"],"validation":[{"expr":"a>0","message":"positive","severity":"error"}],"defaults":[{"field":"status","expr":"'draft'","on":"create"}],"computed":[{"field":"total","expr":"a+b"}],"mutations":[{"dml":"INSERT INTO X"}]}
			}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:   []string{"name"},
					Actions:  []OVAction{},
					Queries:  []OVQuery{{Name: "q1", SOQL: "SELECT Id FROM X"}},
					Computed: []OVViewComputed{{Name: "total", Type: "float", Expr: "a+b"}},
				},
				Edit: &OVEditConfig{
					Fields:     []string{"name"},
					Validation: []OVValidation{{Expr: "a>0", Message: "positive", Severity: "error"}},
					Defaults:   []OVDefault{{Field: "status", Expr: "'draft'", On: "create"}},
					Computed:   []OVComputed{{Field: "total", Expr: "a+b"}},
					Mutations:  []OVMutation{{DML: "INSERT INTO X"}},
				},
			},
		},
		{
			name:  "legacy nested format: read only",
			input: `{"read":{"fields":["name","email"],"actions":[{"key":"send","label":"Send","type":"primary","icon":"mail","visibility_expr":""}]}}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"name", "email"},
					Actions: []OVAction{{Key: "send", Label: "Send", Type: "primary", Icon: "mail"}},
				},
			},
		},
		{
			name:  "legacy flat format: fields and actions only",
			input: `{"fields":["name","email"],"actions":[{"key":"send","label":"Send","type":"primary","icon":"mail","visibility_expr":""}]}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"name", "email"},
					Actions: []OVAction{{Key: "send", Label: "Send", Type: "primary", Icon: "mail"}},
				},
			},
		},
		{
			name:  "legacy flat format: with virtual_fields converted to view.computed",
			input: `{"fields":["amount"],"actions":[],"virtual_fields":[{"name":"total_tax","type":"float","expr":"record.amount * 0.2","when":"has(record.amount)"}]}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"amount"},
					Actions: []OVAction{},
					Computed: []OVViewComputed{
						{Name: "total_tax", Type: "float", Expr: "record.amount * 0.2", When: "has(record.amount)"},
					},
				},
			},
		},
		{
			name: "legacy flat format: with edit concerns creates edit config",
			input: `{
				"fields":["name"],
				"actions":[],
				"validation":[{"expr":"record.name != ''","message":"Name required","severity":"error"}],
				"defaults":[{"field":"status","expr":"'new'","on":"create"}],
				"computed":[{"field":"total","expr":"a+b"}],
				"mutations":[{"dml":"UPDATE X SET y=1"}]
			}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"name"},
					Actions: []OVAction{},
				},
				Edit: &OVEditConfig{
					Validation: []OVValidation{{Expr: "record.name != ''", Message: "Name required", Severity: "error"}},
					Defaults:   []OVDefault{{Field: "status", Expr: "'new'", On: "create"}},
					Computed:   []OVComputed{{Field: "total", Expr: "a+b"}},
					Mutations:  []OVMutation{{DML: "UPDATE X SET y=1"}},
				},
			},
		},
		{
			name:  "legacy flat format: only validation creates edit config",
			input: `{"fields":[],"actions":[],"validation":[{"expr":"a>0","message":"must be positive","severity":"error"}]}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{},
					Actions: []OVAction{},
				},
				Edit: &OVEditConfig{
					Validation: []OVValidation{{Expr: "a>0", Message: "must be positive", Severity: "error"}},
				},
			},
		},
		{
			name:  "legacy flat format: no edit concerns means edit is nil",
			input: `{"fields":["a","b"],"actions":[]}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"a", "b"},
					Actions: []OVAction{},
				},
			},
		},
		{
			name:  "legacy flat format: with queries",
			input: `{"fields":[],"actions":[],"queries":[{"name":"q1","soql":"SELECT Id FROM X","when":"true"}]}`,
			want: OVConfig{
				View: OVViewConfig{
					Fields:  []string{},
					Actions: []OVAction{},
					Queries: []OVQuery{{Name: "q1", SOQL: "SELECT Id FROM X", When: "true"}},
				},
			},
		},
		{
			name:  "empty object treated as legacy flat format",
			input: `{}`,
			want:  OVConfig{},
		},
		{
			name:    "invalid JSON",
			input:   `{broken`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got OVConfig
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOVConfig_MarshalJSON_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input OVConfig
	}{
		{
			name: "view only",
			input: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"name", "email"},
					Actions: []OVAction{{Key: "edit", Label: "Edit", Type: "primary"}},
				},
			},
		},
		{
			name: "view + edit",
			input: OVConfig{
				View: OVViewConfig{
					Fields:   []string{"name"},
					Actions:  []OVAction{},
					Queries:  []OVQuery{{Name: "q1", SOQL: "SELECT Id FROM X"}},
					Computed: []OVViewComputed{{Name: "total", Type: "float", Expr: "a+b", When: "true"}},
				},
				Edit: &OVEditConfig{
					Fields:     []string{"name"},
					Validation: []OVValidation{{Expr: "a>0", Message: "positive", Severity: "error"}},
					Defaults:   []OVDefault{{Field: "status", Expr: "'draft'", On: "create"}},
					Computed:   []OVComputed{{Field: "total", Expr: "a+b"}},
					Mutations:  []OVMutation{{DML: "INSERT INTO X"}},
				},
			},
		},
		{
			name: "edit nil omitted",
			input: OVConfig{
				View: OVViewConfig{
					Fields:  []string{"a"},
					Actions: []OVAction{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.input)
			require.NoError(t, err)

			var got OVConfig
			require.NoError(t, json.Unmarshal(data, &got))

			assert.Equal(t, tt.input, got)
		})
	}
}

func TestConvertVirtualFieldsToViewComputed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []OVVirtualField
		want  []OVViewComputed
	}{
		{
			name:  "nil input returns nil",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty slice returns nil",
			input: []OVVirtualField{},
			want:  nil,
		},
		{
			name: "converts all fields",
			input: []OVVirtualField{
				{Name: "total", Type: "float", Expr: "a+b", When: "true"},
				{Name: "is_active", Type: "bool", Expr: "record.status == 'active'"},
			},
			want: []OVViewComputed{
				{Name: "total", Type: "float", Expr: "a+b", When: "true"},
				{Name: "is_active", Type: "bool", Expr: "record.status == 'active'"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := convertVirtualFieldsToViewComputed(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
