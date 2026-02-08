package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAppError_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{
			name: "formats code and message",
			err:  NotFound("Contact", "abc-123"),
			want: "NOT_FOUND: Contact with id abc-123 not found",
		},
		{
			name: "bad request",
			err:  BadRequest("invalid input"),
			want: "BAD_REQUEST: invalid input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		err        *AppError
		wantCode   Code
		wantStatus int
	}{
		{
			name:       "NotFound",
			err:        NotFound("Account", "id-1"),
			wantCode:   CodeNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "BadRequest",
			err:        BadRequest("bad"),
			wantCode:   CodeBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Forbidden",
			err:        Forbidden("no access"),
			wantCode:   CodeForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Unauthorized",
			err:        Unauthorized("no token"),
			wantCode:   CodeUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Conflict",
			err:        Conflict("duplicate"),
			wantCode:   CodeConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "Internal",
			err:        Internal("oops"),
			wantCode:   CodeInternal,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Validation",
			err:        Validation("field required"),
			wantCode:   CodeValidation,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", tt.err.Code, tt.wantCode)
			}
			if tt.err.HTTPStatus != tt.wantStatus {
				t.Errorf("HTTPStatus = %d, want %d", tt.err.HTTPStatus, tt.wantStatus)
			}
		})
	}
}

func TestRespond(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "AppError returns typed response",
			err:        NotFound("Contact", "abc"),
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":{"code":"NOT_FOUND","message":"Contact with id abc not found"}}`,
		},
		{
			name:       "wrapped AppError is unwrapped",
			err:        fmt.Errorf("service.Get: %w", NotFound("Deal", "xyz")),
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":{"code":"NOT_FOUND","message":"Deal with id xyz not found"}}`,
		},
		{
			name:       "non-AppError returns 500",
			err:        errors.New("unexpected"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"INTERNAL","message":"internal server error"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			Respond(c, tt.err)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if got := w.Body.String(); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}
