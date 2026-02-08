package security

import (
	"time"

	"github.com/google/uuid"
)

// PSType represents permission set type (grant or deny).
type PSType string

const (
	PSTypeGrant PSType = "grant"
	PSTypeDeny  PSType = "deny"
)

// UserRole represents a role in the role hierarchy.
type UserRole struct {
	ID          uuid.UUID  `json:"id"`
	APIName     string     `json:"api_name"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PermissionSet represents a set of permissions (grant or deny).
type PermissionSet struct {
	ID          uuid.UUID `json:"id"`
	APIName     string    `json:"api_name"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	PSType      PSType    `json:"ps_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Profile represents a user profile with a base permission set.
type Profile struct {
	ID                   uuid.UUID `json:"id"`
	APIName              string    `json:"api_name"`
	Label                string    `json:"label"`
	Description          string    `json:"description"`
	BasePermissionSetID  uuid.UUID `json:"base_permission_set_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// User represents a CRM user.
type User struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	ProfileID uuid.UUID  `json:"profile_id"`
	RoleID    *uuid.UUID `json:"role_id"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// PermissionSetToUser represents assignment of a PS to a user.
type PermissionSetToUser struct {
	ID              uuid.UUID `json:"id"`
	PermissionSetID uuid.UUID `json:"permission_set_id"`
	UserID          uuid.UUID `json:"user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// ObjectPermission represents OLS permissions for a PS on an object.
type ObjectPermission struct {
	ID              uuid.UUID `json:"id"`
	PermissionSetID uuid.UUID `json:"permission_set_id"`
	ObjectID        uuid.UUID `json:"object_id"`
	Permissions     int       `json:"permissions"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// FieldPermission represents FLS permissions for a PS on a field.
type FieldPermission struct {
	ID              uuid.UUID `json:"id"`
	PermissionSetID uuid.UUID `json:"permission_set_id"`
	FieldID         uuid.UUID `json:"field_id"`
	Permissions     int       `json:"permissions"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// EffectiveOLS represents cached effective object-level permissions for a user.
type EffectiveOLS struct {
	UserID      uuid.UUID `json:"user_id"`
	ObjectID    uuid.UUID `json:"object_id"`
	Permissions int       `json:"permissions"`
	ComputedAt  time.Time `json:"computed_at"`
}

// EffectiveFLS represents cached effective field-level permissions for a user.
type EffectiveFLS struct {
	UserID      uuid.UUID `json:"user_id"`
	FieldID     uuid.UUID `json:"field_id"`
	Permissions int       `json:"permissions"`
	ComputedAt  time.Time `json:"computed_at"`
}

// EffectiveFieldList represents cached list of accessible field names.
type EffectiveFieldList struct {
	UserID     uuid.UUID `json:"user_id"`
	ObjectID   uuid.UUID `json:"object_id"`
	Mask       int       `json:"mask"`
	FieldNames []string  `json:"field_names"`
	ComputedAt time.Time `json:"computed_at"`
}

// OutboxEvent represents a security change event in the outbox.
type OutboxEvent struct {
	ID          int64      `json:"id"`
	EventType   string     `json:"event_type"`
	EntityType  string     `json:"entity_type"`
	EntityID    uuid.UUID  `json:"entity_id"`
	Payload     []byte     `json:"payload"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

// UserContext holds the security identity of the current request user.
type UserContext struct {
	UserID    uuid.UUID
	ProfileID uuid.UUID
	RoleID    *uuid.UUID
}

// Well-known UUIDs for seed data.
var (
	SystemAdminBasePermissionSetID = uuid.MustParse("00000000-0000-4000-a000-000000000001")
	SystemAdminProfileID           = uuid.MustParse("00000000-0000-4000-a000-000000000002")
	SystemAdminUserID              = uuid.MustParse("00000000-0000-4000-a000-000000000003")
)
