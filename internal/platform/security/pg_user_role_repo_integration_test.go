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

func TestPgUserRoleRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "iam.user_roles")

	repo := security.NewPgUserRoleRepository(pool)
	ctx := context.Background()

	t.Run("Create and GetByID", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		input := security.CreateUserRoleInput{
			APIName:     "sales_rep",
			Label:       "Sales Representative",
			Description: "Handles direct sales",
			ParentID:    nil,
		}
		role, err := repo.Create(ctx, tx, input)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, role.ID)
		assert.Equal(t, "sales_rep", role.APIName)
		assert.Equal(t, "Sales Representative", role.Label)
		assert.Equal(t, "Handles direct sales", role.Description)
		assert.Nil(t, role.ParentID)
		assert.False(t, role.CreatedAt.IsZero())
		assert.False(t, role.UpdatedAt.IsZero())

		fetched, err := repo.GetByID(ctx, role.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, role.ID, fetched.ID)
		assert.Equal(t, role.APIName, fetched.APIName)
		assert.Equal(t, role.Label, fetched.Label)
		assert.Equal(t, role.Description, fetched.Description)
		assert.Nil(t, fetched.ParentID)
	})

	t.Run("GetByAPIName", func(t *testing.T) {
		fetched, err := repo.GetByAPIName(ctx, "sales_rep")
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, "sales_rep", fetched.APIName)
		assert.Equal(t, "Sales Representative", fetched.Label)
	})

	t.Run("GetByID returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("GetByAPIName returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByAPIName(ctx, "nonexistent_role")
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("List with pagination", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		_, err = repo.Create(ctx, tx, security.CreateUserRoleInput{
			APIName:     "sales_manager",
			Label:       "Sales Manager",
			Description: "Manages sales team",
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

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
		existing, err := repo.GetByAPIName(ctx, "sales_rep")
		require.NoError(t, err)
		require.NotNil(t, existing)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		updated, err := repo.Update(ctx, tx, existing.ID, security.UpdateUserRoleInput{
			Label:       "Senior Sales Representative",
			Description: "Senior sales position",
			ParentID:    nil,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, existing.ID, updated.ID)
		assert.Equal(t, "sales_rep", updated.APIName)
		assert.Equal(t, "Senior Sales Representative", updated.Label)
		assert.Equal(t, "Senior sales position", updated.Description)
		assert.True(t, updated.UpdatedAt.After(existing.UpdatedAt) || updated.UpdatedAt.Equal(existing.UpdatedAt))
	})

	t.Run("Delete", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		role, err := repo.Create(ctx, tx, security.CreateUserRoleInput{
			APIName:     "to_delete",
			Label:       "Delete Me",
			Description: "Will be deleted",
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx2, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, tx2, role.ID)
		require.NoError(t, err)
		require.NoError(t, tx2.Commit(ctx))

		fetched, err := repo.GetByID(ctx, role.ID)
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})

	t.Run("Create with ParentID", func(t *testing.T) {
		parent, err := repo.GetByAPIName(ctx, "sales_manager")
		require.NoError(t, err)
		require.NotNil(t, parent)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		child, err := repo.Create(ctx, tx, security.CreateUserRoleInput{
			APIName:     "sales_associate",
			Label:       "Sales Associate",
			Description: "Junior sales role",
			ParentID:    &parent.ID,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotNil(t, child.ParentID)
		assert.Equal(t, parent.ID, *child.ParentID)

		fetched, err := repo.GetByID(ctx, child.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		require.NotNil(t, fetched.ParentID)
		assert.Equal(t, parent.ID, *fetched.ParentID)
	})
}
