package metadata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortalConfig_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input PortalConfig
	}{
		{
			name: "view only with simple fields",
			input: PortalConfig{
				Read: PortalReadConfig{
					Fields:  []PortalViewField{{Name: "name"}, {Name: "email"}},
					Actions: []PortalAction{{Key: "edit", Label: "Edit", Type: "primary", Icon: "pencil"}},
				},
			},
		},
		{
			name: "view with computed fields",
			input: PortalConfig{
				Read: PortalReadConfig{
					Fields: []PortalViewField{
						{Name: "name"},
						{Name: "total", Expr: "record.amount * 1.2", When: "has(record.amount)"},
					},
					Actions: []PortalAction{},
				},
			},
		},
		{
			name: "view with queries",
			input: PortalConfig{
				Read: PortalReadConfig{
					Fields:  []PortalViewField{{Name: "name"}},
					Actions: []PortalAction{},
					Queries: []PortalQuery{
						{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :id"},
						{Name: "contacts", SOQL: "SELECT Id FROM Contact WHERE AccountId = :id"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.input)
			require.NoError(t, err)

			var got PortalConfig
			require.NoError(t, json.Unmarshal(data, &got))

			assert.Equal(t, tt.input, got)
		})
	}
}

func TestFieldNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields []PortalViewField
		want   []string
	}{
		{
			name:   "nil returns nil",
			fields: nil,
			want:   nil,
		},
		{
			name:   "empty returns nil",
			fields: []PortalViewField{},
			want:   nil,
		},
		{
			name:   "extracts names from simple fields",
			fields: []PortalViewField{{Name: "name"}, {Name: "email"}},
			want:   []string{"name", "email"},
		},
		{
			name:   "extracts names from mixed fields",
			fields: []PortalViewField{{Name: "name"}, {Name: "total", Expr: "a+b"}},
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
