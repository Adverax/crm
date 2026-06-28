package metadata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// newPortalCELEnv creates a minimal CEL env for validating portal arg expressions.
func newPortalCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("args", cel.DynType),
		ext.Strings(),
	)
}

// newGateCELEnv creates a CEL env for validating gate body/outcome expressions.
func newGateCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("args", cel.DynType),
		cel.Variable("datasets", cel.DynType),
		cel.Variable("data", cel.DynType),
		cel.Variable("user", cel.DynType),
		ext.Strings(),
	)
}

var (
	gateNameRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	stepNameRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	argNameRegexp  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	validArgTypes  = map[string]bool{"string": true, "int": true, "float": true, "bool": true}
	paramRegexp    = regexp.MustCompile(`:(\w+)`)
)

// validateGateGraphConfig validates the portal gate graph config at save time (ADR-0037).
func validateGateGraphConfig(config PortalConfig) error {
	if err := validateArgs(config.Args); err != nil {
		return err
	}

	if err := validateArgRules("portal", config.ArgRules); err != nil {
		return err
	}

	if config.EntryGate == "" {
		return apperror.BadRequest("entry_gate is required")
	}

	if len(config.Gates) == 0 {
		return apperror.BadRequest("at least one gate is required")
	}

	if len(config.Gates) > 50 {
		return apperror.BadRequest("max 50 gates per portal")
	}

	if _, ok := config.Gates[config.EntryGate]; !ok {
		return apperror.BadRequest(fmt.Sprintf("entry_gate %q is not a key in gates", config.EntryGate))
	}

	// Validate gate names
	for name := range config.Gates {
		if !gateNameRegexp.MatchString(name) {
			return apperror.BadRequest(fmt.Sprintf("gate name %q: must match ^[a-z][a-z0-9_]*$", name))
		}
	}

	// Collect portal-level arg names
	portalArgNames := make(map[string]bool, len(config.Args))
	for _, a := range config.Args {
		portalArgNames[a.Name] = true
	}

	// Validate each gate
	for gateName, gate := range config.Gates {
		if err := validateGate(gateName, gate, portalArgNames, config.Gates); err != nil {
			return err
		}
	}

	// Reachability: BFS from entry_gate
	if err := validateReachability(config.EntryGate, config.Gates); err != nil {
		return err
	}

	return nil
}

// validateGate validates a single gate within the graph.
func validateGate(gateName string, gate PortalGate, portalArgNames map[string]bool, allGates map[string]PortalGate) error {
	// Validate gate-level args
	if err := validateArgs(gate.Args); err != nil {
		return fmt.Errorf("gate %q: %w", gateName, err)
	}

	if err := validateArgRules(fmt.Sprintf("gate %q", gateName), gate.ArgRules); err != nil {
		return err
	}

	if len(gate.Args) > 20 {
		return apperror.BadRequest(fmt.Sprintf("gate %q: max 20 args per gate", gateName))
	}

	// Check arg name collisions with portal-level args
	for _, a := range gate.Args {
		if portalArgNames[a.Name] {
			return apperror.BadRequest(fmt.Sprintf("gate %q: arg %q shadows portal-level arg", gateName, a.Name))
		}
	}

	// Collect all available arg names (portal + gate)
	allArgNames := make(map[string]bool, len(portalArgNames)+len(gate.Args))
	for n := range portalArgNames {
		allArgNames[n] = true
	}
	for _, a := range gate.Args {
		allArgNames[a.Name] = true
	}

	// Validate body steps
	if len(gate.Body) > 10 {
		return apperror.BadRequest(fmt.Sprintf("gate %q: max 10 body steps per gate", gateName))
	}

	stepNames := make(map[string]bool, len(gate.Body))
	for _, step := range gate.Body {
		if err := validateBodyStep(gateName, step, stepNames, allArgNames); err != nil {
			return err
		}
		stepNames[step.Name] = true
	}

	// Validate layout fields
	if gate.Layout != nil && len(gate.Layout.Fields) > 0 {
		if err := validateGateLayoutFields(gateName, gate.Layout.Fields); err != nil {
			return err
		}
	}

	// Validate outcomes
	if len(gate.Outcomes) > 10 {
		return apperror.BadRequest(fmt.Sprintf("gate %q: max 10 outcomes per gate", gateName))
	}

	outcomeNames := make(map[string]bool, len(gate.Outcomes))
	for _, outcome := range gate.Outcomes {
		if err := validateOutcome(gateName, outcome, outcomeNames, allGates); err != nil {
			return err
		}
		outcomeNames[outcome.Name] = true
	}

	return nil
}

// validateBodyStep validates a single body step.
func validateBodyStep(gateName string, step GateBodyStep, existing map[string]bool, argNames map[string]bool) error {
	if step.Name == "" {
		return apperror.BadRequest(fmt.Sprintf("gate %q: body step name is required", gateName))
	}
	if !stepNameRegexp.MatchString(step.Name) {
		return apperror.BadRequest(fmt.Sprintf("gate %q: step name %q must match ^[a-z][a-z0-9_]*$", gateName, step.Name))
	}
	if existing[step.Name] {
		return apperror.BadRequest(fmt.Sprintf("gate %q: duplicate step name: %s", gateName, step.Name))
	}

	if step.Type != "soql" && step.Type != "dml" {
		return apperror.BadRequest(fmt.Sprintf("gate %q: step %q type must be 'soql' or 'dml'", gateName, step.Name))
	}

	if step.Type == "soql" && step.SOQL == "" {
		return apperror.BadRequest(fmt.Sprintf("gate %q: step %q requires soql field for type 'soql'", gateName, step.Name))
	}

	if step.Type == "dml" && step.DML == "" {
		return apperror.BadRequest(fmt.Sprintf("gate %q: step %q requires dml field for type 'dml'", gateName, step.Name))
	}

	if step.PageSize < 0 {
		return apperror.BadRequest(fmt.Sprintf("gate %q: step %q page_size must be non-negative", gateName, step.Name))
	}

	// Validate when condition compiles
	if step.When != "" {
		env, err := newGateCELEnv()
		if err != nil {
			return apperror.BadRequest(fmt.Sprintf("gate %q: step %q: failed to create CEL env: %v", gateName, step.Name, err))
		}
		_, issues := env.Compile(step.When)
		if issues != nil && issues.Err() != nil {
			return apperror.BadRequest(fmt.Sprintf("gate %q: step %q: invalid when expression: %s", gateName, step.Name, issues.Err()))
		}
	}

	// Validate param refs
	var text string
	if step.Type == "soql" {
		text = step.SOQL
	} else {
		text = step.DML
	}
	matches := paramRegexp.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		paramName := m[1]
		if !argNames[paramName] {
			return apperror.BadRequest(fmt.Sprintf("gate %q: step %q: param :%s is not declared in args", gateName, step.Name, paramName))
		}
	}

	return nil
}

// validateGateLayoutFields validates field uniqueness and DAG within a gate layout.
func validateGateLayoutFields(gateName string, fields []PortalViewField) error {
	fieldNames := make(map[string]bool, len(fields))
	for _, f := range fields {
		if f.Name == "" {
			return apperror.BadRequest(fmt.Sprintf("gate %q: field name is required", gateName))
		}
		if fieldNames[f.Name] {
			return apperror.BadRequest(fmt.Sprintf("gate %q: duplicate field name: %s", gateName, f.Name))
		}
		fieldNames[f.Name] = true
	}

	if err := validateFieldDAG(fields); err != nil {
		return fmt.Errorf("gate %q: %w", gateName, err)
	}

	return nil
}

// validateOutcome validates a single outcome declaration.
func validateOutcome(gateName string, outcome GateOutcome, existing map[string]bool, allGates map[string]PortalGate) error {
	if outcome.Name == "" {
		return apperror.BadRequest(fmt.Sprintf("gate %q: outcome name is required", gateName))
	}
	if existing[outcome.Name] {
		return apperror.BadRequest(fmt.Sprintf("gate %q: duplicate outcome name: %s", gateName, outcome.Name))
	}

	if outcome.Gate == "" {
		return apperror.BadRequest(fmt.Sprintf("gate %q: outcome %q must reference a gate", gateName, outcome.Name))
	}
	if _, ok := allGates[outcome.Gate]; !ok {
		return apperror.BadRequest(fmt.Sprintf("gate %q: outcome %q references non-existent gate %q", gateName, outcome.Name, outcome.Gate))
	}

	return nil
}

// validateReachability checks that all gates are reachable from the entry gate via BFS.
func validateReachability(entryGate string, gates map[string]PortalGate) error {
	visited := make(map[string]bool)
	queue := []string{entryGate}
	visited[entryGate] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		gate := gates[current]
		for _, outcome := range gate.Outcomes {
			if !visited[outcome.Gate] {
				visited[outcome.Gate] = true
				queue = append(queue, outcome.Gate)
			}
		}
	}

	if len(visited) < len(gates) {
		var orphans []string
		for name := range gates {
			if !visited[name] {
				orphans = append(orphans, name)
			}
		}
		return apperror.BadRequest(fmt.Sprintf("unreachable gates: %s", strings.Join(orphans, ", ")))
	}

	return nil
}

// validateArgs validates portal arg declarations.
func validateArgs(args []PortalArg) error {
	if len(args) > 50 {
		return apperror.BadRequest("max 50 args per portal")
	}

	names := make(map[string]bool, len(args))
	for _, a := range args {
		if a.Name == "" {
			return apperror.BadRequest("arg name is required")
		}
		if !argNameRegexp.MatchString(a.Name) {
			return apperror.BadRequest(fmt.Sprintf("arg name %q: must match ^[a-z][a-z0-9_]*$", a.Name))
		}
		if names[a.Name] {
			return apperror.BadRequest(fmt.Sprintf("duplicate arg name: %s", a.Name))
		}
		names[a.Name] = true

		if !validArgTypes[a.Type] {
			return apperror.BadRequest(fmt.Sprintf("arg %q: type must be one of: string, int, float, bool", a.Name))
		}

		if a.Default != nil {
			if err := checkArgDefault(a.Name, a.Type, *a.Default); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateArgRules validates a list of arg validation rules within a scope.
func validateArgRules(scope string, rules []PortalArgRule) error {
	if len(rules) > 10 {
		return apperror.BadRequest(fmt.Sprintf("%s: max 10 arg_rules per scope", scope))
	}

	names := make(map[string]bool, len(rules))
	for _, r := range rules {
		if r.Name == "" {
			return apperror.BadRequest(fmt.Sprintf("%s: arg_rule name is required", scope))
		}
		if !argNameRegexp.MatchString(r.Name) {
			return apperror.BadRequest(fmt.Sprintf("%s: arg_rule name %q must match ^[a-z][a-z0-9_]*$", scope, r.Name))
		}
		if names[r.Name] {
			return apperror.BadRequest(fmt.Sprintf("%s: duplicate arg_rule name: %s", scope, r.Name))
		}
		names[r.Name] = true

		if r.Condition == "" {
			return apperror.BadRequest(fmt.Sprintf("%s: arg_rule %q condition is required", scope, r.Name))
		}
		if err := checkCELCondition(scope, r.Name, r.Condition); err != nil {
			return err
		}

		if r.ErrorMessage == "" {
			return apperror.BadRequest(fmt.Sprintf("%s: arg_rule %q error_message is required", scope, r.Name))
		}
	}

	return nil
}

// checkCELCondition checks that a CEL condition compiles in PortalEnv.
func checkCELCondition(scope, name, expr string) error {
	env, err := newPortalCELEnv()
	if err != nil {
		return apperror.BadRequest(fmt.Sprintf("%s: arg_rule %q: failed to create CEL env: %v", scope, name, err))
	}
	_, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return apperror.BadRequest(fmt.Sprintf("%s: arg_rule %q: invalid condition: %s", scope, name, issues.Err()))
	}
	return nil
}

// checkArgDefault validates that a default value is compatible with the declared type.
func checkArgDefault(name, argType, value string) error {
	switch argType {
	case "int":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return apperror.BadRequest(fmt.Sprintf("arg %q: default %q is not a valid int", name, value))
		}
	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return apperror.BadRequest(fmt.Sprintf("arg %q: default %q is not a valid float", name, value))
		}
	case "bool":
		if value != "true" && value != "false" {
			return apperror.BadRequest(fmt.Sprintf("arg %q: default %q is not a valid bool", name, value))
		}
	}
	return nil
}

// validateFieldDAG ensures computed fields form a DAG (no cycles).
// Uses Kahn's algorithm (topological sort).
func validateFieldDAG(fields []PortalViewField) error {
	fieldIndex := make(map[string]int, len(fields))
	for i, f := range fields {
		fieldIndex[f.Name] = i
	}

	inDegree := make([]int, len(fields))
	dependents := make([][]int, len(fields))

	for i, f := range fields {
		if f.Expr == "" {
			continue
		}
		deps := extractFieldRefs(f.Expr, fieldIndex)
		inDegree[i] = len(deps)
		for _, depIdx := range deps {
			dependents[depIdx] = append(dependents[depIdx], i)
		}
	}

	queue := make([]int, 0)
	for i := range fields {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}
	}

	processed := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		processed++

		for _, dep := range dependents[node] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if processed < len(fields) {
		var cycleFields []string
		for i, deg := range inDegree {
			if deg > 0 {
				cycleFields = append(cycleFields, fields[i].Name)
			}
		}
		return apperror.BadRequest(fmt.Sprintf("circular dependency detected among fields: %s", strings.Join(cycleFields, ", ")))
	}

	return nil
}

// extractFieldRefs finds references to other fields in a CEL expression.
// Returns indices of referenced fields.
func extractFieldRefs(expr string, fieldIndex map[string]int) []int {
	var refs []int
	seen := make(map[int]bool)

	for _, token := range strings.Fields(expr) {
		token = strings.Trim(token, "()+-*/!<>=&|,")
		if token == "" {
			continue
		}
		if idx, ok := fieldIndex[token]; ok {
			if !seen[idx] {
				refs = append(refs, idx)
				seen[idx] = true
			}
		}
	}
	return refs
}
