package rls

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

// Enforcer checks row-level security for records.
type Enforcer interface {
	CanReadRecord(ctx context.Context, userID, objectID, recordOwnerID uuid.UUID) error
	CanUpdateRecord(ctx context.Context, userID, objectID, recordOwnerID uuid.UUID) error
	BuildWhereClause(ctx context.Context, userID, objectID uuid.UUID) (string, []interface{}, error)
	GetVisibility(ctx context.Context, objectID uuid.UUID) (string, error)
}

type enforcerImpl struct {
	rlsCacheRepo    security.RLSEffectiveCacheRepository
	metadataAdapter security.MetadataRLSAdapter
}

// NewEnforcer creates a new RLS Enforcer.
func NewEnforcer(
	rlsCacheRepo security.RLSEffectiveCacheRepository,
	metadataAdapter security.MetadataRLSAdapter,
) Enforcer {
	return &enforcerImpl{
		rlsCacheRepo:    rlsCacheRepo,
		metadataAdapter: metadataAdapter,
	}
}

func (e *enforcerImpl) GetVisibility(ctx context.Context, objectID uuid.UUID) (string, error) {
	return e.metadataAdapter.GetObjectVisibility(ctx, objectID)
}

func (e *enforcerImpl) CanReadRecord(ctx context.Context, userID, objectID, recordOwnerID uuid.UUID) error {
	visibility, err := e.metadataAdapter.GetObjectVisibility(ctx, objectID)
	if err != nil {
		return fmt.Errorf("rlsEnforcer.CanReadRecord: %w", err)
	}

	switch visibility {
	case "public_read", "public_read_write":
		return nil
	case "private", "controlled_by_parent":
		return e.checkReadAccess(ctx, userID, objectID, recordOwnerID)
	default:
		return fmt.Errorf("rlsEnforcer.CanReadRecord: unknown visibility %q", visibility)
	}
}

func (e *enforcerImpl) CanUpdateRecord(ctx context.Context, userID, objectID, recordOwnerID uuid.UUID) error {
	visibility, err := e.metadataAdapter.GetObjectVisibility(ctx, objectID)
	if err != nil {
		return fmt.Errorf("rlsEnforcer.CanUpdateRecord: %w", err)
	}

	switch visibility {
	case "public_read_write":
		return nil
	case "public_read", "private", "controlled_by_parent":
		return e.checkWriteAccess(ctx, userID, objectID, recordOwnerID)
	default:
		return fmt.Errorf("rlsEnforcer.CanUpdateRecord: unknown visibility %q", visibility)
	}
}

func (e *enforcerImpl) checkReadAccess(ctx context.Context, userID, objectID, recordOwnerID uuid.UUID) error {
	// 1. Owner can always read their own records
	if userID == recordOwnerID {
		return nil
	}

	// 2. Check role hierarchy (visible owners)
	visibleOwners, err := e.rlsCacheRepo.GetVisibleOwners(ctx, userID)
	if err != nil {
		return fmt.Errorf("rlsEnforcer.checkReadAccess: %w", err)
	}
	for _, ownerID := range visibleOwners {
		if ownerID == recordOwnerID {
			return nil
		}
	}

	// 3. Check share table via effective group memberships
	// (share table query is handled at the SQL level via BuildWhereClause)
	// For point checks, we would query the share table — but that requires the table name
	// This is a simplified check; full enforcement uses BuildWhereClause in SOQL

	return fmt.Errorf("rlsEnforcer.checkReadAccess: %w",
		apperror.Forbidden("insufficient sharing privileges to read this record"))
}

func (e *enforcerImpl) checkWriteAccess(ctx context.Context, userID, objectID, recordOwnerID uuid.UUID) error {
	// Owner can always write their own records
	if userID == recordOwnerID {
		return nil
	}

	// Role hierarchy only grants Read, not Write (ADR-0011)
	// Write access requires explicit share with access_level = 'read_write'
	// Full enforcement is in SOQL query builder

	return fmt.Errorf("rlsEnforcer.checkWriteAccess: %w",
		apperror.Forbidden("insufficient sharing privileges to update this record"))
}

// BuildWhereClause generates a SQL WHERE fragment for RLS filtering.
// Returns the clause (without leading AND/WHERE) and bind parameters.
func (e *enforcerImpl) BuildWhereClause(ctx context.Context, userID, objectID uuid.UUID) (string, []interface{}, error) {
	visibility, err := e.metadataAdapter.GetObjectVisibility(ctx, objectID)
	if err != nil {
		return "", nil, fmt.Errorf("rlsEnforcer.BuildWhereClause: %w", err)
	}

	switch visibility {
	case "public_read_write", "public_read":
		// No RLS filtering needed
		return "TRUE", nil, nil

	case "private", "controlled_by_parent":
		tableName, err := e.metadataAdapter.GetObjectTableName(ctx, objectID)
		if err != nil {
			return "", nil, fmt.Errorf("rlsEnforcer.BuildWhereClause: %w", err)
		}

		// Get visible owners for the user (includes self + role hierarchy subordinates)
		visibleOwners, err := e.rlsCacheRepo.GetVisibleOwners(ctx, userID)
		if err != nil {
			return "", nil, fmt.Errorf("rlsEnforcer.BuildWhereClause: %w", err)
		}

		// Get user's effective group memberships
		groupIDs, err := e.rlsCacheRepo.GetGroupMemberships(ctx, userID)
		if err != nil {
			return "", nil, fmt.Errorf("rlsEnforcer.BuildWhereClause: %w", err)
		}

		var conditions []string
		var params []interface{}
		paramIdx := 1

		// Condition 1: owner_id IN (visible owners)
		if len(visibleOwners) > 0 {
			placeholders := make([]string, len(visibleOwners))
			for i, ownerID := range visibleOwners {
				placeholders[i] = fmt.Sprintf("$%d", paramIdx)
				params = append(params, ownerID)
				paramIdx++
			}
			conditions = append(conditions, fmt.Sprintf("owner_id IN (%s)", strings.Join(placeholders, ",")))
		}

		// Condition 2: record exists in share table for user's groups
		if len(groupIDs) > 0 {
			shareTable := fmt.Sprintf(`"%s__share"`, tableName)
			groupPlaceholders := make([]string, len(groupIDs))
			for i, gid := range groupIDs {
				groupPlaceholders[i] = fmt.Sprintf("$%d", paramIdx)
				params = append(params, gid)
				paramIdx++
			}
			conditions = append(conditions, fmt.Sprintf(
				"id IN (SELECT record_id FROM %s WHERE group_id IN (%s))",
				shareTable, strings.Join(groupPlaceholders, ",")))
		}

		if len(conditions) == 0 {
			// User can only see own records (owner_id = user_id)
			conditions = append(conditions, fmt.Sprintf("owner_id = $%d", paramIdx))
			params = append(params, userID)
		}

		return "(" + strings.Join(conditions, " OR ") + ")", params, nil

	default:
		return "", nil, fmt.Errorf("rlsEnforcer.BuildWhereClause: unknown visibility %q", visibility)
	}
}
