package security

import (
	"time"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/identity"
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
	ID                  uuid.UUID `json:"id"`
	APIName             string    `json:"api_name"`
	Label               string    `json:"label"`
	Description         string    `json:"description"`
	BasePermissionSetID uuid.UUID `json:"base_permission_set_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
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

// UserContext is an alias for identity.UserContext (ADR-0030 shared kernel).
// Existing consumers continue to use security.UserContext without changes.
type UserContext = identity.UserContext

// GroupType represents the type of a group (ADR-0013).
type GroupType string

const (
	GroupTypePersonal            GroupType = "personal"
	GroupTypeRole                GroupType = "role"
	GroupTypeRoleAndSubordinates GroupType = "role_and_subordinates"
	GroupTypePublic              GroupType = "public"
	GroupTypeTerritory           GroupType = "territory"
)

// Group represents a security group.
type Group struct {
	ID            uuid.UUID  `json:"id"`
	APIName       string     `json:"api_name"`
	Label         string     `json:"label"`
	GroupType     GroupType  `json:"group_type"`
	RelatedRoleID *uuid.UUID `json:"related_role_id"`
	RelatedUserID *uuid.UUID `json:"related_user_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// GroupMember represents membership in a group (user XOR nested group).
type GroupMember struct {
	ID            uuid.UUID  `json:"id"`
	GroupID       uuid.UUID  `json:"group_id"`
	MemberUserID  *uuid.UUID `json:"member_user_id"`
	MemberGroupID *uuid.UUID `json:"member_group_id"`
	CreatedAt     time.Time  `json:"created_at"`
}

// RecordShare represents a sharing entry in an object's share table.
type RecordShare struct {
	ID          uuid.UUID `json:"id"`
	RecordID    uuid.UUID `json:"record_id"`
	GroupID     uuid.UUID `json:"group_id"`
	AccessLevel string    `json:"access_level"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

// SharingRule represents a declarative sharing rule.
type SharingRule struct {
	ID            uuid.UUID `json:"id"`
	ObjectID      uuid.UUID `json:"object_id"`
	RuleType      RuleType  `json:"rule_type"`
	SourceGroupID uuid.UUID `json:"source_group_id"`
	TargetGroupID uuid.UUID `json:"target_group_id"`
	AccessLevel   string    `json:"access_level"`
	CriteriaField *string   `json:"criteria_field"`
	CriteriaOp    *string   `json:"criteria_op"`
	CriteriaValue *string   `json:"criteria_value"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RuleType represents the type of a sharing rule.
type RuleType string

const (
	RuleTypeOwnerBased    RuleType = "owner_based"
	RuleTypeCriteriaBased RuleType = "criteria_based"
)

// Effective cache types for RLS (ADR-0012).

// EffectiveRoleHierarchy represents a transitive closure entry in the role tree.
type EffectiveRoleHierarchy struct {
	AncestorRoleID   uuid.UUID `json:"ancestor_role_id"`
	DescendantRoleID uuid.UUID `json:"descendant_role_id"`
	Depth            int       `json:"depth"`
}

// EffectiveVisibleOwner represents which record owners a user can see via role hierarchy.
type EffectiveVisibleOwner struct {
	UserID         uuid.UUID `json:"user_id"`
	VisibleOwnerID uuid.UUID `json:"visible_owner_id"`
}

// EffectiveGroupMember represents flattened group membership (including nested groups).
type EffectiveGroupMember struct {
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
}

// EffectiveObjectHierarchy represents the parent-child closure for controlled_by_parent objects.
type EffectiveObjectHierarchy struct {
	AncestorObjectID   uuid.UUID `json:"ancestor_object_id"`
	DescendantObjectID uuid.UUID `json:"descendant_object_id"`
	Depth              int       `json:"depth"`
}

// Well-known UUIDs for seed data.
var (
	SystemAdminBasePermissionSetID = uuid.MustParse("00000000-0000-4000-a000-000000000001")
	SystemAdminProfileID           = uuid.MustParse("00000000-0000-4000-a000-000000000002")
	SystemAdminUserID              = uuid.MustParse("00000000-0000-4000-a000-000000000003")
)
