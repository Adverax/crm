package metadata

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func validGateGraphConfig() PortalConfig {
	return PortalConfig{
		EntryGate: "main",
		Gates: map[string]PortalGate{
			"main": {
				Label: "Main",
				Body: []GateBodyStep{
					{Name: "data", Type: "soql", SOQL: "SELECT Id FROM Account"},
				},
				Layout: &GateLayout{
					Fields: []PortalViewField{{Name: "Name"}},
				},
				Outcomes: []GateOutcome{
					{Name: "detail", Gate: "detail"},
				},
			},
			"detail": {
				Label: "Detail",
				Args:  []PortalArg{{Name: "id", Type: "string"}},
				Body: []GateBodyStep{
					{Name: "record", Type: "soql", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :id"},
				},
				Layout: &GateLayout{
					Fields: []PortalViewField{{Name: "Name"}},
				},
				Outcomes: []GateOutcome{
					{Name: "back", Gate: "main"},
				},
			},
		},
	}
}

func TestValidateGateGraphConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  PortalConfig
		wantErr string
	}{
		// --- Valid cases ---
		{
			name:   "valid minimal config",
			config: validGateGraphConfig(),
		},
		{
			name: "valid single gate",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "q", Type: "soql", SOQL: "SELECT Id FROM Account"}},
					},
				},
			},
		},
		{
			name: "valid gate with portal args",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "account_id", Type: "string"}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "q", Type: "soql", SOQL: "SELECT Id FROM Account WHERE Id = :account_id"}},
					},
				},
			},
		},
		{
			name: "valid gate with DML step",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "id", Type: "string"}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "del", Type: "dml", DML: "DELETE FROM Account WHERE Id = :id"}},
					},
				},
			},
		},
		{
			name: "valid gate with page_size",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "list", Type: "soql", SOQL: "SELECT Id FROM Account", PageSize: 20}},
					},
				},
			},
		},
		{
			name: "valid gate with restrict_filters",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "list", Type: "soql", SOQL: "SELECT Id FROM Account", PageSize: 20, RestrictFilters: []string{"Id"}}},
					},
				},
			},
		},
		{
			name: "valid CRUD graph with cycle (list→detail→list)",
			config: PortalConfig{
				EntryGate: "list",
				Gates: map[string]PortalGate{
					"list": {
						Label:    "List",
						Body:     []GateBodyStep{{Name: "data", Type: "soql", SOQL: "SELECT Id FROM Account"}},
						Outcomes: []GateOutcome{{Name: "detail", Gate: "detail"}},
					},
					"detail": {
						Label:    "Detail",
						Args:     []PortalArg{{Name: "id", Type: "string"}},
						Body:     []GateBodyStep{{Name: "rec", Type: "soql", SOQL: "SELECT ROW Id FROM Account WHERE Id = :id"}},
						Outcomes: []GateOutcome{{Name: "back", Gate: "list"}},
					},
				},
			},
		},
		{
			name: "valid when condition on body step",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "q", Type: "soql", SOQL: "SELECT 1", When: "args.show == true"}},
					},
				},
			},
		},

		// --- entry_gate errors ---
		{
			name: "missing entry_gate",
			config: PortalConfig{
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "entry_gate is required",
		},
		{
			name: "entry_gate not in gates",
			config: PortalConfig{
				EntryGate: "missing",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: `entry_gate "missing" is not a key in gates`,
		},

		// --- gates count errors ---
		{
			name: "no gates",
			config: PortalConfig{
				EntryGate: "main",
				Gates:     map[string]PortalGate{},
			},
			wantErr: "at least one gate is required",
		},

		// --- gate name errors ---
		{
			name: "invalid gate name uppercase",
			config: PortalConfig{
				EntryGate: "Main",
				Gates: map[string]PortalGate{
					"Main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: `gate name "Main"`,
		},
		{
			name: "invalid gate name starts with number",
			config: PortalConfig{
				EntryGate: "1gate",
				Gates: map[string]PortalGate{
					"1gate": {Label: "Gate", Body: []GateBodyStep{}},
				},
			},
			wantErr: `gate name "1gate"`,
		},

		// --- gate arg errors ---
		{
			name: "gate arg shadows portal arg",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "id", Type: "string"}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Args:  []PortalArg{{Name: "id", Type: "int"}},
						Body:  []GateBodyStep{},
					},
				},
			},
			wantErr: `gate "main": arg "id" shadows portal-level arg`,
		},

		// --- body step errors ---
		{
			name: "step missing name",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Type: "soql", SOQL: "SELECT 1"}},
					},
				},
			},
			wantErr: "body step name is required",
		},
		{
			name: "step invalid name",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "Bad-Name", Type: "soql", SOQL: "SELECT 1"}},
					},
				},
			},
			wantErr: `step name "Bad-Name"`,
		},
		{
			name: "step duplicate name",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body: []GateBodyStep{
							{Name: "data", Type: "soql", SOQL: "SELECT 1"},
							{Name: "data", Type: "soql", SOQL: "SELECT 2"},
						},
					},
				},
			},
			wantErr: "duplicate step name: data",
		},
		{
			name: "step invalid type",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "q", Type: "rest", SOQL: "SELECT 1"}},
					},
				},
			},
			wantErr: "type must be 'soql' or 'dml'",
		},
		{
			name: "soql step missing soql field",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "q", Type: "soql"}},
					},
				},
			},
			wantErr: "requires soql field",
		},
		{
			name: "dml step missing dml field",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "d", Type: "dml"}},
					},
				},
			},
			wantErr: "requires dml field",
		},
		{
			name: "step when expression invalid CEL",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "q", Type: "soql", SOQL: "SELECT 1", When: "!!!invalid"}},
					},
				},
			},
			wantErr: "invalid when expression",
		},
		{
			name: "step param not declared",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{{Name: "q", Type: "soql", SOQL: "SELECT Id FROM Account WHERE Id = :missing_arg"}},
					},
				},
			},
			wantErr: "param :missing_arg is not declared in args",
		},
		{
			name: "too many body steps",
			config: func() PortalConfig {
				steps := make([]GateBodyStep, 11)
				for i := range steps {
					steps[i] = GateBodyStep{Name: fmt.Sprintf("s%d", i), Type: "soql", SOQL: "SELECT 1"}
				}
				return PortalConfig{
					EntryGate: "main",
					Gates: map[string]PortalGate{
						"main": {Label: "Main", Body: steps},
					},
				}
			}(),
			wantErr: "max 10 body steps",
		},

		// --- layout field errors ---
		{
			name: "duplicate field name in layout",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{},
						Layout: &GateLayout{
							Fields: []PortalViewField{{Name: "Name"}, {Name: "Name"}},
						},
					},
				},
			},
			wantErr: "duplicate field name: Name",
		},
		{
			name: "empty field name in layout",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{},
						Layout: &GateLayout{
							Fields: []PortalViewField{{Name: ""}},
						},
					},
				},
			},
			wantErr: "field name is required",
		},
		{
			name: "circular field dependency in layout",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{},
						Layout: &GateLayout{
							Fields: []PortalViewField{
								{Name: "a", Expr: "b + 1"},
								{Name: "b", Expr: "a + 1"},
							},
						},
					},
				},
			},
			wantErr: "circular dependency",
		},

		// --- outcome errors ---
		{
			name: "outcome missing name",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label:    "Main",
						Body:     []GateBodyStep{},
						Outcomes: []GateOutcome{{Gate: "main"}},
					},
				},
			},
			wantErr: "outcome name is required",
		},
		{
			name: "outcome duplicate name",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Body:  []GateBodyStep{},
						Outcomes: []GateOutcome{
							{Name: "back", Gate: "main"},
							{Name: "back", Gate: "main"},
						},
					},
				},
			},
			wantErr: "duplicate outcome name: back",
		},
		{
			name: "outcome references non-existent gate",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label:    "Main",
						Body:     []GateBodyStep{},
						Outcomes: []GateOutcome{{Name: "go", Gate: "nowhere"}},
					},
				},
			},
			wantErr: `references non-existent gate "nowhere"`,
		},
		{
			name: "outcome missing gate",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label:    "Main",
						Body:     []GateBodyStep{},
						Outcomes: []GateOutcome{{Name: "go"}},
					},
				},
			},
			wantErr: "must reference a gate",
		},
		{
			name: "too many outcomes",
			config: func() PortalConfig {
				outcomes := make([]GateOutcome, 11)
				for i := range outcomes {
					outcomes[i] = GateOutcome{Name: fmt.Sprintf("o%d", i), Gate: "main"}
				}
				return PortalConfig{
					EntryGate: "main",
					Gates: map[string]PortalGate{
						"main": {Label: "Main", Body: []GateBodyStep{}, Outcomes: outcomes},
					},
				}
			}(),
			wantErr: "max 10 outcomes",
		},

		// --- reachability errors ---
		{
			name: "orphan gate unreachable",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main":   {Label: "Main", Body: []GateBodyStep{}},
					"orphan": {Label: "Orphan", Body: []GateBodyStep{}},
				},
			},
			wantErr: "unreachable gates",
		},

		// --- arg validation (reused logic) ---
		{
			name: "portal arg empty name",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "", Type: "string"}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "arg name is required",
		},
		{
			name: "portal arg invalid name",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "Bad", Type: "string"}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: `arg name "Bad"`,
		},
		{
			name: "portal arg duplicate name",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "a", Type: "string"}, {Name: "a", Type: "int"}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "duplicate arg name: a",
		},
		{
			name: "portal arg unknown type",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "a", Type: "uuid"}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "type must be one of",
		},
		{
			name: "portal arg invalid default for int",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "n", Type: "int", Default: ptrStr("abc")}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "not a valid int",
		},
		{
			name: "portal arg invalid default for bool",
			config: PortalConfig{
				Args:      []PortalArg{{Name: "b", Type: "bool", Default: ptrStr("yes")}},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "not a valid bool",
		},

		// --- arg_rules validation ---
		{
			name: "valid portal with arg_rules",
			config: PortalConfig{
				Args: []PortalArg{{Name: "n", Type: "int", Default: ptrStr("10")}},
				ArgRules: []PortalArgRule{
					{Name: "limit_range", Condition: "args.n > 0 && args.n <= 100", ErrorMessage: "must be 1-100"},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
		},
		{
			name: "valid gate with arg_rules",
			config: PortalConfig{
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {
						Label: "Main",
						Args:  []PortalArg{{Name: "id", Type: "string"}},
						ArgRules: []PortalArgRule{
							{Name: "id_not_empty", Condition: "args.id != ''", ErrorMessage: "id required"},
						},
						Body: []GateBodyStep{},
					},
				},
			},
		},
		{
			name: "arg_rule empty name",
			config: PortalConfig{
				ArgRules: []PortalArgRule{
					{Name: "", Condition: "true", ErrorMessage: "msg"},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "arg_rule name is required",
		},
		{
			name: "arg_rule invalid name",
			config: PortalConfig{
				ArgRules: []PortalArgRule{
					{Name: "Bad-Name", Condition: "true", ErrorMessage: "msg"},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: `arg_rule name "Bad-Name"`,
		},
		{
			name: "arg_rule duplicate name",
			config: PortalConfig{
				ArgRules: []PortalArgRule{
					{Name: "r1", Condition: "true", ErrorMessage: "msg"},
					{Name: "r1", Condition: "true", ErrorMessage: "msg2"},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "duplicate arg_rule name: r1",
		},
		{
			name: "arg_rule missing condition",
			config: PortalConfig{
				ArgRules: []PortalArgRule{
					{Name: "r1", Condition: "", ErrorMessage: "msg"},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "condition is required",
		},
		{
			name: "arg_rule invalid CEL condition",
			config: PortalConfig{
				ArgRules: []PortalArgRule{
					{Name: "r1", Condition: "!!!bad", ErrorMessage: "msg"},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "invalid condition",
		},
		{
			name: "arg_rule missing error_message",
			config: PortalConfig{
				ArgRules: []PortalArgRule{
					{Name: "r1", Condition: "true", ErrorMessage: ""},
				},
				EntryGate: "main",
				Gates: map[string]PortalGate{
					"main": {Label: "Main", Body: []GateBodyStep{}},
				},
			},
			wantErr: "error_message is required",
		},
		{
			name: "too many arg_rules",
			config: func() PortalConfig {
				rules := make([]PortalArgRule, 11)
				for i := range rules {
					rules[i] = PortalArgRule{Name: fmt.Sprintf("r%d", i), Condition: "true", ErrorMessage: "msg"}
				}
				return PortalConfig{
					ArgRules:  rules,
					EntryGate: "main",
					Gates: map[string]PortalGate{
						"main": {Label: "Main", Body: []GateBodyStep{}},
					},
				}
			}(),
			wantErr: "max 10 arg_rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateGateGraphConfig(tt.config)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}
