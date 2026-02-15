package dml

import (
	"context"
	"fmt"
	"strings"

	celengine "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
)

// CELRuleValidator evaluates CEL-based validation rules from MetadataCache.
type CELRuleValidator struct {
	cache    *metadata.MetadataCache
	celCache *celengine.ProgramCache
}

// NewCELRuleValidator creates a new CELRuleValidator with an optional FunctionRegistry.
// If registry is nil, a plain StandardEnv is used.
func NewCELRuleValidator(cache *metadata.MetadataCache, registry *celengine.FunctionRegistry) (*CELRuleValidator, error) {
	env, err := buildStandardEnv(registry)
	if err != nil {
		return nil, fmt.Errorf("newCELRuleValidator: %w", err)
	}
	return &CELRuleValidator{
		cache:    cache,
		celCache: celengine.NewProgramCache(env),
	}, nil
}

// RebuildEnv rebuilds the CEL environment with an updated FunctionRegistry.
func (v *CELRuleValidator) RebuildEnv(registry *celengine.FunctionRegistry) error {
	env, err := buildStandardEnv(registry)
	if err != nil {
		return fmt.Errorf("celRuleValidator.RebuildEnv: %w", err)
	}
	v.celCache.Reset(env)
	return nil
}

func buildStandardEnv(registry *celengine.FunctionRegistry) (*celengine.Env, error) {
	if registry != nil {
		return celengine.StandardEnvWithFunctions(registry)
	}
	return celengine.StandardEnv()
}

// ValidateRules implements engine.RuleValidator.
func (v *CELRuleValidator) ValidateRules(
	ctx context.Context,
	object *engine.ObjectMeta,
	operation engine.Operation,
	record, old map[string]any,
) ([]engine.ValidationRuleError, error) {
	// Find object ID from cache
	objDef, ok := v.cache.GetObjectByAPIName(object.Name)
	if !ok {
		return nil, nil
	}

	rules := v.cache.GetValidationRules(objDef.ID)
	if len(rules) == 0 {
		return nil, nil
	}

	userVars := buildUserVars(ctx)
	celVars := celengine.ValidationVars(record, old, userVars)

	var violations []engine.ValidationRuleError
	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}
		if !matchesAppliesTo(rule.AppliesTo, operation) {
			continue
		}

		// Evaluate when_expression (conditional rule)
		if rule.WhenExpression != nil && *rule.WhenExpression != "" {
			whenResult, err := v.celCache.EvaluateBool(*rule.WhenExpression, celVars)
			if err != nil {
				// when_expression eval error — skip rule (log would be ideal, but we don't import slog)
				continue
			}
			if !whenResult {
				continue
			}
		}

		// Evaluate main expression
		result, err := v.celCache.EvaluateBool(rule.Expression, celVars)
		if err != nil {
			violations = append(violations, engine.ValidationRuleError{
				RuleID:   rule.ID.String(),
				Code:     "cel_error",
				Message:  fmt.Sprintf("rule %s: expression evaluation failed: %s", rule.APIName, err),
				Severity: rule.Severity,
			})
			continue
		}

		// Rule expression should return true when valid, false when violated
		if !result {
			violations = append(violations, engine.ValidationRuleError{
				RuleID:   rule.ID.String(),
				Code:     rule.ErrorCode,
				Message:  rule.ErrorMessage,
				Severity: rule.Severity,
			})
		}
	}

	return violations, nil
}

// matchesAppliesTo checks if a rule's applies_to matches the operation.
func matchesAppliesTo(appliesTo string, operation engine.Operation) bool {
	switch operation {
	case engine.OperationInsert, engine.OperationUpsert:
		return strings.Contains(appliesTo, "create")
	case engine.OperationUpdate:
		return strings.Contains(appliesTo, "update")
	default:
		return false
	}
}
