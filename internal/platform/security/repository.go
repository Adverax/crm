package security

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// UserRoleRepository defines data access for user roles.
type UserRoleRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreateUserRoleInput) (*UserRole, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserRole, error)
	GetByAPIName(ctx context.Context, apiName string) (*UserRole, error)
	List(ctx context.Context, limit, offset int32) ([]UserRole, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateUserRoleInput) (*UserRole, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// PermissionSetRepository defines data access for permission sets.
type PermissionSetRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreatePermissionSetInput) (*PermissionSet, error)
	GetByID(ctx context.Context, id uuid.UUID) (*PermissionSet, error)
	GetByAPIName(ctx context.Context, apiName string) (*PermissionSet, error)
	List(ctx context.Context, limit, offset int32) ([]PermissionSet, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdatePermissionSetInput) (*PermissionSet, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// ProfileRepository defines data access for profiles.
type ProfileRepository interface {
	Create(ctx context.Context, tx pgx.Tx, profile *Profile) (*Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	GetByAPIName(ctx context.Context, apiName string) (*Profile, error)
	List(ctx context.Context, limit, offset int32) ([]Profile, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateProfileInput) (*Profile, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// UserRepository defines data access for users.
type UserRepository interface {
	Create(ctx context.Context, tx pgx.Tx, input CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context, limit, offset int32) ([]User, error)
	Update(ctx context.Context, tx pgx.Tx, id uuid.UUID, input UpdateUserInput) (*User, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// PermissionSetToUserRepository defines data access for PS-to-user assignments.
type PermissionSetToUserRepository interface {
	Assign(ctx context.Context, tx pgx.Tx, psID, userID uuid.UUID) (*PermissionSetToUser, error)
	Revoke(ctx context.Context, tx pgx.Tx, psID, userID uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]PermissionSetToUser, error)
	ListByPermissionSetID(ctx context.Context, psID uuid.UUID) ([]PermissionSetToUser, error)
}

// ObjectPermissionRepository defines data access for object permissions.
type ObjectPermissionRepository interface {
	Upsert(ctx context.Context, tx pgx.Tx, psID, objectID uuid.UUID, permissions int) (*ObjectPermission, error)
	GetByPSAndObject(ctx context.Context, psID, objectID uuid.UUID) (*ObjectPermission, error)
	ListByPermissionSetID(ctx context.Context, psID uuid.UUID) ([]ObjectPermission, error)
	Delete(ctx context.Context, tx pgx.Tx, psID, objectID uuid.UUID) error
}

// FieldPermissionRepository defines data access for field permissions.
type FieldPermissionRepository interface {
	Upsert(ctx context.Context, tx pgx.Tx, psID, fieldID uuid.UUID, permissions int) (*FieldPermission, error)
	GetByPSAndField(ctx context.Context, psID, fieldID uuid.UUID) (*FieldPermission, error)
	ListByPermissionSetID(ctx context.Context, psID uuid.UUID) ([]FieldPermission, error)
	Delete(ctx context.Context, tx pgx.Tx, psID, fieldID uuid.UUID) error
}

// EffectivePermissionRepository defines data access for effective permission caches.
type EffectivePermissionRepository interface {
	GetOLS(ctx context.Context, userID, objectID uuid.UUID) (*EffectiveOLS, error)
	GetFLS(ctx context.Context, userID, fieldID uuid.UUID) (*EffectiveFLS, error)
	GetFieldList(ctx context.Context, userID, objectID uuid.UUID, mask int) (*EffectiveFieldList, error)
	UpsertOLS(ctx context.Context, tx pgx.Tx, userID, objectID uuid.UUID, permissions int) error
	UpsertFLS(ctx context.Context, tx pgx.Tx, userID, fieldID uuid.UUID, permissions int) error
	UpsertFieldList(ctx context.Context, tx pgx.Tx, userID, objectID uuid.UUID, mask int, fieldNames []string) error
	DeleteByUserID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
}

// OutboxRepository defines data access for the security outbox.
type OutboxRepository interface {
	Insert(ctx context.Context, tx pgx.Tx, event OutboxEvent) error
	ListUnprocessed(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkProcessed(ctx context.Context, id int64) error
}

// AllObjectPermissions returns all object permissions across all PSes for given PS IDs.
type AllObjectPermissions interface {
	ListByPermissionSetIDs(ctx context.Context, psIDs []uuid.UUID) ([]ObjectPermission, error)
}

// AllFieldPermissions returns all field permissions across all PSes for given PS IDs.
type AllFieldPermissions interface {
	ListByPermissionSetIDs(ctx context.Context, psIDs []uuid.UUID) ([]FieldPermission, error)
}
