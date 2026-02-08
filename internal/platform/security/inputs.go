package security

import "github.com/google/uuid"

// CreateUserRoleInput contains input data for creating a user role.
type CreateUserRoleInput struct {
	APIName     string     `json:"api_name"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
}

// UpdateUserRoleInput contains input data for updating a user role.
type UpdateUserRoleInput struct {
	Label       string     `json:"label"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
}

// CreatePermissionSetInput contains input data for creating a permission set.
type CreatePermissionSetInput struct {
	APIName     string `json:"api_name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	PSType      PSType `json:"ps_type"`
}

// UpdatePermissionSetInput contains input data for updating a permission set.
type UpdatePermissionSetInput struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// CreateProfileInput contains input data for creating a profile.
type CreateProfileInput struct {
	APIName     string `json:"api_name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// UpdateProfileInput contains input data for updating a profile.
type UpdateProfileInput struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// CreateUserInput contains input data for creating a user.
type CreateUserInput struct {
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	ProfileID uuid.UUID  `json:"profile_id"`
	RoleID    *uuid.UUID `json:"role_id"`
}

// UpdateUserInput contains input data for updating a user.
type UpdateUserInput struct {
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	ProfileID uuid.UUID  `json:"profile_id"`
	RoleID    *uuid.UUID `json:"role_id"`
	IsActive  bool       `json:"is_active"`
}

// SetObjectPermissionInput contains input for setting OLS permissions.
type SetObjectPermissionInput struct {
	ObjectID    uuid.UUID `json:"object_id"`
	Permissions int       `json:"permissions"`
}

// SetFieldPermissionInput contains input for setting FLS permissions.
type SetFieldPermissionInput struct {
	FieldID     uuid.UUID `json:"field_id"`
	Permissions int       `json:"permissions"`
}
