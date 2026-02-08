package security

import (
	"context"

	"github.com/google/uuid"
)

// UserRoleService defines business logic for user roles.
type UserRoleService interface {
	Create(ctx context.Context, input CreateUserRoleInput) (*UserRole, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserRole, error)
	List(ctx context.Context, page, perPage int32) ([]UserRole, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserRoleInput) (*UserRole, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// PermissionSetService defines business logic for permission sets.
type PermissionSetService interface {
	Create(ctx context.Context, input CreatePermissionSetInput) (*PermissionSet, error)
	GetByID(ctx context.Context, id uuid.UUID) (*PermissionSet, error)
	List(ctx context.Context, page, perPage int32) ([]PermissionSet, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdatePermissionSetInput) (*PermissionSet, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProfileService defines business logic for profiles.
type ProfileService interface {
	Create(ctx context.Context, input CreateProfileInput) (*Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	List(ctx context.Context, page, perPage int32) ([]Profile, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateProfileInput) (*Profile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserService defines business logic for users.
type UserService interface {
	Create(ctx context.Context, input CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, page, perPage int32) ([]User, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AssignPermissionSet(ctx context.Context, userID, psID uuid.UUID) error
	RevokePermissionSet(ctx context.Context, userID, psID uuid.UUID) error
	ListPermissionSets(ctx context.Context, userID uuid.UUID) ([]PermissionSetToUser, error)
}

// PermissionService defines business logic for OLS/FLS permission management.
type PermissionService interface {
	SetObjectPermission(ctx context.Context, psID uuid.UUID, input SetObjectPermissionInput) (*ObjectPermission, error)
	ListObjectPermissions(ctx context.Context, psID uuid.UUID) ([]ObjectPermission, error)
	RemoveObjectPermission(ctx context.Context, psID, objectID uuid.UUID) error
	SetFieldPermission(ctx context.Context, psID uuid.UUID, input SetFieldPermissionInput) (*FieldPermission, error)
	ListFieldPermissions(ctx context.Context, psID uuid.UUID) ([]FieldPermission, error)
	RemoveFieldPermission(ctx context.Context, psID, fieldID uuid.UUID) error
}

// EffectiveComputer recomputes effective permission caches.
type EffectiveComputer interface {
	RecomputeForUser(ctx context.Context, userID uuid.UUID) error
	RecomputeForPermissionSet(ctx context.Context, psID uuid.UUID) error
	RecomputeAll(ctx context.Context) error
}

// MetadataFieldLister provides field definitions from the metadata cache.
// This interface breaks the circular dependency between security and metadata packages.
type MetadataFieldLister interface {
	ListFieldsByObjectID(ctx context.Context, objectID uuid.UUID) ([]FieldInfo, error)
	ListAllObjectIDs(ctx context.Context) ([]uuid.UUID, error)
}

// FieldInfo is a minimal view of a field for security computation.
type FieldInfo struct {
	ID      uuid.UUID
	APIName string
}
