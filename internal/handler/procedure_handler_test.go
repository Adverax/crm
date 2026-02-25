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

type mockProcedureService struct {
	createFn             func(ctx context.Context, input metadata.CreateProcedureInput) (*metadata.ProcedureWithVersions, error)
	getByIDFn            func(ctx context.Context, id uuid.UUID) (*metadata.ProcedureWithVersions, error)
	getByCodeFn          func(ctx context.Context, code string) (*metadata.Procedure, error)
	listAllFn            func(ctx context.Context) ([]metadata.Procedure, error)
	deleteFn             func(ctx context.Context, id uuid.UUID) error
	updateMetadataFn     func(ctx context.Context, id uuid.UUID, input metadata.UpdateProcedureMetadataInput) (*metadata.Procedure, error)
	saveDraftFn          func(ctx context.Context, id uuid.UUID, input metadata.SaveDraftInput) (*metadata.ProcedureVersion, error)
	discardDraftFn       func(ctx context.Context, id uuid.UUID) error
	createDraftFromPubFn func(ctx context.Context, id uuid.UUID) (*metadata.ProcedureVersion, error)
	publishFn            func(ctx context.Context, id uuid.UUID) (*metadata.ProcedureVersion, error)
	rollbackFn           func(ctx context.Context, id uuid.UUID) (*metadata.ProcedureVersion, error)
	listVersionsFn       func(ctx context.Context, id uuid.UUID) ([]metadata.ProcedureVersion, error)
	getPublishedDefFn    func(ctx context.Context, code string) (*metadata.ProcedureDefinition, error)
}

func (m *mockProcedureService) Create(ctx context.Context, input metadata.CreateProcedureInput) (*metadata.ProcedureWithVersions, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &metadata.ProcedureWithVersions{
		Procedure: metadata.Procedure{ID: uuid.New(), Code: input.Code, Name: input.Name},
	}, nil
}

func (m *mockProcedureService) GetByID(ctx context.Context, id uuid.UUID) (*metadata.ProcedureWithVersions, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("procedure", id.String()))
}

func (m *mockProcedureService) GetByCode(ctx context.Context, code string) (*metadata.Procedure, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(ctx, code)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("procedure", code))
}

func (m *mockProcedureService) ListAll(ctx context.Context) ([]metadata.Procedure, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return []metadata.Procedure{}, nil
}

func (m *mockProcedureService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockProcedureService) UpdateMetadata(ctx context.Context, id uuid.UUID, input metadata.UpdateProcedureMetadataInput) (*metadata.Procedure, error) {
	if m.updateMetadataFn != nil {
		return m.updateMetadataFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("procedure", id.String()))
}

func (m *mockProcedureService) SaveDraft(ctx context.Context, id uuid.UUID, input metadata.SaveDraftInput) (*metadata.ProcedureVersion, error) {
	if m.saveDraftFn != nil {
		return m.saveDraftFn(ctx, id, input)
	}
	return &metadata.ProcedureVersion{ID: uuid.New(), Version: 1}, nil
}

func (m *mockProcedureService) DiscardDraft(ctx context.Context, id uuid.UUID) error {
	if m.discardDraftFn != nil {
		return m.discardDraftFn(ctx, id)
	}
	return nil
}

func (m *mockProcedureService) CreateDraftFromPublished(ctx context.Context, id uuid.UUID) (*metadata.ProcedureVersion, error) {
	if m.createDraftFromPubFn != nil {
		return m.createDraftFromPubFn(ctx, id)
	}
	return &metadata.ProcedureVersion{ID: uuid.New(), Version: 2, Status: metadata.VersionStatusDraft}, nil
}

func (m *mockProcedureService) Publish(ctx context.Context, id uuid.UUID) (*metadata.ProcedureVersion, error) {
	if m.publishFn != nil {
		return m.publishFn(ctx, id)
	}
	return &metadata.ProcedureVersion{ID: uuid.New(), Version: 1, Status: metadata.VersionStatusPublished}, nil
}

func (m *mockProcedureService) Rollback(ctx context.Context, id uuid.UUID) (*metadata.ProcedureVersion, error) {
	if m.rollbackFn != nil {
		return m.rollbackFn(ctx, id)
	}
	return &metadata.ProcedureVersion{ID: uuid.New(), Version: 1, Status: metadata.VersionStatusPublished}, nil
}

func (m *mockProcedureService) ListVersions(ctx context.Context, id uuid.UUID) ([]metadata.ProcedureVersion, error) {
	if m.listVersionsFn != nil {
		return m.listVersionsFn(ctx, id)
	}
	return []metadata.ProcedureVersion{}, nil
}

func (m *mockProcedureService) GetPublishedDefinition(ctx context.Context, code string) (*metadata.ProcedureDefinition, error) {
	if m.getPublishedDefFn != nil {
		return m.getPublishedDefFn(ctx, code)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("procedure", code))
}

func setupProcedureRouter(t *testing.T, svc *mockProcedureService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/api/v1/admin")
	h := NewProcedureHandler(svc, nil)
	h.RegisterRoutes(admin)
	return r
}

func TestProcedureHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setupSvc   func(*mockProcedureService)
		wantStatus int
	}{
		{
			name: "creates procedure successfully",
			body: map[string]interface{}{
				"code": "send_welcome",
				"name": "Send Welcome",
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
				"code": "test",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 409 for duplicate code",
			body: map[string]interface{}{
				"code": "existing",
				"name": "Existing",
			},
			setupSvc: func(m *mockProcedureService) {
				m.createFn = func(_ context.Context, _ metadata.CreateProcedureInput) (*metadata.ProcedureWithVersions, error) {
					return nil, fmt.Errorf("%w", apperror.Conflict("procedure already exists"))
				}
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/procedures", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestProcedureHandler_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupSvc   func(*mockProcedureService)
		wantStatus int
		wantCount  int
	}{
		{
			name:       "returns empty list",
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "returns procedures",
			setupSvc: func(m *mockProcedureService) {
				m.listAllFn = func(_ context.Context) ([]metadata.Procedure, error) {
					return []metadata.Procedure{
						{ID: uuid.New(), Code: "a"},
						{ID: uuid.New(), Code: "b"},
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
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/procedures", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var resp struct {
				Data []metadata.Procedure `json:"data"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Len(t, resp.Data, tt.wantCount)
		})
	}
}

func TestProcedureHandler_Get(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockProcedureService)
		wantStatus int
	}{
		{
			name: "returns procedure with versions",
			id:   existingID.String(),
			setupSvc: func(m *mockProcedureService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*metadata.ProcedureWithVersions, error) {
					return &metadata.ProcedureWithVersions{
						Procedure: metadata.Procedure{ID: existingID, Code: "test"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/procedures/"+tt.id, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestProcedureHandler_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockProcedureService)
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
			setupSvc: func(m *mockProcedureService) {
				m.deleteFn = func(_ context.Context, id uuid.UUID) error {
					return fmt.Errorf("%w", apperror.NotFound("procedure", id.String()))
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/procedures/"+tt.id, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestProcedureHandler_SaveDraft(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		body       interface{}
		setupSvc   func(*mockProcedureService)
		wantStatus int
	}{
		{
			name: "saves draft successfully",
			id:   uuid.New().String(),
			body: map[string]interface{}{
				"definition": map[string]interface{}{
					"commands": []map[string]interface{}{
						{"type": "compute.transform", "as": "s1"},
					},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid UUID",
			id:         "bad-uuid",
			body:       map[string]interface{}{"definition": map[string]interface{}{"commands": []interface{}{}}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			id:         uuid.New().String(),
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/procedures/"+tt.id+"/draft", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestProcedureHandler_Publish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockProcedureService)
		wantStatus int
	}{
		{
			name:       "publishes successfully",
			id:         uuid.New().String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid UUID",
			id:         "bad-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 when no draft",
			id:   uuid.New().String(),
			setupSvc: func(m *mockProcedureService) {
				m.publishFn = func(_ context.Context, _ uuid.UUID) (*metadata.ProcedureVersion, error) {
					return nil, fmt.Errorf("%w", apperror.BadRequest("no draft to publish"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/procedures/"+tt.id+"/publish", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestProcedureHandler_Rollback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockProcedureService)
		wantStatus int
	}{
		{
			name:       "rolls back successfully",
			id:         uuid.New().String(),
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 400 when no previous version",
			id:   uuid.New().String(),
			setupSvc: func(m *mockProcedureService) {
				m.rollbackFn = func(_ context.Context, _ uuid.UUID) (*metadata.ProcedureVersion, error) {
					return nil, fmt.Errorf("%w", apperror.BadRequest("no previous version"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/procedures/"+tt.id+"/rollback", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestProcedureHandler_ListVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupSvc   func(*mockProcedureService)
		wantStatus int
	}{
		{
			name: "returns versions",
			id:   uuid.New().String(),
			setupSvc: func(m *mockProcedureService) {
				m.listVersionsFn = func(_ context.Context, _ uuid.UUID) ([]metadata.ProcedureVersion, error) {
					return []metadata.ProcedureVersion{
						{ID: uuid.New(), Version: 1},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid UUID",
			id:         "bad-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &mockProcedureService{}
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			r := setupProcedureRouter(t, svc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/procedures/"+tt.id+"/versions", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}
