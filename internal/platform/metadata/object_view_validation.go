package metadata

import (
	"fmt"
	"strings"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// validateViewConfig validates the OV view config at save time.
// Checks: query uniqueness, at most one default, valid query types,
// field uniqueness, valid query references, DAG (no cycles).
func validateViewConfig(config OVConfig) error {
	view := config.View

	if err := validateQueries(view.Queries); err != nil {
		return err
	}

	if err := validateFields(view.Fields, view.Queries); err != nil {
		return err
	}

	return nil
}

func validateQueries(queries []OVQuery) error {
	names := make(map[string]bool, len(queries))
	defaultCount := 0

	for _, q := range queries {
		if q.Name == "" {
			return apperror.BadRequest("query name is required")
		}
		if names[q.Name] {
			return apperror.BadRequest(fmt.Sprintf("duplicate query name: %s", q.Name))
		}
		names[q.Name] = true

		if q.Type != "scalar" && q.Type != "list" {
			return apperror.BadRequest(fmt.Sprintf("query %q: type must be 'scalar' or 'list', got %q", q.Name, q.Type))
		}

		if q.Default {
			defaultCount++
		}
	}

	if defaultCount > 1 {
		return apperror.BadRequest("at most one query can be marked as default")
	}

	return nil
}

func validateFields(fields []OVViewField, queries []OVQuery) error {
	queryNames := make(map[string]bool, len(queries))
	for _, q := range queries {
		queryNames[q.Name] = true
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
			if err := validateExprQueryRefs(f.Name, f.Expr, queryNames, fieldNames); err != nil {
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
// refer to existing queries. Uses simple prefix matching: "queryName.something".
func validateExprQueryRefs(fieldName string, expr string, queryNames map[string]bool, fieldNames map[string]bool) error {
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
		// If it looks like a query reference, validate it exists
		if !queryNames[prefix] && !fieldNames[prefix] {
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
