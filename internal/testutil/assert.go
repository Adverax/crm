package testutil

import (
	"errors"
	"strings"
	"testing"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// AssertAppError checks that err contains an *apperror.AppError with the expected code.
func AssertAppError(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected AppError with code %s, got nil", code)
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperror.AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Errorf("expected error code %s, got %s (message: %s)", code, appErr.Code, appErr.Message)
	}
}

// AssertAppErrorContains checks code and that the message contains substr.
func AssertAppErrorContains(t *testing.T, err error, code apperror.Code, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected AppError with code %s, got nil", code)
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperror.AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Errorf("expected error code %s, got %s", code, appErr.Code)
	}
	if !strings.Contains(appErr.Message, substr) {
		t.Errorf("expected message to contain %q, got %q", substr, appErr.Message)
	}
}

// RequireNoError calls t.Fatal if err is not nil.
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
