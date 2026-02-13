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

func TestPgUserRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "iam.users", "iam.profiles", "iam.permission_sets", "iam.user_roles")

	ctx := context.Background()
	psRepo := security.NewPgPermissionSetRepository(pool)
	profileRepo := security.NewPgProfileRepository(pool)
	roleRepo := security.NewPgUserRoleRepository(pool)
	repo := security.NewPgUserRepository(pool)

	// Create prerequisite permission set, profile, and role.
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	basePS, err := psRepo.Create(ctx, tx, security.CreatePermissionSetInput{
		APIName:     "user_test_base_ps",
		Label:       "User Test Base PS",
		Description: "Base PS for user integration tests",
		PSType:      security.PSTypeGrant,
	})
	require.NoError(t, err)

	profile, err := profileRepo.Create(ctx, tx, &security.Profile{
		APIName:             "user_test_profile",
		Label:               "User Test Profile",
		Description:         "Profile for user integration tests",
		BasePermissionSetID: basePS.ID,
	})
	require.NoError(t, err)

	role, err := roleRepo.Create(ctx, tx, security.CreateUserRoleInput{
		APIName:     "user_test_role",
		Label:       "User Test Role",
		Description: "Role for user integration tests",
	})
	require.NoError(t, err)

	require.NoError(t, tx.Commit(ctx))

	t.Run("Create and GetByID", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		input := security.CreateUserInput{
			Username:  "jdoe",
			Email:     "jdoe@example.com",
			FirstName: "John",
			LastName:  "Doe",
			ProfileID: profile.ID,
			RoleID:    &role.ID,
		}
		user, err := repo.Create(ctx, tx, input)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "jdoe", user.Username)
		assert.Equal(t, "jdoe@example.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, profile.ID, user.ProfileID)
		require.NotNil(t, user.RoleID)
		assert.Equal(t, role.ID, *user.RoleID)
		assert.True(t, user.IsActive)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())

		fetched, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, user.ID, fetched.ID)
		assert.Equal(t, user.Username, fetched.Username)
		assert.Equal(t, user.Email, fetched.Email)
		assert.Equal(t, user.FirstName, fetched.FirstName)
		assert.Equal(t, user.LastName, fetched.LastName)
		assert.Equal(t, profile.ID, fetched.ProfileID)
		require.NotNil(t, fetched.RoleID)
		assert.Equal(t, role.ID, *fetched.RoleID)
		assert.True(t, fetched.IsActive)
	})

	t.Run("Create without RoleID", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		user, err := repo.Create(ctx, tx, security.CreateUserInput{
			Username:  "norole",
			Email:     "norole@example.com",
			FirstName: "No",
			LastName:  "Role",
			ProfileID: profile.ID,
			RoleID:    nil,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Nil(t, user.RoleID)

		fetched, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Nil(t, fetched.RoleID)
	})

	t.Run("GetByUsername", func(t *testing.T) {
		fetched, err := repo.GetByUsername(ctx, "jdoe")
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, "jdoe", fetched.Username)
		assert.Equal(t, "jdoe@example.com", fetched.Email)
	})

	t.Run("GetByID returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("GetByUsername returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByUsername(ctx, "nonexistent_user")
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("List with pagination", func(t *testing.T) {
		all, err := repo.List(ctx, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 2)

		page, err := repo.List(ctx, 1, 0)
		require.NoError(t, err)
		assert.Len(t, page, 1)

		page2, err := repo.List(ctx, 1, 1)
		require.NoError(t, err)
		assert.Len(t, page2, 1)
		assert.NotEqual(t, page[0].ID, page2[0].ID)
	})

	t.Run("Update", func(t *testing.T) {
		existing, err := repo.GetByUsername(ctx, "jdoe")
		require.NoError(t, err)
		require.NotNil(t, existing)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		updated, err := repo.Update(ctx, tx, existing.ID, security.UpdateUserInput{
			Email:     "john.doe@example.com",
			FirstName: "Jonathan",
			LastName:  "Doe",
			ProfileID: profile.ID,
			RoleID:    &role.ID,
			IsActive:  true,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, existing.ID, updated.ID)
		assert.Equal(t, "jdoe", updated.Username)
		assert.Equal(t, "john.doe@example.com", updated.Email)
		assert.Equal(t, "Jonathan", updated.FirstName)
		assert.True(t, updated.IsActive)
		assert.True(t, updated.UpdatedAt.After(existing.UpdatedAt) || updated.UpdatedAt.Equal(existing.UpdatedAt))
	})

	t.Run("Update IsActive to false", func(t *testing.T) {
		existing, err := repo.GetByUsername(ctx, "jdoe")
		require.NoError(t, err)
		require.NotNil(t, existing)
		assert.True(t, existing.IsActive)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		updated, err := repo.Update(ctx, tx, existing.ID, security.UpdateUserInput{
			Email:     existing.Email,
			FirstName: existing.FirstName,
			LastName:  existing.LastName,
			ProfileID: existing.ProfileID,
			RoleID:    existing.RoleID,
			IsActive:  false,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.False(t, updated.IsActive)

		fetched, err := repo.GetByID(ctx, existing.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.False(t, fetched.IsActive)
	})

	t.Run("Delete", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		user, err := repo.Create(ctx, tx, security.CreateUserInput{
			Username:  "to_delete_user",
			Email:     "delete@example.com",
			FirstName: "Delete",
			LastName:  "Me",
			ProfileID: profile.ID,
			RoleID:    nil,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx2, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, tx2, user.ID)
		require.NoError(t, err)
		require.NoError(t, tx2.Commit(ctx))

		fetched, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})
}
