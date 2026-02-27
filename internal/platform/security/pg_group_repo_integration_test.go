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

func TestPgGroupRepo_Integration(t *testing.T) {
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

	repo := security.NewPgGroupRepository(pool)

	// --- Prerequisites: create a role, permission set, profile, and user ---

	roleRepo := security.NewPgUserRoleRepository(pool)
	psRepo := security.NewPgPermissionSetRepository(pool)
	profileRepo := security.NewPgProfileRepository(pool)
	userRepo := security.NewPgUserRepository(pool)

	// Create permission set (needed for profile).
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	ps, err := psRepo.Create(ctx, tx, security.CreatePermissionSetInput{
		APIName:     "test_group_ps",
		Label:       "Test Group PS",
		Description: "PS for group integration tests",
		PSType:      security.PSTypeGrant,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create profile (needs base_permission_set_id).
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	profile, err := profileRepo.Create(ctx, tx, &security.Profile{
		APIName:             "test_group_profile",
		Label:               "Test Group Profile",
		Description:         "Profile for group integration tests",
		BasePermissionSetID: ps.ID,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create role.
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	role, err := roleRepo.Create(ctx, tx, security.CreateUserRoleInput{
		APIName:     "test_group_role",
		Label:       "Test Group Role",
		Description: "Role for group integration tests",
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create user (needs profile_id and role_id).
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	user, err := userRepo.Create(ctx, tx, security.CreateUserInput{
		Username:  "grouptest_user",
		Email:     "grouptest@example.com",
		FirstName: "Group",
		LastName:  "Tester",
		ProfileID: profile.ID,
		RoleID:    &role.ID,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// --- Test: Create public group (no related IDs) ---
	t.Run("create public group", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		g, err := repo.Create(ctx, tx, security.CreateGroupInput{
			APIName:   "all_users",
			Label:     "All Users",
			GroupType: security.GroupTypePublic,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, g.ID)
		assert.Equal(t, "all_users", g.APIName)
		assert.Equal(t, "All Users", g.Label)
		assert.Equal(t, security.GroupTypePublic, g.GroupType)
		assert.Nil(t, g.RelatedRoleID)
		assert.Nil(t, g.RelatedUserID)
		assert.False(t, g.CreatedAt.IsZero())
		assert.False(t, g.UpdatedAt.IsZero())
	})

	// --- Test: Create role group with RelatedRoleID ---
	var roleGroup *security.Group
	t.Run("create role group with related role ID", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		g, err := repo.Create(ctx, tx, security.CreateGroupInput{
			APIName:       "role_group_test_group_role",
			Label:         "Role Group: Test Group Role",
			GroupType:     security.GroupTypeRole,
			RelatedRoleID: &role.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, g.ID)
		assert.Equal(t, security.GroupTypeRole, g.GroupType)
		require.NotNil(t, g.RelatedRoleID)
		assert.Equal(t, role.ID, *g.RelatedRoleID)
		assert.Nil(t, g.RelatedUserID)
		roleGroup = g
	})

	// --- Test: Create personal group with RelatedUserID ---
	var personalGroup *security.Group
	t.Run("create personal group with related user ID", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		g, err := repo.Create(ctx, tx, security.CreateGroupInput{
			APIName:       "personal_group_grouptest_user",
			Label:         "Personal Group: grouptest_user",
			GroupType:     security.GroupTypePersonal,
			RelatedUserID: &user.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, g.ID)
		assert.Equal(t, security.GroupTypePersonal, g.GroupType)
		assert.Nil(t, g.RelatedRoleID)
		require.NotNil(t, g.RelatedUserID)
		assert.Equal(t, user.ID, *g.RelatedUserID)
		personalGroup = g
	})

	// --- Test: GetByID ---
	t.Run("GetByID returns existing group", func(t *testing.T) {
		g, err := repo.GetByID(ctx, roleGroup.ID)
		require.NoError(t, err)
		require.NotNil(t, g)
		assert.Equal(t, roleGroup.ID, g.ID)
		assert.Equal(t, "RoleGroup_TestGroupRole", g.APIName)
		assert.Equal(t, security.GroupTypeRole, g.GroupType)
	})

	// --- Test: GetByAPIName ---
	t.Run("GetByAPIName returns existing group", func(t *testing.T) {
		g, err := repo.GetByAPIName(ctx, "AllUsers")
		require.NoError(t, err)
		require.NotNil(t, g)
		assert.Equal(t, "AllUsers", g.APIName)
		assert.Equal(t, security.GroupTypePublic, g.GroupType)
	})

	// --- Test: GetByRelatedRoleID ---
	t.Run("GetByRelatedRoleID returns matching group", func(t *testing.T) {
		g, err := repo.GetByRelatedRoleID(ctx, role.ID, security.GroupTypeRole)
		require.NoError(t, err)
		require.NotNil(t, g)
		assert.Equal(t, roleGroup.ID, g.ID)
		assert.Equal(t, security.GroupTypeRole, g.GroupType)
		require.NotNil(t, g.RelatedRoleID)
		assert.Equal(t, role.ID, *g.RelatedRoleID)
	})

	t.Run("GetByRelatedRoleID returns nil for wrong group type", func(t *testing.T) {
		g, err := repo.GetByRelatedRoleID(ctx, role.ID, security.GroupTypeRoleAndSubordinates)
		require.NoError(t, err)
		assert.Nil(t, g)
	})

	// --- Test: GetByRelatedUserID ---
	t.Run("GetByRelatedUserID returns personal group", func(t *testing.T) {
		g, err := repo.GetByRelatedUserID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, g)
		assert.Equal(t, personalGroup.ID, g.ID)
		assert.Equal(t, security.GroupTypePersonal, g.GroupType)
		require.NotNil(t, g.RelatedUserID)
		assert.Equal(t, user.ID, *g.RelatedUserID)
	})

	// --- Test: GetByID returns nil for non-existent ---
	t.Run("GetByID returns nil for non-existent ID", func(t *testing.T) {
		g, err := repo.GetByID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, g)
	})

	// --- Test: List ---
	t.Run("List returns all groups", func(t *testing.T) {
		groups, err := repo.List(ctx, 100, 0)
		require.NoError(t, err)
		assert.Len(t, groups, 3) // public + role + personal
	})

	t.Run("List respects limit and offset", func(t *testing.T) {
		groups, err := repo.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Len(t, groups, 2)

		groups, err = repo.List(ctx, 100, 2)
		require.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	// --- Test: Count ---
	t.Run("Count returns correct count", func(t *testing.T) {
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	// --- Test: Delete ---
	t.Run("Delete removes group", func(t *testing.T) {
		// Get the public group to delete it.
		pubGroup, err := repo.GetByAPIName(ctx, "AllUsers")
		require.NoError(t, err)
		require.NotNil(t, pubGroup)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, tx, pubGroup.ID)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		// Verify it's gone.
		g, err := repo.GetByID(ctx, pubGroup.ID)
		require.NoError(t, err)
		assert.Nil(t, g)

		// Count should be 2 now.
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}
