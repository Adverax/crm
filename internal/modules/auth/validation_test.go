package auth

import (
	"strings"
	"testing"
)

func TestValidateLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   LoginInput
		wantErr bool
		errMsg  string
	}{
		{name: "valid input", input: LoginInput{Username: "admin", Password: "secret"}, wantErr: false},
		{name: "empty username", input: LoginInput{Username: "", Password: "secret"}, wantErr: true, errMsg: "username is required"},
		{name: "whitespace username", input: LoginInput{Username: "  ", Password: "secret"}, wantErr: true, errMsg: "username is required"},
		{name: "empty password", input: LoginInput{Username: "admin", Password: ""}, wantErr: true, errMsg: "password is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateLogin(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateLogin() error = %v, want containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   RefreshInput
		wantErr bool
	}{
		{name: "valid token", input: RefreshInput{RefreshToken: "abc123"}, wantErr: false},
		{name: "empty token", input: RefreshInput{RefreshToken: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRefresh(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRefresh() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{name: "valid 8 chars", password: "12345678", wantErr: false},
		{name: "valid long", password: strings.Repeat("a", 128), wantErr: false},
		{name: "too short", password: "1234567", wantErr: true, errMsg: "at least 8"},
		{name: "too long", password: strings.Repeat("a", 129), wantErr: true, errMsg: "at most 128"},
		{name: "empty", password: "", wantErr: true, errMsg: "at least 8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidatePassword() error = %v, want containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateForgotPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   ForgotPasswordInput
		wantErr bool
		errMsg  string
	}{
		{name: "valid email", input: ForgotPasswordInput{Email: "user@example.com"}, wantErr: false},
		{name: "empty email", input: ForgotPasswordInput{Email: ""}, wantErr: true, errMsg: "email is required"},
		{name: "invalid email", input: ForgotPasswordInput{Email: "notanemail"}, wantErr: true, errMsg: "valid email"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateForgotPassword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateForgotPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateForgotPassword() error = %v, want containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateResetPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   ResetPasswordInput
		wantErr bool
		errMsg  string
	}{
		{name: "valid input", input: ResetPasswordInput{Token: "abc", Password: "newpass123"}, wantErr: false},
		{name: "empty token", input: ResetPasswordInput{Token: "", Password: "newpass123"}, wantErr: true, errMsg: "token is required"},
		{name: "short password", input: ResetPasswordInput{Token: "abc", Password: "short"}, wantErr: true, errMsg: "at least 8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateResetPassword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResetPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateResetPassword() error = %v, want containing %q", err, tt.errMsg)
			}
		})
	}
}
