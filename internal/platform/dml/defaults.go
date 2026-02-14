package dml

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	celengine "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/security"
)

// CELDefaultResolver resolves static and dynamic (CEL) default values for fields.
type CELDefaultResolver struct {
	celCache *celengine.ProgramCache
}

// NewCELDefaultResolver creates a new CELDefaultResolver.
func NewCELDefaultResolver() (*CELDefaultResolver, error) {
	env, err := celengine.DefaultEnv()
	if err != nil {
		return nil, fmt.Errorf("newCELDefaultResolver: %w", err)
	}
	return &CELDefaultResolver{
		celCache: celengine.NewProgramCache(env),
	}, nil
}

// ResolveDefaults implements engine.DefaultResolver.
func (r *CELDefaultResolver) ResolveDefaults(
	ctx context.Context,
	object *engine.ObjectMeta,
	operation engine.Operation,
	providedFields []string,
) (map[string]any, error) {
	provided := make(map[string]bool, len(providedFields))
	for _, f := range providedFields {
		provided[f] = true
	}

	// Build CEL variables
	userVars := buildUserVars(ctx)
	celVars := celengine.DefaultVars(make(celengine.RecordMap), userVars)

	defaults := make(map[string]any)
	for _, field := range object.Fields {
		if provided[field.Name] {
			continue
		}
		if field.ReadOnly || field.Calculated {
			continue
		}
		if !matchesDefaultOn(field.DefaultOn, operation) {
			continue
		}

		// Dynamic default (CEL expression) takes priority
		if field.DefaultExpr != nil && *field.DefaultExpr != "" {
			result, err := r.celCache.EvaluateAny(*field.DefaultExpr, celVars)
			if err != nil {
				return nil, &engine.DefaultEvalError{
					Field:      field.Name,
					Expression: *field.DefaultExpr,
					Cause:      err,
				}
			}
			defaults[field.Name] = result
			continue
		}

		// Static default
		if field.DefaultValue != nil && *field.DefaultValue != "" {
			converted := convertStaticDefault(*field.DefaultValue, field.Type)
			defaults[field.Name] = converted
		}
	}

	return defaults, nil
}

// matchesDefaultOn checks if a field's default_on setting matches the operation.
func matchesDefaultOn(defaultOn *string, operation engine.Operation) bool {
	if defaultOn == nil || *defaultOn == "" {
		// Default: only on create
		return operation == engine.OperationInsert || operation == engine.OperationUpsert
	}

	on := *defaultOn
	switch operation {
	case engine.OperationInsert, engine.OperationUpsert:
		return strings.Contains(on, "create")
	case engine.OperationUpdate:
		return strings.Contains(on, "update")
	default:
		return false
	}
}

// convertStaticDefault parses a string default value to a typed Go value.
func convertStaticDefault(value string, fieldType engine.FieldType) any {
	switch fieldType {
	case engine.FieldTypeInteger:
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		return value
	case engine.FieldTypeFloat:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
		return value
	case engine.FieldTypeBoolean:
		return strings.EqualFold(value, "true")
	case engine.FieldTypeDateTime:
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			return t
		}
		return value
	case engine.FieldTypeDate:
		if t, err := time.Parse("2006-01-02", value); err == nil {
			return t
		}
		return value
	default:
		return value
	}
}

// buildUserVars extracts user variables from the context for CEL expressions.
func buildUserVars(ctx context.Context) map[string]any {
	uc, ok := security.UserFromContext(ctx)
	if !ok {
		return map[string]any{
			"id":         "",
			"profile_id": "",
			"role_id":    "",
		}
	}
	roleID := uc.RoleID
	if roleID == nil {
		return celengine.UserVars(uc.UserID, uc.ProfileID, uuid.Nil)
	}
	return celengine.UserVars(uc.UserID, uc.ProfileID, *roleID)
}
