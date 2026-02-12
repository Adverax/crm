package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetadata() *StaticMetadataProvider {
	account := NewObjectMeta("Account", "accounts").
		RequiredField("Name", "name", FieldTypeString).
		Field("Industry", "industry", FieldTypeString).
		Field("Description", "description", FieldTypeString).
		ExternalIdField("ExternalId", "external_id", FieldTypeString).
		ReadOnlyField("CreatedAt", "created_at", FieldTypeDateTime).
		Build()

	contact := NewObjectMeta("Contact", "contacts").
		RequiredField("FirstName", "first_name", FieldTypeString).
		Field("LastName", "last_name", FieldTypeString).
		Field("Email", "email", FieldTypeString).
		Field("Age", "age", FieldTypeInteger).
		Field("IsActive", "is_active", FieldTypeBoolean).
		Field("AccountId", "account_id", FieldTypeID).
		Field("Status", "status", FieldTypeString).
		Build()

	task := NewObjectMeta("Task", "tasks").
		Field("Subject", "subject", FieldTypeString).
		Field("Status", "status", FieldTypeString).
		Field("DueDate", "due_date", FieldTypeDate).
		Field("Priority", "priority", FieldTypeString).
		Build()

	return NewStaticMetadataProvider(map[string]*ObjectMeta{
		"Account": account,
		"Contact": contact,
		"Task":    task,
	})
}

func TestValidateInsert(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	t.Run("valid insert", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech')")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.Equal(t, OperationInsert, validated.Operation)
		assert.Equal(t, "Account", validated.Object.Name)
		assert.Len(t, validated.Fields, 2)
		assert.Equal(t, 1, validated.RowCount)
	})

	t.Run("multi-row insert", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech'), ('Globex', 'Finance')")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.Equal(t, 2, validated.RowCount)
	})

	t.Run("unknown object error", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Unknown (Name) VALUES ('Test')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeUnknownObject, ve.Code)
	})

	t.Run("unknown field error", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Unknown) VALUES ('Test', 'X')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeUnknownField, ve.Code)
	})

	t.Run("read-only field error", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, CreatedAt) VALUES ('Test', 2024-01-15T10:00:00Z)")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeReadOnlyField, ve.Code)
	})

	t.Run("missing required field error", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Industry) VALUES ('Tech')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeMissingRequired, ve.Code)
	})

	t.Run("duplicate field error", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Name) VALUES ('Test', 'Test')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeDuplicateField, ve.Code)
	})

	t.Run("wrong value count error", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name, Industry) VALUES ('Test')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
	})
}

func TestValidateUpdate(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	t.Run("valid update with where", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET Status = 'Active' WHERE Email = 'test@example.com'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.Equal(t, OperationUpdate, validated.Operation)
		assert.Equal(t, "Contact", validated.Object.Name)
		assert.Len(t, validated.Assignments, 1)
		assert.True(t, validated.HasWhere)
	})

	t.Run("update without where allowed by default", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET Status = 'Active'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.False(t, validated.HasWhere)
	})

	t.Run("update with multiple assignments", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET FirstName = 'John', LastName = 'Doe'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.Len(t, validated.Assignments, 2)
	})

	t.Run("unknown field in SET error", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET Unknown = 'Test'")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
	})

	t.Run("duplicate field in SET error", func(t *testing.T) {
		ast, err := Parse("UPDATE Contact SET FirstName = 'A', FirstName = 'B'")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeDuplicateField, ve.Code)
	})
}

func TestValidateDelete(t *testing.T) {
	metadata := newTestMetadata()

	t.Run("delete with where allowed", func(t *testing.T) {
		validator := NewValidator(metadata, nil, &DefaultLimits)
		ctx := context.Background()

		ast, err := Parse("DELETE FROM Task WHERE Status = 'Completed'")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.Equal(t, OperationDelete, validated.Operation)
		assert.Equal(t, "Task", validated.Object.Name)
		assert.True(t, validated.HasWhere)
	})

	t.Run("delete without where requires error with default limits", func(t *testing.T) {
		validator := NewValidator(metadata, nil, &DefaultLimits)
		ctx := context.Background()

		ast, err := Parse("DELETE FROM Task")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeDeleteRequiresWhere, ve.Code)
	})

	t.Run("delete without where allowed with NoLimits", func(t *testing.T) {
		validator := NewValidator(metadata, nil, &NoLimits)
		ctx := context.Background()

		ast, err := Parse("DELETE FROM Task")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.False(t, validated.HasWhere)
	})

	t.Run("unknown field in WHERE error", func(t *testing.T) {
		validator := NewValidator(metadata, nil, &NoLimits)
		ctx := context.Background()

		ast, err := Parse("DELETE FROM Task WHERE Unknown = 'Test'")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
	})
}

func TestValidateUpsert(t *testing.T) {
	metadata := newTestMetadata()
	validator := NewValidator(metadata, nil, nil)
	ctx := context.Background()

	t.Run("valid upsert", func(t *testing.T) {
		ast, err := Parse("UPSERT Account (ExternalId, Name, Industry) VALUES ('ext-001', 'Acme', 'Tech') ON ExternalId")
		require.NoError(t, err)

		validated, err := validator.Validate(ctx, ast)
		require.NoError(t, err)
		assert.Equal(t, OperationUpsert, validated.Operation)
		assert.Equal(t, "Account", validated.Object.Name)
		assert.Len(t, validated.Fields, 3)
		assert.Equal(t, "ExternalId", validated.ExternalIdField.Name)
	})

	t.Run("external id not in field list error", func(t *testing.T) {
		ast, err := Parse("UPSERT Account (Name, Industry) VALUES ('Acme', 'Tech') ON ExternalId")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsValidationError(err))
	})

	t.Run("non-external-id field error", func(t *testing.T) {
		ast, err := Parse("UPSERT Account (Name, Industry) VALUES ('Acme', 'Tech') ON Name")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, ErrCodeExternalIdNotFound, ve.Code)
	})
}

func TestValidateWithAccessControl(t *testing.T) {
	metadata := newTestMetadata()
	ctx := context.Background()

	t.Run("object access denied", func(t *testing.T) {
		denyAccess := &DenyAllWriteAccessController{}
		validator := NewValidator(metadata, denyAccess, nil)

		ast, err := Parse("INSERT INTO Account (Name) VALUES ('Test')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsAccessError(err))
	})

	t.Run("field access denied", func(t *testing.T) {
		fieldDenyAccess := &FuncWriteAccessController{
			ObjectFunc: func(ctx context.Context, object string, op Operation) error {
				return nil // Allow object access
			},
			FieldFunc: func(ctx context.Context, object string, fields []string) error {
				for _, f := range fields {
					if f == "Industry" {
						return NewFieldWriteAccessError(object, f)
					}
				}
				return nil
			},
		}
		validator := NewValidator(metadata, fieldDenyAccess, nil)

		ast, err := Parse("INSERT INTO Account (Name, Industry) VALUES ('Test', 'Tech')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsAccessError(err))
	})
}

func TestValidateBatchLimits(t *testing.T) {
	metadata := newTestMetadata()
	limits := &Limits{
		MaxBatchSize:         2,
		RequireWhereOnDelete: false,
	}
	validator := NewValidator(metadata, nil, limits)
	ctx := context.Background()

	t.Run("batch within limit", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES ('A'), ('B')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.NoError(t, err)
	})

	t.Run("batch exceeds limit", func(t *testing.T) {
		ast, err := Parse("INSERT INTO Account (Name) VALUES ('A'), ('B'), ('C')")
		require.NoError(t, err)

		_, err = validator.Validate(ctx, ast)
		require.Error(t, err)
		assert.True(t, IsLimitError(err))
	})
}
