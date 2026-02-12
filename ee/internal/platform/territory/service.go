//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import (
	"context"

	"github.com/google/uuid"
)

// ModelService defines business logic for territory models.
type ModelService interface {
	Create(ctx context.Context, input CreateModelInput) (*TerritoryModel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*TerritoryModel, error)
	GetActive(ctx context.Context) (*TerritoryModel, error)
	List(ctx context.Context, page, perPage int32) ([]TerritoryModel, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateModelInput) (*TerritoryModel, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	Archive(ctx context.Context, id uuid.UUID) error
}

// TerritoryService defines business logic for territories.
type TerritoryService interface {
	Create(ctx context.Context, input CreateTerritoryInput) (*Territory, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Territory, error)
	ListByModelID(ctx context.Context, modelID uuid.UUID) ([]Territory, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateTerritoryInput) (*Territory, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ObjectDefaultService defines business logic for territory object defaults.
type ObjectDefaultService interface {
	Set(ctx context.Context, input SetObjectDefaultInput) (*TerritoryObjectDefault, error)
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]TerritoryObjectDefault, error)
	Remove(ctx context.Context, territoryID, objectID uuid.UUID) error
}

// UserAssignmentService defines business logic for user-territory assignments.
type UserAssignmentService interface {
	Assign(ctx context.Context, input AssignUserInput) (*UserTerritoryAssignment, error)
	Unassign(ctx context.Context, userID, territoryID uuid.UUID) error
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]UserTerritoryAssignment, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]UserTerritoryAssignment, error)
}

// RecordAssignmentService defines business logic for record-territory assignments.
type RecordAssignmentService interface {
	Assign(ctx context.Context, input AssignRecordInput) (*RecordTerritoryAssignment, error)
	Unassign(ctx context.Context, recordID, objectID, territoryID uuid.UUID) error
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]RecordTerritoryAssignment, error)
	ListByRecordID(ctx context.Context, recordID, objectID uuid.UUID) ([]RecordTerritoryAssignment, error)
}

// AssignmentRuleService defines business logic for territory assignment rules.
type AssignmentRuleService interface {
	Create(ctx context.Context, input CreateAssignmentRuleInput) (*AssignmentRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AssignmentRule, error)
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]AssignmentRule, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateAssignmentRuleInput) (*AssignmentRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ObjectDefinitionLookup provides object metadata needed by territory services.
type ObjectDefinitionLookup interface {
	GetTableName(ctx context.Context, objectID uuid.UUID) (string, error)
}
