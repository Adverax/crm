//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import (
	"fmt"
	"regexp"

	"github.com/adverax/crm/internal/pkg/apperror"
)

var apiNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,98}[a-zA-Z0-9]$`)

// ValidateCreateModel validates input for creating a territory model.
func ValidateCreateModel(input CreateModelInput) error {
	if err := validateAPIName(input.APIName); err != nil {
		return err
	}
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateUpdateModel validates input for updating a territory model.
func ValidateUpdateModel(input UpdateModelInput) error {
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateCreateTerritory validates input for creating a territory.
func ValidateCreateTerritory(input CreateTerritoryInput) error {
	if err := validateAPIName(input.APIName); err != nil {
		return err
	}
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateUpdateTerritory validates input for updating a territory.
func ValidateUpdateTerritory(input UpdateTerritoryInput) error {
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateSetObjectDefault validates input for setting a territory object default.
func ValidateSetObjectDefault(input SetObjectDefaultInput) error {
	if input.AccessLevel != "read" && input.AccessLevel != "read_write" {
		return apperror.Validation("access_level must be 'read' or 'read_write'")
	}
	return nil
}

// ValidateCreateAssignmentRule validates input for creating an assignment rule.
func ValidateCreateAssignmentRule(input CreateAssignmentRuleInput) error {
	if input.CriteriaField == "" {
		return apperror.Validation("criteria_field is required")
	}
	switch input.CriteriaOp {
	case "eq", "neq", "in", "gt", "lt", "contains":
	default:
		return apperror.Validation(fmt.Sprintf("criteria_op must be one of: eq, neq, in, gt, lt, contains; got '%s'", input.CriteriaOp))
	}
	if input.CriteriaValue == "" {
		return apperror.Validation("criteria_value is required")
	}
	return nil
}

// ValidateUpdateAssignmentRule validates input for updating an assignment rule.
func ValidateUpdateAssignmentRule(input UpdateAssignmentRuleInput) error {
	if input.CriteriaField == "" {
		return apperror.Validation("criteria_field is required")
	}
	switch input.CriteriaOp {
	case "eq", "neq", "in", "gt", "lt", "contains":
	default:
		return apperror.Validation(fmt.Sprintf("criteria_op must be one of: eq, neq, in, gt, lt, contains; got '%s'", input.CriteriaOp))
	}
	if input.CriteriaValue == "" {
		return apperror.Validation("criteria_value is required")
	}
	return nil
}

// ValidateAssignRecord validates input for assigning a record to a territory.
func ValidateAssignRecord(input AssignRecordInput) error {
	if input.Reason != "" && input.Reason != "manual" && input.Reason != "assignment_rule" {
		return apperror.Validation("reason must be 'manual' or 'assignment_rule'")
	}
	return nil
}

func validateAPIName(apiName string) error {
	if apiName == "" {
		return apperror.Validation("api_name is required")
	}
	if len(apiName) > 100 {
		return apperror.Validation("api_name must be at most 100 characters")
	}
	if !apiNamePattern.MatchString(apiName) {
		return apperror.Validation("api_name must start with a letter, contain only letters, digits, and underscores, and end with a letter or digit")
	}
	return nil
}
