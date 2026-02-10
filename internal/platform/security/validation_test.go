package security

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateCreateUserRole(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   CreateUserRoleInput
		wantErr bool
	}{
		{
			name:    "valid input",
			input:   CreateUserRoleInput{APIName: "sales_manager", Label: "Sales Manager"},
			wantErr: false,
		},
		{
			name:    "empty api_name",
			input:   CreateUserRoleInput{APIName: "", Label: "X"},
			wantErr: true,
		},
		{
			name:    "empty label",
			input:   CreateUserRoleInput{APIName: "test", Label: ""},
			wantErr: true,
		},
		{
			name:    "invalid api_name characters",
			input:   CreateUserRoleInput{APIName: "test-role", Label: "Test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCreateUserRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateUserRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreatePermissionSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   CreatePermissionSetInput
		wantErr bool
	}{
		{
			name:    "valid grant",
			input:   CreatePermissionSetInput{APIName: "sales_ps", Label: "Sales", PSType: PSTypeGrant},
			wantErr: false,
		},
		{
			name:    "valid deny",
			input:   CreatePermissionSetInput{APIName: "deny_export", Label: "Deny Export", PSType: PSTypeDeny},
			wantErr: false,
		},
		{
			name:    "invalid ps_type",
			input:   CreatePermissionSetInput{APIName: "test", Label: "Test", PSType: "invalid"},
			wantErr: true,
		},
		{
			name:    "empty label",
			input:   CreatePermissionSetInput{APIName: "test", Label: "", PSType: PSTypeGrant},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCreatePermissionSet(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreatePermissionSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreateUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   CreateUserInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: CreateUserInput{
				Username:  "jdoe",
				Email:     "jdoe@test.com",
				ProfileID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "empty username",
			input: CreateUserInput{
				Username:  "",
				Email:     "jdoe@test.com",
				ProfileID: uuid.New(),
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			input: CreateUserInput{
				Username:  "jdoe",
				Email:     "invalid",
				ProfileID: uuid.New(),
			},
			wantErr: true,
		},
		{
			name: "empty email",
			input: CreateUserInput{
				Username:  "jdoe",
				Email:     "",
				ProfileID: uuid.New(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCreateUser(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSetObjectPermission(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   SetObjectPermissionInput
		wantErr bool
	}{
		{
			name:    "valid read only",
			input:   SetObjectPermissionInput{ObjectID: uuid.New(), Permissions: OLSRead},
			wantErr: false,
		},
		{
			name:    "valid all",
			input:   SetObjectPermissionInput{ObjectID: uuid.New(), Permissions: OLSAll},
			wantErr: false,
		},
		{
			name:    "zero permissions is valid",
			input:   SetObjectPermissionInput{ObjectID: uuid.New(), Permissions: 0},
			wantErr: false,
		},
		{
			name:    "negative permissions",
			input:   SetObjectPermissionInput{ObjectID: uuid.New(), Permissions: -1},
			wantErr: true,
		},
		{
			name:    "permissions exceed max",
			input:   SetObjectPermissionInput{ObjectID: uuid.New(), Permissions: 16},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateSetObjectPermission(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSetObjectPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSetFieldPermission(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   SetFieldPermissionInput
		wantErr bool
	}{
		{
			name:    "valid read",
			input:   SetFieldPermissionInput{FieldID: uuid.New(), Permissions: FLSRead},
			wantErr: false,
		},
		{
			name:    "valid all",
			input:   SetFieldPermissionInput{FieldID: uuid.New(), Permissions: FLSAll},
			wantErr: false,
		},
		{
			name:    "negative",
			input:   SetFieldPermissionInput{FieldID: uuid.New(), Permissions: -1},
			wantErr: true,
		},
		{
			name:    "exceed max",
			input:   SetFieldPermissionInput{FieldID: uuid.New(), Permissions: 4},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateSetFieldPermission(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSetFieldPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUpdateUserRole(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   UpdateUserRoleInput
		wantErr bool
	}{
		{name: "valid input", input: UpdateUserRoleInput{Label: "Updated"}, wantErr: false},
		{name: "empty label", input: UpdateUserRoleInput{Label: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateUpdateUserRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateUpdatePermissionSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   UpdatePermissionSetInput
		wantErr bool
	}{
		{name: "valid input", input: UpdatePermissionSetInput{Label: "Updated"}, wantErr: false},
		{name: "empty label", input: UpdatePermissionSetInput{Label: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateUpdatePermissionSet(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateCreateProfile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   CreateProfileInput
		wantErr bool
	}{
		{name: "valid input", input: CreateProfileInput{APIName: "sales_profile", Label: "Sales"}, wantErr: false},
		{name: "empty api_name", input: CreateProfileInput{APIName: "", Label: "Bad"}, wantErr: true},
		{name: "empty label", input: CreateProfileInput{APIName: "good_name", Label: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCreateProfile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateUpdateProfile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   UpdateProfileInput
		wantErr bool
	}{
		{name: "valid input", input: UpdateProfileInput{Label: "Updated"}, wantErr: false},
		{name: "empty label", input: UpdateProfileInput{Label: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateUpdateProfile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateUpdateUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   UpdateUserInput
		wantErr bool
	}{
		{name: "valid input", input: UpdateUserInput{Email: "j@test.com"}, wantErr: false},
		{name: "empty email", input: UpdateUserInput{Email: ""}, wantErr: true},
		{name: "invalid email", input: UpdateUserInput{Email: "no-at"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateUpdateUser(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateCreateGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   CreateGroupInput
		wantErr bool
	}{
		{name: "valid public group", input: CreateGroupInput{APIName: "test_group", Label: "Test", GroupType: GroupTypePublic}, wantErr: false},
		{name: "valid personal group", input: CreateGroupInput{APIName: "personal_john", Label: "John", GroupType: GroupTypePersonal}, wantErr: false},
		{name: "valid role group", input: CreateGroupInput{APIName: "role_sales", Label: "Sales", GroupType: GroupTypeRole}, wantErr: false},
		{name: "invalid group_type", input: CreateGroupInput{APIName: "bad", Label: "Bad", GroupType: "invalid"}, wantErr: true},
		{name: "empty api_name", input: CreateGroupInput{APIName: "", Label: "Bad", GroupType: GroupTypePublic}, wantErr: true},
		{name: "empty label", input: CreateGroupInput{APIName: "good_name", Label: "", GroupType: GroupTypePublic}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCreateGroup(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateAddGroupMember(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	gid := uuid.New()
	tests := []struct {
		name    string
		input   AddGroupMemberInput
		wantErr bool
	}{
		{name: "valid user member", input: AddGroupMemberInput{GroupID: uuid.New(), MemberUserID: &uid}, wantErr: false},
		{name: "valid group member", input: AddGroupMemberInput{GroupID: uuid.New(), MemberGroupID: &gid}, wantErr: false},
		{name: "both set", input: AddGroupMemberInput{GroupID: uuid.New(), MemberUserID: &uid, MemberGroupID: &gid}, wantErr: true},
		{name: "neither set", input: AddGroupMemberInput{GroupID: uuid.New()}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAddGroupMember(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateCreateSharingRule(t *testing.T) {
	t.Parallel()
	cf := "status"
	cop := "eq"
	cv := "active"
	tests := []struct {
		name    string
		input   CreateSharingRuleInput
		wantErr bool
	}{
		{
			name: "valid owner_based",
			input: CreateSharingRuleInput{
				ObjectID: uuid.New(), RuleType: RuleTypeOwnerBased,
				SourceGroupID: uuid.New(), TargetGroupID: uuid.New(), AccessLevel: "read",
			},
			wantErr: false,
		},
		{
			name: "valid criteria_based",
			input: CreateSharingRuleInput{
				ObjectID: uuid.New(), RuleType: RuleTypeCriteriaBased,
				SourceGroupID: uuid.New(), TargetGroupID: uuid.New(), AccessLevel: "read_write",
				CriteriaField: &cf, CriteriaOp: &cop, CriteriaValue: &cv,
			},
			wantErr: false,
		},
		{
			name: "invalid rule_type",
			input: CreateSharingRuleInput{
				ObjectID: uuid.New(), RuleType: "invalid",
				SourceGroupID: uuid.New(), TargetGroupID: uuid.New(), AccessLevel: "read",
			},
			wantErr: true,
		},
		{
			name: "invalid access_level",
			input: CreateSharingRuleInput{
				ObjectID: uuid.New(), RuleType: RuleTypeOwnerBased,
				SourceGroupID: uuid.New(), TargetGroupID: uuid.New(), AccessLevel: "write",
			},
			wantErr: true,
		},
		{
			name: "criteria_based without criteria",
			input: CreateSharingRuleInput{
				ObjectID: uuid.New(), RuleType: RuleTypeCriteriaBased,
				SourceGroupID: uuid.New(), TargetGroupID: uuid.New(), AccessLevel: "read",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCreateSharingRule(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateUpdateSharingRule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   UpdateSharingRuleInput
		wantErr bool
	}{
		{name: "valid read", input: UpdateSharingRuleInput{TargetGroupID: uuid.New(), AccessLevel: "read"}, wantErr: false},
		{name: "valid read_write", input: UpdateSharingRuleInput{TargetGroupID: uuid.New(), AccessLevel: "read_write"}, wantErr: false},
		{name: "invalid access_level", input: UpdateSharingRuleInput{TargetGroupID: uuid.New(), AccessLevel: "write"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateUpdateSharingRule(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}
