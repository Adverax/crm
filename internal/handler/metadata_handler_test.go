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

	"github.com/adverax/crm/internal/api"
	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

type mockObjectService struct {
	createFn  func(ctx context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*metadata.ObjectDefinition, error)
	listFn    func(ctx context.Context, filter metadata.ObjectFilter) ([]metadata.ObjectDefinition, int64, error)
	updateFn  func(ctx context.Context, id uuid.UUID, input metadata.UpdateObjectInput) (*metadata.ObjectDefinition, error)
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockObjectService) Create(ctx context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &metadata.ObjectDefinition{ID: uuid.New(), APIName: input.APIName, Label: input.Label}, nil
}

func (m *mockObjectService) GetByID(ctx context.Context, id uuid.UUID) (*metadata.ObjectDefinition, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("ObjectDefinition", id.String()))
}

func (m *mockObjectService) List(ctx context.Context, filter metadata.ObjectFilter) ([]metadata.ObjectDefinition, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return []metadata.ObjectDefinition{}, 0, nil
}

func (m *mockObjectService) Update(ctx context.Context, id uuid.UUID, input metadata.UpdateObjectInput) (*metadata.ObjectDefinition, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("ObjectDefinition", id.String()))
}

func (m *mockObjectService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockFieldService struct {
	createFn    func(ctx context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error)
	getByIDFn   func(ctx context.Context, id uuid.UUID) (*metadata.FieldDefinition, error)
	listByObjFn func(ctx context.Context, objectID uuid.UUID) ([]metadata.FieldDefinition, error)
	updateFn    func(ctx context.Context, id uuid.UUID, input metadata.UpdateFieldInput) (*metadata.FieldDefinition, error)
	deleteFn    func(ctx context.Context, id uuid.UUID) error
}

func (m *mockFieldService) Create(ctx context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &metadata.FieldDefinition{ID: uuid.New(), APIName: input.APIName}, nil
}

func (m *mockFieldService) GetByID(ctx context.Context, id uuid.UUID) (*metadata.FieldDefinition, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("FieldDefinition", id.String()))
}

func (m *mockFieldService) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]metadata.FieldDefinition, error) {
	if m.listByObjFn != nil {
		return m.listByObjFn(ctx, objectID)
	}
	return []metadata.FieldDefinition{}, nil
}

func (m *mockFieldService) Update(ctx context.Context, id uuid.UUID, input metadata.UpdateFieldInput) (*metadata.FieldDefinition, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("FieldDefinition", id.String()))
}

func (m *mockFieldService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func setupRouter(h *MetadataHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api.RegisterHandlers(r, h)
	return r
}

func TestMetadataHandlerCreateObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setupObj   func(*mockObjectService)
		wantStatus int
	}{
		{
			name: "creates object successfully",
			body: api.CreateObjectRequest{
				ApiName:     "Invoice__c",
				Label:       "Invoice",
				PluralLabel: "Invoices",
				ObjectType:  "custom",
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
			body: api.CreateObjectRequest{
				ApiName:     "Bad",
				Label:       "Bad",
				PluralLabel: "Bads",
				ObjectType:  "custom",
			},
			setupObj: func(m *mockObjectService) {
				m.createFn = func(_ context.Context, _ metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
					return nil, fmt.Errorf("%w", apperror.Validation("validation failed"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objSvc := &mockObjectService{}
			if tt.setupObj != nil {
				tt.setupObj(objSvc)
			}
			h := NewMetadataHandler(objSvc, &mockFieldService{})
			r := setupRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/metadata/objects", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestMetadataHandlerGetObject(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         string
		setupObj   func(*mockObjectService)
		wantStatus int
	}{
		{
			name: "returns object",
			id:   existingID.String(),
			setupObj: func(m *mockObjectService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*metadata.ObjectDefinition, error) {
					return &metadata.ObjectDefinition{ID: existingID, APIName: "Account", Label: "Account"}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 404 for nonexistent",
			id:         uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objSvc := &mockObjectService{}
			if tt.setupObj != nil {
				tt.setupObj(objSvc)
			}
			h := NewMetadataHandler(objSvc, &mockFieldService{})
			r := setupRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/metadata/objects/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestMetadataHandlerListObjects(t *testing.T) {
	t.Parallel()

	objSvc := &mockObjectService{
		listFn: func(_ context.Context, _ metadata.ObjectFilter) ([]metadata.ObjectDefinition, int64, error) {
			return []metadata.ObjectDefinition{
				{ID: uuid.New(), APIName: "Account"},
			}, 1, nil
		},
	}
	h := NewMetadataHandler(objSvc, &mockFieldService{})
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/metadata/objects?page=1&per_page=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp api.ObjectListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Data == nil || len(*resp.Data) != 1 {
		t.Errorf("expected 1 object, got %v", resp.Data)
	}
	if resp.Pagination == nil || resp.Pagination.Total == nil || *resp.Pagination.Total != 1 {
		t.Errorf("expected total=1, got %v", resp.Pagination)
	}
}

func TestMetadataHandlerDeleteObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupObj   func(*mockObjectService)
		wantStatus int
	}{
		{
			name:       "deletes successfully",
			id:         uuid.New().String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name: "returns 404",
			id:   uuid.New().String(),
			setupObj: func(m *mockObjectService) {
				m.deleteFn = func(_ context.Context, id uuid.UUID) error {
					return fmt.Errorf("%w", apperror.NotFound("ObjectDefinition", id.String()))
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns 403 for non-deleteable",
			id:   uuid.New().String(),
			setupObj: func(m *mockObjectService) {
				m.deleteFn = func(_ context.Context, _ uuid.UUID) error {
					return fmt.Errorf("%w", apperror.Forbidden("cannot delete"))
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			objSvc := &mockObjectService{}
			if tt.setupObj != nil {
				tt.setupObj(objSvc)
			}
			h := NewMetadataHandler(objSvc, &mockFieldService{})
			r := setupRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/metadata/objects/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestMetadataHandlerCreateField(t *testing.T) {
	t.Parallel()

	objectID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		setupField func(*mockFieldService)
		wantStatus int
	}{
		{
			name: "creates field successfully",
			body: api.CreateFieldRequest{
				ApiName:   "first_name__c",
				Label:     "First Name",
				FieldType: "text",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "bad",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fieldSvc := &mockFieldService{}
			if tt.setupField != nil {
				tt.setupField(fieldSvc)
			}
			h := NewMetadataHandler(&mockObjectService{}, fieldSvc)
			r := setupRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost,
				"/api/v1/admin/metadata/objects/"+objectID.String()+"/fields",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestMetadataHandlerGetField(t *testing.T) {
	t.Parallel()

	objectID := uuid.New()
	fieldID := uuid.New()

	tests := []struct {
		name       string
		setupField func(*mockFieldService)
		wantStatus int
	}{
		{
			name: "returns field",
			setupField: func(m *mockFieldService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*metadata.FieldDefinition, error) {
					return &metadata.FieldDefinition{ID: fieldID, ObjectID: objectID, APIName: "name"}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 404",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fieldSvc := &mockFieldService{}
			if tt.setupField != nil {
				tt.setupField(fieldSvc)
			}
			h := NewMetadataHandler(&mockObjectService{}, fieldSvc)
			r := setupRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet,
				"/api/v1/admin/metadata/objects/"+objectID.String()+"/fields/"+fieldID.String(), nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestMetadataHandlerDeleteField(t *testing.T) {
	t.Parallel()

	objectID := uuid.New()
	fieldID := uuid.New()

	h := NewMetadataHandler(&mockObjectService{}, &mockFieldService{})
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete,
		"/api/v1/admin/metadata/objects/"+objectID.String()+"/fields/"+fieldID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestMetadataHandlerHealthCheck(t *testing.T) {
	t.Parallel()

	h := NewMetadataHandler(&mockObjectService{}, &mockFieldService{})
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
