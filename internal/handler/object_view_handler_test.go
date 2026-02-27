package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// --- Mock ObjectViewService ---

type mockObjectViewService struct {
	createFn   func(ctx context.Context, input metadata.CreateObjectViewInput) (*metadata.ObjectView, error)
	getByIDFn  func(ctx context.Context, id uuid.UUID) (*metadata.ObjectView, error)
	getByAPIFn func(ctx context.Context, apiName string) (*metadata.ObjectView, error)
	listAllFn  func(ctx context.Context) ([]metadata.ObjectView, error)
	updateFn   func(ctx context.Context, id uuid.UUID, input metadata.UpdateObjectViewInput) (*metadata.ObjectView, error)
	deleteFn   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockObjectViewService) Create(ctx context.Context, input metadata.CreateObjectViewInput) (*metadata.ObjectView, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &metadata.ObjectView{
		ID:      uuid.New(),
		APIName: input.APIName,
		Label:   input.Label,
	}, nil
}

func (m *mockObjectViewService) GetByID(ctx context.Context, id uuid.UUID) (*metadata.ObjectView, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("object_view", id.String()))
}

func (m *mockObjectViewService) GetByAPIName(ctx context.Context, apiName string) (*metadata.ObjectView, error) {
	if m.getByAPIFn != nil {
		return m.getByAPIFn(ctx, apiName)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("object_view", apiName))
}

func (m *mockObjectViewService) ListAll(ctx context.Context) ([]metadata.ObjectView, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return []metadata.ObjectView{}, nil
}

func (m *mockObjectViewService) Update(ctx context.Context, id uuid.UUID, input metadata.UpdateObjectViewInput) (*metadata.ObjectView, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("object_view", id.String()))
}

func (m *mockObjectViewService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func setupObjectViewRouter(t *testing.T, svc *mockObjectViewService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/api/v1/admin")
	h := NewObjectViewHandler(svc)
	h.RegisterRoutes(admin)
	return r
}

func TestObjectViewHandler_Create(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		setupSvc   func(*mockObjectViewService)
		wantStatus int
	}{
		{
			name: "creates object view successfully",
			body: map[string]interface{}{
				"api_name": "default_view",
				"label":    "Default View",
				"config": map[string]interface{}{
					"view": map[string]interface{}{
						"fields": []map[string]string{{"name": "first_name"}, {"name": "last_name"}},
					},
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "creates object view with profile_id",
			body: map[string]interface{}{
				"profile_id": profileID.String(),
				"api_name":   "sales_view",
				"label":      "Sales View",
			},
			setupSvc: func(m *mockObjectViewService) {
				m.createFn = func(_ context.Context, input metadata.CreateObjectViewInput) (*metadata.ObjectView, error) {
					if input.ProfileID == nil {
						return nil, fmt.Errorf("expected profile_id to be set")
					}
					if *input.ProfileID != profileID {
						return nil, fmt.Errorf("expected profile_id %s, got %s", profileID, *input.ProfileID)
					}
					return &metadata.ObjectView{
						ID:        uuid.New(),
						ProfileID: input.ProfileID,
						APIName:   input.APIName,
						Label:     input.Label,
					}, nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "passes description to service",
			body: map[string]interface{}{
				"api_name":    "detailed_view",
				"label":       "Detailed View",
				"description": "A detailed view for contacts",
			},
			setupSvc: func(m *mockObjectViewService) {
				m.createFn = func(_ context.Context, input metadata.CreateObjectViewInput) (*metadata.ObjectView, error) {
					if input.Description != "A detailed view for contacts" {
						return nil, fmt.Errorf("expected description 'A detailed view for contacts', got %q", input.Description)
					}
					return &metadata.ObjectView{
						ID:          uuid.New(),
						APIName:     input.APIName,
						Label:       input.Label,
						Description: input.Description,
					}, nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for missing api_name",
			body: map[string]interface{}{
				"label": "Test View",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for missing label",
			body: map[string]interface{}{
				"api_name": "test_view",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for invalid profile_id",
			body: map[string]interface{}{
				"profile_id": "bad-uuid",
				"api_name":   "test_view",
				"label":      "Test View",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns error from service",
			body: map[string]interface{}{
				"api_name": "test_view",
				"label":    "Test View",
			},
			setupSvc: func(m *mockObjectViewService) {
				m.createFn = func(_ context.Context, _ metadata.CreateObjectViewInput) (*metadata.ObjectView, error) {
					return nil, fmt.Errorf("%w", apperror.BadRequest("api_name must match ^[a-z][a-z0-9_]*$"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 409 for duplicate api_name",
			body: map[string]interface{}{
				"api_name": "existing_view",
				"label":    "Existing View",
			},
			setupSvc: func(m *mockObjectViewService) {
				m.createFn = func(_ context.Context, _ metadata.CreateObjectViewInput) (*metadata.ObjectView, error) {
					return nil, fmt.Errorf("%w", apperror.Conflict("object view with this api_name already exists"))
				}
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockObjectViewService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupObjectViewRouter(t, svc)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/object-views", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestObjectViewHandler_List(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		setupSvc   func(*mockObjectViewService)
		wantStatus int
		wantCount  int
	}{
		{
			name:       "returns empty list",
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "returns all object views",
			setupSvc: func(m *mockObjectViewService) {
				m.listAllFn = func(_ context.Context) ([]metadata.ObjectView, error) {
					return []metadata.ObjectView{
						{
							ID:        uuid.New(),
							APIName:   "view_a",
							Label:     "View A",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        uuid.New(),
							APIName:   "view_b",
							Label:     "View B",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "returns 500 on service error",
			setupSvc: func(m *mockObjectViewService) {
				m.listAllFn = func(_ context.Context) ([]metadata.ObjectView, error) {
					return nil, fmt.Errorf("database connection failed")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockObjectViewService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupObjectViewRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/object-views", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Data []metadata.ObjectView `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Len(t, resp.Data, tt.wantCount)
			}
		})
	}
}

func TestObjectViewHandler_Get(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()
	now := time.Now()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockObjectViewService)
		wantStatus int
	}{
		{
			name: "returns object view",
			id:   existingID.String(),
			setupSvc: func(m *mockObjectViewService) {
				m.getByIDFn = func(_ context.Context, id uuid.UUID) (*metadata.ObjectView, error) {
					return &metadata.ObjectView{
						ID:      id,
						APIName: "default_view",
						Label:   "Default View",
						Config: metadata.OVConfig{
							View: metadata.OVViewConfig{
								Fields: []metadata.OVViewField{{Name: "name"}},
							},
						},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 404 for nonexistent",
			id:         uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 for invalid UUID",
			id:         "not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			id:   existingID.String(),
			setupSvc: func(m *mockObjectViewService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*metadata.ObjectView, error) {
					return nil, fmt.Errorf("unexpected database error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockObjectViewService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupObjectViewRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/object-views/"+tt.id, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Data metadata.ObjectView `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, existingID, resp.Data.ID)
				assert.Equal(t, "default_view", resp.Data.APIName)
			}
		})
	}
}

func TestObjectViewHandler_Update(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()
	now := time.Now()

	tests := []struct {
		name       string
		id         string
		body       interface{}
		setupSvc   func(*mockObjectViewService)
		wantStatus int
	}{
		{
			name: "updates successfully",
			id:   existingID.String(),
			body: map[string]interface{}{
				"label": "Updated View",
				"config": map[string]interface{}{
					"view": map[string]interface{}{
						"fields": []map[string]string{{"name": "email"}, {"name": "phone"}},
					},
				},
			},
			setupSvc: func(m *mockObjectViewService) {
				m.updateFn = func(_ context.Context, id uuid.UUID, input metadata.UpdateObjectViewInput) (*metadata.ObjectView, error) {
					return &metadata.ObjectView{
						ID:        id,
						APIName:   "default_view",
						Label:     input.Label,
						Config:    input.Config,
						CreatedAt: now,
						UpdatedAt: time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "passes description to service",
			id:   existingID.String(),
			body: map[string]interface{}{
				"label":       "Updated View",
				"description": "Updated description",
			},
			setupSvc: func(m *mockObjectViewService) {
				m.updateFn = func(_ context.Context, id uuid.UUID, input metadata.UpdateObjectViewInput) (*metadata.ObjectView, error) {
					if input.Description != "Updated description" {
						return nil, fmt.Errorf("expected description 'Updated description', got %q", input.Description)
					}
					return &metadata.ObjectView{
						ID:          id,
						APIName:     "test_view",
						Label:       input.Label,
						Description: input.Description,
						CreatedAt:   now,
						UpdatedAt:   time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid UUID",
			id:         "bad-uuid",
			body:       map[string]interface{}{"label": "Test"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			id:         existingID.String(),
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for missing label",
			id:         existingID.String(),
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 404 for nonexistent",
			id:   uuid.New().String(),
			body: map[string]interface{}{
				"label": "Updated View",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns error from service",
			id:   existingID.String(),
			body: map[string]interface{}{
				"label": "Updated View",
			},
			setupSvc: func(m *mockObjectViewService) {
				m.updateFn = func(_ context.Context, _ uuid.UUID, _ metadata.UpdateObjectViewInput) (*metadata.ObjectView, error) {
					return nil, fmt.Errorf("%w", apperror.BadRequest("label must be at most 255 characters"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockObjectViewService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupObjectViewRouter(t, svc)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/object-views/"+tt.id, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestObjectViewHandler_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockObjectViewService)
		wantStatus int
	}{
		{
			name:       "deletes successfully",
			id:         uuid.New().String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for invalid UUID",
			id:         "bad-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 404 for nonexistent",
			id:   uuid.New().String(),
			setupSvc: func(m *mockObjectViewService) {
				m.deleteFn = func(_ context.Context, id uuid.UUID) error {
					return fmt.Errorf("%w", apperror.NotFound("object_view", id.String()))
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns 409 when view is in use",
			id:   uuid.New().String(),
			setupSvc: func(m *mockObjectViewService) {
				m.deleteFn = func(_ context.Context, _ uuid.UUID) error {
					return fmt.Errorf("%w", apperror.Conflict("object view is referenced by layouts"))
				}
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "returns 500 on service error",
			id:   uuid.New().String(),
			setupSvc: func(m *mockObjectViewService) {
				m.deleteFn = func(_ context.Context, _ uuid.UUID) error {
					return fmt.Errorf("unexpected database error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockObjectViewService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupObjectViewRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/object-views/"+tt.id, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}
