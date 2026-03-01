package metadata

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock PortalRepository ---

type mockPortalRepository struct {
	views     map[uuid.UUID]*Portal
	byAPIName map[string]*Portal
}

func newMockOVRepo() *mockPortalRepository {
	return &mockPortalRepository{
		views:     make(map[uuid.UUID]*Portal),
		byAPIName: make(map[string]*Portal),
	}
}

func (m *mockPortalRepository) Create(_ context.Context, input CreatePortalInput) (*Portal, error) {
	if _, exists := m.byAPIName[input.APIName]; exists {
		return nil, &duplicateError{apiName: input.APIName}
	}

	now := time.Now()
	ov := &Portal{
		ID:          uuid.New(),
		ProfileID:   input.ProfileID,
		APIName:     input.APIName,
		Label:       input.Label,
		Description: input.Description,
		Config:      input.Config,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.views[ov.ID] = ov
	m.byAPIName[input.APIName] = ov
	return ov, nil
}

func (m *mockPortalRepository) GetByID(_ context.Context, id uuid.UUID) (*Portal, error) {
	return m.views[id], nil
}

func (m *mockPortalRepository) GetByAPIName(_ context.Context, apiName string) (*Portal, error) {
	return m.byAPIName[apiName], nil
}

func (m *mockPortalRepository) ListAll(_ context.Context) ([]Portal, error) {
	result := make([]Portal, 0, len(m.views))
	for _, ov := range m.views {
		result = append(result, *ov)
	}
	return result, nil
}

func (m *mockPortalRepository) Update(_ context.Context, id uuid.UUID, input UpdatePortalInput) (*Portal, error) {
	ov := m.views[id]
	if ov == nil {
		return nil, nil
	}
	ov.Label = input.Label
	ov.Description = input.Description
	ov.Config = input.Config
	ov.UpdatedAt = time.Now()
	return ov, nil
}

func (m *mockPortalRepository) Delete(_ context.Context, id uuid.UUID) error {
	ov := m.views[id]
	if ov != nil {
		delete(m.byAPIName, ov.APIName)
		delete(m.views, id)
	}
	return nil
}

// duplicateError is used by the mock to simulate DB unique constraint violation.
type duplicateError struct {
	apiName string
}

func (e *duplicateError) Error() string {
	return "duplicate api_name: " + e.apiName
}

// --- Mock CacheLoader ---

type mockOVCacheLoader struct {
	objects []ObjectDefinition
	portals []Portal
}

func (m *mockOVCacheLoader) LoadAllObjects(_ context.Context) ([]ObjectDefinition, error) {
	return m.objects, nil
}

func (m *mockOVCacheLoader) LoadAllFields(_ context.Context) ([]FieldDefinition, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) LoadRelationships(_ context.Context) ([]RelationshipInfo, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) LoadAllValidationRules(_ context.Context) ([]ValidationRule, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) LoadAllFunctions(_ context.Context) ([]Function, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) LoadAllPortals(_ context.Context) ([]Portal, error) {
	return m.portals, nil
}

func (m *mockOVCacheLoader) LoadAllProcedures(_ context.Context) ([]Procedure, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) LoadAllAutomationRules(_ context.Context) ([]AutomationRule, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) LoadAllLayouts(_ context.Context) ([]Layout, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) LoadAllSharedLayouts(_ context.Context) ([]SharedLayout, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

// --- Test helpers ---

func setupOVServiceTest(t *testing.T) (PortalService, *mockPortalRepository, *mockOVCacheLoader) {
	t.Helper()

	repo := newMockOVRepo()
	loader := &mockOVCacheLoader{
		objects: []ObjectDefinition{
			{ID: uuid.New(), APIName: "contacts"},
		},
	}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	svc := NewPortalService(nil, repo, cache)
	return svc, repo, loader
}

func validCreateInput() CreatePortalInput {
	return CreatePortalInput{
		APIName: "default_view",
		Label:   "Default View",
		Config: PortalConfig{
			Read: PortalReadConfig{
				Fields: []PortalViewField{{Name: "first_name"}, {Name: "last_name"}},
			},
		},
	}
}

// --- Tests ---

func TestPortalService_Create(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()

	tests := []struct {
		name    string
		input   CreatePortalInput
		wantErr bool
		errMsg  string
	}{
		{
			name:  "creates successfully with valid input",
			input: validCreateInput(),
		},
		{
			name: "creates with profile_id",
			input: CreatePortalInput{
				ProfileID: &profileID,
				APIName:   "sales_view",
				Label:     "Sales View",
			},
		},
		{
			name: "rejects empty api_name",
			input: CreatePortalInput{
				APIName: "",
				Label:   "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects uppercase api_name",
			input: CreatePortalInput{
				APIName: "DefaultView",
				Label:   "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name starting with number",
			input: CreatePortalInput{
				APIName: "1view",
				Label:   "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name with special characters",
			input: CreatePortalInput{
				APIName: "my-view",
				Label:   "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name longer than 100 characters",
			input: CreatePortalInput{
				APIName: strings.Repeat("a", 101),
				Label:   "Test",
			},
			wantErr: true,
			errMsg:  "api_name must be at most 100 characters",
		},
		{
			name: "rejects empty label",
			input: CreatePortalInput{
				APIName: "my_view",
				Label:   "",
			},
			wantErr: true,
			errMsg:  "label is required",
		},
		{
			name: "rejects label longer than 255 characters",
			input: CreatePortalInput{
				APIName: "my_view",
				Label:   strings.Repeat("x", 256),
			},
			wantErr: true,
			errMsg:  "label must be at most 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)

			ov, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, ov)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ov)
			assert.NotEqual(t, uuid.Nil, ov.ID)
			assert.Equal(t, tt.input.APIName, ov.APIName)
			assert.Equal(t, tt.input.Label, ov.Label)
		})
	}
}

func TestPortalService_Create_DuplicateAPIName(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)

	// Same api_name = duplicate
	_, err = svc.Create(ctx, CreatePortalInput{
		APIName: "default_view",
		Label:   "Another Label",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestPortalService_Create_CacheRefresh(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	// Create triggers cache.LoadPortals
	ov, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)
	require.NotNil(t, ov)
}

func TestPortalService_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(svc PortalService) uuid.UUID
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing view",
			setup: func(svc PortalService) uuid.UUID {
				ov, _ := svc.Create(context.Background(), validCreateInput())
				return ov.ID
			},
		},
		{
			name: "returns not found for nonexistent id",
			setup: func(_ PortalService) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)
			id := tt.setup(svc)

			ov, err := svc.GetByID(context.Background(), id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ov)
			assert.Equal(t, id, ov.ID)
		})
	}
}

func TestPortalService_GetByAPIName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(svc PortalService)
		apiName string
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing view by api_name",
			setup: func(svc PortalService) {
				_, _ = svc.Create(context.Background(), validCreateInput())
			},
			apiName: "default_view",
		},
		{
			name:    "returns not found for nonexistent api_name",
			setup:   func(_ PortalService) {},
			apiName: "nonexistent",
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)
			tt.setup(svc)

			ov, err := svc.GetByAPIName(context.Background(), tt.apiName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ov)
			assert.Equal(t, tt.apiName, ov.APIName)
		})
	}
}

func TestPortalService_ListAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(svc PortalService)
		wantCount int
	}{
		{
			name:      "returns empty list when no views exist",
			setup:     func(_ PortalService) {},
			wantCount: 0,
		},
		{
			name: "returns all views",
			setup: func(svc PortalService) {
				ctx := context.Background()
				_, _ = svc.Create(ctx, CreatePortalInput{
					APIName: "view_a", Label: "View A",
				})
				_, _ = svc.Create(ctx, CreatePortalInput{
					APIName: "view_b", Label: "View B",
				})
				_, _ = svc.Create(ctx, CreatePortalInput{
					APIName: "view_c", Label: "View C",
				})
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)
			tt.setup(svc)

			views, err := svc.ListAll(context.Background())
			require.NoError(t, err)
			assert.Len(t, views, tt.wantCount)
		})
	}
}

func TestPortalService_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   UpdatePortalInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "updates label successfully",
			input: UpdatePortalInput{
				Label:       "Updated Label",
				Description: "Updated desc",
			},
		},
		{
			name: "updates config successfully",
			input: UpdatePortalInput{
				Label: "With Config",
				Config: PortalConfig{
					Read: PortalReadConfig{
						Fields: []PortalViewField{{Name: "email"}},
					},
				},
			},
		},
		{
			name: "rejects empty label",
			input: UpdatePortalInput{
				Label: "",
			},
			wantErr: true,
			errMsg:  "label is required",
		},
		{
			name: "rejects label longer than 255 characters",
			input: UpdatePortalInput{
				Label: strings.Repeat("x", 256),
			},
			wantErr: true,
			errMsg:  "label must be at most 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)
			ctx := context.Background()

			created, err := svc.Create(ctx, validCreateInput())
			require.NoError(t, err)

			ov, err := svc.Update(ctx, created.ID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ov)
			assert.Equal(t, tt.input.Label, ov.Label)
			assert.Equal(t, tt.input.Description, ov.Description)
		})
	}
}

func TestPortalService_Update_NotFound(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)

	_, err := svc.Update(context.Background(), uuid.New(), UpdatePortalInput{
		Label: "Updated",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortalService_Delete(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	ov, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)

	err = svc.Delete(ctx, ov.ID)
	require.NoError(t, err)

	// Verify it no longer exists
	_, err = svc.GetByID(ctx, ov.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortalService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)

	err := svc.Delete(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortalService_Create_ValidAPINamePatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		apiName string
		wantErr bool
	}{
		{name: "simple lowercase", apiName: "view", wantErr: false},
		{name: "with underscores", apiName: "my_custom_view", wantErr: false},
		{name: "with numbers", apiName: "view2", wantErr: false},
		{name: "single char", apiName: "v", wantErr: false},
		{name: "starts with underscore", apiName: "_view", wantErr: true},
		{name: "contains dash", apiName: "my-view", wantErr: true},
		{name: "contains space", apiName: "my view", wantErr: true},
		{name: "contains dot", apiName: "my.view", wantErr: true},
		{name: "uppercase", apiName: "MyView", wantErr: true},
		{name: "empty", apiName: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)

			_, err := svc.Create(context.Background(), CreatePortalInput{
				APIName: tt.apiName,
				Label:   "Test Label",
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPortalService_ErrorWrapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operation  func(svc PortalService) error
		wantPrefix string
	}{
		{
			name: "Create wraps with method name",
			operation: func(svc PortalService) error {
				_, err := svc.Create(context.Background(), CreatePortalInput{
					APIName: "",
					Label:   "Test",
				})
				return err
			},
			wantPrefix: "portalService.Create:",
		},
		{
			name: "GetByID wraps with method name",
			operation: func(svc PortalService) error {
				_, err := svc.GetByID(context.Background(), uuid.New())
				return err
			},
			wantPrefix: "portalService.GetByID:",
		},
		{
			name: "GetByAPIName wraps with method name",
			operation: func(svc PortalService) error {
				_, err := svc.GetByAPIName(context.Background(), "nonexistent")
				return err
			},
			wantPrefix: "portalService.GetByAPIName:",
		},
		{
			name: "Update wraps with method name",
			operation: func(svc PortalService) error {
				_, err := svc.Update(context.Background(), uuid.New(), UpdatePortalInput{Label: "Test"})
				return err
			},
			wantPrefix: "portalService.Update:",
		},
		{
			name: "Delete wraps with method name",
			operation: func(svc PortalService) error {
				return svc.Delete(context.Background(), uuid.New())
			},
			wantPrefix: "portalService.Delete:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)

			err := tt.operation(svc)

			if tt.wantPrefix == "" {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantPrefix)
		})
	}
}

func TestPortalService_Update_PreservesAPIName(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, UpdatePortalInput{
		Label:       "New Label",
		Description: "New Desc",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	// API name should remain unchanged
	assert.Equal(t, created.APIName, updated.APIName)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "New Label", updated.Label)
}

func TestPortalService_Delete_ThenGetReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	ov, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)

	err = svc.Delete(ctx, ov.ID)
	require.NoError(t, err)

	result, err := svc.GetByID(ctx, ov.ID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortalService_ListAll_AfterCreateAndDelete(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	ov1, err := svc.Create(ctx, CreatePortalInput{
		APIName: "view_one", Label: "View One",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreatePortalInput{
		APIName: "view_two", Label: "View Two",
	})
	require.NoError(t, err)

	views, err := svc.ListAll(ctx)
	require.NoError(t, err)
	assert.Len(t, views, 2)

	err = svc.Delete(ctx, ov1.ID)
	require.NoError(t, err)

	views, err = svc.ListAll(ctx)
	require.NoError(t, err)
	assert.Len(t, views, 1)
	assert.Equal(t, "View Two", views[0].Label)
}
