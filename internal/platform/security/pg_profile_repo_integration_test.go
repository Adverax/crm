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

func TestPgProfileRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "iam.profiles", "iam.permission_sets")

	ctx := context.Background()
	psRepo := security.NewPgPermissionSetRepository(pool)
	repo := security.NewPgProfileRepository(pool)

	// Create prerequisite permission set.
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	basePS, err := psRepo.Create(ctx, tx, security.CreatePermissionSetInput{
		APIName:     "profile_base_ps",
		Label:       "Profile Base PS",
		Description: "Base PS for profile tests",
		PSType:      security.PSTypeGrant,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	t.Run("Create and GetByID", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		profile, err := repo.Create(ctx, tx, &security.Profile{
			APIName:             "standard_user",
			Label:               "Standard User",
			Description:         "Default profile for regular users",
			BasePermissionSetID: basePS.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, profile.ID)
		assert.Equal(t, "standard_user", profile.APIName)
		assert.Equal(t, "Standard User", profile.Label)
		assert.Equal(t, "Default profile for regular users", profile.Description)
		assert.Equal(t, basePS.ID, profile.BasePermissionSetID)
		assert.False(t, profile.CreatedAt.IsZero())
		assert.False(t, profile.UpdatedAt.IsZero())

		fetched, err := repo.GetByID(ctx, profile.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, profile.ID, fetched.ID)
		assert.Equal(t, profile.APIName, fetched.APIName)
		assert.Equal(t, profile.Label, fetched.Label)
		assert.Equal(t, profile.Description, fetched.Description)
		assert.Equal(t, basePS.ID, fetched.BasePermissionSetID)
	})

	t.Run("GetByAPIName", func(t *testing.T) {
		fetched, err := repo.GetByAPIName(ctx, "standard_user")
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, "standard_user", fetched.APIName)
		assert.Equal(t, "Standard User", fetched.Label)
	})

	t.Run("GetByID returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("GetByAPIName returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByAPIName(ctx, "nonexistent_profile")
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("List with pagination", func(t *testing.T) {
		// Create a second profile.
		tx2, err := pool.Begin(ctx)
		require.NoError(t, err)

		secondPS, err := psRepo.Create(ctx, tx2, security.CreatePermissionSetInput{
			APIName:     "admin_base_ps",
			Label:       "Admin Base PS",
			Description: "Base PS for admin profile",
			PSType:      security.PSTypeGrant,
		})
		require.NoError(t, err)

		_, err = repo.Create(ctx, tx2, &security.Profile{
			APIName:             "system_admin",
			Label:               "System Administrator",
			Description:         "Full access profile",
			BasePermissionSetID: secondPS.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx2.Commit(ctx))

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
		existing, err := repo.GetByAPIName(ctx, "standard_user")
		require.NoError(t, err)
		require.NotNil(t, existing)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		updated, err := repo.Update(ctx, tx, existing.ID, security.UpdateProfileInput{
			Label:       "Updated Standard User",
			Description: "Updated description for standard users",
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, existing.ID, updated.ID)
		assert.Equal(t, "standard_user", updated.APIName)
		assert.Equal(t, "Updated Standard User", updated.Label)
		assert.Equal(t, "Updated description for standard users", updated.Description)
		assert.Equal(t, basePS.ID, updated.BasePermissionSetID)
		assert.True(t, updated.UpdatedAt.After(existing.UpdatedAt) || updated.UpdatedAt.Equal(existing.UpdatedAt))
	})

	t.Run("Delete", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		delPS, err := psRepo.Create(ctx, tx, security.CreatePermissionSetInput{
			APIName:     "del_profile_ps",
			Label:       "Deletable PS",
			Description: "PS for deletable profile",
			PSType:      security.PSTypeGrant,
		})
		require.NoError(t, err)

		profile, err := repo.Create(ctx, tx, &security.Profile{
			APIName:             "to_delete_profile",
			Label:               "Delete Me",
			Description:         "Will be deleted",
			BasePermissionSetID: delPS.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx2, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, tx2, profile.ID)
		require.NoError(t, err)
		require.NoError(t, tx2.Commit(ctx))

		fetched, err := repo.GetByID(ctx, profile.ID)
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})
}
