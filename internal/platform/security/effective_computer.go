package security

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type effectiveComputerImpl struct {
	txBeginner      TxBeginner
	userRepo        UserRepository
	profileRepo     ProfileRepository
	psToUserRepo    PermissionSetToUserRepository
	psRepo          PermissionSetRepository
	objPermRepo     AllObjectPermissions
	fieldPermRepo   AllFieldPermissions
	effectiveRepo   EffectivePermissionRepository
	metadataLister  MetadataFieldLister
}

// NewEffectiveComputer creates a new EffectiveComputer.
func NewEffectiveComputer(
	txBeginner TxBeginner,
	userRepo UserRepository,
	profileRepo ProfileRepository,
	psToUserRepo PermissionSetToUserRepository,
	psRepo PermissionSetRepository,
	objPermRepo AllObjectPermissions,
	fieldPermRepo AllFieldPermissions,
	effectiveRepo EffectivePermissionRepository,
	metadataLister MetadataFieldLister,
) EffectiveComputer {
	return &effectiveComputerImpl{
		txBeginner:     txBeginner,
		userRepo:       userRepo,
		profileRepo:    profileRepo,
		psToUserRepo:   psToUserRepo,
		psRepo:         psRepo,
		objPermRepo:    objPermRepo,
		fieldPermRepo:  fieldPermRepo,
		effectiveRepo:  effectiveRepo,
		metadataLister: metadataLister,
	}
}

func (c *effectiveComputerImpl) RecomputeForUser(ctx context.Context, userID uuid.UUID) error {
	user, err := c.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: get user: %w", err)
	}
	if user == nil {
		slog.Warn("effectiveComputer: user not found, skipping", "user_id", userID)
		return nil
	}

	profile, err := c.profileRepo.GetByID(ctx, user.ProfileID)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: get profile: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: profile not found for user %s", userID)
	}

	// Collect all PS IDs for this user: base PS from profile + assigned PSes
	assignments, err := c.psToUserRepo.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: list assignments: %w", err)
	}

	psIDs := make([]uuid.UUID, 0, len(assignments)+1)
	psIDs = append(psIDs, profile.BasePermissionSetID)
	for _, a := range assignments {
		psIDs = append(psIDs, a.PermissionSetID)
	}

	// Load PS metadata to partition by type
	grants := make([]uuid.UUID, 0, len(psIDs))
	denies := make([]uuid.UUID, 0)
	for _, psID := range psIDs {
		ps, err := c.psRepo.GetByID(ctx, psID)
		if err != nil {
			return fmt.Errorf("effectiveComputer.RecomputeForUser: get PS %s: %w", psID, err)
		}
		if ps == nil {
			continue
		}
		if ps.PSType == PSTypeDeny {
			denies = append(denies, psID)
		} else {
			grants = append(grants, psID)
		}
	}

	// Load all object permissions for grant and deny PSes
	grantObjPerms, err := c.objPermRepo.ListByPermissionSetIDs(ctx, grants)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: grant obj perms: %w", err)
	}
	denyObjPerms, err := c.objPermRepo.ListByPermissionSetIDs(ctx, denies)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: deny obj perms: %w", err)
	}

	// Load all field permissions
	grantFieldPerms, err := c.fieldPermRepo.ListByPermissionSetIDs(ctx, grants)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: grant field perms: %w", err)
	}
	denyFieldPerms, err := c.fieldPermRepo.ListByPermissionSetIDs(ctx, denies)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: deny field perms: %w", err)
	}

	// Compute effective OLS per object
	objGrants := groupByObject(grantObjPerms)
	objDenies := groupByObject(denyObjPerms)
	effectiveOLS := make(map[uuid.UUID]int)
	for objectID, perms := range objGrants {
		effectiveOLS[objectID] = ComputeEffective(perms, objDenies[objectID])
	}

	// Compute effective FLS per field
	fieldGrants := groupByField(grantFieldPerms)
	fieldDenies := groupByField(denyFieldPerms)
	effectiveFLS := make(map[uuid.UUID]int)
	for fieldID, perms := range fieldGrants {
		effectiveFLS[fieldID] = ComputeEffective(perms, fieldDenies[fieldID])
	}

	// Build field lists per object
	objectIDs, err := c.metadataLister.ListAllObjectIDs(ctx)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForUser: list objects: %w", err)
	}

	type fieldListEntry struct {
		objectID   uuid.UUID
		mask       int
		fieldNames []string
	}
	var fieldLists []fieldListEntry

	for _, objectID := range objectIDs {
		fields, err := c.metadataLister.ListFieldsByObjectID(ctx, objectID)
		if err != nil {
			return fmt.Errorf("effectiveComputer.RecomputeForUser: list fields for %s: %w", objectID, err)
		}

		var readableNames []string
		var writableNames []string
		for _, f := range fields {
			perm := effectiveFLS[f.ID]
			if HasFLS(perm, FLSRead) {
				readableNames = append(readableNames, f.APIName)
			}
			if HasFLS(perm, FLSWrite) {
				writableNames = append(writableNames, f.APIName)
			}
		}

		if len(readableNames) > 0 {
			fieldLists = append(fieldLists, fieldListEntry{objectID, FLSRead, readableNames})
		}
		if len(writableNames) > 0 {
			fieldLists = append(fieldLists, fieldListEntry{objectID, FLSWrite, writableNames})
		}
	}

	// Write everything in one transaction
	return withTx(ctx, c.txBeginner, func(tx pgx.Tx) error {
		if err := c.effectiveRepo.DeleteByUserID(ctx, tx, userID); err != nil {
			return fmt.Errorf("effectiveComputer.RecomputeForUser: clear caches: %w", err)
		}

		for objectID, perm := range effectiveOLS {
			if err := c.effectiveRepo.UpsertOLS(ctx, tx, userID, objectID, perm); err != nil {
				return fmt.Errorf("effectiveComputer.RecomputeForUser: upsert OLS: %w", err)
			}
		}

		for fieldID, perm := range effectiveFLS {
			if err := c.effectiveRepo.UpsertFLS(ctx, tx, userID, fieldID, perm); err != nil {
				return fmt.Errorf("effectiveComputer.RecomputeForUser: upsert FLS: %w", err)
			}
		}

		for _, fl := range fieldLists {
			if err := c.effectiveRepo.UpsertFieldList(ctx, tx, userID, fl.objectID, fl.mask, fl.fieldNames); err != nil {
				return fmt.Errorf("effectiveComputer.RecomputeForUser: upsert field list: %w", err)
			}
		}

		return nil
	})
}

func (c *effectiveComputerImpl) RecomputeForPermissionSet(ctx context.Context, psID uuid.UUID) error {
	// Find all users who have this PS assigned (directly or via profile)
	assignments, err := c.psToUserRepo.ListByPermissionSetID(ctx, psID)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForPermissionSet: list assignments: %w", err)
	}

	userIDs := make(map[uuid.UUID]bool)
	for _, a := range assignments {
		userIDs[a.UserID] = true
	}

	// Also find users whose profile's base PS is this one
	// For MVP, iterate all users (small scale)
	allUsers, err := c.userRepo.List(ctx, 10000, 0)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeForPermissionSet: list users: %w", err)
	}
	for _, u := range allUsers {
		profile, err := c.profileRepo.GetByID(ctx, u.ProfileID)
		if err != nil {
			continue
		}
		if profile != nil && profile.BasePermissionSetID == psID {
			userIDs[u.ID] = true
		}
	}

	for uid := range userIDs {
		if err := c.RecomputeForUser(ctx, uid); err != nil {
			slog.Error("effectiveComputer: failed to recompute for user",
				"user_id", uid, "ps_id", psID, "error", err)
		}
	}

	return nil
}

func (c *effectiveComputerImpl) RecomputeAll(ctx context.Context) error {
	users, err := c.userRepo.List(ctx, 10000, 0)
	if err != nil {
		return fmt.Errorf("effectiveComputer.RecomputeAll: list users: %w", err)
	}

	for _, u := range users {
		if err := c.RecomputeForUser(ctx, u.ID); err != nil {
			slog.Error("effectiveComputer: failed to recompute for user",
				"user_id", u.ID, "error", err)
		}
	}

	return nil
}

func groupByObject(perms []ObjectPermission) map[uuid.UUID][]int {
	result := make(map[uuid.UUID][]int)
	for _, p := range perms {
		result[p.ObjectID] = append(result[p.ObjectID], p.Permissions)
	}
	return result
}

func groupByField(perms []FieldPermission) map[uuid.UUID][]int {
	result := make(map[uuid.UUID][]int)
	for _, p := range perms {
		result[p.FieldID] = append(result[p.FieldID], p.Permissions)
	}
	return result
}
