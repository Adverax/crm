package metadata

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

var fnRefRegex = regexp.MustCompile(`\bfn\.([a-z][a-z0-9_]*)\b`)

// ExtractFnReferences extracts all fn.* references from a CEL expression.
func ExtractFnReferences(expression string) []string {
	matches := fnRefRegex.FindAllStringSubmatch(expression, -1)
	seen := make(map[string]bool, len(matches))
	var refs []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			refs = append(refs, name)
		}
	}
	return refs
}

// DetectCycles checks for cyclic dependencies in a set of functions.
// Returns an error describing the cycle if one is found.
func DetectCycles(functions []Function) error {
	graph := make(map[string][]string, len(functions))
	for _, fn := range functions {
		graph[fn.Name] = ExtractFnReferences(fn.Body)
	}

	const (
		white = 0 // unvisited
		gray  = 1 // in current path
		black = 2 // done
	)

	color := make(map[string]int, len(functions))
	var path []string

	var dfs func(node string) error
	dfs = func(node string) error {
		color[node] = gray
		path = append(path, node)

		for _, dep := range graph[node] {
			switch color[dep] {
			case gray:
				return fmt.Errorf("cycle detected: %s -> %s", node, dep)
			case white:
				if _, exists := graph[dep]; !exists {
					continue // reference to unknown function — will be caught at compile time
				}
				if err := dfs(dep); err != nil {
					return err
				}
			}
		}

		color[node] = black
		path = path[:len(path)-1]
		return nil
	}

	for name := range graph {
		if color[name] == white {
			if err := dfs(name); err != nil {
				return err
			}
		}
	}
	return nil
}

// DetectNestingDepth checks if adding a function would exceed the max nesting depth (3 levels).
func DetectNestingDepth(functions []Function, maxDepth int) error {
	graph := make(map[string][]string, len(functions))
	for _, fn := range functions {
		graph[fn.Name] = ExtractFnReferences(fn.Body)
	}

	var depth func(name string, visited map[string]bool) int
	depth = func(name string, visited map[string]bool) int {
		if visited[name] {
			return 0 // cycle — handled by DetectCycles
		}
		visited[name] = true

		maxChild := 0
		for _, dep := range graph[name] {
			d := depth(dep, visited)
			if d > maxChild {
				maxChild = d
			}
		}

		delete(visited, name)
		return maxChild + 1
	}

	for name := range graph {
		d := depth(name, make(map[string]bool))
		if d > maxDepth {
			return fmt.Errorf("function %q exceeds max nesting depth %d (actual: %d)", name, maxDepth, d)
		}
	}
	return nil
}

// FunctionUsage describes where a function is referenced.
type FunctionUsage struct {
	Entity string `json:"entity"` // "function", "validation_rule", "field_definition"
	Name   string `json:"name"`
	ID     string `json:"id"`
}

// FindUsages finds all places where a named function is used.
func FindUsages(ctx context.Context, pool *pgxpool.Pool, name string) ([]FunctionUsage, error) {
	pattern := "fn." + name

	var usages []FunctionUsage

	// Check in other functions
	rows, err := pool.Query(ctx, `
		SELECT id::text, name FROM metadata.functions
		WHERE body LIKE '%' || $1 || '%' AND name != $2`, pattern, name)
	if err != nil {
		return nil, fmt.Errorf("findUsages.functions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u FunctionUsage
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("findUsages.functions.scan: %w", err)
		}
		u.Entity = "function"
		usages = append(usages, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("findUsages.functions.rows: %w", err)
	}

	// Check in validation rules
	rows2, err := pool.Query(ctx, `
		SELECT id::text, api_name FROM metadata.validation_rules
		WHERE expression LIKE '%' || $1 || '%'
		   OR (when_expression IS NOT NULL AND when_expression LIKE '%' || $1 || '%')`, pattern)
	if err != nil {
		return nil, fmt.Errorf("findUsages.validation_rules: %w", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var u FunctionUsage
		if err := rows2.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("findUsages.validation_rules.scan: %w", err)
		}
		u.Entity = "validation_rule"
		usages = append(usages, u)
	}
	if err := rows2.Err(); err != nil {
		return nil, fmt.Errorf("findUsages.validation_rules.rows: %w", err)
	}

	// Check in field definitions (default_expr in config JSONB)
	rows3, err := pool.Query(ctx, `
		SELECT id::text, api_name FROM metadata.field_definitions
		WHERE config->>'default_expr' LIKE '%' || $1 || '%'`, pattern)
	if err != nil {
		return nil, fmt.Errorf("findUsages.field_definitions: %w", err)
	}
	defer rows3.Close()

	for rows3.Next() {
		var u FunctionUsage
		if err := rows3.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("findUsages.field_definitions.scan: %w", err)
		}
		u.Entity = "field_definition"
		usages = append(usages, u)
	}
	if err := rows3.Err(); err != nil {
		return nil, fmt.Errorf("findUsages.field_definitions.rows: %w", err)
	}

	return usages, nil
}
