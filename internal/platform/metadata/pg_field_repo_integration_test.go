//go:build integration

package metadata

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/testutil/pgtest"
)

func ptr[T any](v T) *T { return &v }

func TestPgFieldRepo_Integration(t *testing.T) {
	pool := pgtest.SetupTestPool(t)
	pgtest.TruncateTables(t, pool, "metadata.field_definitions", "metadata.object_definitions")

	ctx := context.Background()
	objRepo := NewPgObjectRepository(pool)
	fieldRepo := NewPgFieldRepository(pool)

	// Create a parent object as prerequisite for all field tests.
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	parentObj, err := objRepo.Create(ctx, tx, CreateObjectInput{
		APIName:    "field_test_obj",
		Label:      "Field Test Object",
		ObjectType: ObjectTypeStandard,
		Visibility: VisibilityPrivate,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	subtypePlain := SubtypePlain

	t.Run("Create text field with subtype plain and MaxLength config", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		field, err := fieldRepo.Create(ctx, tx, CreateFieldInput{
			ObjectID:     parentObj.ID,
			APIName:      "first_name",
			Label:        "First Name",
			Description:  "Contact first name",
			HelpText:     "Enter the first name",
			FieldType:    FieldTypeText,
			FieldSubtype: &subtypePlain,
			IsRequired:   true,
			IsUnique:     false,
			Config: FieldConfig{
				MaxLength: ptr(255),
			},
			IsCustom:  false,
			SortOrder: 1,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.NotEqual(t, uuid.Nil, field.ID)
		assert.Equal(t, parentObj.ID, field.ObjectID)
		assert.Equal(t, "first_name", field.APIName)
		assert.Equal(t, "First Name", field.Label)
		assert.Equal(t, "Contact first name", field.Description)
		assert.Equal(t, "Enter the first name", field.HelpText)
		assert.Equal(t, FieldTypeText, field.FieldType)
		require.NotNil(t, field.FieldSubtype)
		assert.Equal(t, SubtypePlain, *field.FieldSubtype)
		assert.Nil(t, field.ReferencedObjectID)
		assert.True(t, field.IsRequired)
		assert.False(t, field.IsUnique)
		require.NotNil(t, field.Config.MaxLength)
		assert.Equal(t, 255, *field.Config.MaxLength)
		assert.False(t, field.IsSystemField)
		assert.False(t, field.IsCustom)
		assert.False(t, field.IsPlatformManaged)
		assert.Equal(t, 1, field.SortOrder)
		assert.False(t, field.CreatedAt.IsZero())
		assert.False(t, field.UpdatedAt.IsZero())
	})

	t.Run("GetByID returns correct field with all fields scanned", func(t *testing.T) {
		// Create a field to look up.
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := fieldRepo.Create(ctx, tx, CreateFieldInput{
			ObjectID:     parentObj.ID,
			APIName:      "get_by_id_field",
			Label:        "GetByID Field",
			FieldType:    FieldTypeText,
			FieldSubtype: &subtypePlain,
			Config: FieldConfig{
				MaxLength: ptr(100),
			},
			SortOrder: 10,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		found, err := fieldRepo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "get_by_id_field", found.APIName)
		assert.Equal(t, "GetByID Field", found.Label)
		assert.Equal(t, FieldTypeText, found.FieldType)
		require.NotNil(t, found.FieldSubtype)
		assert.Equal(t, SubtypePlain, *found.FieldSubtype)
		require.NotNil(t, found.Config.MaxLength)
		assert.Equal(t, 100, *found.Config.MaxLength)
		assert.Equal(t, 10, found.SortOrder)
	})

	t.Run("GetByObjectAndName finds the field", func(t *testing.T) {
		found, err := fieldRepo.GetByObjectAndName(ctx, parentObj.ID, "first_name")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "first_name", found.APIName)
		assert.Equal(t, parentObj.ID, found.ObjectID)
	})

	t.Run("GetByObjectAndName returns nil for wrong name", func(t *testing.T) {
		found, err := fieldRepo.GetByObjectAndName(ctx, parentObj.ID, "nonexistent_field")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Create number field with integer subtype and Precision config", func(t *testing.T) {
		subtypeInt := SubtypeInteger

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		field, err := fieldRepo.Create(ctx, tx, CreateFieldInput{
			ObjectID:     parentObj.ID,
			APIName:      "employee_count",
			Label:        "Employee Count",
			Description:  "Number of employees",
			FieldType:    FieldTypeNumber,
			FieldSubtype: &subtypeInt,
			Config: FieldConfig{
				Precision: ptr(10),
			},
			SortOrder: 2,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, FieldTypeNumber, field.FieldType)
		require.NotNil(t, field.FieldSubtype)
		assert.Equal(t, SubtypeInteger, *field.FieldSubtype)
		require.NotNil(t, field.Config.Precision)
		assert.Equal(t, 10, *field.Config.Precision)
	})

	t.Run("ListByObjectID returns fields ordered by sort_order", func(t *testing.T) {
		fields, err := fieldRepo.ListByObjectID(ctx, parentObj.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(fields), 2)

		// Verify sort_order is non-decreasing.
		for i := 1; i < len(fields); i++ {
			assert.GreaterOrEqual(t, fields[i].SortOrder, fields[i-1].SortOrder,
				"fields should be ordered by sort_order")
		}
	})

	t.Run("ListAll returns all fields", func(t *testing.T) {
		all, err := fieldRepo.ListAll(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 2)
	})

	t.Run("Update changes label and description", func(t *testing.T) {
		// Create a field to update.
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := fieldRepo.Create(ctx, tx, CreateFieldInput{
			ObjectID:     parentObj.ID,
			APIName:      "update_me",
			Label:        "Original Label",
			Description:  "Original Description",
			FieldType:    FieldTypeText,
			FieldSubtype: &subtypePlain,
			Config: FieldConfig{
				MaxLength: ptr(50),
			},
			SortOrder: 5,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		txUpd, err := pool.Begin(ctx)
		require.NoError(t, err)
		updated, err := fieldRepo.Update(ctx, txUpd, created.ID, UpdateFieldInput{
			Label:       "Updated Label",
			Description: "Updated Description",
			HelpText:    "New help text",
			IsRequired:  true,
			IsUnique:    true,
			Config: FieldConfig{
				MaxLength: ptr(100),
			},
			SortOrder: 99,
		})
		require.NoError(t, err)
		require.NoError(t, txUpd.Commit(ctx))

		assert.Equal(t, "Updated Label", updated.Label)
		assert.Equal(t, "Updated Description", updated.Description)
		assert.Equal(t, "New help text", updated.HelpText)
		assert.True(t, updated.IsRequired)
		assert.True(t, updated.IsUnique)
		require.NotNil(t, updated.Config.MaxLength)
		assert.Equal(t, 100, *updated.Config.MaxLength)
		assert.Equal(t, 99, updated.SortOrder)
		assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))
	})

	t.Run("Delete removes field", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		created, err := fieldRepo.Create(ctx, tx, CreateFieldInput{
			ObjectID:     parentObj.ID,
			APIName:      "delete_me",
			Label:        "Delete Me",
			FieldType:    FieldTypeText,
			FieldSubtype: &subtypePlain,
			SortOrder:    50,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		txDel, err := pool.Begin(ctx)
		require.NoError(t, err)
		err = fieldRepo.Delete(ctx, txDel, created.ID)
		require.NoError(t, err)
		require.NoError(t, txDel.Commit(ctx))

		found, err := fieldRepo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Create boolean field with no subtype", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		field, err := fieldRepo.Create(ctx, tx, CreateFieldInput{
			ObjectID:     parentObj.ID,
			APIName:      "is_active",
			Label:        "Is Active",
			Description:  "Whether the record is active",
			FieldType:    FieldTypeBoolean,
			FieldSubtype: nil,
			SortOrder:    3,
		})
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		assert.Equal(t, FieldTypeBoolean, field.FieldType)
		assert.Nil(t, field.FieldSubtype)

		// Verify via GetByID that nil subtype round-trips correctly.
		found, err := fieldRepo.GetByID(ctx, field.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Nil(t, found.FieldSubtype)
	})
}
