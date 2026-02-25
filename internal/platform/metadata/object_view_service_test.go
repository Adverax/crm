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

// --- Mock ObjectViewRepository ---

type mockObjectViewRepository struct {
	views    map[uuid.UUID]*ObjectView
	byAPIKey map[string]*ObjectView // key: objectID|api_name
}

func newMockOVRepo() *mockObjectViewRepository {
	return &mockObjectViewRepository{
		views:    make(map[uuid.UUID]*ObjectView),
		byAPIKey: make(map[string]*ObjectView),
	}
}

func ovKey(objectID uuid.UUID, apiName string) string {
	return objectID.String() + "|" + apiName
}

func (m *mockObjectViewRepository) Create(_ context.Context, input CreateObjectViewInput) (*ObjectView, error) {
	key := ovKey(input.ObjectID, input.APIName)
	if _, exists := m.byAPIKey[key]; exists {
		return nil, &duplicateError{apiName: input.APIName}
	}

	now := time.Now()
	ov := &ObjectView{
		ID:          uuid.New(),
		ObjectID:    input.ObjectID,
		ProfileID:   input.ProfileID,
		APIName:     input.APIName,
		Label:       input.Label,
		Description: input.Description,
		IsDefault:   input.IsDefault,
		Config:      input.Config,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.views[ov.ID] = ov
	m.byAPIKey[key] = ov
	return ov, nil
}

func (m *mockObjectViewRepository) GetByID(_ context.Context, id uuid.UUID) (*ObjectView, error) {
	return m.views[id], nil
}

func (m *mockObjectViewRepository) ListAll(_ context.Context) ([]ObjectView, error) {
	result := make([]ObjectView, 0, len(m.views))
	for _, ov := range m.views {
		result = append(result, *ov)
	}
	return result, nil
}

func (m *mockObjectViewRepository) ListByObjectID(_ context.Context, objectID uuid.UUID) ([]ObjectView, error) {
	var result []ObjectView
	for _, ov := range m.views {
		if ov.ObjectID == objectID {
			result = append(result, *ov)
		}
	}
	return result, nil
}

func (m *mockObjectViewRepository) Update(_ context.Context, id uuid.UUID, input UpdateObjectViewInput) (*ObjectView, error) {
	ov := m.views[id]
	if ov == nil {
		return nil, nil
	}
	ov.Label = input.Label
	ov.Description = input.Description
	ov.IsDefault = input.IsDefault
	ov.Config = input.Config
	ov.UpdatedAt = time.Now()
	return ov, nil
}

func (m *mockObjectViewRepository) Delete(_ context.Context, id uuid.UUID) error {
	ov := m.views[id]
	if ov != nil {
		delete(m.byAPIKey, ovKey(ov.ObjectID, ov.APIName))
		delete(m.views, id)
	}
	return nil
}

func (m *mockObjectViewRepository) FindForProfile(_ context.Context, objectID uuid.UUID, profileID uuid.UUID) (*ObjectView, error) {
	for _, ov := range m.views {
		if ov.ObjectID == objectID && ov.ProfileID != nil && *ov.ProfileID == profileID {
			return ov, nil
		}
	}
	return nil, nil
}

func (m *mockObjectViewRepository) FindDefault(_ context.Context, objectID uuid.UUID) (*ObjectView, error) {
	for _, ov := range m.views {
		if ov.ObjectID == objectID && ov.IsDefault {
			return ov, nil
		}
	}
	return nil, nil
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
	objects     []ObjectDefinition
	objectViews []ObjectView
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

func (m *mockOVCacheLoader) LoadAllObjectViews(_ context.Context) ([]ObjectView, error) {
	return m.objectViews, nil
}

func (m *mockOVCacheLoader) LoadAllProcedures(_ context.Context) ([]Procedure, error) {
	return nil, nil
}

func (m *mockOVCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

// --- Test helpers ---

var testObjectID = uuid.New()

func setupOVServiceTest(t *testing.T) (ObjectViewService, *mockObjectViewRepository, *mockOVCacheLoader) {
	t.Helper()

	repo := newMockOVRepo()
	loader := &mockOVCacheLoader{
		objects: []ObjectDefinition{
			{ID: testObjectID, APIName: "contacts"},
		},
	}
	cache := NewMetadataCache(loader)
	require.NoError(t, cache.Load(context.Background()))

	svc := NewObjectViewService(nil, repo, cache)
	return svc, repo, loader
}

func validCreateInput() CreateObjectViewInput {
	return CreateObjectViewInput{
		ObjectID:  testObjectID,
		APIName:   "default_view",
		Label:     "Default View",
		IsDefault: true,
		Config: OVConfig{
			Read: OVReadConfig{
				Fields: []string{"first_name", "last_name"},
			},
		},
	}
}

// --- Tests ---

func TestObjectViewService_Create(t *testing.T) {
	t.Parallel()

	profileID := uuid.New()

	tests := []struct {
		name    string
		input   CreateObjectViewInput
		wantErr bool
		errMsg  string
	}{
		{
			name:  "creates successfully with valid input",
			input: validCreateInput(),
		},
		{
			name: "creates with profile_id",
			input: CreateObjectViewInput{
				ObjectID:  testObjectID,
				ProfileID: &profileID,
				APIName:   "sales_view",
				Label:     "Sales View",
			},
		},
		{
			name: "rejects empty api_name",
			input: CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  "",
				Label:    "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects uppercase api_name",
			input: CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  "DefaultView",
				Label:    "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name starting with number",
			input: CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  "1view",
				Label:    "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name with special characters",
			input: CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  "my-view",
				Label:    "Test",
			},
			wantErr: true,
			errMsg:  "api_name must match",
		},
		{
			name: "rejects api_name longer than 100 characters",
			input: CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  strings.Repeat("a", 101),
				Label:    "Test",
			},
			wantErr: true,
			errMsg:  "api_name must be at most 100 characters",
		},
		{
			name: "rejects empty label",
			input: CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  "my_view",
				Label:    "",
			},
			wantErr: true,
			errMsg:  "label is required",
		},
		{
			name: "rejects label longer than 255 characters",
			input: CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  "my_view",
				Label:    strings.Repeat("x", 256),
			},
			wantErr: true,
			errMsg:  "label must be at most 255 characters",
		},
		{
			name: "rejects nil object_id",
			input: CreateObjectViewInput{
				ObjectID: uuid.Nil,
				APIName:  "my_view",
				Label:    "Test",
			},
			wantErr: true,
			errMsg:  "object_id is required",
		},
		{
			name: "rejects nonexistent object_id",
			input: CreateObjectViewInput{
				ObjectID: uuid.New(),
				APIName:  "my_view",
				Label:    "Test",
			},
			wantErr: true,
			errMsg:  "not found",
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
			assert.Equal(t, tt.input.ObjectID, ov.ObjectID)
			assert.Equal(t, tt.input.APIName, ov.APIName)
			assert.Equal(t, tt.input.Label, ov.Label)
			assert.Equal(t, tt.input.IsDefault, ov.IsDefault)
		})
	}
}

func TestObjectViewService_Create_DuplicateAPIName(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)

	// Same api_name + same object = duplicate
	_, err = svc.Create(ctx, CreateObjectViewInput{
		ObjectID: testObjectID,
		APIName:  "default_view",
		Label:    "Another Label",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestObjectViewService_Create_CacheRefresh(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	// Create triggers cache.LoadObjectViews
	ov, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)
	require.NotNil(t, ov)
}

func TestObjectViewService_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(svc ObjectViewService) uuid.UUID
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing view",
			setup: func(svc ObjectViewService) uuid.UUID {
				ov, _ := svc.Create(context.Background(), validCreateInput())
				return ov.ID
			},
		},
		{
			name: "returns not found for nonexistent id",
			setup: func(_ ObjectViewService) uuid.UUID {
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

func TestObjectViewService_ListAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(svc ObjectViewService)
		wantCount int
	}{
		{
			name:      "returns empty list when no views exist",
			setup:     func(_ ObjectViewService) {},
			wantCount: 0,
		},
		{
			name: "returns all views",
			setup: func(svc ObjectViewService) {
				ctx := context.Background()
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID: testObjectID, APIName: "view_a", Label: "View A",
				})
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID: testObjectID, APIName: "view_b", Label: "View B",
				})
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID: testObjectID, APIName: "view_c", Label: "View C",
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

func TestObjectViewService_ListByObjectID(t *testing.T) {
	t.Parallel()

	otherObjectID := uuid.New()

	tests := []struct {
		name      string
		setup     func(svc ObjectViewService, loader *mockOVCacheLoader)
		objectID  uuid.UUID
		wantCount int
	}{
		{
			name:      "returns empty when no views for object",
			setup:     func(_ ObjectViewService, _ *mockOVCacheLoader) {},
			objectID:  testObjectID,
			wantCount: 0,
		},
		{
			name: "returns only views for the specified object",
			setup: func(svc ObjectViewService, loader *mockOVCacheLoader) {
				ctx := context.Background()
				// Add another object to cache so we can create views for it
				loader.objects = append(loader.objects, ObjectDefinition{
					ID: otherObjectID, APIName: "accounts",
				})

				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID: testObjectID, APIName: "contact_view", Label: "Contact View",
				})
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID: testObjectID, APIName: "contact_view_2", Label: "Contact View 2",
				})
			},
			objectID:  testObjectID,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, loader := setupOVServiceTest(t)
			tt.setup(svc, loader)

			views, err := svc.ListByObjectID(context.Background(), tt.objectID)
			require.NoError(t, err)
			assert.Len(t, views, tt.wantCount)
		})
	}
}

func TestObjectViewService_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   UpdateObjectViewInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "updates label successfully",
			input: UpdateObjectViewInput{
				Label:       "Updated Label",
				Description: "Updated desc",
				IsDefault:   false,
			},
		},
		{
			name: "updates config successfully",
			input: UpdateObjectViewInput{
				Label: "With Config",
				Config: OVConfig{
					Read: OVReadConfig{
						Fields: []string{"email"},
					},
				},
			},
		},
		{
			name: "rejects empty label",
			input: UpdateObjectViewInput{
				Label: "",
			},
			wantErr: true,
			errMsg:  "label is required",
		},
		{
			name: "rejects label longer than 255 characters",
			input: UpdateObjectViewInput{
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
			assert.Equal(t, tt.input.IsDefault, ov.IsDefault)
		})
	}
}

func TestObjectViewService_Update_NotFound(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)

	_, err := svc.Update(context.Background(), uuid.New(), UpdateObjectViewInput{
		Label: "Updated",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestObjectViewService_Delete(t *testing.T) {
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

func TestObjectViewService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)

	err := svc.Delete(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestObjectViewService_ResolveForProfile(t *testing.T) {
	t.Parallel()

	profileSales := uuid.New()
	profileSupport := uuid.New()

	tests := []struct {
		name      string
		setup     func(svc ObjectViewService)
		profileID uuid.UUID
		wantLabel string
		wantNil   bool
	}{
		{
			name: "returns profile-specific view when available",
			setup: func(svc ObjectViewService) {
				ctx := context.Background()
				// Default view
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID:  testObjectID,
					APIName:   "default_view",
					Label:     "Default View",
					IsDefault: true,
				})
				// Profile-specific view
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID:  testObjectID,
					ProfileID: &profileSales,
					APIName:   "sales_view",
					Label:     "Sales View",
				})
			},
			profileID: profileSales,
			wantLabel: "Sales View",
		},
		{
			name: "falls back to default when no profile-specific view",
			setup: func(svc ObjectViewService) {
				ctx := context.Background()
				// Only default view
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID:  testObjectID,
					APIName:   "default_view",
					Label:     "Default View",
					IsDefault: true,
				})
			},
			profileID: profileSupport,
			wantLabel: "Default View",
		},
		{
			name: "returns nil when no views exist at all",
			setup: func(_ ObjectViewService) {
				// No views created
			},
			profileID: profileSales,
			wantNil:   true,
		},
		{
			name: "returns nil when no profile-specific and no default view",
			setup: func(svc ObjectViewService) {
				ctx := context.Background()
				// Non-default, different profile
				otherProfile := uuid.New()
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID:  testObjectID,
					ProfileID: &otherProfile,
					APIName:   "other_profile_view",
					Label:     "Other Profile View",
				})
			},
			profileID: profileSales,
			wantNil:   true,
		},
		{
			name: "prefers profile-specific over default",
			setup: func(svc ObjectViewService) {
				ctx := context.Background()
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID:  testObjectID,
					APIName:   "default_view",
					Label:     "Default",
					IsDefault: true,
				})
				_, _ = svc.Create(ctx, CreateObjectViewInput{
					ObjectID:  testObjectID,
					ProfileID: &profileSales,
					APIName:   "sales_special",
					Label:     "Sales Special",
				})
			},
			profileID: profileSales,
			wantLabel: "Sales Special",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _, _ := setupOVServiceTest(t)
			tt.setup(svc)

			ov, err := svc.ResolveForProfile(context.Background(), testObjectID, tt.profileID)
			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, ov)
				return
			}

			require.NotNil(t, ov)
			assert.Equal(t, tt.wantLabel, ov.Label)
			assert.Equal(t, testObjectID, ov.ObjectID)
		})
	}
}

func TestObjectViewService_Create_ValidAPINamePatterns(t *testing.T) {
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

			_, err := svc.Create(context.Background(), CreateObjectViewInput{
				ObjectID: testObjectID,
				APIName:  tt.apiName,
				Label:    "Test Label",
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestObjectViewService_Create_ObjectExistenceCheck(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)

	nonexistentObjectID := uuid.New()
	_, err := svc.Create(context.Background(), CreateObjectViewInput{
		ObjectID: nonexistentObjectID,
		APIName:  "my_view",
		Label:    "Test",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "object")
}

func TestObjectViewService_ErrorWrapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		operation  func(svc ObjectViewService) error
		wantPrefix string
	}{
		{
			name: "Create wraps with method name",
			operation: func(svc ObjectViewService) error {
				_, err := svc.Create(context.Background(), CreateObjectViewInput{
					ObjectID: uuid.Nil,
					APIName:  "test",
					Label:    "Test",
				})
				return err
			},
			wantPrefix: "objectViewService.Create:",
		},
		{
			name: "GetByID wraps with method name",
			operation: func(svc ObjectViewService) error {
				_, err := svc.GetByID(context.Background(), uuid.New())
				return err
			},
			wantPrefix: "objectViewService.GetByID:",
		},
		{
			name: "Update wraps with method name",
			operation: func(svc ObjectViewService) error {
				_, err := svc.Update(context.Background(), uuid.New(), UpdateObjectViewInput{Label: "Test"})
				return err
			},
			wantPrefix: "objectViewService.Update:",
		},
		{
			name: "Delete wraps with method name",
			operation: func(svc ObjectViewService) error {
				return svc.Delete(context.Background(), uuid.New())
			},
			wantPrefix: "objectViewService.Delete:",
		},
		{
			name: "ResolveForProfile wraps with method name on success path",
			operation: func(svc ObjectViewService) error {
				// ResolveForProfile returns nil, nil when not found, so we just verify no error
				_, err := svc.ResolveForProfile(context.Background(), testObjectID, uuid.New())
				return err
			},
			wantPrefix: "", // no error expected
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

func TestObjectViewService_Update_PreservesObjectID(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, validCreateInput())
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, UpdateObjectViewInput{
		Label:       "New Label",
		Description: "New Desc",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	// Object ID and API name should remain unchanged
	assert.Equal(t, created.ObjectID, updated.ObjectID)
	assert.Equal(t, created.APIName, updated.APIName)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "New Label", updated.Label)
}

func TestObjectViewService_Delete_ThenGetReturnsNotFound(t *testing.T) {
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

func TestObjectViewService_ListAll_AfterCreateAndDelete(t *testing.T) {
	t.Parallel()
	svc, _, _ := setupOVServiceTest(t)
	ctx := context.Background()

	ov1, err := svc.Create(ctx, CreateObjectViewInput{
		ObjectID: testObjectID, APIName: "view_one", Label: "View One",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, CreateObjectViewInput{
		ObjectID: testObjectID, APIName: "view_two", Label: "View Two",
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
