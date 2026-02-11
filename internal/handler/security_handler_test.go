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

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

// --- Mocks ---

type mockRoleService struct {
	createFn  func(ctx context.Context, input security.CreateUserRoleInput) (*security.UserRole, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*security.UserRole, error)
	listFn    func(ctx context.Context, page, perPage int32) ([]security.UserRole, int64, error)
	updateFn  func(ctx context.Context, id uuid.UUID, input security.UpdateUserRoleInput) (*security.UserRole, error)
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRoleService) Create(ctx context.Context, input security.CreateUserRoleInput) (*security.UserRole, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &security.UserRole{ID: uuid.New(), Label: input.Label}, nil
}

func (m *mockRoleService) GetByID(ctx context.Context, id uuid.UUID) (*security.UserRole, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("UserRole", id.String()))
}

func (m *mockRoleService) List(ctx context.Context, page, perPage int32) ([]security.UserRole, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return []security.UserRole{}, 0, nil
}

func (m *mockRoleService) Update(ctx context.Context, id uuid.UUID, input security.UpdateUserRoleInput) (*security.UserRole, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("UserRole", id.String()))
}

func (m *mockRoleService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockPSService struct {
	createFn  func(ctx context.Context, input security.CreatePermissionSetInput) (*security.PermissionSet, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*security.PermissionSet, error)
	listFn    func(ctx context.Context, page, perPage int32) ([]security.PermissionSet, int64, error)
	updateFn  func(ctx context.Context, id uuid.UUID, input security.UpdatePermissionSetInput) (*security.PermissionSet, error)
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPSService) Create(ctx context.Context, input security.CreatePermissionSetInput) (*security.PermissionSet, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &security.PermissionSet{ID: uuid.New(), Label: input.Label}, nil
}

func (m *mockPSService) GetByID(ctx context.Context, id uuid.UUID) (*security.PermissionSet, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("PermissionSet", id.String()))
}

func (m *mockPSService) List(ctx context.Context, page, perPage int32) ([]security.PermissionSet, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return []security.PermissionSet{}, 0, nil
}

func (m *mockPSService) Update(ctx context.Context, id uuid.UUID, input security.UpdatePermissionSetInput) (*security.PermissionSet, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("PermissionSet", id.String()))
}

func (m *mockPSService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockProfileService struct {
	createFn  func(ctx context.Context, input security.CreateProfileInput) (*security.Profile, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*security.Profile, error)
	listFn    func(ctx context.Context, page, perPage int32) ([]security.Profile, int64, error)
	updateFn  func(ctx context.Context, id uuid.UUID, input security.UpdateProfileInput) (*security.Profile, error)
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockProfileService) Create(ctx context.Context, input security.CreateProfileInput) (*security.Profile, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &security.Profile{ID: uuid.New(), Label: input.Label}, nil
}

func (m *mockProfileService) GetByID(ctx context.Context, id uuid.UUID) (*security.Profile, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("Profile", id.String()))
}

func (m *mockProfileService) List(ctx context.Context, page, perPage int32) ([]security.Profile, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return []security.Profile{}, 0, nil
}

func (m *mockProfileService) Update(ctx context.Context, id uuid.UUID, input security.UpdateProfileInput) (*security.Profile, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("Profile", id.String()))
}

func (m *mockProfileService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockUserService struct {
	createFn   func(ctx context.Context, input security.CreateUserInput) (*security.User, error)
	getByIDFn  func(ctx context.Context, id uuid.UUID) (*security.User, error)
	listFn     func(ctx context.Context, page, perPage int32) ([]security.User, int64, error)
	updateFn   func(ctx context.Context, id uuid.UUID, input security.UpdateUserInput) (*security.User, error)
	deleteFn   func(ctx context.Context, id uuid.UUID) error
	assignPSFn func(ctx context.Context, userID, psID uuid.UUID) error
	revokePSFn func(ctx context.Context, userID, psID uuid.UUID) error
	listPSFn   func(ctx context.Context, userID uuid.UUID) ([]security.PermissionSetToUser, error)
}

func (m *mockUserService) Create(ctx context.Context, input security.CreateUserInput) (*security.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &security.User{ID: uuid.New(), Username: input.Username, Email: input.Email}, nil
}

func (m *mockUserService) GetByID(ctx context.Context, id uuid.UUID) (*security.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("User", id.String()))
}

func (m *mockUserService) List(ctx context.Context, page, perPage int32) ([]security.User, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return []security.User{}, 0, nil
}

func (m *mockUserService) Update(ctx context.Context, id uuid.UUID, input security.UpdateUserInput) (*security.User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("User", id.String()))
}

func (m *mockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockUserService) AssignPermissionSet(ctx context.Context, userID, psID uuid.UUID) error {
	if m.assignPSFn != nil {
		return m.assignPSFn(ctx, userID, psID)
	}
	return nil
}

func (m *mockUserService) RevokePermissionSet(ctx context.Context, userID, psID uuid.UUID) error {
	if m.revokePSFn != nil {
		return m.revokePSFn(ctx, userID, psID)
	}
	return nil
}

func (m *mockUserService) ListPermissionSets(ctx context.Context, userID uuid.UUID) ([]security.PermissionSetToUser, error) {
	if m.listPSFn != nil {
		return m.listPSFn(ctx, userID)
	}
	return []security.PermissionSetToUser{}, nil
}

type mockPermissionService struct {
	setObjPermFn    func(ctx context.Context, psID uuid.UUID, input security.SetObjectPermissionInput) (*security.ObjectPermission, error)
	listObjPermsFn  func(ctx context.Context, psID uuid.UUID) ([]security.ObjectPermission, error)
	removeObjPermFn func(ctx context.Context, psID, objectID uuid.UUID) error
	setFldPermFn    func(ctx context.Context, psID uuid.UUID, input security.SetFieldPermissionInput) (*security.FieldPermission, error)
	listFldPermsFn  func(ctx context.Context, psID uuid.UUID) ([]security.FieldPermission, error)
	removeFldPermFn func(ctx context.Context, psID, fieldID uuid.UUID) error
}

func (m *mockPermissionService) SetObjectPermission(ctx context.Context, psID uuid.UUID, input security.SetObjectPermissionInput) (*security.ObjectPermission, error) {
	if m.setObjPermFn != nil {
		return m.setObjPermFn(ctx, psID, input)
	}
	return &security.ObjectPermission{}, nil
}

func (m *mockPermissionService) ListObjectPermissions(ctx context.Context, psID uuid.UUID) ([]security.ObjectPermission, error) {
	if m.listObjPermsFn != nil {
		return m.listObjPermsFn(ctx, psID)
	}
	return []security.ObjectPermission{}, nil
}

func (m *mockPermissionService) RemoveObjectPermission(ctx context.Context, psID, objectID uuid.UUID) error {
	if m.removeObjPermFn != nil {
		return m.removeObjPermFn(ctx, psID, objectID)
	}
	return nil
}

func (m *mockPermissionService) SetFieldPermission(ctx context.Context, psID uuid.UUID, input security.SetFieldPermissionInput) (*security.FieldPermission, error) {
	if m.setFldPermFn != nil {
		return m.setFldPermFn(ctx, psID, input)
	}
	return &security.FieldPermission{}, nil
}

func (m *mockPermissionService) ListFieldPermissions(ctx context.Context, psID uuid.UUID) ([]security.FieldPermission, error) {
	if m.listFldPermsFn != nil {
		return m.listFldPermsFn(ctx, psID)
	}
	return []security.FieldPermission{}, nil
}

func (m *mockPermissionService) RemoveFieldPermission(ctx context.Context, psID, fieldID uuid.UUID) error {
	if m.removeFldPermFn != nil {
		return m.removeFldPermFn(ctx, psID, fieldID)
	}
	return nil
}

type mockGroupService struct {
	createFn       func(ctx context.Context, input security.CreateGroupInput) (*security.Group, error)
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*security.Group, error)
	listFn         func(ctx context.Context, page, perPage int32) ([]security.Group, int64, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
	addMemberFn    func(ctx context.Context, input security.AddGroupMemberInput) (*security.GroupMember, error)
	removeMemberFn func(ctx context.Context, groupID uuid.UUID, memberUserID *uuid.UUID, memberGroupID *uuid.UUID) error
	listMembersFn  func(ctx context.Context, groupID uuid.UUID) ([]security.GroupMember, error)
}

func (m *mockGroupService) Create(ctx context.Context, input security.CreateGroupInput) (*security.Group, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &security.Group{ID: uuid.New(), Label: input.Label}, nil
}

func (m *mockGroupService) GetByID(ctx context.Context, id uuid.UUID) (*security.Group, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("Group", id.String()))
}

func (m *mockGroupService) List(ctx context.Context, page, perPage int32) ([]security.Group, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return []security.Group{}, 0, nil
}

func (m *mockGroupService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockGroupService) AddMember(ctx context.Context, input security.AddGroupMemberInput) (*security.GroupMember, error) {
	if m.addMemberFn != nil {
		return m.addMemberFn(ctx, input)
	}
	return &security.GroupMember{ID: uuid.New()}, nil
}

func (m *mockGroupService) RemoveMember(ctx context.Context, groupID uuid.UUID, memberUserID *uuid.UUID, memberGroupID *uuid.UUID) error {
	if m.removeMemberFn != nil {
		return m.removeMemberFn(ctx, groupID, memberUserID, memberGroupID)
	}
	return nil
}

func (m *mockGroupService) ListMembers(ctx context.Context, groupID uuid.UUID) ([]security.GroupMember, error) {
	if m.listMembersFn != nil {
		return m.listMembersFn(ctx, groupID)
	}
	return []security.GroupMember{}, nil
}

type mockSharingRuleService struct {
	createFn       func(ctx context.Context, input security.CreateSharingRuleInput) (*security.SharingRule, error)
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*security.SharingRule, error)
	listByObjectFn func(ctx context.Context, objectID uuid.UUID) ([]security.SharingRule, error)
	updateFn       func(ctx context.Context, id uuid.UUID, input security.UpdateSharingRuleInput) (*security.SharingRule, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockSharingRuleService) Create(ctx context.Context, input security.CreateSharingRuleInput) (*security.SharingRule, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &security.SharingRule{ID: uuid.New()}, nil
}

func (m *mockSharingRuleService) GetByID(ctx context.Context, id uuid.UUID) (*security.SharingRule, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("SharingRule", id.String()))
}

func (m *mockSharingRuleService) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]security.SharingRule, error) {
	if m.listByObjectFn != nil {
		return m.listByObjectFn(ctx, objectID)
	}
	return []security.SharingRule{}, nil
}

func (m *mockSharingRuleService) Update(ctx context.Context, id uuid.UUID, input security.UpdateSharingRuleInput) (*security.SharingRule, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, fmt.Errorf("%w", apperror.NotFound("SharingRule", id.String()))
}

func (m *mockSharingRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

// --- Helpers ---

func setupSecurityRouter(h *SecurityHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/api/v1/admin")
	h.RegisterRoutes(admin)
	return r
}

func newSecurityHandler(
	roles *mockRoleService,
	ps *mockPSService,
	profiles *mockProfileService,
	users *mockUserService,
	perms *mockPermissionService,
	groups *mockGroupService,
	sharing *mockSharingRuleService,
) *SecurityHandler {
	if roles == nil {
		roles = &mockRoleService{}
	}
	if ps == nil {
		ps = &mockPSService{}
	}
	if profiles == nil {
		profiles = &mockProfileService{}
	}
	if users == nil {
		users = &mockUserService{}
	}
	if perms == nil {
		perms = &mockPermissionService{}
	}
	if groups == nil {
		groups = &mockGroupService{}
	}
	if sharing == nil {
		sharing = &mockSharingRuleService{}
	}
	return NewSecurityHandler(roles, ps, profiles, users, perms, groups, sharing)
}

// --- Tests ---

func TestSecurityHandler_CreateRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setup      func(*mockRoleService)
		wantStatus int
	}{
		{
			name:       "creates role successfully",
			body:       security.CreateUserRoleInput{Label: "Sales Rep", APIName: "sales_rep"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "bad json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns error from service",
			body: security.CreateUserRoleInput{Label: "Bad", APIName: "bad"},
			setup: func(m *mockRoleService) {
				m.createFn = func(_ context.Context, _ security.CreateUserRoleInput) (*security.UserRole, error) {
					return nil, fmt.Errorf("%w", apperror.Validation("validation failed"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			roles := &mockRoleService{}
			if tt.setup != nil {
				tt.setup(roles)
			}
			h := newSecurityHandler(roles, nil, nil, nil, nil, nil, nil)
			r := setupSecurityRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/security/roles", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_GetRole(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         string
		setup      func(*mockRoleService)
		wantStatus int
	}{
		{
			name: "returns role",
			id:   existingID.String(),
			setup: func(m *mockRoleService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*security.UserRole, error) {
					return &security.UserRole{ID: existingID, Label: "Sales Rep"}, nil
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
			roles := &mockRoleService{}
			if tt.setup != nil {
				tt.setup(roles)
			}
			h := newSecurityHandler(roles, nil, nil, nil, nil, nil, nil)
			r := setupSecurityRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/security/roles/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_ListRoles(t *testing.T) {
	t.Parallel()

	roles := &mockRoleService{
		listFn: func(_ context.Context, _, _ int32) ([]security.UserRole, int64, error) {
			return []security.UserRole{{ID: uuid.New(), Label: "Admin"}}, 1, nil
		},
	}
	h := newSecurityHandler(roles, nil, nil, nil, nil, nil, nil)
	r := setupSecurityRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/security/roles?page=1&per_page=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestSecurityHandler_DeleteRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setup      func(*mockRoleService)
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
			setup: func(m *mockRoleService) {
				m.deleteFn = func(_ context.Context, id uuid.UUID) error {
					return fmt.Errorf("%w", apperror.NotFound("UserRole", id.String()))
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			roles := &mockRoleService{}
			if tt.setup != nil {
				tt.setup(roles)
			}
			h := newSecurityHandler(roles, nil, nil, nil, nil, nil, nil)
			r := setupSecurityRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/security/roles/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_CreateUser(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()
	roleID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		setup      func(*mockUserService)
		wantStatus int
	}{
		{
			name: "creates user successfully",
			body: security.CreateUserInput{
				Username:  "jdoe",
				Email:     "jdoe@example.com",
				FirstName: "John",
				LastName:  "Doe",
				ProfileID: profileID,
				RoleID:    &roleID,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns error on validation failure",
			body: security.CreateUserInput{Username: "", Email: "bad"},
			setup: func(m *mockUserService) {
				m.createFn = func(_ context.Context, _ security.CreateUserInput) (*security.User, error) {
					return nil, fmt.Errorf("%w", apperror.Validation("invalid input"))
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			users := &mockUserService{}
			if tt.setup != nil {
				tt.setup(users)
			}
			h := newSecurityHandler(nil, nil, nil, users, nil, nil, nil)
			r := setupSecurityRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/security/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_GetUser(t *testing.T) {
	t.Parallel()

	existingID := uuid.New()

	tests := []struct {
		name       string
		id         string
		setup      func(*mockUserService)
		wantStatus int
	}{
		{
			name: "returns user",
			id:   existingID.String(),
			setup: func(m *mockUserService) {
				m.getByIDFn = func(_ context.Context, _ uuid.UUID) (*security.User, error) {
					return &security.User{ID: existingID, Username: "admin"}, nil
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
			users := &mockUserService{}
			if tt.setup != nil {
				tt.setup(users)
			}
			h := newSecurityHandler(nil, nil, nil, users, nil, nil, nil)
			r := setupSecurityRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/security/users/"+tt.id, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_AssignPermissionSet(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	psID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		setup      func(*mockUserService)
		wantStatus int
	}{
		{
			name:       "assigns PS successfully",
			body:       assignPSRequest{PermissionSetID: psID},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns error from service",
			body: assignPSRequest{PermissionSetID: psID},
			setup: func(m *mockUserService) {
				m.assignPSFn = func(_ context.Context, _, _ uuid.UUID) error {
					return fmt.Errorf("%w", apperror.NotFound("User", userID.String()))
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			users := &mockUserService{}
			if tt.setup != nil {
				tt.setup(users)
			}
			h := newSecurityHandler(nil, nil, nil, users, nil, nil, nil)
			r := setupSecurityRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost,
				"/api/v1/admin/security/users/"+userID.String()+"/permission-sets",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_CreatePermissionSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setup      func(*mockPSService)
		wantStatus int
	}{
		{
			name:       "creates PS successfully",
			body:       security.CreatePermissionSetInput{Label: "Read Only", PSType: "grant"},
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
			ps := &mockPSService{}
			if tt.setup != nil {
				tt.setup(ps)
			}
			h := newSecurityHandler(nil, ps, nil, nil, nil, nil, nil)
			r := setupSecurityRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/security/permission-sets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_SetObjectPermission(t *testing.T) {
	t.Parallel()

	psID := uuid.New()
	objID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		setup      func(*mockPermissionService)
		wantStatus int
	}{
		{
			name:       "sets permission successfully",
			body:       security.SetObjectPermissionInput{ObjectID: objID, Permissions: 5},
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
			perms := &mockPermissionService{}
			if tt.setup != nil {
				tt.setup(perms)
			}
			h := newSecurityHandler(nil, nil, nil, nil, perms, nil, nil)
			r := setupSecurityRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut,
				"/api/v1/admin/security/permission-sets/"+psID.String()+"/object-permissions",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_CreateGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		setup      func(*mockGroupService)
		wantStatus int
	}{
		{
			name:       "creates group successfully",
			body:       security.CreateGroupInput{Label: "Engineering", GroupType: "public"},
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
			groups := &mockGroupService{}
			if tt.setup != nil {
				tt.setup(groups)
			}
			h := newSecurityHandler(nil, nil, nil, nil, nil, groups, nil)
			r := setupSecurityRouter(h)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/security/groups", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestSecurityHandler_ListSharingRules(t *testing.T) {
	t.Parallel()

	objectID := uuid.New()

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "returns rules for valid object_id",
			query:      "?object_id=" + objectID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 when object_id missing",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid object_id",
			query:      "?object_id=not-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newSecurityHandler(nil, nil, nil, nil, nil, nil, &mockSharingRuleService{})
			r := setupSecurityRouter(h)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/security/sharing-rules"+tt.query, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
