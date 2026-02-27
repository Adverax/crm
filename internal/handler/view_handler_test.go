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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/platform/metadata"
)

func buildViewHandlerTestCache(ovs []metadata.ObjectView) *metadata.MetadataCache {
	loader := &stubDescribeCacheLoader{}
	cache := metadata.NewMetadataCache(loader)
	if err := cache.Load(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to load test cache: %v", err))
	}

	// Load OVs into cache via the loader override
	loader2 := &stubDescribeCacheLoader{}
	loader2Ovs := ovs
	_ = loader2
	_ = loader2Ovs

	// Use a custom loader that returns the OVs
	ovLoader := &ovCacheLoaderForView{objectViews: ovs}
	cache2 := metadata.NewMetadataCache(ovLoader)
	if err := cache2.Load(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to load test cache: %v", err))
	}
	return cache2
}

type ovCacheLoaderForView struct {
	objectViews []metadata.ObjectView
}

func (l *ovCacheLoaderForView) LoadAllObjects(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadAllFields(_ context.Context) ([]metadata.FieldDefinition, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadRelationships(_ context.Context) ([]metadata.RelationshipInfo, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadAllValidationRules(_ context.Context) ([]metadata.ValidationRule, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadAllFunctions(_ context.Context) ([]metadata.Function, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadAllObjectViews(_ context.Context) ([]metadata.ObjectView, error) {
	return l.objectViews, nil
}
func (l *ovCacheLoaderForView) LoadAllProcedures(_ context.Context) ([]metadata.Procedure, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadAllAutomationRules(_ context.Context) ([]metadata.AutomationRule, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadAllLayouts(_ context.Context) ([]metadata.Layout, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) LoadAllSharedLayouts(_ context.Context) ([]metadata.SharedLayout, error) {
	return nil, nil
}
func (l *ovCacheLoaderForView) RefreshMaterializedView(_ context.Context) error {
	return nil
}

func setupViewRouter(t *testing.T, cache *metadata.MetadataCache) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	h := NewViewHandler(cache)
	h.RegisterRoutes(api)
	return r
}

func TestViewHandler_GetByAPIName(t *testing.T) {
	t.Parallel()

	testOV := metadata.ObjectView{
		ID:      uuid.New(),
		APIName: "sales_dashboard",
		Label:   "Sales Dashboard",
		Config: metadata.OVConfig{
			View: metadata.OVViewConfig{
				Fields: []string{"name", "amount"},
			},
		},
	}

	tests := []struct {
		name       string
		apiName    string
		ovs        []metadata.ObjectView
		wantStatus int
		wantLabel  string
	}{
		{
			name:       "returns OV config when found",
			apiName:    "sales_dashboard",
			ovs:        []metadata.ObjectView{testOV},
			wantStatus: http.StatusOK,
			wantLabel:  "Sales Dashboard",
		},
		{
			name:       "returns 404 when not found",
			apiName:    "nonexistent",
			ovs:        []metadata.ObjectView{testOV},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 404 when no views at all",
			apiName:    "anything",
			ovs:        nil,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := buildViewHandlerTestCache(tt.ovs)
			r := setupViewRouter(t, cache)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/view/"+tt.apiName, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Data metadata.ObjectView `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, tt.wantLabel, resp.Data.Label)
				assert.Equal(t, tt.apiName, resp.Data.APIName)
			}
		})
	}
}
