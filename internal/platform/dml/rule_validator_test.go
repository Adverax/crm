package dml

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
)

// mockCacheLoader implements metadata.CacheLoader for testing.
type mockCacheLoader struct {
	objects []metadata.ObjectDefinition
	fields  []metadata.FieldDefinition
	rules   []metadata.ValidationRule
}

func (m *mockCacheLoader) LoadAllObjects(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return m.objects, nil
}

func (m *mockCacheLoader) LoadAllFields(_ context.Context) ([]metadata.FieldDefinition, error) {
	return m.fields, nil
}

func (m *mockCacheLoader) LoadRelationships(_ context.Context) ([]metadata.RelationshipInfo, error) {
	return nil, nil
}

func (m *mockCacheLoader) LoadAllValidationRules(_ context.Context) ([]metadata.ValidationRule, error) {
	return m.rules, nil
}

func (m *mockCacheLoader) LoadAllFunctions(_ context.Context) ([]metadata.Function, error) {
	return nil, nil
}

func (m *mockCacheLoader) LoadAllObjectViews(_ context.Context) ([]metadata.ObjectView, error) {
	return nil, nil
}

func (m *mockCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

func setupTestCache(t *testing.T, objID uuid.UUID, rules []metadata.ValidationRule) *metadata.MetadataCache {
	t.Helper()
	loader := &mockCacheLoader{
		objects: []metadata.ObjectDefinition{
			{ID: objID, APIName: "Account", TableName: "obj_account"},
		},
		rules: rules,
	}
	cache := metadata.NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))
	return cache
}

func TestCELRuleValidator_ValidateRules(t *testing.T) {
	objID := uuid.New()

	tests := []struct {
		name           string
		rules          []metadata.ValidationRule
		record         map[string]any
		operation      engine.Operation
		wantViolations int
	}{
		{
			name: "rule passes",
			rules: []metadata.ValidationRule{{
				ID: uuid.New(), ObjectID: objID, APIName: "name_required",
				Expression: `size(record.Name) > 0`, ErrorMessage: "Name required",
				ErrorCode: "name_required", Severity: "error",
				AppliesTo: "create,update", IsActive: true,
			}},
			record:         map[string]any{"Name": "Acme"},
			operation:      engine.OperationInsert,
			wantViolations: 0,
		},
		{
			name: "rule fails",
			rules: []metadata.ValidationRule{{
				ID: uuid.New(), ObjectID: objID, APIName: "name_required",
				Expression: `size(record.Name) > 0`, ErrorMessage: "Name required",
				ErrorCode: "name_required", Severity: "error",
				AppliesTo: "create,update", IsActive: true,
			}},
			record:         map[string]any{"Name": ""},
			operation:      engine.OperationInsert,
			wantViolations: 1,
		},
		{
			name: "multiple rules - AND behavior",
			rules: []metadata.ValidationRule{
				{
					ID: uuid.New(), ObjectID: objID, APIName: "name_required",
					Expression: `size(record.Name) > 0`, ErrorMessage: "Name required",
					ErrorCode: "name_required", Severity: "error",
					AppliesTo: "create,update", IsActive: true,
				},
				{
					ID: uuid.New(), ObjectID: objID, APIName: "amount_positive",
					Expression: `record.Amount > 0`, ErrorMessage: "Amount must be positive",
					ErrorCode: "amount_positive", Severity: "error",
					AppliesTo: "create,update", IsActive: true,
				},
			},
			record:         map[string]any{"Name": "", "Amount": int64(0)},
			operation:      engine.OperationInsert,
			wantViolations: 2,
		},
		{
			name: "inactive rule skipped",
			rules: []metadata.ValidationRule{{
				ID: uuid.New(), ObjectID: objID, APIName: "name_required",
				Expression: `false`, ErrorMessage: "Always fails",
				ErrorCode: "always_fail", Severity: "error",
				AppliesTo: "create,update", IsActive: false,
			}},
			record:         map[string]any{"Name": "Acme"},
			operation:      engine.OperationInsert,
			wantViolations: 0,
		},
		{
			name: "applies_to=update only, insert skipped",
			rules: []metadata.ValidationRule{{
				ID: uuid.New(), ObjectID: objID, APIName: "update_only",
				Expression: `false`, ErrorMessage: "Only on update",
				ErrorCode: "update_only", Severity: "error",
				AppliesTo: "update", IsActive: true,
			}},
			record:         map[string]any{"Name": "Acme"},
			operation:      engine.OperationInsert,
			wantViolations: 0,
		},
		{
			name: "warning severity still reported",
			rules: []metadata.ValidationRule{{
				ID: uuid.New(), ObjectID: objID, APIName: "name_warning",
				Expression: `size(record.Name) > 3`, ErrorMessage: "Name too short",
				ErrorCode: "name_short", Severity: "warning",
				AppliesTo: "create,update", IsActive: true,
			}},
			record:         map[string]any{"Name": "AB"},
			operation:      engine.OperationInsert,
			wantViolations: 1,
		},
		{
			name: "when_expression false skips rule",
			rules: []metadata.ValidationRule{func() metadata.ValidationRule {
				when := `record.Type == "Premium"`
				return metadata.ValidationRule{
					ID: uuid.New(), ObjectID: objID, APIName: "premium_check",
					Expression: `record.Amount > 1000`, ErrorMessage: "Premium needs high amount",
					ErrorCode: "premium_amount", Severity: "error",
					WhenExpression: &when,
					AppliesTo:      "create,update", IsActive: true,
				}
			}()},
			record:         map[string]any{"Type": "Basic", "Amount": int64(10)},
			operation:      engine.OperationInsert,
			wantViolations: 0,
		},
		{
			name: "when_expression true applies rule",
			rules: []metadata.ValidationRule{func() metadata.ValidationRule {
				when := `record.Type == "Premium"`
				return metadata.ValidationRule{
					ID: uuid.New(), ObjectID: objID, APIName: "premium_check",
					Expression: `record.Amount > 1000`, ErrorMessage: "Premium needs high amount",
					ErrorCode: "premium_amount", Severity: "error",
					WhenExpression: &when,
					AppliesTo:      "create,update", IsActive: true,
				}
			}()},
			record:         map[string]any{"Type": "Premium", "Amount": int64(10)},
			operation:      engine.OperationInsert,
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := setupTestCache(t, objID, tt.rules)
			validator, err := NewCELRuleValidator(cache, nil)
			require.NoError(t, err)

			obj := engine.NewObjectMeta("Account", "obj_account").Build()
			violations, err := validator.ValidateRules(context.Background(), obj, tt.operation, tt.record, nil)
			require.NoError(t, err)
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}
