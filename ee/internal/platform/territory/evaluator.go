//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// AssignmentEvaluator evaluates territory assignment rules against record field values.
type AssignmentEvaluator struct {
	ruleRepo AssignmentRuleRepository
}

// NewAssignmentEvaluator creates a new AssignmentEvaluator.
func NewAssignmentEvaluator(ruleRepo AssignmentRuleRepository) *AssignmentEvaluator {
	return &AssignmentEvaluator{ruleRepo: ruleRepo}
}

// MatchResult represents a successful rule match for a record.
type MatchResult struct {
	TerritoryID uuid.UUID
	RuleID      uuid.UUID
}

// Evaluate checks all active assignment rules for the given object and returns
// matching territory IDs based on the record's field values.
func (e *AssignmentEvaluator) Evaluate(ctx context.Context, objectID uuid.UUID, fieldValues map[string]string) ([]MatchResult, error) {
	rules, err := e.ruleRepo.ListActiveByObjectID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("assignmentEvaluator.Evaluate: %w", err)
	}

	matches := make([]MatchResult, 0)
	for _, rule := range rules {
		val, ok := fieldValues[rule.CriteriaField]
		if !ok {
			continue
		}

		if matchesCriteria(val, rule.CriteriaOp, rule.CriteriaValue) {
			matches = append(matches, MatchResult{
				TerritoryID: rule.TerritoryID,
				RuleID:      rule.ID,
			})
		}
	}

	return matches, nil
}

func matchesCriteria(fieldValue, op, criteriaValue string) bool {
	switch op {
	case "eq":
		return fieldValue == criteriaValue
	case "neq":
		return fieldValue != criteriaValue
	case "in":
		for _, v := range strings.Split(criteriaValue, ",") {
			if strings.TrimSpace(v) == fieldValue {
				return true
			}
		}
		return false
	case "gt":
		return compareNumeric(fieldValue, criteriaValue) > 0
	case "lt":
		return compareNumeric(fieldValue, criteriaValue) < 0
	case "contains":
		return strings.Contains(fieldValue, criteriaValue)
	default:
		return false
	}
}

func compareNumeric(a, b string) int {
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA != nil || errB != nil {
		return strings.Compare(a, b)
	}
	switch {
	case fa < fb:
		return -1
	case fa > fb:
		return 1
	default:
		return 0
	}
}
