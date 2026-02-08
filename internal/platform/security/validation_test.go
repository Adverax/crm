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
