//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import "github.com/google/uuid"

// CreateModelInput contains input data for creating a territory model.
type CreateModelInput struct {
	APIName     string `json:"api_name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// UpdateModelInput contains input data for updating a territory model.
type UpdateModelInput struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// CreateTerritoryInput contains input data for creating a territory.
type CreateTerritoryInput struct {
	ModelID     uuid.UUID  `json:"model_id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	APIName     string     `json:"api_name"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
}

// UpdateTerritoryInput contains input data for updating a territory.
type UpdateTerritoryInput struct {
	ParentID    *uuid.UUID `json:"parent_id"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
}

// SetObjectDefaultInput contains input for setting a territory object default.
type SetObjectDefaultInput struct {
	TerritoryID uuid.UUID `json:"territory_id"`
	ObjectID    uuid.UUID `json:"object_id"`
	AccessLevel string    `json:"access_level"`
}

// AssignUserInput contains input for assigning a user to a territory.
type AssignUserInput struct {
	UserID      uuid.UUID `json:"user_id"`
	TerritoryID uuid.UUID `json:"territory_id"`
}

// AssignRecordInput contains input for assigning a record to a territory.
type AssignRecordInput struct {
	RecordID    uuid.UUID `json:"record_id"`
	ObjectID    uuid.UUID `json:"object_id"`
	TerritoryID uuid.UUID `json:"territory_id"`
	Reason      string    `json:"reason"`
}

// CreateAssignmentRuleInput contains input for creating an assignment rule.
type CreateAssignmentRuleInput struct {
	TerritoryID   uuid.UUID `json:"territory_id"`
	ObjectID      uuid.UUID `json:"object_id"`
	IsActive      bool      `json:"is_active"`
	RuleOrder     int       `json:"rule_order"`
	CriteriaField string    `json:"criteria_field"`
	CriteriaOp    string    `json:"criteria_op"`
	CriteriaValue string    `json:"criteria_value"`
}

// UpdateAssignmentRuleInput contains input for updating an assignment rule.
type UpdateAssignmentRuleInput struct {
	IsActive      bool   `json:"is_active"`
	RuleOrder     int    `json:"rule_order"`
	CriteriaField string `json:"criteria_field"`
	CriteriaOp    string `json:"criteria_op"`
	CriteriaValue string `json:"criteria_value"`
}
