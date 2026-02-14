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
			h := NewDescribeHandler(cache, olsEnf, &mockFLSEnforcer{})

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
			h := NewDescribeHandler(cache, olsEnf, flsEnf)

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
