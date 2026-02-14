package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/templates"
)

// --- template-specific mocks (prefixed to avoid collisions) ---

type tmplMockObjectService struct{}

func (m *tmplMockObjectService) Create(_ context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
	return &metadata.ObjectDefinition{ID: uuid.New(), APIName: input.APIName}, nil
}
func (m *tmplMockObjectService) GetByID(_ context.Context, _ uuid.UUID) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectService) List(_ context.Context, _ metadata.ObjectFilter) ([]metadata.ObjectDefinition, int64, error) {
	return nil, 0, nil
}
func (m *tmplMockObjectService) Update(_ context.Context, _ uuid.UUID, _ metadata.UpdateObjectInput) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectService) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type tmplMockFieldService struct{}

func (m *tmplMockFieldService) Create(_ context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error) {
	return &metadata.FieldDefinition{ID: uuid.New(), APIName: input.APIName}, nil
}
func (m *tmplMockFieldService) GetByID(_ context.Context, _ uuid.UUID) (*metadata.FieldDefinition, error) {
	return nil, nil
}
func (m *tmplMockFieldService) ListByObjectID(_ context.Context, _ uuid.UUID) ([]metadata.FieldDefinition, error) {
	return nil, nil
}
func (m *tmplMockFieldService) Update(_ context.Context, _ uuid.UUID, _ metadata.UpdateFieldInput) (*metadata.FieldDefinition, error) {
	return nil, nil
}
func (m *tmplMockFieldService) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type tmplMockObjectRepo struct {
	countResult int64
}

func (m *tmplMockObjectRepo) Count(_ context.Context) (int64, error) { return m.countResult, nil }
func (m *tmplMockObjectRepo) Create(_ context.Context, _ pgx.Tx, _ metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectRepo) GetByID(_ context.Context, _ uuid.UUID) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectRepo) GetByAPIName(_ context.Context, _ string) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectRepo) List(_ context.Context, _, _ int32) ([]metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectRepo) ListAll(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectRepo) Update(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ metadata.UpdateObjectInput) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *tmplMockObjectRepo) Delete(_ context.Context, _ pgx.Tx, _ uuid.UUID) error { return nil }

type tmplMockPermissionService struct{}

func (m *tmplMockPermissionService) SetObjectPermission(_ context.Context, _ uuid.UUID, _ security.SetObjectPermissionInput) (*security.ObjectPermission, error) {
	return &security.ObjectPermission{ID: uuid.New()}, nil
}
func (m *tmplMockPermissionService) ListObjectPermissions(_ context.Context, _ uuid.UUID) ([]security.ObjectPermission, error) {
	return nil, nil
}
func (m *tmplMockPermissionService) RemoveObjectPermission(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *tmplMockPermissionService) SetFieldPermission(_ context.Context, _ uuid.UUID, _ security.SetFieldPermissionInput) (*security.FieldPermission, error) {
	return &security.FieldPermission{ID: uuid.New()}, nil
}
func (m *tmplMockPermissionService) ListFieldPermissions(_ context.Context, _ uuid.UUID) ([]security.FieldPermission, error) {
	return nil, nil
}
func (m *tmplMockPermissionService) RemoveFieldPermission(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

type tmplMockCacheInvalidator struct{}

func (m *tmplMockCacheInvalidator) Invalidate(_ context.Context) error { return nil }

// --- tests ---

func setupTemplateHandler(countResult int64) (*TemplateHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	registry := templates.BuildRegistry()
	applier := templates.NewApplier(
		&tmplMockObjectService{},
		&tmplMockFieldService{},
		&tmplMockObjectRepo{countResult: countResult},
		&tmplMockPermissionService{},
		&tmplMockCacheInvalidator{},
	)

	h := NewTemplateHandler(registry, applier)

	router := gin.New()
	group := router.Group("/api/v1/admin")
	h.RegisterRoutes(group)

	return h, router
}

func TestTemplateHandler_ListTemplates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "returns all registered templates",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, router := setupTemplateHandler(0)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/templates", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var resp struct {
				Data []templates.TemplateInfo `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if len(resp.Data) != tt.expectedCount {
				t.Errorf("expected %d templates, got %d", tt.expectedCount, len(resp.Data))
			}

			for _, tmpl := range resp.Data {
				if tmpl.ID == "" {
					t.Error("template ID is empty")
				}
				if tmpl.Label == "" {
					t.Error("template Label is empty")
				}
				if tmpl.Objects == 0 {
					t.Error("template Objects count is 0")
				}
				if tmpl.Fields == 0 {
					t.Error("template Fields count is 0")
				}
			}
		})
	}
}

func TestTemplateHandler_ApplyTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		templateID     string
		countResult    int64
		expectedStatus int
	}{
		{
			name:           "applies template successfully",
			templateID:     "sales_crm",
			countResult:    0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "returns not found for unknown template",
			templateID:     "nonexistent",
			countResult:    0,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "returns conflict when objects exist",
			templateID:     "sales_crm",
			countResult:    5,
			expectedStatus: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, router := setupTemplateHandler(tt.countResult)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/templates/"+tt.templateID+"/apply", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
