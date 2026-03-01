package metadata

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/soql/engine"
)

var actionKeyRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// validateViewConfig validates the OV view config at save time.
// Checks: query uniqueness, valid query types,
// field uniqueness, valid query references, DAG (no cycles).
func validateViewConfig(config OVConfig) error {
	view := config.Read

	if err := validateQueries(view.Queries); err != nil {
		return err
	}

	if err := validateFields(view.Fields, view.Queries); err != nil {
		return err
	}

	if err := validateActions(view.Actions); err != nil {
		return err
	}

	return nil
}

func validateQueries(queries []OVQuery) error {
	names := make(map[string]bool, len(queries))

	for _, q := range queries {
		if q.Name == "" {
			return apperror.BadRequest("query name is required")
		}
		if names[q.Name] {
			return apperror.BadRequest(fmt.Sprintf("duplicate query name: %s", q.Name))
		}
		names[q.Name] = true
	}

	return nil
}

func validateFields(fields []OVViewField, queries []OVQuery) error {
	queryTypes := make(map[string]string, len(queries))
	for _, q := range queries {
		if engine.IsRowQuery(q.SOQL) {
			queryTypes[q.Name] = "scalar"
		} else {
			queryTypes[q.Name] = "list"
		}
	}

	fieldNames := make(map[string]bool, len(fields))
	for _, f := range fields {
		if f.Name == "" {
			return apperror.BadRequest("field name is required")
		}
		if fieldNames[f.Name] {
			return apperror.BadRequest(fmt.Sprintf("duplicate field name: %s", f.Name))
		}
		fieldNames[f.Name] = true

		// Validate query references in expr
		if f.Expr != "" {
			if err := validateExprQueryRefs(f.Name, f.Expr, queryTypes, fieldNames); err != nil {
				return err
			}
		}
	}

	// DAG validation: detect cycles among fields with expressions
	if err := validateFieldDAG(fields); err != nil {
		return err
	}

	return nil
}

// validateExprQueryRefs checks that any query.field references in an expression
// refer to existing scalar queries. List queries cannot be referenced from field
// expressions — they are only used as data sources for related lists and tables.
// Uses simple prefix matching: "queryName.something".
func validateExprQueryRefs(fieldName string, expr string, queryTypes map[string]string, fieldNames map[string]bool) error {
	// Simple heuristic: find identifiers that look like "word.word"
	// This is not a full expression parser, just basic validation.
	for _, token := range strings.Fields(expr) {
		// Clean common operators
		token = strings.Trim(token, "()+-*/!<>=&|,")
		if token == "" {
			continue
		}
		parts := strings.SplitN(token, ".", 2)
		if len(parts) != 2 {
			continue
		}
		prefix := parts[0]
		// Skip well-known CEL prefixes
		if prefix == "record" || prefix == "size" || prefix == "has" || prefix == "int" || prefix == "double" || prefix == "string" || prefix == "bool" {
			continue
		}
		// If it looks like a query reference, validate it exists and is scalar
		qType, isQuery := queryTypes[prefix]
		if isQuery {
			if qType == "list" {
				return apperror.BadRequest(fmt.Sprintf("field %q: expr references list query %q, only scalar queries are allowed in field expressions", fieldName, prefix))
			}
			continue
		}
		if !fieldNames[prefix] {
			return apperror.BadRequest(fmt.Sprintf("field %q: expr references unknown query %q", fieldName, prefix))
		}
	}
	return nil
}

// validateFieldDAG ensures computed fields form a DAG (no cycles).
// Uses Kahn's algorithm (topological sort).
func validateFieldDAG(fields []OVViewField) error {
	// Build adjacency: field → set of fields it depends on
	fieldIndex := make(map[string]int, len(fields))
	for i, f := range fields {
		fieldIndex[f.Name] = i
	}

	// inDegree[i] = number of fields that field i depends on (among computed fields)
	inDegree := make([]int, len(fields))
	// dependents[i] = list of field indices that depend on field i
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

	// Kahn's algorithm
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
		// Find cycle participants for error message
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

// validateActions validates the actions list at save time (ADR-0036).
func validateActions(actions []OVAction) error {
	if len(actions) > 20 {
		return apperror.BadRequest("max 20 actions per object view")
	}

	keys := make(map[string]bool, len(actions))
	for _, a := range actions {
		if a.Key == "" {
			return apperror.BadRequest("action key is required")
		}
		if !actionKeyRegexp.MatchString(a.Key) {
			return apperror.BadRequest(fmt.Sprintf("action key %q: must match ^[a-z][a-z0-9_]*$", a.Key))
		}
		if keys[a.Key] {
			return apperror.BadRequest(fmt.Sprintf("duplicate action key: %s", a.Key))
		}
		keys[a.Key] = true

		if len(a.Validation) > 20 {
			return apperror.BadRequest(fmt.Sprintf("action %q: max 20 validation rules", a.Key))
		}

		if a.Apply != nil {
			if err := validateActionApply(a.Key, a.Apply); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateActionApply(key string, apply *OVActionApply) error {
	if apply.Type != "dml" && apply.Type != "scenario" {
		return apperror.BadRequest(fmt.Sprintf("action %q: apply.type must be 'dml' or 'scenario'", key))
	}

	if apply.Type == "dml" {
		if len(apply.DML) == 0 {
			return apperror.BadRequest(fmt.Sprintf("action %q: apply.dml must not be empty for type 'dml'", key))
		}
		if len(apply.DML) > 10 {
			return apperror.BadRequest(fmt.Sprintf("action %q: max 10 DML queries per action", key))
		}
	}

	if apply.Type == "scenario" {
		if apply.Scenario == nil {
			return apperror.BadRequest(fmt.Sprintf("action %q: apply.scenario is required for type 'scenario'", key))
		}
		if apply.Scenario.APIName == "" {
			return apperror.BadRequest(fmt.Sprintf("action %q: apply.scenario.api_name is required", key))
		}
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
		// Direct field reference (no dot)
		if idx, ok := fieldIndex[token]; ok {
			if !seen[idx] {
				refs = append(refs, idx)
				seen[idx] = true
			}
		}
	}
	return refs
}
