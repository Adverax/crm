package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
)

// --- Mock ObjectViewService ---

type mockOVService struct {
	resolveFn func(ctx context.Context, objectID, profileID uuid.UUID) (*metadata.ObjectView, error)
}

func (m *mockOVService) Create(_ context.Context, _ metadata.CreateObjectViewInput) (*metadata.ObjectView, error) {
	return nil, nil
}
func (m *mockOVService) GetByID(_ context.Context, _ uuid.UUID) (*metadata.ObjectView, error) {
	return nil, nil
}
func (m *mockOVService) ListAll(_ context.Context) ([]metadata.ObjectView, error) { return nil, nil }
func (m *mockOVService) ListByObjectID(_ context.Context, _ uuid.UUID) ([]metadata.ObjectView, error) {
	return nil, nil
}
func (m *mockOVService) Update(_ context.Context, _ uuid.UUID, _ metadata.UpdateObjectViewInput) (*metadata.ObjectView, error) {
	return nil, nil
}
func (m *mockOVService) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockOVService) ResolveForProfile(ctx context.Context, objectID, profileID uuid.UUID) (*metadata.ObjectView, error) {
	if m.resolveFn != nil {
		return m.resolveFn(ctx, objectID, profileID)
	}
	return nil, nil
}

// --- Mock OLS/FLS enforcers ---

type mockOLSEnforcer struct {
	canReadFn func(ctx context.Context, userID, objectID uuid.UUID) error
}

func (m *mockOLSEnforcer) CanRead(ctx context.Context, userID, objectID uuid.UUID) error {
	if m.canReadFn != nil {
		return m.canReadFn(ctx, userID, objectID)
	}
	return nil
}

func (m *mockOLSEnforcer) CanCreate(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockOLSEnforcer) CanUpdate(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockOLSEnforcer) CanDelete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockOLSEnforcer) GetPermissions(_ context.Context, _, _ uuid.UUID) (int, error) {
	return 15, nil
}

type mockFLSEnforcer struct {
	canReadFieldFn func(ctx context.Context, userID, fieldID uuid.UUID) error
}

func (m *mockFLSEnforcer) CanReadField(ctx context.Context, userID, fieldID uuid.UUID) error {
	if m.canReadFieldFn != nil {
		return m.canReadFieldFn(ctx, userID, fieldID)
	}
	return nil
}

func (m *mockFLSEnforcer) CanWriteField(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockFLSEnforcer) GetReadableFields(_ context.Context, _, _ uuid.UUID) ([]string, error) {
	return nil, nil
}
func (m *mockFLSEnforcer) GetWritableFields(_ context.Context, _, _ uuid.UUID) ([]string, error) {
	return nil, nil
}

// --- Test helpers ---

type stubDescribeCacheLoader struct {
	objects []metadata.ObjectDefinition
	fields  []metadata.FieldDefinition
}

func (s *stubDescribeCacheLoader) LoadAllObjects(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return s.objects, nil
}

func (s *stubDescribeCacheLoader) LoadAllFields(_ context.Context) ([]metadata.FieldDefinition, error) {
	return s.fields, nil
}

func (s *stubDescribeCacheLoader) LoadRelationships(_ context.Context) ([]metadata.RelationshipInfo, error) {
	return nil, nil
}

func (s *stubDescribeCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

func (s *stubDescribeCacheLoader) LoadAllValidationRules(_ context.Context) ([]metadata.ValidationRule, error) {
	return nil, nil
}

func (s *stubDescribeCacheLoader) LoadAllFunctions(_ context.Context) ([]metadata.Function, error) {
	return nil, nil
}

func (s *stubDescribeCacheLoader) LoadAllObjectViews(_ context.Context) ([]metadata.ObjectView, error) {
	return nil, nil
}

func (s *stubDescribeCacheLoader) LoadAllProcedures(_ context.Context) ([]metadata.Procedure, error) {
	return nil, nil
}

func buildDescribeTestCache(objID uuid.UUID, apiName, tableName string) *metadata.MetadataCache {
	fieldID := uuid.New()
	loader := &stubDescribeCacheLoader{
		objects: []metadata.ObjectDefinition{
			{
				ID:           objID,
				APIName:      apiName,
				TableName:    tableName,
				Label:        apiName,
				PluralLabel:  apiName + "s",
				IsCreateable: true,
				IsUpdateable: true,
				IsDeleteable: true,
				IsQueryable:  true,
			},
		},
		fields: []metadata.FieldDefinition{
			{
				ID:         fieldID,
				ObjectID:   objID,
				APIName:    "Name",
				Label:      "Название",
				FieldType:  metadata.FieldTypeText,
				IsRequired: true,
				SortOrder:  1,
			},
		},
	}
	cache := metadata.NewMetadataCache(loader)
	if err := cache.Load(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to load test cache: %v", err))
	}
	return cache
}

func setupDescribeRouter(t *testing.T, h *DescribeHandler, userID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	r.Use(func(c *gin.Context) {
		ctx := security.ContextWithUser(c.Request.Context(), security.UserContext{UserID: userID})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func setupDescribeRouterWithProfile(t *testing.T, h *DescribeHandler, userID, profileID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	r.Use(func(c *gin.Context) {
		ctx := security.ContextWithUser(c.Request.Context(), security.UserContext{
			UserID:    userID,
			ProfileID: profileID,
		})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func setupDescribeRouterNoAuth(t *testing.T, h *DescribeHandler) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

// --- Tests ---

func TestDescribeHandler_ListObjects(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	userID := uuid.New()
	cache := buildDescribeTestCache(objID, "Account", "obj_account")

	tests := []struct {
		name       string
		setupOLS   func(*mockOLSEnforcer)
		noAuth     bool
		wantStatus int
		wantCount  int
	}{
		{
			name:       "returns objects the user can read",
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name: "filters out objects without OLS read",
			setupOLS: func(m *mockOLSEnforcer) {
				m.canReadFn = func(_ context.Context, _, _ uuid.UUID) error {
					return apperror.Forbidden("no read")
				}
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name:       "returns 401 without user context",
			noAuth:     true,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			olsEnf := &mockOLSEnforcer{}
			if tt.setupOLS != nil {
				tt.setupOLS(olsEnf)
			}
			h := NewDescribeHandler(cache, olsEnf, &mockFLSEnforcer{}, nil)

			var r *gin.Engine
			if tt.noAuth {
				r = setupDescribeRouterNoAuth(t, h)
			} else {
				r = setupDescribeRouter(t, h, userID)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/describe", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				var resp map[string]json.RawMessage
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				var items []objectNavItem
				if err := json.Unmarshal(resp["data"], &items); err != nil {
					t.Fatalf("failed to parse data: %v", err)
				}
				if len(items) != tt.wantCount {
					t.Errorf("expected %d items, got %d", tt.wantCount, len(items))
				}
			}
		})
	}
}

func TestDescribeHandler_DescribeObject(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	userID := uuid.New()
	cache := buildDescribeTestCache(objID, "Account", "obj_account")

	tests := []struct {
		name         string
		objectName   string
		setupOLS     func(*mockOLSEnforcer)
		setupFLS     func(*mockFLSEnforcer)
		noAuth       bool
		wantStatus   int
		wantFieldMin int
	}{
		{
			name:         "returns object description with fields",
			objectName:   "Account",
			wantStatus:   http.StatusOK,
			wantFieldMin: 7, // 6 system fields + 1 user field (Name)
		},
		{
			name:       "returns 404 for unknown object",
			objectName: "Unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 403 when OLS denies read",
			objectName: "Account",
			setupOLS: func(m *mockOLSEnforcer) {
				m.canReadFn = func(_ context.Context, _, _ uuid.UUID) error {
					return apperror.Forbidden("no read")
				}
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "filters out fields without FLS read",
			objectName: "Account",
			setupFLS: func(m *mockFLSEnforcer) {
				m.canReadFieldFn = func(_ context.Context, _, _ uuid.UUID) error {
					return apperror.Forbidden("no field read")
				}
			},
			wantStatus:   http.StatusOK,
			wantFieldMin: 6, // only 6 system fields, user field filtered out
		},
		{
			name:       "returns 401 without user context",
			objectName: "Account",
			noAuth:     true,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			olsEnf := &mockOLSEnforcer{}
			if tt.setupOLS != nil {
				tt.setupOLS(olsEnf)
			}
			flsEnf := &mockFLSEnforcer{}
			if tt.setupFLS != nil {
				tt.setupFLS(flsEnf)
			}
			h := NewDescribeHandler(cache, olsEnf, flsEnf, nil)

			var r *gin.Engine
			if tt.noAuth {
				r = setupDescribeRouterNoAuth(t, h)
			} else {
				r = setupDescribeRouter(t, h, userID)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/describe/"+tt.objectName, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantStatus == http.StatusOK && tt.wantFieldMin > 0 {
				var resp map[string]json.RawMessage
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				var desc objectDescribe
				if err := json.Unmarshal(resp["data"], &desc); err != nil {
					t.Fatalf("failed to parse data: %v", err)
				}
				if len(desc.Fields) < tt.wantFieldMin {
					t.Errorf("expected at least %d fields, got %d", tt.wantFieldMin, len(desc.Fields))
				}
			}
		})
	}
}

func TestDescribeHandler_DescribeObject_WithObjectView(t *testing.T) {
	t.Parallel()

	objID := uuid.New()
	userID := uuid.New()
	profileID := uuid.New()

	// Build cache with two user fields: Name and Email
	emailFieldID := uuid.New()
	nameFieldID := uuid.New()
	loader := &stubDescribeCacheLoader{
		objects: []metadata.ObjectDefinition{
			{
				ID:           objID,
				APIName:      "Contact",
				TableName:    "obj_contact",
				Label:        "Contact",
				PluralLabel:  "Contacts",
				IsCreateable: true,
				IsUpdateable: true,
				IsDeleteable: true,
				IsQueryable:  true,
			},
		},
		fields: []metadata.FieldDefinition{
			{
				ID:         nameFieldID,
				ObjectID:   objID,
				APIName:    "Name",
				Label:      "Name",
				FieldType:  metadata.FieldTypeText,
				IsRequired: true,
				SortOrder:  1,
			},
			{
				ID:         emailFieldID,
				ObjectID:   objID,
				APIName:    "Email",
				Label:      "Email",
				FieldType:  metadata.FieldTypeText,
				IsRequired: false,
				SortOrder:  2,
			},
		},
	}
	cache := metadata.NewMetadataCache(loader)
	if err := cache.Load(context.Background()); err != nil {
		t.Fatalf("failed to load cache: %v", err)
	}

	testOV := &metadata.ObjectView{
		ID:       uuid.New(),
		ObjectID: objID,
		APIName:  "contact_sales",
		Label:    "Sales View",
		Config: metadata.OVConfig{
			Read: metadata.OVReadConfig{
				Fields: []string{"Name", "Email"},
				Actions: []metadata.OVAction{
					{Key: "send_email", Label: "Send Email", Type: "primary", Icon: "mail", VisibilityExpr: "record.Email != ''"},
				},
			},
		},
	}

	tests := []struct {
		name           string
		ov             *metadata.ObjectView
		setupFLS       func(*mockFLSEnforcer)
		wantFormFields []string
		wantActions    int
		wantHighlights int
		wantSections   int
		wantListFields int
	}{
		{
			name:           "form built from OV read config",
			ov:             testOV,
			wantFormFields: []string{"Name", "Email"},
			wantActions:    1,
			wantHighlights: 2,
			wantSections:   1,
			wantListFields: 2,
		},
		{
			name: "FLS intersection excludes inaccessible OV fields",
			ov:   testOV,
			setupFLS: func(m *mockFLSEnforcer) {
				m.canReadFieldFn = func(_ context.Context, _, fieldID uuid.UUID) error {
					if fieldID == emailFieldID {
						return apperror.Forbidden("no read on Email")
					}
					return nil
				}
			},
			wantFormFields: []string{"Name"},
			wantActions:    1,
			wantHighlights: 1,
			wantSections:   1,
			wantListFields: 1,
		},
		{
			name: "OV with actions passed through",
			ov: &metadata.ObjectView{
				ID:       uuid.New(),
				ObjectID: objID,
				APIName:  "contact_actions",
				Label:    "Actions View",
				Config: metadata.OVConfig{
					Read: metadata.OVReadConfig{
						Fields: []string{"Name"},
						Actions: []metadata.OVAction{
							{Key: "call", Label: "Call", Type: "primary"},
							{Key: "email", Label: "Email", Type: "secondary"},
						},
					},
				},
			},
			wantFormFields: []string{"Name"},
			wantActions:    2,
			wantHighlights: 1,
			wantSections:   1,
			wantListFields: 1,
		},
		{
			name: "OV with empty fields produces no sections",
			ov: &metadata.ObjectView{
				ID:       uuid.New(),
				ObjectID: objID,
				APIName:  "contact_empty",
				Label:    "Empty View",
				Config: metadata.OVConfig{
					Read: metadata.OVReadConfig{
						Fields:  []string{},
						Actions: []metadata.OVAction{},
					},
				},
			},
			wantFormFields: []string{},
			wantActions:    0,
			wantHighlights: 0,
			wantSections:   0,
			wantListFields: 0,
		},
		{
			name: "highlight limited to first 3 fields",
			ov: &metadata.ObjectView{
				ID:       uuid.New(),
				ObjectID: objID,
				APIName:  "contact_many",
				Label:    "Many Fields",
				Config: metadata.OVConfig{
					Read: metadata.OVReadConfig{
						Fields:  []string{"Name", "Email", "Id", "CreatedAt", "OwnerId"},
						Actions: []metadata.OVAction{},
					},
				},
			},
			wantFormFields: nil, // don't check exact list
			wantActions:    0,
			wantHighlights: 3,
			wantSections:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flsEnf := &mockFLSEnforcer{}
			if tt.setupFLS != nil {
				tt.setupFLS(flsEnf)
			}

			ovSvc := &mockOVService{
				resolveFn: func(_ context.Context, _, _ uuid.UUID) (*metadata.ObjectView, error) {
					return tt.ov, nil
				},
			}

			h := NewDescribeHandler(cache, &mockOLSEnforcer{}, flsEnf, ovSvc)
			r := setupDescribeRouterWithProfile(t, h, userID, profileID)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/describe/Contact", nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
			}

			var resp map[string]json.RawMessage
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			var desc objectDescribe
			if err := json.Unmarshal(resp["data"], &desc); err != nil {
				t.Fatalf("failed to parse data: %v", err)
			}

			if desc.Form == nil {
				t.Fatal("expected form in response, got nil")
			}

			if tt.wantFormFields != nil {
				// Extract field names from sections
				var sectionFields []string
				for _, s := range desc.Form.Sections {
					sectionFields = append(sectionFields, s.Fields...)
				}
				if len(sectionFields) != len(tt.wantFormFields) {
					t.Errorf("form section fields = %v, want %v", sectionFields, tt.wantFormFields)
				}
				for i, f := range tt.wantFormFields {
					if i < len(sectionFields) && sectionFields[i] != f {
						t.Errorf("form section field[%d] = %q, want %q", i, sectionFields[i], f)
					}
				}
			}

			if len(desc.Form.Actions) != tt.wantActions {
				t.Errorf("actions count = %d, want %d", len(desc.Form.Actions), tt.wantActions)
			}

			if len(desc.Form.HighlightFields) != tt.wantHighlights {
				t.Errorf("highlight_fields count = %d, want %d", len(desc.Form.HighlightFields), tt.wantHighlights)
			}

			if len(desc.Form.Sections) != tt.wantSections {
				t.Errorf("sections count = %d, want %d", len(desc.Form.Sections), tt.wantSections)
			}

			if tt.wantListFields > 0 && len(desc.Form.ListFields) != tt.wantListFields {
				t.Errorf("list_fields count = %d, want %d", len(desc.Form.ListFields), tt.wantListFields)
			}
		})
	}
}
