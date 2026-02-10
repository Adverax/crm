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

// CreateGroupInput contains input data for creating a group.
type CreateGroupInput struct {
	APIName       string     `json:"api_name"`
	Label         string     `json:"label"`
	GroupType     GroupType  `json:"group_type"`
	RelatedRoleID *uuid.UUID `json:"related_role_id"`
	RelatedUserID *uuid.UUID `json:"related_user_id"`
}

// AddGroupMemberInput contains input for adding a member to a group.
type AddGroupMemberInput struct {
	GroupID       uuid.UUID  `json:"group_id"`
	MemberUserID  *uuid.UUID `json:"member_user_id"`
	MemberGroupID *uuid.UUID `json:"member_group_id"`
}

// CreateSharingRuleInput contains input data for creating a sharing rule.
type CreateSharingRuleInput struct {
	ObjectID      uuid.UUID `json:"object_id"`
	RuleType      RuleType  `json:"rule_type"`
	SourceGroupID uuid.UUID `json:"source_group_id"`
	TargetGroupID uuid.UUID `json:"target_group_id"`
	AccessLevel   string    `json:"access_level"`
	CriteriaField *string   `json:"criteria_field"`
	CriteriaOp    *string   `json:"criteria_op"`
	CriteriaValue *string   `json:"criteria_value"`
}

// UpdateSharingRuleInput contains input data for updating a sharing rule.
type UpdateSharingRuleInput struct {
	TargetGroupID uuid.UUID `json:"target_group_id"`
	AccessLevel   string    `json:"access_level"`
	CriteriaField *string   `json:"criteria_field"`
	CriteriaOp    *string   `json:"criteria_op"`
	CriteriaValue *string   `json:"criteria_value"`
}

// ShareRecordInput contains input for manually sharing a record.
type ShareRecordInput struct {
	RecordID    uuid.UUID `json:"record_id"`
	GroupID     uuid.UUID `json:"group_id"`
	AccessLevel string    `json:"access_level"`
}

// RevokeShareInput contains input for revoking a manual share.
type RevokeShareInput struct {
	RecordID uuid.UUID `json:"record_id"`
	GroupID  uuid.UUID `json:"group_id"`
}
