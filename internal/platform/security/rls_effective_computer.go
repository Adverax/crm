package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type rlsEffectiveComputerImpl struct {
	txBeginner      TxBeginner
	roleRepo        UserRoleRepository
	userRepo        UserRepository
	groupRepo       GroupRepository
	memberRepo      GroupMemberRepository
	rlsCacheRepo    RLSEffectiveCacheRepository
	metadataAdapter MetadataRLSAdapter
}

// MetadataRLSAdapter provides metadata queries needed for RLS computation.
type MetadataRLSAdapter interface {
	GetObjectVisibility(ctx context.Context, objectID uuid.UUID) (string, error)
	GetObjectTableName(ctx context.Context, objectID uuid.UUID) (string, error)
	ListCompositionFields(ctx context.Context) ([]CompositionFieldInfo, error)
}

// CompositionFieldInfo holds minimal info about a composition reference field.
type CompositionFieldInfo struct {
	ChildObjectID  uuid.UUID
	ParentObjectID uuid.UUID
}

// NewRLSEffectiveComputer creates a new RLSEffectiveComputer.
func NewRLSEffectiveComputer(
	txBeginner TxBeginner,
	roleRepo UserRoleRepository,
	userRepo UserRepository,
	groupRepo GroupRepository,
	memberRepo GroupMemberRepository,
	rlsCacheRepo RLSEffectiveCacheRepository,
	metadataAdapter MetadataRLSAdapter,
) RLSEffectiveComputer {
	return &rlsEffectiveComputerImpl{
		txBeginner:      txBeginner,
		roleRepo:        roleRepo,
		userRepo:        userRepo,
		groupRepo:       groupRepo,
		memberRepo:      memberRepo,
		rlsCacheRepo:    rlsCacheRepo,
		metadataAdapter: metadataAdapter,
	}
}

func (c *rlsEffectiveComputerImpl) RecomputeRoleHierarchy(ctx context.Context) error {
	roles, err := c.roleRepo.List(ctx, 10000, 0)
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeRoleHierarchy: list roles: %w", err)
	}

	// Build parent map
	parentMap := make(map[uuid.UUID]*uuid.UUID)
	for _, r := range roles {
		parentMap[r.ID] = r.ParentID
	}

	// Compute transitive closure
	var entries []EffectiveRoleHierarchy
	for _, role := range roles {
		// Self entry (depth 0)
		entries = append(entries, EffectiveRoleHierarchy{
			AncestorRoleID:   role.ID,
			DescendantRoleID: role.ID,
			Depth:            0,
		})

		// Walk up the parent chain
		current := role.ID
		depth := 0
		visited := map[uuid.UUID]bool{role.ID: true}
		for {
			pid := parentMap[current]
			if pid == nil {
				break
			}
			if visited[*pid] {
				break // prevent infinite loops
			}
			visited[*pid] = true
			depth++
			entries = append(entries, EffectiveRoleHierarchy{
				AncestorRoleID:   *pid,
				DescendantRoleID: role.ID,
				Depth:            depth,
			})
			current = *pid
		}
	}

	return withTx(ctx, c.txBeginner, func(tx pgx.Tx) error {
		return c.rlsCacheRepo.ReplaceRoleHierarchy(ctx, tx, entries)
	})
}

func (c *rlsEffectiveComputerImpl) RecomputeVisibleOwnersForUser(ctx context.Context, userID uuid.UUID) error {
	user, err := c.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeVisibleOwnersForUser: get user: %w", err)
	}
	if user == nil || user.RoleID == nil {
		// User without role can only see own records — store self as visible owner
		entries := []EffectiveVisibleOwner{
			{UserID: userID, VisibleOwnerID: userID},
		}
		return withTx(ctx, c.txBeginner, func(tx pgx.Tx) error {
			return c.rlsCacheRepo.ReplaceVisibleOwners(ctx, tx, userID, entries)
		})
	}

	// Get all descendant roles (role hierarchy gives Read access)
	descendantRoleIDs, err := c.rlsCacheRepo.GetRoleDescendants(ctx, *user.RoleID)
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeVisibleOwnersForUser: get descendants: %w", err)
	}

	// Collect all role IDs (own role + descendants)
	allRoleIDs := append([]uuid.UUID{*user.RoleID}, descendantRoleIDs...)

	// Find all users with those roles
	allUsers, err := c.userRepo.List(ctx, 10000, 0)
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeVisibleOwnersForUser: list users: %w", err)
	}

	roleSet := make(map[uuid.UUID]bool)
	for _, rid := range allRoleIDs {
		roleSet[rid] = true
	}

	var entries []EffectiveVisibleOwner
	// Self is always visible
	entries = append(entries, EffectiveVisibleOwner{UserID: userID, VisibleOwnerID: userID})

	for _, u := range allUsers {
		if u.ID == userID {
			continue
		}
		if u.RoleID != nil && roleSet[*u.RoleID] {
			entries = append(entries, EffectiveVisibleOwner{
				UserID:         userID,
				VisibleOwnerID: u.ID,
			})
		}
	}

	return withTx(ctx, c.txBeginner, func(tx pgx.Tx) error {
		return c.rlsCacheRepo.ReplaceVisibleOwners(ctx, tx, userID, entries)
	})
}

func (c *rlsEffectiveComputerImpl) RecomputeVisibleOwnersAll(ctx context.Context) error {
	users, err := c.userRepo.List(ctx, 10000, 0)
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeVisibleOwnersAll: list users: %w", err)
	}

	for _, u := range users {
		if err := c.RecomputeVisibleOwnersForUser(ctx, u.ID); err != nil {
			return fmt.Errorf("rlsComputer.RecomputeVisibleOwnersAll: user %s: %w", u.ID, err)
		}
	}
	return nil
}

func (c *rlsEffectiveComputerImpl) RecomputeGroupMembersForGroup(ctx context.Context, groupID uuid.UUID) error {
	entries, err := c.flattenGroupMembers(ctx, groupID, make(map[uuid.UUID]bool))
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeGroupMembersForGroup: %w", err)
	}

	return withTx(ctx, c.txBeginner, func(tx pgx.Tx) error {
		return c.rlsCacheRepo.ReplaceGroupMembers(ctx, tx, groupID, entries)
	})
}

func (c *rlsEffectiveComputerImpl) RecomputeGroupMembersAll(ctx context.Context) error {
	groups, err := c.groupRepo.List(ctx, 10000, 0)
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeGroupMembersAll: list groups: %w", err)
	}

	var allEntries []EffectiveGroupMember
	for _, g := range groups {
		entries, err := c.flattenGroupMembers(ctx, g.ID, make(map[uuid.UUID]bool))
		if err != nil {
			return fmt.Errorf("rlsComputer.RecomputeGroupMembersAll: group %s: %w", g.ID, err)
		}
		allEntries = append(allEntries, entries...)
	}

	return withTx(ctx, c.txBeginner, func(tx pgx.Tx) error {
		return c.rlsCacheRepo.ReplaceGroupMembersAll(ctx, tx, allEntries)
	})
}

// flattenGroupMembers recursively resolves nested group memberships.
func (c *rlsEffectiveComputerImpl) flattenGroupMembers(ctx context.Context, groupID uuid.UUID, visited map[uuid.UUID]bool) ([]EffectiveGroupMember, error) {
	if visited[groupID] {
		return nil, nil // prevent infinite recursion
	}
	visited[groupID] = true

	members, err := c.memberRepo.ListByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var entries []EffectiveGroupMember
	for _, m := range members {
		if m.MemberUserID != nil {
			entries = append(entries, EffectiveGroupMember{
				GroupID: groupID,
				UserID:  *m.MemberUserID,
			})
		}
		if m.MemberGroupID != nil {
			// Recursively flatten nested group
			nested, err := c.flattenGroupMembers(ctx, *m.MemberGroupID, visited)
			if err != nil {
				return nil, err
			}
			// All users from nested group are also members of this group
			for _, n := range nested {
				entries = append(entries, EffectiveGroupMember{
					GroupID: groupID,
					UserID:  n.UserID,
				})
			}
		}
	}

	return entries, nil
}

func (c *rlsEffectiveComputerImpl) RecomputeObjectHierarchy(ctx context.Context) error {
	compositionFields, err := c.metadataAdapter.ListCompositionFields(ctx)
	if err != nil {
		return fmt.Errorf("rlsComputer.RecomputeObjectHierarchy: %w", err)
	}

	// Build parent map: child -> parent
	parentMap := make(map[uuid.UUID]uuid.UUID)
	allObjects := make(map[uuid.UUID]bool)
	for _, f := range compositionFields {
		parentMap[f.ChildObjectID] = f.ParentObjectID
		allObjects[f.ChildObjectID] = true
		allObjects[f.ParentObjectID] = true
	}

	// Compute transitive closure
	var entries []EffectiveObjectHierarchy
	for objID := range allObjects {
		// Self
		entries = append(entries, EffectiveObjectHierarchy{
			AncestorObjectID:   objID,
			DescendantObjectID: objID,
			Depth:              0,
		})

		// Walk up
		current := objID
		depth := 0
		visited := map[uuid.UUID]bool{objID: true}
		for {
			pid, ok := parentMap[current]
			if !ok {
				break
			}
			if visited[pid] {
				break
			}
			visited[pid] = true
			depth++
			entries = append(entries, EffectiveObjectHierarchy{
				AncestorObjectID:   pid,
				DescendantObjectID: objID,
				Depth:              depth,
			})
			current = pid
		}
	}

	return withTx(ctx, c.txBeginner, func(tx pgx.Tx) error {
		return c.rlsCacheRepo.ReplaceObjectHierarchy(ctx, tx, entries)
	})
}
