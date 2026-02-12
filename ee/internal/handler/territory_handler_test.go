//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.

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

	"github.com/adverax/crm/ee/internal/platform/territory"
	"github.com/adverax/crm/internal/pkg/apperror"
)

// --- Mock Services ---

type mockModelService struct {
	createFn    func(ctx context.Context, input territory.CreateModelInput) (*territory.TerritoryModel, error)
	getByIDFn   func(ctx context.Context, id uuid.UUID) (*territory.TerritoryModel, error)
	getActiveFn func(ctx context.Context) (*territory.TerritoryModel, error)
	listFn      func(ctx context.Context, page, perPage int32) ([]territory.TerritoryModel, int64, error)
	updateFn    func(ctx context.Context, id uuid.UUID, input territory.UpdateModelInput) (*territory.TerritoryModel, error)
	deleteFn    func(ctx context.Context, id uuid.UUID) error
	activateFn  func(ctx context.Context, id uuid.UUID) error
	archiveFn   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockModelService) Create(ctx context.Context, input territory.CreateModelInput) (*territory.TerritoryModel, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &territory.TerritoryModel{ID: uuid.New(), APIName: input.APIName, Label: input.Label}, nil
}

func (m *mockModelService) GetByID(ctx context.Context, id uuid.UUID) (*territory.TerritoryModel, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("TerritoryModel", id.String()))
}

func (m *mockModelService) GetActive(ctx context.Context) (*territory.TerritoryModel, error) {
	if m.getActiveFn != nil {
		return m.getActiveFn(ctx)
	}
	return nil, nil
}

func (m *mockModelService) List(ctx context.Context, page, perPage int32) ([]territory.TerritoryModel, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return []territory.TerritoryModel{}, 0, nil
}

func (m *mockModelService) Update(ctx context.Context, id uuid.UUID, input territory.UpdateModelInput) (*territory.TerritoryModel, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("TerritoryModel", id.String()))
}

func (m *mockModelService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockModelService) Activate(ctx context.Context, id uuid.UUID) error {
	if m.activateFn != nil {
		return m.activateFn(ctx, id)
	}
	return nil
}

func (m *mockModelService) Archive(ctx context.Context, id uuid.UUID) error {
	if m.archiveFn != nil {
		return m.archiveFn(ctx, id)
	}
	return nil
}

type mockTerritoryService struct {
	createFn      func(ctx context.Context, input territory.CreateTerritoryInput) (*territory.Territory, error)
	getByIDFn     func(ctx context.Context, id uuid.UUID) (*territory.Territory, error)
	listByModelFn func(ctx context.Context, modelID uuid.UUID) ([]territory.Territory, error)
	updateFn      func(ctx context.Context, id uuid.UUID, input territory.UpdateTerritoryInput) (*territory.Territory, error)
	deleteFn      func(ctx context.Context, id uuid.UUID) error
}

func (m *mockTerritoryService) Create(ctx context.Context, input territory.CreateTerritoryInput) (*territory.Territory, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &territory.Territory{ID: uuid.New(), APIName: input.APIName, Label: input.Label}, nil
}

func (m *mockTerritoryService) GetByID(ctx context.Context, id uuid.UUID) (*territory.Territory, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("Territory", id.String()))
}

func (m *mockTerritoryService) ListByModelID(ctx context.Context, modelID uuid.UUID) ([]territory.Territory, error) {
	if m.listByModelFn != nil {
		return m.listByModelFn(ctx, modelID)
	}
	return []territory.Territory{}, nil
}

func (m *mockTerritoryService) Update(ctx context.Context, id uuid.UUID, input territory.UpdateTerritoryInput) (*territory.Territory, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("Territory", id.String()))
}

func (m *mockTerritoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockObjDefaultService struct {
	setFn    func(ctx context.Context, input territory.SetObjectDefaultInput) (*territory.TerritoryObjectDefault, error)
	listFn   func(ctx context.Context, territoryID uuid.UUID) ([]territory.TerritoryObjectDefault, error)
	removeFn func(ctx context.Context, territoryID, objectID uuid.UUID) error
}

func (m *mockObjDefaultService) Set(ctx context.Context, input territory.SetObjectDefaultInput) (*territory.TerritoryObjectDefault, error) {
	if m.setFn != nil {
		return m.setFn(ctx, input)
	}
	return &territory.TerritoryObjectDefault{ID: uuid.New()}, nil
}

func (m *mockObjDefaultService) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]territory.TerritoryObjectDefault, error) {
	if m.listFn != nil {
		return m.listFn(ctx, territoryID)
	}
	return []territory.TerritoryObjectDefault{}, nil
}

func (m *mockObjDefaultService) Remove(ctx context.Context, territoryID, objectID uuid.UUID) error {
	if m.removeFn != nil {
		return m.removeFn(ctx, territoryID, objectID)
	}
	return nil
}

type mockUserAssignService struct {
	assignFn     func(ctx context.Context, input territory.AssignUserInput) (*territory.UserTerritoryAssignment, error)
	unassignFn   func(ctx context.Context, userID, territoryID uuid.UUID) error
	listByTerrFn func(ctx context.Context, territoryID uuid.UUID) ([]territory.UserTerritoryAssignment, error)
	listByUserFn func(ctx context.Context, userID uuid.UUID) ([]territory.UserTerritoryAssignment, error)
}

func (m *mockUserAssignService) Assign(ctx context.Context, input territory.AssignUserInput) (*territory.UserTerritoryAssignment, error) {
	if m.assignFn != nil {
		return m.assignFn(ctx, input)
	}
	return &territory.UserTerritoryAssignment{ID: uuid.New()}, nil
}

func (m *mockUserAssignService) Unassign(ctx context.Context, userID, territoryID uuid.UUID) error {
	if m.unassignFn != nil {
		return m.unassignFn(ctx, userID, territoryID)
	}
	return nil
}

func (m *mockUserAssignService) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]territory.UserTerritoryAssignment, error) {
	if m.listByTerrFn != nil {
		return m.listByTerrFn(ctx, territoryID)
	}
	return []territory.UserTerritoryAssignment{}, nil
}

func (m *mockUserAssignService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]territory.UserTerritoryAssignment, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID)
	}
	return []territory.UserTerritoryAssignment{}, nil
}

type mockRecordAssignService struct {
	assignFn     func(ctx context.Context, input territory.AssignRecordInput) (*territory.RecordTerritoryAssignment, error)
	unassignFn   func(ctx context.Context, recordID, objectID, territoryID uuid.UUID) error
	listByTerrFn func(ctx context.Context, territoryID uuid.UUID) ([]territory.RecordTerritoryAssignment, error)
	listByRecFn  func(ctx context.Context, recordID, objectID uuid.UUID) ([]territory.RecordTerritoryAssignment, error)
}

func (m *mockRecordAssignService) Assign(ctx context.Context, input territory.AssignRecordInput) (*territory.RecordTerritoryAssignment, error) {
	if m.assignFn != nil {
		return m.assignFn(ctx, input)
	}
	return &territory.RecordTerritoryAssignment{ID: uuid.New()}, nil
}

func (m *mockRecordAssignService) Unassign(ctx context.Context, recordID, objectID, territoryID uuid.UUID) error {
	if m.unassignFn != nil {
		return m.unassignFn(ctx, recordID, objectID, territoryID)
	}
	return nil
}

func (m *mockRecordAssignService) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]territory.RecordTerritoryAssignment, error) {
	if m.listByTerrFn != nil {
		return m.listByTerrFn(ctx, territoryID)
	}
	return []territory.RecordTerritoryAssignment{}, nil
}

func (m *mockRecordAssignService) ListByRecordID(ctx context.Context, recordID, objectID uuid.UUID) ([]territory.RecordTerritoryAssignment, error) {
	if m.listByRecFn != nil {
		return m.listByRecFn(ctx, recordID, objectID)
	}
	return []territory.RecordTerritoryAssignment{}, nil
}

type mockRuleService struct {
	createFn     func(ctx context.Context, input territory.CreateAssignmentRuleInput) (*territory.AssignmentRule, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*territory.AssignmentRule, error)
	listByTerrFn func(ctx context.Context, territoryID uuid.UUID) ([]territory.AssignmentRule, error)
	updateFn     func(ctx context.Context, id uuid.UUID, input territory.UpdateAssignmentRuleInput) (*territory.AssignmentRule, error)
	deleteFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRuleService) Create(ctx context.Context, input territory.CreateAssignmentRuleInput) (*territory.AssignmentRule, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &territory.AssignmentRule{ID: uuid.New()}, nil
}

func (m *mockRuleService) GetByID(ctx context.Context, id uuid.UUID) (*territory.AssignmentRule, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("AssignmentRule", id.String()))
}

func (m *mockRuleService) ListByTerritoryID(ctx context.Context, territoryID uuid.UUID) ([]territory.AssignmentRule, error) {
	if m.listByTerrFn != nil {
		return m.listByTerrFn(ctx, territoryID)
	}
	return []territory.AssignmentRule{}, nil
}

func (m *mockRuleService) Update(ctx context.Context, id uuid.UUID, input territory.UpdateAssignmentRuleInput) (*territory.AssignmentRule, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("AssignmentRule", id.String()))
}

func (m *mockRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

// --- Helpers ---

func setupTerritoryRouter(h *TerritoryHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/api/v1/admin")
	h.RegisterRoutes(admin)
	return r
}

func newTestHandler(
	models *mockModelService,
	territories *mockTerritoryService,
	objDefaults *mockObjDefaultService,
	userAssign *mockUserAssignService,
	recAssign *mockRecordAssignService,
	rules *mockRuleService,
) *TerritoryHandler {
	if models == nil {
		models = &mockModelService{}
	}
	if territories == nil {
		territories = &mockTerritoryService{}
	}
	if objDefaults == nil {
		objDefaults = &mockObjDefaultService{}
	}
	if userAssign == nil {
		userAssign = &mockUserAssignService{}
	}
	if recAssign == nil {
		recAssign = &mockRecordAssignService{}
	}
	if rules == nil {
		rules = &mockRuleService{}
	}
	return NewTerritoryHandler(models, territories, objDefaults, userAssign, recAssign, rules)
}

// --- Tests: Models ---

func TestTerritoryHandler_CreateModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setup      func(*mockModelService)
		wantStatus int
	}{
		{
			name:       "creates model successfully",
			body:       territory.CreateModelInput{APIName: "q1_2026", Label: "Q1 2026"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "bad json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns error from service",
			body: territory.CreateModelInput{APIName: "bad", Label: "Bad"},
			setup: func(m *mockModelService) {
				m.createFn = func(_ context.Context, _ territory.CreateModelInput) (*territory.TerritoryModel, error) {
					return nil, fmt.Errorf("%w", apperror.Validation("validation failed"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			models := &mockModelService{}
			if tt.setup != nil {
				tt.setup(models)
			}
			h := newTestHandler(models, nil, nil, nil, nil, nil)
			r := setupTerritoryRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/territory/models", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTerritoryHandler_GetModel(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         string
		setup      func(*mockModelService)
		wantStatus int
	}{
		{
			name: "returns model",
			id:   existingID.String(),
			setup: func(m *mockModelService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*territory.TerritoryModel, error) {
					return &territory.TerritoryModel{ID: existingID, APIName: "q1"}, nil
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
			models := &mockModelService{}
			if tt.setup != nil {
				tt.setup(models)
			}
			h := newTestHandler(models, nil, nil, nil, nil, nil)
			r := setupTerritoryRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/territory/models/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTerritoryHandler_ListModels(t *testing.T) {
	t.Parallel()

	models := &mockModelService{
		listFn: func(_ context.Context, _, _ int32) ([]territory.TerritoryModel, int64, error) {
			return []territory.TerritoryModel{{ID: uuid.New(), APIName: "q1"}}, 1, nil
		},
	}
	h := newTestHandler(models, nil, nil, nil, nil, nil)
	r := setupTerritoryRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/territory/models?page=1&per_page=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTerritoryHandler_DeleteModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setup      func(*mockModelService)
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
			setup: func(m *mockModelService) {
				m.deleteFn = func(_ context.Context, id uuid.UUID) error {
					return fmt.Errorf("%w", apperror.NotFound("TerritoryModel", id.String()))
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			models := &mockModelService{}
			if tt.setup != nil {
				tt.setup(models)
			}
			h := newTestHandler(models, nil, nil, nil, nil, nil)
			r := setupTerritoryRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/territory/models/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTerritoryHandler_ActivateModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setup      func(*mockModelService)
		wantStatus int
	}{
		{
			name:       "activates successfully",
			id:         uuid.New().String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name: "returns 400 on validation error",
			id:   uuid.New().String(),
			setup: func(m *mockModelService) {
				m.activateFn = func(_ context.Context, _ uuid.UUID) error {
					return fmt.Errorf("%w", apperror.Validation("only planning models can be activated"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			models := &mockModelService{}
			if tt.setup != nil {
				tt.setup(models)
			}
			h := newTestHandler(models, nil, nil, nil, nil, nil)
			r := setupTerritoryRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/territory/models/"+tt.id+"/activate", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- Tests: Territories ---

func TestTerritoryHandler_CreateTerritory(t *testing.T) {
	t.Parallel()

	modelID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "creates territory successfully",
			body:       territory.CreateTerritoryInput{ModelID: modelID, APIName: "north", Label: "North"},
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
			h := newTestHandler(nil, &mockTerritoryService{}, nil, nil, nil, nil)
			r := setupTerritoryRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/territory/territories", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTerritoryHandler_ListTerritories(t *testing.T) {
	t.Parallel()

	modelID := uuid.New()

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "returns territories for valid model_id",
			query:      "?model_id=" + modelID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 when model_id missing",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid model_id",
			query:      "?model_id=not-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newTestHandler(nil, &mockTerritoryService{}, nil, nil, nil, nil)
			r := setupTerritoryRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/territory/territories"+tt.query, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- Tests: User Assignments ---

func TestTerritoryHandler_AssignUser(t *testing.T) {
	t.Parallel()

	territoryID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "assigns user successfully",
			body:       assignUserRequest{UserID: userID},
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
			h := newTestHandler(nil, nil, nil, &mockUserAssignService{}, nil, nil)
			r := setupTerritoryRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost,
				"/api/v1/admin/territory/territories/"+territoryID.String()+"/users",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- Tests: Assignment Rules ---

func TestTerritoryHandler_CreateAssignmentRule(t *testing.T) {
	t.Parallel()

	territoryID := uuid.New()
	objectID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name: "creates rule successfully",
			body: territory.CreateAssignmentRuleInput{
				TerritoryID: territoryID, ObjectID: objectID,
				IsActive: true, CriteriaField: "region", CriteriaOp: "eq", CriteriaValue: "US",
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
			h := newTestHandler(nil, nil, nil, nil, nil, &mockRuleService{})
			r := setupTerritoryRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/territory/assignment-rules", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTerritoryHandler_ListAssignmentRules(t *testing.T) {
	t.Parallel()

	territoryID := uuid.New()

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "returns rules for valid territory_id",
			query:      "?territory_id=" + territoryID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 when territory_id missing",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid territory_id",
			query:      "?territory_id=not-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newTestHandler(nil, nil, nil, nil, nil, &mockRuleService{})
			r := setupTerritoryRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/territory/assignment-rules"+tt.query, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- Tests: Object Defaults ---

func TestTerritoryHandler_SetObjectDefault(t *testing.T) {
	t.Parallel()

	territoryID := uuid.New()
	objectID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "sets object default successfully",
			body:       setObjectDefaultRequest{ObjectID: objectID, AccessLevel: "read"},
			wantStatus: http.StatusOK,
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
			h := newTestHandler(nil, nil, &mockObjDefaultService{}, nil, nil, nil)
			r := setupTerritoryRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost,
				"/api/v1/admin/territory/territories/"+territoryID.String()+"/object-defaults",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
