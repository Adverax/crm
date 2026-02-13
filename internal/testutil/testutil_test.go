package testutil

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

func TestAssertAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     apperror.Code
		wantFail bool
	}{
		{
			name: "direct AppError matches",
			err:  apperror.NotFound("contact", "123"),
			code: apperror.CodeNotFound,
		},
		{
			name: "wrapped AppError matches",
			err:  fmt.Errorf("service.GetByID: %w", apperror.NotFound("contact", "123")),
			code: apperror.CodeNotFound,
		},
		{
			name: "double-wrapped AppError matches",
			err:  fmt.Errorf("handler: %w", fmt.Errorf("service: %w", apperror.Forbidden("no access"))),
			code: apperror.CodeForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertAppError(t, tt.err, tt.code)
		})
	}
}

func TestAssertAppErrorContains(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		code   apperror.Code
		substr string
	}{
		{
			name:   "message contains substring",
			err:    apperror.NotFound("contact", "abc-123"),
			code:   apperror.CodeNotFound,
			substr: "abc-123",
		},
		{
			name:   "wrapped error message contains substring",
			err:    fmt.Errorf("svc: %w", apperror.BadRequest("invalid email format")),
			code:   apperror.CodeBadRequest,
			substr: "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertAppErrorContains(t, tt.err, tt.code, tt.substr)
		})
	}
}

func TestRequireNoError(t *testing.T) {
	RequireNoError(t, nil)
}

func TestIssueTestJWT(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		profileID uuid.UUID
		roleID    *uuid.UUID
		wantRID   bool
	}{
		{
			name:      "with role ID",
			userID:    TestUserID,
			profileID: TestProfileID,
			roleID:    &TestRoleID,
			wantRID:   true,
		},
		{
			name:      "without role ID",
			userID:    AdminUserID,
			profileID: AdminProfileID,
			roleID:    nil,
			wantRID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStr := IssueTestJWT(tt.userID, tt.profileID, tt.roleID)
			if tokenStr == "" {
				t.Fatal("expected non-empty token")
			}

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(TestJWTSecret), nil
			})
			if err != nil {
				t.Fatalf("failed to parse token: %v", err)
			}
			if !token.Valid {
				t.Fatal("token is not valid")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatal("expected MapClaims")
			}

			sub, _ := claims.GetSubject()
			if sub != tt.userID.String() {
				t.Errorf("sub = %s, want %s", sub, tt.userID)
			}

			pid, _ := claims["pid"].(string)
			if pid != tt.profileID.String() {
				t.Errorf("pid = %s, want %s", pid, tt.profileID)
			}

			rid, hasRID := claims["rid"]
			if tt.wantRID {
				if !hasRID {
					t.Error("expected rid claim, not found")
				} else if rid != tt.roleID.String() {
					t.Errorf("rid = %s, want %s", rid, tt.roleID)
				}
			} else {
				if hasRID {
					t.Errorf("unexpected rid claim: %v", rid)
				}
			}
		})
	}
}

func TestIssueExpiredJWT(t *testing.T) {
	tokenStr := IssueExpiredJWT(TestUserID, TestProfileID)
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	_, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(TestJWTSecret), nil
	})
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestContextWithTestUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		profileID uuid.UUID
		roleID    *uuid.UUID
	}{
		{
			name:      "with role",
			userID:    TestUserID,
			profileID: TestProfileID,
			roleID:    &TestRoleID,
		},
		{
			name:      "without role",
			userID:    AdminUserID,
			profileID: AdminProfileID,
			roleID:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ContextWithTestUser(tt.userID, tt.profileID, tt.roleID)

			uc, ok := security.UserFromContext(ctx)
			if !ok {
				t.Fatal("UserContext not found in context")
			}
			if uc.UserID != tt.userID {
				t.Errorf("UserID = %s, want %s", uc.UserID, tt.userID)
			}
			if uc.ProfileID != tt.profileID {
				t.Errorf("ProfileID = %s, want %s", uc.ProfileID, tt.profileID)
			}
			if tt.roleID == nil && uc.RoleID != nil {
				t.Errorf("RoleID = %v, want nil", uc.RoleID)
			}
			if tt.roleID != nil && (uc.RoleID == nil || *uc.RoleID != *tt.roleID) {
				t.Errorf("RoleID = %v, want %s", uc.RoleID, tt.roleID)
			}
		})
	}
}

func TestGinContextWithTestUser(t *testing.T) {
	c, w := GinContextWithTestUser(t, http.MethodGet, "/api/test", nil,
		TestUserID, TestProfileID, &TestRoleID)

	if c.Request == nil {
		t.Fatal("expected non-nil Request")
	}
	if c.Request.Method != http.MethodGet {
		t.Errorf("Method = %s, want GET", c.Request.Method)
	}
	if w == nil {
		t.Fatal("expected non-nil ResponseRecorder")
	}

	val, exists := c.Get("user_context")
	if !exists {
		t.Fatal("user_context not set in gin context")
	}
	uc, ok := val.(security.UserContext)
	if !ok {
		t.Fatal("user_context is not security.UserContext")
	}
	if uc.UserID != TestUserID {
		t.Errorf("UserID = %s, want %s", uc.UserID, TestUserID)
	}

	ucFromStd, ok := security.UserFromContext(c.Request.Context())
	if !ok {
		t.Fatal("UserContext not found in request context")
	}
	if ucFromStd.UserID != TestUserID {
		t.Errorf("std context UserID = %s, want %s", ucFromStd.UserID, TestUserID)
	}
}

func TestGinContextAnonymous(t *testing.T) {
	c, w := GinContextAnonymous(t, http.MethodPost, "/auth/login", nil)

	if c.Request == nil {
		t.Fatal("expected non-nil Request")
	}
	if c.Request.Method != http.MethodPost {
		t.Errorf("Method = %s, want POST", c.Request.Method)
	}
	if w == nil {
		t.Fatal("expected non-nil ResponseRecorder")
	}

	_, exists := c.Get("user_context")
	if exists {
		t.Error("expected no user_context in anonymous gin context")
	}
}
