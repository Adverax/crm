//go:build integration

package metadata

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/testutil"
)

func TestPgObjectRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "metadata.object_definitions")

	repo := NewPgObjectRepository(pool)
	ctx := context.Background()

	standardInput := CreateObjectInput{
		APIName:               "TestAccount",
		Label:                 "Test Account",
		PluralLabel:           "Test Accounts",
		Description:           "Account for integration tests",
		ObjectType:            ObjectTypeStandard,
		IsVisibleInSetup:      true,
		IsCustomFieldsAllowed: true,
		IsDeleteableObject:    false,
		IsCreateable:          true,
		IsUpdateable:          true,
		IsDeleteable:          true,
		IsQueryable:           true,
		IsSearchable:          false,
		HasActivities:         true,
		HasNotes:              false,
		HasHistoryTracking:    true,
		HasSharingRules:       false,
		Visibility:            VisibilityPrivate,
	}

	t.Run("Create standard object and verify all fields", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		obj, err := repo.Create(ctx, tx, standardInput)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, obj.ID)
		assert.Equal(t, "TestAccount", obj.APIName)
		assert.Equal(t, "Test Account", obj.Label)
		assert.Equal(t, "Test Accounts", obj.PluralLabel)
		assert.Equal(t, "Account for integration tests", obj.Description)
		assert.Equal(t, "obj_testaccount", obj.TableName)
		assert.Equal(t, ObjectTypeStandard, obj.ObjectType)
		assert.False(t, obj.IsPlatformManaged)
		assert.True(t, obj.IsVisibleInSetup)
		assert.True(t, obj.IsCustomFieldsAllowed)
		assert.False(t, obj.IsDeleteableObject)
		assert.True(t, obj.IsCreateable)
		assert.True(t, obj.IsUpdateable)
		assert.True(t, obj.IsDeleteable)
		assert.True(t, obj.IsQueryable)
		assert.False(t, obj.IsSearchable)
		assert.True(t, obj.HasActivities)
		assert.False(t, obj.HasNotes)
		assert.True(t, obj.HasHistoryTracking)
		assert.False(t, obj.HasSharingRules)
		assert.Equal(t, VisibilityPrivate, obj.Visibility)
		assert.False(t, obj.CreatedAt.IsZero())
		assert.False(t, obj.UpdatedAt.IsZero())
	})

	t.Run("GetByID finds created object", func(t *testing.T) {
		// Create a fresh object to get its ID.
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := repo.Create(ctx, tx, CreateObjectInput{
			APIName:    "GetByIDObj",
			Label:      "GetByID Object",
			ObjectType: ObjectTypeStandard,
			Visibility: VisibilityPrivate,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		found, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "GetByIDObj", found.APIName)
	})

	t.Run("GetByAPIName finds created object", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := repo.Create(ctx, tx, CreateObjectInput{
			APIName:    "FindByName",
			Label:      "Find By Name",
			ObjectType: ObjectTypeStandard,
			Visibility: VisibilityPrivate,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		found, err := repo.GetByAPIName(ctx, "FindByName")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "FindByName", found.APIName)
	})

	t.Run("GetByID returns nil for non-existent UUID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("List returns objects with pagination", func(t *testing.T) {
		// Truncate and create 3 objects for deterministic results.
		testutil.TruncateTables(t, pool, "metadata.object_definitions")

		for _, name := range []string{"ListObj1", "ListObj2", "ListObj3"} {
			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			_, err = repo.Create(ctx, tx, CreateObjectInput{
				APIName:    name,
				Label:      name,
				ObjectType: ObjectTypeStandard,
				Visibility: VisibilityPrivate,
			})
			require.NoError(t, err)
			require.NoError(t, tx.Commit(ctx))
		}

		// Page 1: limit 2, offset 0.
		page1, err := repo.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		// Page 2: limit 2, offset 2.
		page2, err := repo.List(ctx, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 1)
	})

	t.Run("ListAll returns all objects", func(t *testing.T) {
		all, err := repo.ListAll(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 1)
	})

	t.Run("Update changes label description and flags", func(t *testing.T) {
		testutil.TruncateTables(t, pool, "metadata.object_definitions")

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := repo.Create(ctx, tx, CreateObjectInput{
			APIName:      "UpdObj",
			Label:        "Original Label",
			Description:  "Original Desc",
			ObjectType:   ObjectTypeStandard,
			IsSearchable: false,
			Visibility:   VisibilityPrivate,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx2, err := pool.Begin(ctx)
		require.NoError(t, err)
		updated, err := repo.Update(ctx, tx2, created.ID, UpdateObjectInput{
			Label:        "Updated Label",
			Description:  "Updated Desc",
			IsSearchable: true,
			Visibility:   VisibilityPublicRead,
		})
		require.NoError(t, err)
		require.NoError(t, tx2.Commit(ctx))

		assert.Equal(t, "Updated Label", updated.Label)
		assert.Equal(t, "Updated Desc", updated.Description)
		assert.True(t, updated.IsSearchable)
		assert.Equal(t, VisibilityPublicRead, updated.Visibility)
		assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))
	})

	t.Run("Delete removes the object", func(t *testing.T) {
		testutil.TruncateTables(t, pool, "metadata.object_definitions")

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := repo.Create(ctx, tx, CreateObjectInput{
			APIName:    "DelObj",
			Label:      "Delete Me",
			ObjectType: ObjectTypeStandard,
			Visibility: VisibilityPrivate,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		txDel, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, txDel, created.ID)
		require.NoError(t, err)
		require.NoError(t, txDel.Commit(ctx))

		found, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Count is correct after create and delete", func(t *testing.T) {
		testutil.TruncateTables(t, pool, "metadata.object_definitions")

		count0, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count0)

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := repo.Create(ctx, tx, CreateObjectInput{
			APIName:    "CountObj",
			Label:      "Count Object",
			ObjectType: ObjectTypeStandard,
			Visibility: VisibilityPrivate,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		count1, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count1)

		txDel, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = repo.Delete(ctx, txDel, created.ID)
		require.NoError(t, err)
		require.NoError(t, txDel.Commit(ctx))

		count2, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count2)
	})

	t.Run("Create custom object strips __c suffix in table name", func(t *testing.T) {
		testutil.TruncateTables(t, pool, "metadata.object_definitions")

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		obj, err := repo.Create(ctx, tx, CreateObjectInput{
			APIName:    "MyCustom__c",
			Label:      "My Custom Object",
			ObjectType: ObjectTypeCustom,
			Visibility: VisibilityPrivate,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, "obj_mycustom", obj.TableName)
		assert.Equal(t, ObjectTypeCustom, obj.ObjectType)
	})
}
