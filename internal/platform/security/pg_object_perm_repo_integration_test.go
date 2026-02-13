//go:build integration

package security_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metadata "github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/testutil"
)

func TestPgObjectPermRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	ctx := context.Background()

	testutil.TruncateTables(t, pool,
		"security.object_permissions",
		"iam.permission_sets",
		"metadata.field_definitions",
		"metadata.object_definitions",
	)

	repo := security.NewPgObjectPermissionRepository(pool)

	// --- Prerequisites: create a permission set and a metadata object ---

	psRepo := security.NewPgPermissionSetRepository(pool)

	// Create first permission set.
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	ps1, err := psRepo.Create(ctx, tx, security.CreatePermissionSetInput{
		APIName:     "TestObjPermPS1",
		Label:       "Test Object Perm PS 1",
		Description: "PS1 for object permission integration tests",
		PSType:      security.PSTypeGrant,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create second permission set (for ListByPermissionSetIDs test).
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	ps2, err := psRepo.Create(ctx, tx, security.CreatePermissionSetInput{
		APIName:     "TestObjPermPS2",
		Label:       "Test Object Perm PS 2",
		Description: "PS2 for object permission integration tests",
		PSType:      security.PSTypeGrant,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	// Create metadata object via metadata repo.
	metaRepo := metadata.NewPgObjectRepository(pool)
	tx, err = pool.Begin(ctx)
	require.NoError(t, err)
	metaObj, err := metaRepo.Create(ctx, tx, metadata.CreateObjectInput{
		APIName:     "TestAccount",
		Label:       "Test Account",
		PluralLabel: "Test Accounts",
		ObjectType:  metadata.ObjectTypeStandard,
		IsQueryable: true,
		Visibility:  metadata.VisibilityPrivate,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	objectID := metaObj.ID

	// --- Test: Upsert (insert) and GetByPSAndObject ---
	t.Run("upsert inserts new permission and get returns it", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		op, err := repo.Upsert(ctx, tx, ps1.ID, objectID, 0b1111) // CRUD: all 4 bits
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, op.ID)
		assert.Equal(t, ps1.ID, op.PermissionSetID)
		assert.Equal(t, objectID, op.ObjectID)
		assert.Equal(t, 0b1111, op.Permissions)
		assert.False(t, op.CreatedAt.IsZero())
		assert.False(t, op.UpdatedAt.IsZero())

		// Read it back.
		fetched, err := repo.GetByPSAndObject(ctx, ps1.ID, objectID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, op.ID, fetched.ID)
		assert.Equal(t, 0b1111, fetched.Permissions)
	})

	// --- Test: Upsert (update) same key with different permissions ---
	t.Run("upsert updates existing permission", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		op, err := repo.Upsert(ctx, tx, ps1.ID, objectID, 0b0101) // only read + delete
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, 0b0101, op.Permissions)

		// Verify via read.
		fetched, err := repo.GetByPSAndObject(ctx, ps1.ID, objectID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, 0b0101, fetched.Permissions)
	})

	// --- Test: GetByPSAndObject returns nil for non-existent ---
	t.Run("GetByPSAndObject returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByPSAndObject(ctx, uuid.New(), objectID)
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("GetByPSAndObject returns nil for non-existent object", func(t *testing.T) {
		fetched, err := repo.GetByPSAndObject(ctx, ps1.ID, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	// --- Test: ListByPermissionSetID ---
	t.Run("ListByPermissionSetID returns permissions for PS", func(t *testing.T) {
		perms, err := repo.ListByPermissionSetID(ctx, ps1.ID)
		require.NoError(t, err)
		require.Len(t, perms, 1)
		assert.Equal(t, ps1.ID, perms[0].PermissionSetID)
		assert.Equal(t, objectID, perms[0].ObjectID)
	})

	t.Run("ListByPermissionSetID returns empty for unknown PS", func(t *testing.T) {
		perms, err := repo.ListByPermissionSetID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, perms)
	})

	// --- Test: ListByPermissionSetIDs with multiple PS ---
	t.Run("ListByPermissionSetIDs returns permissions for multiple PSes", func(t *testing.T) {
		// Add a permission for ps2.
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		_, err = repo.Upsert(ctx, tx, ps2.ID, objectID, 0b0011) // read + create
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		perms, err := repo.ListByPermissionSetIDs(ctx, []uuid.UUID{ps1.ID, ps2.ID})
		require.NoError(t, err)
		assert.Len(t, perms, 2)

		// Build a map for easier verification.
		permMap := make(map[uuid.UUID]int)
		for _, p := range perms {
			permMap[p.PermissionSetID] = p.Permissions
		}
		assert.Equal(t, 0b0101, permMap[ps1.ID])
		assert.Equal(t, 0b0011, permMap[ps2.ID])
	})

	t.Run("ListByPermissionSetIDs returns empty for empty input", func(t *testing.T) {
		perms, err := repo.ListByPermissionSetIDs(ctx, []uuid.UUID{})
		require.NoError(t, err)
		assert.Empty(t, perms)
	})

	// --- Test: Delete ---
	t.Run("Delete removes object permission", func(t *testing.T) {
		// Delete ps1's permission on the object.
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, tx, ps1.ID, objectID)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		// Verify it's gone.
		fetched, err := repo.GetByPSAndObject(ctx, ps1.ID, objectID)
		require.NoError(t, err)
		assert.Nil(t, fetched)

		// ps2's permission should still be there.
		fetched, err = repo.GetByPSAndObject(ctx, ps2.ID, objectID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, 0b0011, fetched.Permissions)
	})

	t.Run("Delete is idempotent for non-existent permission", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, tx, ps1.ID, objectID)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
	})
}
