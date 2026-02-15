package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

type mockFunctionService struct {
	createFn  func(ctx context.Context, input metadata.CreateFunctionInput) (*metadata.Function, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*metadata.Function, error)
	listAllFn func(ctx context.Context) ([]metadata.Function, error)
	updateFn  func(ctx context.Context, id uuid.UUID, input metadata.UpdateFunctionInput) (*metadata.Function, error)
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockFunctionService) Create(ctx context.Context, input metadata.CreateFunctionInput) (*metadata.Function, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &metadata.Function{ID: uuid.New(), Name: input.Name, Body: input.Body}, nil
}

func (m *mockFunctionService) GetByID(ctx context.Context, id uuid.UUID) (*metadata.Function, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("function", id.String()))
}

func (m *mockFunctionService) ListAll(ctx context.Context) ([]metadata.Function, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return []metadata.Function{}, nil
}

func (m *mockFunctionService) Update(ctx context.Context, id uuid.UUID, input metadata.UpdateFunctionInput) (*metadata.Function, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("function", id.String()))
}

func (m *mockFunctionService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func setupFunctionRouter(t *testing.T, svc *mockFunctionService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	admin := r.Group("/api/v1/admin")
	h := NewFunctionHandler(svc)
	h.RegisterRoutes(admin)
	return r
}

func TestFunctionHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setupSvc   func(*mockFunctionService)
		wantStatus int
	}{
		{
			name: "creates function successfully",
			body: map[string]interface{}{
				"name": "double",
				"body": "x * 2",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for missing required fields",
			body: map[string]interface{}{
				"name": "test",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns error from service",
			body: map[string]interface{}{
				"name": "bad",
				"body": "x",
			},
			setupSvc: func(m *mockFunctionService) {
				m.createFn = func(_ context.Context, _ metadata.CreateFunctionInput) (*metadata.Function, error) {
					return nil, fmt.Errorf("%w", apperror.BadRequest("name must match"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 409 for duplicate name",
			body: map[string]interface{}{
				"name": "existing",
				"body": "x",
			},
			setupSvc: func(m *mockFunctionService) {
				m.createFn = func(_ context.Context, _ metadata.CreateFunctionInput) (*metadata.Function, error) {
					return nil, fmt.Errorf("%w", apperror.Conflict("function already exists"))
				}
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "passes description and return_type to service",
			body: map[string]interface{}{
				"name":        "calc",
				"body":        "x * 2",
				"description": "Calculation",
				"return_type": "number",
				"params":      []map[string]string{{"name": "x", "type": "number"}},
			},
			setupSvc: func(m *mockFunctionService) {
				m.createFn = func(_ context.Context, input metadata.CreateFunctionInput) (*metadata.Function, error) {
					if input.Description != "Calculation" {
						return nil, fmt.Errorf("expected description 'Calculation', got %q", input.Description)
					}
					if input.ReturnType != "number" {
						return nil, fmt.Errorf("expected return_type 'number', got %q", input.ReturnType)
					}
					return &metadata.Function{ID: uuid.New(), Name: input.Name, Body: input.Body}, nil
				}
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockFunctionService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupFunctionRouter(t, svc)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/functions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestFunctionHandler_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupSvc   func(*mockFunctionService)
		wantStatus int
		wantCount  int
	}{
		{
			name:       "returns empty list",
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "returns functions",
			setupSvc: func(m *mockFunctionService) {
				m.listAllFn = func(_ context.Context) ([]metadata.Function, error) {
					return []metadata.Function{
						{ID: uuid.New(), Name: "fn_a", Body: "x"},
						{ID: uuid.New(), Name: "fn_b", Body: "y"},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockFunctionService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupFunctionRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/functions", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var resp struct {
				Data []metadata.Function `json:"data"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Len(t, resp.Data, tt.wantCount)
		})
	}
}

func TestFunctionHandler_Get(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockFunctionService)
		wantStatus int
	}{
		{
			name: "returns function",
			id:   existingID.String(),
			setupSvc: func(m *mockFunctionService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*metadata.Function, error) {
					return &metadata.Function{ID: existingID, Name: "double", Body: "x * 2"}, nil
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockFunctionService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupFunctionRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/functions/"+tt.id, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestFunctionHandler_Update(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         string
		body       interface{}
		setupSvc   func(*mockFunctionService)
		wantStatus int
	}{
		{
			name: "updates successfully",
			id:   existingID.String(),
			body: map[string]interface{}{
				"body": "x * 3",
			},
			setupSvc: func(m *mockFunctionService) {
				m.updateFn = func(_ context.Context, _ uuid.UUID, input metadata.UpdateFunctionInput) (*metadata.Function, error) {
					return &metadata.Function{ID: existingID, Name: "test", Body: input.Body}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid UUID",
			id:         "bad-uuid",
			body:       map[string]interface{}{"body": "x"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			id:         existingID.String(),
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 404 for nonexistent",
			id:   uuid.New().String(),
			body: map[string]interface{}{
				"body": "x * 2",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "passes description and return_type",
			id:   existingID.String(),
			body: map[string]interface{}{
				"body":        "x * 2",
				"description": "Updated desc",
				"return_type": "number",
			},
			setupSvc: func(m *mockFunctionService) {
				m.updateFn = func(_ context.Context, _ uuid.UUID, input metadata.UpdateFunctionInput) (*metadata.Function, error) {
					if input.Description != "Updated desc" {
						return nil, fmt.Errorf("expected description")
					}
					if input.ReturnType != "number" {
						return nil, fmt.Errorf("expected return_type")
					}
					return &metadata.Function{ID: existingID, Name: "test", Body: input.Body}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockFunctionService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupFunctionRouter(t, svc)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/functions/"+tt.id, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestFunctionHandler_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockFunctionService)
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
			setupSvc: func(m *mockFunctionService) {
				m.deleteFn = func(_ context.Context, id uuid.UUID) error {
					return fmt.Errorf("%w", apperror.NotFound("function", id.String()))
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns 409 when function is in use",
			id:   uuid.New().String(),
			setupSvc: func(m *mockFunctionService) {
				m.deleteFn = func(_ context.Context, _ uuid.UUID) error {
					return fmt.Errorf("%w", apperror.Conflict("function is used in 2 place(s)"))
				}
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockFunctionService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupFunctionRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/functions/"+tt.id, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}
