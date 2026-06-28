package metadata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrStr(s string) *string { return &s }

func TestPortalConfig_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input PortalConfig
	}{
		{
			name: "single gate with simple fields",
			input: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body: []GateBodyStep{
							{Name: "accounts", Type: "soql", SOQL: "SELECT Id, Name FROM Account"},
						},
						Layout: &GateLayout{
							Fields: []PortalViewField{{Name: "Name"}, {Name: "Industry"}},
						},
						Outcomes: []GateOutcome{
							{Name: "detail", Gate: "detail", Label: "View"},
						},
					},
				},
			},
		},
		{
			name: "gate with computed fields",
			input: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{},
						Layout: &GateLayout{
							Fields: []PortalViewField{
								{Name: "name"},
								{Name: "total", Expr: "record.amount * 1.2", When: "has(record.amount)"},
							},
						},
					},
				},
			},
		},
		{
			name: "multi-gate CRUD graph",
			input: PortalConfig{
				EntryGate: "list",
				Gates: map[string]PortalGate{
					"list": {
						Label: "Accounts",
						Body: []GateBodyStep{
							{Name: "accounts", Type: "soql", SOQL: "SELECT Id, Name FROM Account", PageSize: 20},
						},
						Layout: &GateLayout{
							Fields: []PortalViewField{{Name: "Name"}},
						},
						Outcomes: []GateOutcome{
							{Name: "create", Gate: "create", Label: "New", Type: "primary"},
							{Name: "detail", Gate: "detail"},
						},
					},
					"detail": {
						Label: "Account Detail",
						Args:  []PortalArg{{Name: "id", Type: "string"}},
						Body: []GateBodyStep{
							{Name: "account", Type: "soql", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :id"},
						},
						Layout: &GateLayout{
							Fields: []PortalViewField{{Name: "Name"}},
						},
						Outcomes: []GateOutcome{
							{Name: "edit", Gate: "edit", ArgsTemplate: map[string]string{"id": "args.id"}},
							{Name: "delete", Gate: "delete", Type: "danger"},
						},
					},
					"create": {
						Label: "New Account",
						Body: []GateBodyStep{
							{Name: "insert", Type: "dml", DML: "INSERT INTO Account (Name) VALUES (:name)"},
						},
						Layout: &GateLayout{
							Fields: []PortalViewField{{Name: "Name"}},
						},
						Outcomes: []GateOutcome{
							{Name: "list", Gate: "list", Label: "Cancel"},
						},
					},
					"edit": {
						Label: "Edit Account",
						Args:  []PortalArg{{Name: "id", Type: "string"}},
						Body: []GateBodyStep{
							{Name: "account", Type: "soql", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :id"},
							{Name: "save", Type: "dml", DML: "UPDATE Account SET Name = :name WHERE Id = :id"},
						},
						Layout: &GateLayout{
							Fields: []PortalViewField{{Name: "Name"}},
						},
						Outcomes: []GateOutcome{
							{Name: "detail", Gate: "detail", ArgsTemplate: map[string]string{"id": "args.id"}},
						},
					},
					"delete": {
						Label: "Delete",
						Args:  []PortalArg{{Name: "id", Type: "string"}},
						Body: []GateBodyStep{
							{Name: "delete", Type: "dml", DML: "DELETE FROM Account WHERE Id = :id"},
						},
						Outcomes: []GateOutcome{
							{Name: "list", Gate: "list"},
						},
					},
				},
			},
		},
		{
			name: "portal with args",
			input: PortalConfig{
				Args: []PortalArg{
					{Name: "account_id", Type: "string"},
					{Name: "limit", Type: "int", Default: ptrStr("10")},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{},
					},
				},
			},
		},
		{
			name: "portal with arg_rules",
			input: PortalConfig{
				Args: []PortalArg{
					{Name: "limit", Type: "int", Default: ptrStr("10")},
				},
				ArgRules: []PortalArgRule{
					{Name: "limit_range", Condition: "args.limit > 0 && args.limit <= 100", ErrorMessage: "Limit must be 1-100"},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{},
						ArgRules: []PortalArgRule{
							{Name: "gate_check", Condition: "args.limit > 5", ErrorMessage: "Gate requires limit > 5"},
						},
					},
				},
			},
		},
		{
			name: "gate with body step when condition and page_size",
			input: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body: []GateBodyStep{
							{
								Name:            "accounts",
								Type:            "soql",
								SOQL:            "SELECT Id FROM Account",
								When:            "args.show_all == true",
								PageSize:        25,
								RestrictFilters: []string{"Id"},
							},
						},
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
