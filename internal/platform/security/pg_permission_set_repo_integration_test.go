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

func TestPgPermissionSetRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "iam.permission_sets")

	repo := security.NewPgPermissionSetRepository(pool)
	ctx := context.Background()

	t.Run("Create grant PS and GetByID", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		input := security.CreatePermissionSetInput{
			APIName:     "sales_grant_ps",
			Label:       "Sales Grant PS",
			Description: "Grants sales permissions",
			PSType:      security.PSTypeGrant,
		}
		ps, err := repo.Create(ctx, tx, input)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, ps.ID)
		assert.Equal(t, "sales_grant_ps", ps.APIName)
		assert.Equal(t, "Sales Grant PS", ps.Label)
		assert.Equal(t, "Grants sales permissions", ps.Description)
		assert.Equal(t, security.PSTypeGrant, ps.PSType)
		assert.False(t, ps.CreatedAt.IsZero())
		assert.False(t, ps.UpdatedAt.IsZero())

		fetched, err := repo.GetByID(ctx, ps.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, ps.ID, fetched.ID)
		assert.Equal(t, security.PSTypeGrant, fetched.PSType)
	})

	t.Run("Create deny PS", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		input := security.CreatePermissionSetInput{
			APIName:     "restrict_deny_ps",
			Label:       "Restrict Deny PS",
			Description: "Denies restricted permissions",
			PSType:      security.PSTypeDeny,
		}
		ps, err := repo.Create(ctx, tx, input)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, security.PSTypeDeny, ps.PSType)

		fetched, err := repo.GetByID(ctx, ps.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, security.PSTypeDeny, fetched.PSType)
	})

	t.Run("GetByAPIName", func(t *testing.T) {
		fetched, err := repo.GetByAPIName(ctx, "sales_grant_ps")
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, "sales_grant_ps", fetched.APIName)
		assert.Equal(t, security.PSTypeGrant, fetched.PSType)
	})

	t.Run("GetByID returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("GetByAPIName returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByAPIName(ctx, "nonexistent_ps")
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
		existing, err := repo.GetByAPIName(ctx, "sales_grant_ps")
		require.NoError(t, err)
		require.NotNil(t, existing)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		updated, err := repo.Update(ctx, tx, existing.ID, security.UpdatePermissionSetInput{
			Label:       "Updated Sales Grant PS",
			Description: "Updated description",
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, existing.ID, updated.ID)
		assert.Equal(t, "sales_grant_ps", updated.APIName)
		assert.Equal(t, "Updated Sales Grant PS", updated.Label)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, security.PSTypeGrant, updated.PSType)
		assert.True(t, updated.UpdatedAt.After(existing.UpdatedAt) || updated.UpdatedAt.Equal(existing.UpdatedAt))
	})

	t.Run("Delete", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		ps, err := repo.Create(ctx, tx, security.CreatePermissionSetInput{
			APIName:     "to_delete_ps",
			Label:       "Delete Me",
			Description: "Will be deleted",
			PSType:      security.PSTypeGrant,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx2, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, tx2, ps.ID)
		require.NoError(t, err)
		require.NoError(t, tx2.Commit(ctx))

		fetched, err := repo.GetByID(ctx, ps.ID)
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})
}
