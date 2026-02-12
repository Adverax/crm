package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/adverax/crm/internal/pkg/apperror"
)

var apiNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,98}[a-zA-Z0-9]$`)

// ValidateCreateUserRole validates input for creating a user role.
func ValidateCreateUserRole(input CreateUserRoleInput) error {
	if err := validateAPIName(input.APIName); err != nil {
		return err
	}
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateUpdateUserRole validates input for updating a user role.
func ValidateUpdateUserRole(input UpdateUserRoleInput) error {
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateCreatePermissionSet validates input for creating a permission set.
func ValidateCreatePermissionSet(input CreatePermissionSetInput) error {
	if err := validateAPIName(input.APIName); err != nil {
		return err
	}
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	if input.PSType != PSTypeGrant && input.PSType != PSTypeDeny {
		return apperror.Validation(fmt.Sprintf("ps_type must be '%s' or '%s'", PSTypeGrant, PSTypeDeny))
	}
	return nil
}

// ValidateUpdatePermissionSet validates input for updating a permission set.
func ValidateUpdatePermissionSet(input UpdatePermissionSetInput) error {
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateCreateProfile validates input for creating a profile.
func ValidateCreateProfile(input CreateProfileInput) error {
	if err := validateAPIName(input.APIName); err != nil {
		return err
	}
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateUpdateProfile validates input for updating a profile.
func ValidateUpdateProfile(input UpdateProfileInput) error {
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	return nil
}

// ValidateCreateUser validates input for creating a user.
func ValidateCreateUser(input CreateUserInput) error {
	if input.Username == "" {
		return apperror.Validation("username is required")
	}
	if len(input.Username) > 100 {
		return apperror.Validation("username must be at most 100 characters")
	}
	if input.Email == "" {
		return apperror.Validation("email is required")
	}
	if !strings.Contains(input.Email, "@") {
		return apperror.Validation("email must be a valid email address")
	}
	return nil
}

// ValidateUpdateUser validates input for updating a user.
func ValidateUpdateUser(input UpdateUserInput) error {
	if input.Email == "" {
		return apperror.Validation("email is required")
	}
	if !strings.Contains(input.Email, "@") {
		return apperror.Validation("email must be a valid email address")
	}
	return nil
}

// ValidateSetObjectPermission validates OLS permission input.
func ValidateSetObjectPermission(input SetObjectPermissionInput) error {
	if input.Permissions < 0 || input.Permissions > OLSAll {
		return apperror.Validation(fmt.Sprintf("permissions must be between 0 and %d", OLSAll))
	}
	return nil
}

// ValidateSetFieldPermission validates FLS permission input.
func ValidateSetFieldPermission(input SetFieldPermissionInput) error {
	if input.Permissions < 0 || input.Permissions > FLSAll {
		return apperror.Validation(fmt.Sprintf("permissions must be between 0 and %d", FLSAll))
	}
	return nil
}

// ValidateCreateGroup validates input for creating a group.
func ValidateCreateGroup(input CreateGroupInput) error {
	if err := validateAPIName(input.APIName); err != nil {
		return err
	}
	if input.Label == "" {
		return apperror.Validation("label is required")
	}
	switch input.GroupType {
	case GroupTypePersonal, GroupTypeRole, GroupTypeRoleAndSubordinates, GroupTypePublic, GroupTypeTerritory:
	default:
		return apperror.Validation(fmt.Sprintf("group_type must be one of: personal, role, role_and_subordinates, public, territory; got '%s'", input.GroupType))
	}
	return nil
}

// ValidateAddGroupMember validates input for adding a group member.
func ValidateAddGroupMember(input AddGroupMemberInput) error {
	if input.MemberUserID == nil && input.MemberGroupID == nil {
		return apperror.Validation("either member_user_id or member_group_id is required")
	}
	if input.MemberUserID != nil && input.MemberGroupID != nil {
		return apperror.Validation("only one of member_user_id or member_group_id can be set")
	}
	return nil
}

// ValidateCreateSharingRule validates input for creating a sharing rule.
func ValidateCreateSharingRule(input CreateSharingRuleInput) error {
	switch input.RuleType {
	case RuleTypeOwnerBased, RuleTypeCriteriaBased:
	default:
		return apperror.Validation(fmt.Sprintf("rule_type must be 'owner_based' or 'criteria_based'; got '%s'", input.RuleType))
	}
	if input.AccessLevel != "read" && input.AccessLevel != "read_write" {
		return apperror.Validation("access_level must be 'read' or 'read_write'")
	}
	if input.RuleType == RuleTypeCriteriaBased {
		if input.CriteriaField == nil || input.CriteriaOp == nil || input.CriteriaValue == nil {
			return apperror.Validation("criteria_field, criteria_op, and criteria_value are required for criteria_based rules")
		}
	}
	return nil
}

// ValidateUpdateSharingRule validates input for updating a sharing rule.
func ValidateUpdateSharingRule(input UpdateSharingRuleInput) error {
	if input.AccessLevel != "read" && input.AccessLevel != "read_write" {
		return apperror.Validation("access_level must be 'read' or 'read_write'")
	}
	return nil
}

func validateAPIName(apiName string) error {
	if apiName == "" {
		return apperror.Validation("api_name is required")
	}
	if len(apiName) > 100 {
		return apperror.Validation("api_name must be at most 100 characters")
	}
	if !apiNamePattern.MatchString(apiName) {
		return apperror.Validation("api_name must start with a letter, contain only letters, digits, and underscores, and end with a letter or digit")
	}
	return nil
}
