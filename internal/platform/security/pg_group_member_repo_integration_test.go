//go:build integration

package security_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/testutil"
)

func TestPgGroupMemberRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	ctx := context.Background()

	testutil.TruncateTables(t, pool,
		"iam.group_members",
		"iam.groups",
		"iam.users",
		"iam.profiles",
		"iam.permission_sets",
		"iam.user_roles",
	)

	memberRepo := security.NewPgGroupMemberRepository(pool)
	groupRepo := security.NewPgGroupRepository(pool)

	// --- Prerequisites: create role, PS, profile, user, and two groups ---

	roleRepo := security.NewPgUserRoleRepository(pool)
	psRepo := security.NewPgPermissionSetRepository(pool)
	profileRepo := security.NewPgProfileRepository(pool)
	userRepo := security.NewPgUserRepository(pool)

	// Create permission set.
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	ps, err := psRepo.Create(ctx, tx, security.CreatePermissionSetInput{
		APIName:     "TestMemberPS",
		Label:       "Test Member PS",
		Description: "PS for group member integration tests",
		PSType:      security.PSTypeGrant,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create profile.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	profile, err := profileRepo.Create(ctx, tx, &security.Profile{
		APIName:             "TestMemberProfile",
		Label:               "Test Member Profile",
		Description:         "Profile for group member integration tests",
		BasePermissionSetID: ps.ID,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create role.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	role, err := roleRepo.Create(ctx, tx, security.CreateUserRoleInput{
		APIName:     "TestMemberRole",
		Label:       "Test Member Role",
		Description: "Role for group member integration tests",
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create user.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	user, err := userRepo.Create(ctx, tx, security.CreateUserInput{
		Username:  "membertest_user",
		Email:     "membertest@example.com",
		FirstName: "Member",
		LastName:  "Tester",
		ProfileID: profile.ID,
		RoleID:    &role.ID,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create two groups: one public, one role.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	publicGroup, err := groupRepo.Create(ctx, tx, security.CreateGroupInput{
		APIName:   "MemberTestPublic",
		Label:     "Member Test Public",
		GroupType: security.GroupTypePublic,
	})
	require.NoError(t, err)

	roleGroup, err := groupRepo.Create(ctx, tx, security.CreateGroupInput{
		APIName:       "MemberTestRole",
		Label:         "Member Test Role",
		GroupType:     security.GroupTypeRole,
		RelatedRoleID: &role.ID,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// --- Test: Add user member to group ---
	var userMember *security.GroupMember
	t.Run("add user member to group", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		m, err := memberRepo.Add(ctx, tx, security.AddGroupMemberInput{
			GroupID:      publicGroup.ID,
			MemberUserID: &user.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, m.ID)
		assert.Equal(t, publicGroup.ID, m.GroupID)
		require.NotNil(t, m.MemberUserID)
		assert.Equal(t, user.ID, *m.MemberUserID)
		assert.Nil(t, m.MemberGroupID)
		assert.False(t, m.CreatedAt.IsZero())
		userMember = m
	})

	// --- Test: Add nested group member ---
	t.Run("add nested group member", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		m, err := memberRepo.Add(ctx, tx, security.AddGroupMemberInput{
			GroupID:       publicGroup.ID,
			MemberGroupID: &roleGroup.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, m.ID)
		assert.Equal(t, publicGroup.ID, m.GroupID)
		assert.Nil(t, m.MemberUserID)
		require.NotNil(t, m.MemberGroupID)
		assert.Equal(t, roleGroup.ID, *m.MemberGroupID)
	})

	// --- Test: ListByGroupID ---
	t.Run("ListByGroupID returns correct members", func(t *testing.T) {
		members, err := memberRepo.ListByGroupID(ctx, publicGroup.ID)
		require.NoError(t, err)
		assert.Len(t, members, 2)

		// Verify we have both a user member and a group member.
		var hasUserMember, hasGroupMember bool
		for _, m := range members {
			if m.MemberUserID != nil && *m.MemberUserID == user.ID {
				hasUserMember = true
			}
			if m.MemberGroupID != nil && *m.MemberGroupID == roleGroup.ID {
				hasGroupMember = true
			}
		}
		assert.True(t, hasUserMember, "expected user member in group")
		assert.True(t, hasGroupMember, "expected nested group member in group")
	})

	// --- Test: ListByUserID ---
	t.Run("ListByUserID returns groups the user belongs to", func(t *testing.T) {
		members, err := memberRepo.ListByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, members, 1)
		assert.Equal(t, publicGroup.ID, members[0].GroupID)
		require.NotNil(t, members[0].MemberUserID)
		assert.Equal(t, user.ID, *members[0].MemberUserID)
	})

	t.Run("ListByUserID returns empty for user not in any group", func(t *testing.T) {
		members, err := memberRepo.ListByUserID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, members)
	})

	// --- Test: Remove user member ---
	t.Run("remove user member from group", func(t *testing.T) {
		_ = userMember // ensure it was set

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = memberRepo.Remove(ctx, tx, publicGroup.ID, &user.ID, nil)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		// Only the nested group member should remain.
		members, err := memberRepo.ListByGroupID(ctx, publicGroup.ID)
		require.NoError(t, err)
		assert.Len(t, members, 1)
		require.NotNil(t, members[0].MemberGroupID)
		assert.Equal(t, roleGroup.ID, *members[0].MemberGroupID)
	})

	// --- Test: DeleteByGroupID ---
	t.Run("DeleteByGroupID removes all members", func(t *testing.T) {
		// Re-add the user member so there are 2 members again.
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		_, err = memberRepo.Add(ctx, tx, security.AddGroupMemberInput{
			GroupID:      publicGroup.ID,
			MemberUserID: &user.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		members, err := memberRepo.ListByGroupID(ctx, publicGroup.ID)
		require.NoError(t, err)
		assert.Len(t, members, 2)

		// Now delete all members.
		tx, err = pool.Begin(ctx)
		require.NoError(t, err)
		err = memberRepo.DeleteByGroupID(ctx, tx, publicGroup.ID)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		// Verify empty.
		members, err = memberRepo.ListByGroupID(ctx, publicGroup.ID)
		require.NoError(t, err)
		assert.Empty(t, members)
	})
}
