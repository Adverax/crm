package metadata

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ObjectRepository defines the data access interface for object definitions.
type ObjectRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreateObjectInput) (*ObjectDefinition, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ObjectDefinition, error)
	GetByAPIName(ctx context.Context, apiName string) (*ObjectDefinition, error)
	List(ctx context.Context, limit, offset int32) ([]ObjectDefinition, error)
	ListAll(ctx context.Context) ([]ObjectDefinition, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateObjectInput) (*ObjectDefinition, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// FieldRepository defines the data access interface for field definitions.
type FieldRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreateFieldInput) (*FieldDefinition, error)
	GetByID(ctx context.Context, id uuid.UUID) (*FieldDefinition, error)
	GetByObjectAndName(ctx context.Context, objectID uuid.UUID, apiName string) (*FieldDefinition, error)
	ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]FieldDefinition, error)
	ListAll(ctx context.Context) ([]FieldDefinition, error)
	ListReferenceFields(ctx context.Context) ([]FieldDefinition, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateFieldInput) (*FieldDefinition, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}

// PolymorphicTargetRepository defines the data access interface for polymorphic targets.
type PolymorphicTargetRepository interface {
	Create(ctx context.Context, tx pgx.Tx, fieldID, objectID uuid.UUID) (*PolymorphicTarget, error)
	ListByFieldID(ctx context.Context, fieldID uuid.UUID) ([]PolymorphicTarget, error)
	ListAll(ctx context.Context) ([]PolymorphicTarget, error)
	DeleteByFieldID(ctx context.Context, tx pgx.Tx, fieldID uuid.UUID) error
}

// DDLExecutor executes DDL statements within a transaction.
type DDLExecutor interface {
	ExecInTx(ctx context.Context, tx pgx.Tx, statements []string) error
}
