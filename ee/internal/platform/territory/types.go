//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

package territory

import (
	"time"

	"github.com/google/uuid"
)

// ModelStatus represents the lifecycle status of a territory model.
type ModelStatus string

const (
	ModelStatusPlanning ModelStatus = "planning"
	ModelStatusActive   ModelStatus = "active"
	ModelStatusArchived ModelStatus = "archived"
)

// TerritoryModel represents a named territory model with lifecycle.
type TerritoryModel struct {
	ID          uuid.UUID   `json:"id"`
	APIName     string      `json:"api_name"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Status      ModelStatus `json:"status"`
	ActivatedAt *time.Time  `json:"activated_at"`
	ArchivedAt  *time.Time  `json:"archived_at"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Territory represents a node in the territory hierarchy.
type Territory struct {
	ID          uuid.UUID  `json:"id"`
	ModelID     uuid.UUID  `json:"model_id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	APIName     string     `json:"api_name"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TerritoryObjectDefault defines the access level for an object within a territory.
type TerritoryObjectDefault struct {
	ID          uuid.UUID `json:"id"`
	TerritoryID uuid.UUID `json:"territory_id"`
	ObjectID    uuid.UUID `json:"object_id"`
	AccessLevel string    `json:"access_level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserTerritoryAssignment represents a user assigned to a territory.
type UserTerritoryAssignment struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	TerritoryID uuid.UUID `json:"territory_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// RecordTerritoryAssignment represents a record assigned to a territory.
type RecordTerritoryAssignment struct {
	ID          uuid.UUID `json:"id"`
	RecordID    uuid.UUID `json:"record_id"`
	ObjectID    uuid.UUID `json:"object_id"`
	TerritoryID uuid.UUID `json:"territory_id"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

// AssignmentRule defines a criteria-based rule for automatic record assignment.
type AssignmentRule struct {
	ID            uuid.UUID `json:"id"`
	TerritoryID   uuid.UUID `json:"territory_id"`
	ObjectID      uuid.UUID `json:"object_id"`
	IsActive      bool      `json:"is_active"`
	RuleOrder     int       `json:"rule_order"`
	CriteriaField string    `json:"criteria_field"`
	CriteriaOp    string    `json:"criteria_op"`
	CriteriaValue string    `json:"criteria_value"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
