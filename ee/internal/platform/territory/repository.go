//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ModelRepository defines data access for territory models.
type ModelRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreateModelInput) (*TerritoryModel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*TerritoryModel, error)
	GetByAPIName(ctx context.Context, apiName string) (*TerritoryModel, error)
	GetActive(ctx context.Context) (*TerritoryModel, error)
	List(ctx context.Context, limit, offset int32) ([]TerritoryModel, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateModelInput) (*TerritoryModel, error)
	UpdateStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status ModelStatus) error
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// TerritoryRepository defines data access for territories within a model.
type TerritoryRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreateTerritoryInput) (*Territory, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Territory, error)
	GetByAPIName(ctx context.Context, modelID uuid.UUID, apiName string) (*Territory, error)
	ListByModelID(ctx context.Context, modelID uuid.UUID) ([]Territory, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateTerritoryInput) (*Territory, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}

// ObjectDefaultRepository defines data access for territory object defaults.
type ObjectDefaultRepository interface {
	Upsert(ctx context.Context, tx pgx.Tx, input SetObjectDefaultInput) (*TerritoryObjectDefault, error)
	GetByTerritoryAndObject(ctx context.Context, territoryID, objectID uuid.UUID) (*TerritoryObjectDefault, error)
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]TerritoryObjectDefault, error)
	Delete(ctx context.Context, tx pgx.Tx, territoryID, objectID uuid.UUID) error
}

// UserAssignmentRepository defines data access for user-territory assignments.
type UserAssignmentRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input AssignUserInput) (*UserTerritoryAssignment, error)
	Delete(ctx context.Context, tx pgx.Tx, userID, territoryID uuid.UUID) error
	GetByUserAndTerritory(ctx context.Context, userID, territoryID uuid.UUID) (*UserTerritoryAssignment, error)
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]UserTerritoryAssignment, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]UserTerritoryAssignment, error)
}

// RecordAssignmentRepository defines data access for record-territory assignments.
type RecordAssignmentRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input AssignRecordInput) (*RecordTerritoryAssignment, error)
	Delete(ctx context.Context, tx pgx.Tx, recordID, objectID, territoryID uuid.UUID) error
	GetByRecordAndTerritory(ctx context.Context, recordID, objectID, territoryID uuid.UUID) (*RecordTerritoryAssignment, error)
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]RecordTerritoryAssignment, error)
	ListByRecordID(ctx context.Context, recordID, objectID uuid.UUID) ([]RecordTerritoryAssignment, error)
}

// AssignmentRuleRepository defines data access for territory assignment rules.
type AssignmentRuleRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreateAssignmentRuleInput) (*AssignmentRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AssignmentRule, error)
	ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]AssignmentRule, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]AssignmentRule, error)
	ListActiveByObjectID(ctx context.Context, objectID uuid.UUID) ([]AssignmentRule, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateAssignmentRuleInput) (*AssignmentRule, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}

// EffectiveRepository defines data access for territory effective caches and stored functions.
type EffectiveRepository interface {
	RebuildHierarchy(ctx context.Context, tx pgx.Tx, modelID uuid.UUID) error
	GenerateRecordShareEntries(ctx context.Context, tx pgx.Tx, recordID, objectID, territoryID uuid.UUID, shareTable string) error
	ActivateModel(ctx context.Context, tx pgx.Tx, modelID uuid.UUID) error
	GetUserTerritories(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetTerritoryGroupIDs(ctx context.Context, territoryIDs []uuid.UUID) ([]uuid.UUID, error)
}
