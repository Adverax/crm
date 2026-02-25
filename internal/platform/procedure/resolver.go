package procedure

import (
	"fmt"
	"strings"

	celengine "github.com/adverax/crm/internal/platform/cel"
)

// ExpressionResolver resolves $.* expressions from the execution context.
type ExpressionResolver struct {
	cache *celengine.ProgramCache
}

// NewExpressionResolver creates a new ExpressionResolver.
func NewExpressionResolver(cache *celengine.ProgramCache) *ExpressionResolver {
	return &ExpressionResolver{cache: cache}
}

// ResolveString resolves a string value. If it starts with "$.", it's evaluated as
// a CEL expression against the execution context variables. Otherwise, returned as-is.
func (r *ExpressionResolver) ResolveString(expr string, execCtx *ExecutionContext) (string, error) {
	if !strings.HasPrefix(expr, "$.") {
		return expr, nil
	}

	celExpr := strings.TrimPrefix(expr, "$.")
	result, err := r.cache.EvaluateAny(celExpr, execCtx.Vars)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", expr, err)
	}

	return fmt.Sprintf("%v", result), nil
}

// ResolveAny resolves a value. If it's a string starting with "$.", evaluate as CEL.
// Otherwise, return as-is.
func (r *ExpressionResolver) ResolveAny(expr string, execCtx *ExecutionContext) (any, error) {
	if !strings.HasPrefix(expr, "$.") {
		return expr, nil
	}

	celExpr := strings.TrimPrefix(expr, "$.")
	result, err := r.cache.EvaluateAny(celExpr, execCtx.Vars)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", expr, err)
	}

	return result, nil
}

// ResolveMap resolves all values in a string map through the expression resolver.
func (r *ExpressionResolver) ResolveMap(data map[string]string, execCtx *ExecutionContext) (map[string]any, error) {
	result := make(map[string]any, len(data))
	for k, v := range data {
		resolved, err := r.ResolveAny(v, execCtx)
		if err != nil {
			return nil, fmt.Errorf("resolveMap[%s]: %w", k, err)
		}
		result[k] = resolved
	}
	return result, nil
}

// ResolveBool resolves a CEL expression to a boolean value.
func (r *ExpressionResolver) ResolveBool(expr string, execCtx *ExecutionContext) (bool, error) {
	if expr == "" {
		return true, nil
	}

	celExpr := expr
	if strings.HasPrefix(expr, "$.") {
		celExpr = strings.TrimPrefix(expr, "$.")
	}

	result, err := r.cache.EvaluateBool(celExpr, execCtx.Vars)
	if err != nil {
		return false, fmt.Errorf("resolveBool %q: %w", expr, err)
	}
	return result, nil
}
