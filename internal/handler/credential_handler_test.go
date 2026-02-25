package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/platform/credential"
)

func setupCredentialRouter(svc credential.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCredentialHandler(svc)
	admin := r.Group("/api/v1/admin")
	h.RegisterRoutes(admin)
	return r
}

func TestCredentialHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       map[string]any
		wantStatus int
	}{
		{
			name: "creates credential",
			body: map[string]any{
				"code":     "test_api",
				"name":     "Test API",
				"type":     "api_key",
				"base_url": "https://api.example.com",
				"auth_data": map[string]string{
					"header": "X-API-Key",
					"value":  "sk-12345",
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "rejects missing code",
			body: map[string]any{
				"name":     "Test",
				"type":     "api_key",
				"base_url": "https://api.example.com",
				"auth_data": map[string]string{
					"header": "X-Key",
					"value":  "val",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &stubCredentialService{}
			router := setupCredentialRouter(svc)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/credentials", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestCredentialHandler_List(t *testing.T) {
	t.Parallel()

	svc := &stubCredentialService{}
	router := setupCredentialRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/credentials", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
}

func TestCredentialHandler_Get(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	svc := &stubCredentialService{
		getByIDResult: &credential.Credential{ID: id, Code: "test", Name: "Test"},
	}
	router := setupCredentialRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/credentials/"+id.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCredentialHandler_Delete(t *testing.T) {
	t.Parallel()

	svc := &stubCredentialService{}
	router := setupCredentialRouter(svc)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/credentials/"+id.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCredentialHandler_TestConnection(t *testing.T) {
	t.Parallel()

	svc := &stubCredentialService{}
	router := setupCredentialRouter(svc)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/credentials/"+id.String()+"/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCredentialHandler_ActivateDeactivate(t *testing.T) {
	t.Parallel()

	svc := &stubCredentialService{}
	router := setupCredentialRouter(svc)

	id := uuid.New()

	// Deactivate
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/credentials/"+id.String()+"/deactivate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Activate
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/credentials/"+id.String()+"/activate", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCredentialHandler_UsageLog(t *testing.T) {
	t.Parallel()

	svc := &stubCredentialService{}
	router := setupCredentialRouter(svc)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/credentials/"+id.String()+"/usage", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// stubCredentialService implements credential.Service for handler tests.
type stubCredentialService struct {
	getByIDResult *credential.Credential
}

func (s *stubCredentialService) Create(_ context.Context, input credential.CreateCredentialInput) (*credential.Credential, error) {
	return &credential.Credential{
		ID:       uuid.New(),
		Code:     input.Code,
		Name:     input.Name,
		Type:     input.Type,
		BaseURL:  input.BaseURL,
		IsActive: true,
	}, nil
}

func (s *stubCredentialService) GetByID(_ context.Context, id uuid.UUID) (*credential.Credential, error) {
	if s.getByIDResult != nil {
		return s.getByIDResult, nil
	}
	return &credential.Credential{ID: id, Code: "test", Name: "Test"}, nil
}

func (s *stubCredentialService) GetByCode(_ context.Context, code string) (*credential.Credential, error) {
	return &credential.Credential{Code: code, Name: "Test"}, nil
}

func (s *stubCredentialService) ListAll(_ context.Context) ([]credential.Credential, error) {
	return []credential.Credential{}, nil
}

func (s *stubCredentialService) Update(_ context.Context, id uuid.UUID, input credential.UpdateCredentialInput) (*credential.Credential, error) {
	return &credential.Credential{ID: id, Name: input.Name, BaseURL: input.BaseURL}, nil
}

func (s *stubCredentialService) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *stubCredentialService) Activate(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *stubCredentialService) Deactivate(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *stubCredentialService) TestConnection(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *stubCredentialService) GetUsageLog(_ context.Context, _ uuid.UUID, _ int) ([]credential.UsageLogEntry, error) {
	return []credential.UsageLogEntry{}, nil
}

func (s *stubCredentialService) ResolveAuth(_ context.Context, _ string) (string, string, string, error) {
	return "", "", "", nil
}

func (s *stubCredentialService) LogUsage(_ context.Context, _ *credential.UsageLogEntry) error {
	return nil
}
