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
	"github.com/adverax/crm/internal/platform/soql"
)

func buildViewHandlerTestCache(ovs []metadata.ObjectView) *metadata.MetadataCache {
	ovLoader := &ovCacheLoaderForView{objectViews: ovs}
	cache := metadata.NewMetadataCache(ovLoader)
	if err := cache.Load(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to load test cache: %v", err))
	}
	return cache
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

func setupViewRouter(t *testing.T, cache *metadata.MetadataCache, soqlSvc soql.QueryService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	h := NewViewHandler(cache, soqlSvc, nil)
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
			Read: metadata.OVReadConfig{
				Fields: []metadata.OVViewField{{Name: "name"}, {Name: "amount"}},
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
			r := setupViewRouter(t, cache, nil)

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

// --- Mock SOQL service ---

type mockSOQLService struct {
	executeFn func(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error)
}

func (m *mockSOQLService) Execute(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, query, params)
	}
	return &soql.QueryResult{}, nil
}

func TestViewHandler_ExecuteQuery(t *testing.T) {
	t.Parallel()

	testOV := metadata.ObjectView{
		ID:      uuid.New(),
		APIName: "account_view",
		Label:   "Account View",
		Config: metadata.OVConfig{
			Read: metadata.OVReadConfig{
				Fields: []metadata.OVViewField{{Name: "name"}},
				Queries: []metadata.OVQuery{
					{Name: "main", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :id"},
					{Name: "contacts", SOQL: "SELECT Id, Name FROM Contact WHERE AccountId = :id"},
				},
			},
		},
	}

	tests := []struct {
		name       string
		ovAPIName  string
		queryName  string
		queryStr   string
		ovs        []metadata.ObjectView
		setupSOQL  func(m *mockSOQLService)
		wantStatus int
	}{
		{
			name:      "executes scalar query successfully",
			ovAPIName: "account_view",
			queryName: "main",
			queryStr:  "id=abc-123",
			ovs:       []metadata.ObjectView{testOV},
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error) {
					assert.Contains(t, query, "'abc-123'")
					return &soql.QueryResult{
						TotalSize: 1,
						Done:      true,
						Records:   []map[string]any{{"Id": "abc-123", "Name": "Acme"}},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "executes list query",
			ovAPIName: "account_view",
			queryName: "contacts",
			queryStr:  "id=abc-123&per_page=10",
			ovs:       []metadata.ObjectView{testOV},
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, _ string, params *soql.QueryParams) (*soql.QueryResult, error) {
					assert.Equal(t, 10, params.PageSize)
					return &soql.QueryResult{
						TotalSize: 2,
						Done:      true,
						Records:   []map[string]any{{"Id": "c1"}, {"Id": "c2"}},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 404 for unknown OV",
			ovAPIName:  "nonexistent",
			queryName:  "main",
			ovs:        []metadata.ObjectView{testOV},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 404 for unknown query",
			ovAPIName:  "account_view",
			queryName:  "nonexistent",
			ovs:        []metadata.ObjectView{testOV},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := buildViewHandlerTestCache(tt.ovs)
			soqlSvc := &mockSOQLService{}
			if tt.setupSOQL != nil {
				tt.setupSOQL(soqlSvc)
			}
			r := setupViewRouter(t, cache, soqlSvc)

			url := fmt.Sprintf("/api/v1/view/%s/query/%s", tt.ovAPIName, tt.queryName)
			if tt.queryStr != "" {
				url += "?" + tt.queryStr
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}
