//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/ee/internal/platform/territory"
)

// --- Mock AssignmentRuleRepository ---

type mockAssignmentRuleRepo struct {
	rules    map[uuid.UUID]*territory.AssignmentRule
	byObject map[uuid.UUID][]territory.AssignmentRule
}

func newMockAssignmentRuleRepo() *mockAssignmentRuleRepo {
	return &mockAssignmentRuleRepo{
		rules:    make(map[uuid.UUID]*territory.AssignmentRule),
		byObject: make(map[uuid.UUID][]territory.AssignmentRule),
	}
}

func (r *mockAssignmentRuleRepo) Create(_ context.Context, _ pgx.Tx, input territory.CreateAssignmentRuleInput) (*territory.AssignmentRule, error) {
	rule := &territory.AssignmentRule{
		ID:            uuid.New(),
		TerritoryID:   input.TerritoryID,
		ObjectID:      input.ObjectID,
		IsActive:      input.IsActive,
		RuleOrder:     input.RuleOrder,
		CriteriaField: input.CriteriaField,
		CriteriaOp:    input.CriteriaOp,
		CriteriaValue: input.CriteriaValue,
	}
	r.rules[rule.ID] = rule
	return rule, nil
}

func (r *mockAssignmentRuleRepo) GetByID(_ context.Context, id uuid.UUID) (*territory.AssignmentRule, error) {
	return r.rules[id], nil
}

func (r *mockAssignmentRuleRepo) ListByTerritoryID(_ context.Context, territoryID uuid.UUID) ([]territory.AssignmentRule, error) {
	var result []territory.AssignmentRule
	for _, rule := range r.rules {
		if rule.TerritoryID == territoryID {
			result = append(result, *rule)
		}
	}
	return result, nil
}

func (r *mockAssignmentRuleRepo) ListByObjectID(_ context.Context, objectID uuid.UUID) ([]territory.AssignmentRule, error) {
	return r.byObject[objectID], nil
}

func (r *mockAssignmentRuleRepo) ListActiveByObjectID(_ context.Context, objectID uuid.UUID) ([]territory.AssignmentRule, error) {
	var result []territory.AssignmentRule
	for _, rule := range r.byObject[objectID] {
		if rule.IsActive {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (r *mockAssignmentRuleRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input territory.UpdateAssignmentRuleInput) (*territory.AssignmentRule, error) {
	rule := r.rules[id]
	if rule != nil {
		rule.IsActive = input.IsActive
		rule.CriteriaField = input.CriteriaField
		rule.CriteriaOp = input.CriteriaOp
		rule.CriteriaValue = input.CriteriaValue
	}
	return rule, nil
}

func (r *mockAssignmentRuleRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(r.rules, id)
	return nil
}

// --- Tests ---

func TestAssignmentEvaluator_Evaluate(t *testing.T) {
	objectID := uuid.New()
	territory1ID := uuid.New()
	territory2ID := uuid.New()
	territory3ID := uuid.New()

	tests := []struct {
		name        string
		fieldValues map[string]string
		rules       []territory.AssignmentRule
		wantCount   int
	}{
		{
			name:        "matches eq operator",
			fieldValues: map[string]string{"region": "US"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "region", CriteriaOp: "eq", CriteriaValue: "US"},
			},
			wantCount: 1,
		},
		{
			name:        "does not match eq when value differs",
			fieldValues: map[string]string{"region": "EU"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "region", CriteriaOp: "eq", CriteriaValue: "US"},
			},
			wantCount: 0,
		},
		{
			name:        "matches neq operator",
			fieldValues: map[string]string{"region": "EU"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "region", CriteriaOp: "neq", CriteriaValue: "US"},
			},
			wantCount: 1,
		},
		{
			name:        "matches in operator",
			fieldValues: map[string]string{"region": "US"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "region", CriteriaOp: "in", CriteriaValue: "US,EU,APAC"},
			},
			wantCount: 1,
		},
		{
			name:        "matches contains operator",
			fieldValues: map[string]string{"name": "Acme Corporation"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "name", CriteriaOp: "contains", CriteriaValue: "Acme"},
			},
			wantCount: 1,
		},
		{
			name:        "matches gt operator (numeric)",
			fieldValues: map[string]string{"amount": "150"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "amount", CriteriaOp: "gt", CriteriaValue: "100"},
			},
			wantCount: 1,
		},
		{
			name:        "matches lt operator (numeric)",
			fieldValues: map[string]string{"amount": "50"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "amount", CriteriaOp: "lt", CriteriaValue: "100"},
			},
			wantCount: 1,
		},
		{
			name:        "matches multiple territories",
			fieldValues: map[string]string{"region": "US", "tier": "enterprise"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "region", CriteriaOp: "eq", CriteriaValue: "US"},
				{ID: uuid.New(), TerritoryID: territory2ID, ObjectID: objectID, IsActive: true, CriteriaField: "tier", CriteriaOp: "eq", CriteriaValue: "enterprise"},
				{ID: uuid.New(), TerritoryID: territory3ID, ObjectID: objectID, IsActive: true, CriteriaField: "tier", CriteriaOp: "eq", CriteriaValue: "smb"},
			},
			wantCount: 2,
		},
		{
			name:        "skips inactive rules",
			fieldValues: map[string]string{"region": "US"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: false, CriteriaField: "region", CriteriaOp: "eq", CriteriaValue: "US"},
			},
			wantCount: 0,
		},
		{
			name:        "skips when field not in record",
			fieldValues: map[string]string{"name": "Acme"},
			rules: []territory.AssignmentRule{
				{ID: uuid.New(), TerritoryID: territory1ID, ObjectID: objectID, IsActive: true, CriteriaField: "region", CriteriaOp: "eq", CriteriaValue: "US"},
			},
			wantCount: 0,
		},
		{
			name:        "empty rules returns empty matches",
			fieldValues: map[string]string{"region": "US"},
			rules:       nil,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleRepo := newMockAssignmentRuleRepo()
			ruleRepo.byObject[objectID] = tt.rules

			evaluator := territory.NewAssignmentEvaluator(ruleRepo)
			matches, err := evaluator.Evaluate(context.Background(), objectID, tt.fieldValues)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(matches) != tt.wantCount {
				t.Errorf("expected %d matches, got %d", tt.wantCount, len(matches))
			}
		})
	}
}
