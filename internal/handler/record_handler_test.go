package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/service"
)

type mockRecordService struct {
	listFn    func(ctx context.Context, objectName string, params service.ListParams) (*service.RecordListResult, error)
	getByIDFn func(ctx context.Context, objectName string, recordID string) (map[string]any, error)
	createFn  func(ctx context.Context, objectName string, fields map[string]any) (*service.CreateResult, error)
	updateFn  func(ctx context.Context, objectName string, recordID string, fields map[string]any) error
	deleteFn  func(ctx context.Context, objectName string, recordID string) error
}

func (m *mockRecordService) List(ctx context.Context, objectName string, params service.ListParams) (*service.RecordListResult, error) {
	if m.listFn != nil {
		return m.listFn(ctx, objectName, params)
	}
	return &service.RecordListResult{}, nil
}

func (m *mockRecordService) GetByID(ctx context.Context, objectName string, recordID string) (map[string]any, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, objectName, recordID)
	}
	return nil, apperror.NotFound("record", recordID)
}

func (m *mockRecordService) Create(ctx context.Context, objectName string, fields map[string]any) (*service.CreateResult, error) {
	if m.createFn != nil {
		return m.createFn(ctx, objectName, fields)
	}
	return &service.CreateResult{ID: "new-id"}, nil
}

func (m *mockRecordService) Update(ctx context.Context, objectName string, recordID string, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, objectName, recordID, fields)
	}
	return nil
}

func (m *mockRecordService) Delete(ctx context.Context, objectName string, recordID string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, objectName, recordID)
	}
	return nil
}

func setupRecordRouter(t *testing.T, h *RecordHandler) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func TestRecordHandler_ListRecords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		setupSvc   func(*mockRecordService)
		wantStatus int
	}{
		{
			name: "returns records successfully",
			url:  "/api/v1/records/Account?page=1&per_page=10",
			setupSvc: func(m *mockRecordService) {
				m.listFn = func(_ context.Context, _ string, _ service.ListParams) (*service.RecordListResult, error) {
					return &service.RecordListResult{
						Data: []map[string]any{{"Id": "id1", "Name": "Acme"}},
						Pagination: service.PaginationMeta{
							Page: 1, PerPage: 10, Total: 1, TotalPages: 1,
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 404 for unknown object",
			url:  "/api/v1/records/Unknown",
			setupSvc: func(m *mockRecordService) {
				m.listFn = func(_ context.Context, _ string, _ service.ListParams) (*service.RecordListResult, error) {
					return nil, apperror.NotFound("object", "Unknown")
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "uses default pagination",
			url:        "/api/v1/records/Account",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockRecordService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewRecordHandler(svc)
			r := setupRecordRouter(t, h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestRecordHandler_GetRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		setupSvc   func(*mockRecordService)
		wantStatus int
	}{
		{
			name: "returns record successfully",
			url:  "/api/v1/records/Account/some-id",
			setupSvc: func(m *mockRecordService) {
				m.getByIDFn = func(_ context.Context, _ string, _ string) (map[string]any, error) {
					return map[string]any{"Id": "some-id", "Name": "Acme"}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 404 when not found",
			url:  "/api/v1/records/Account/missing-id",
			setupSvc: func(m *mockRecordService) {
				m.getByIDFn = func(_ context.Context, _ string, _ string) (map[string]any, error) {
					return nil, apperror.NotFound("record", "missing-id")
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns 400 for invalid UUID",
			url:  "/api/v1/records/Account/bad-uuid",
			setupSvc: func(m *mockRecordService) {
				m.getByIDFn = func(_ context.Context, _ string, _ string) (map[string]any, error) {
					return nil, apperror.BadRequest("invalid UUID format")
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockRecordService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewRecordHandler(svc)
			r := setupRecordRouter(t, h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestRecordHandler_CreateRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setupSvc   func(*mockRecordService)
		wantStatus int
	}{
		{
			name: "creates record successfully",
			body: map[string]any{"Name": "Test Corp"},
			setupSvc: func(m *mockRecordService) {
				m.createFn = func(_ context.Context, _ string, _ map[string]any) (*service.CreateResult, error) {
					return &service.CreateResult{ID: "new-id"}, nil
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
			name: "returns error from service",
			body: map[string]any{"Name": "Test"},
			setupSvc: func(m *mockRecordService) {
				m.createFn = func(_ context.Context, _ string, _ map[string]any) (*service.CreateResult, error) {
					return nil, apperror.Forbidden("not createable")
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockRecordService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewRecordHandler(svc)
			r := setupRecordRouter(t, h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/records/Account", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestRecordHandler_UpdateRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setupSvc   func(*mockRecordService)
		wantStatus int
	}{
		{
			name: "updates record successfully",
			body: map[string]any{"Name": "Updated"},
			setupSvc: func(m *mockRecordService) {
				m.updateFn = func(_ context.Context, _ string, _ string, _ map[string]any) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 404 from service",
			body: map[string]any{"Name": "Updated"},
			setupSvc: func(m *mockRecordService) {
				m.updateFn = func(_ context.Context, _ string, _ string, _ map[string]any) error {
					return apperror.NotFound("record", "id")
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockRecordService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewRecordHandler(svc)
			r := setupRecordRouter(t, h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/api/v1/records/Account/some-id", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestRecordHandler_DeleteRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupSvc   func(*mockRecordService)
		wantStatus int
	}{
		{
			name: "deletes record successfully",
			setupSvc: func(m *mockRecordService) {
				m.deleteFn = func(_ context.Context, _ string, _ string) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "returns 404 from service",
			setupSvc: func(m *mockRecordService) {
				m.deleteFn = func(_ context.Context, _ string, _ string) error {
					return apperror.NotFound("record", "id")
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns 403 for forbidden",
			setupSvc: func(m *mockRecordService) {
				m.deleteFn = func(_ context.Context, _ string, _ string) error {
					return apperror.Forbidden("not deleteable")
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockRecordService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewRecordHandler(svc)
			r := setupRecordRouter(t, h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/records/Account/some-id", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
