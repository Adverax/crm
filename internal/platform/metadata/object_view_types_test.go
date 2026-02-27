package metadata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOVConfig_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input OVConfig
	}{
		{
			name: "view only with simple fields",
			input: OVConfig{
				View: OVViewConfig{
					Fields:  []OVViewField{{Name: "name"}, {Name: "email"}},
					Actions: []OVAction{{Key: "edit", Label: "Edit", Type: "primary", Icon: "pencil"}},
				},
			},
		},
		{
			name: "view with computed fields",
			input: OVConfig{
				View: OVViewConfig{
					Fields: []OVViewField{
						{Name: "name"},
						{Name: "total", Type: "float", Expr: "record.amount * 1.2", When: "has(record.amount)"},
					},
					Actions: []OVAction{},
				},
			},
		},
		{
			name: "view with queries",
			input: OVConfig{
				View: OVViewConfig{
					Fields:  []OVViewField{{Name: "name"}},
					Actions: []OVAction{},
					Queries: []OVQuery{
						{Name: "main", SOQL: "SELECT Id FROM Account WHERE Id = :id", Type: "scalar", Default: true},
						{Name: "contacts", SOQL: "SELECT Id FROM Contact WHERE AccountId = :id", Type: "list"},
					},
				},
			},
		},
		{
			name: "view with single default query",
			input: OVConfig{
				View: OVViewConfig{
					Fields:  []OVViewField{{Name: "name"}},
					Actions: []OVAction{},
					Queries: []OVQuery{{Name: "q1", SOQL: "SELECT Id FROM X", Type: "list", Default: true}},
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

func TestFieldNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields []OVViewField
		want   []string
	}{
		{
			name:   "nil returns nil",
			fields: nil,
			want:   nil,
		},
		{
			name:   "empty returns nil",
			fields: []OVViewField{},
			want:   nil,
		},
		{
			name:   "extracts names from simple fields",
			fields: []OVViewField{{Name: "name"}, {Name: "email"}},
			want:   []string{"name", "email"},
		},
		{
			name:   "extracts names from mixed fields",
			fields: []OVViewField{{Name: "name"}, {Name: "total", Expr: "a+b"}},
			want:   []string{"name", "total"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FieldNames(tt.fields)
			assert.Equal(t, tt.want, got)
		})
	}
}
