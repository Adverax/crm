package testutil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/platform/security"
)

// ContextWithTestUser creates a context.Context with security.UserContext set.
func ContextWithTestUser(userID, profileID uuid.UUID, roleID *uuid.UUID) context.Context {
	uc := security.UserContext{
		UserID:    userID,
		ProfileID: profileID,
		RoleID:    roleID,
	}
	return security.ContextWithUser(context.Background(), uc)
}

// GinContextWithTestUser creates a *gin.Context with a pre-set UserContext and HTTP request.
// Returns the gin context and a response recorder for assertions.
func GinContextWithTestUser(
	t *testing.T,
	method, path string,
	body io.Reader,
	userID, profileID uuid.UUID,
	roleID *uuid.UUID,
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")

	uc := security.UserContext{
		UserID:    userID,
		ProfileID: profileID,
		RoleID:    roleID,
	}
	req = req.WithContext(security.ContextWithUser(req.Context(), uc))

	c.Request = req
	c.Set("user_context", uc)

	return c, w
}

// GinContextAnonymous creates a *gin.Context without any user context (for auth endpoints).
func GinContextAnonymous(t *testing.T, method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatalf("testutil.GinContextAnonymous: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, w
}
