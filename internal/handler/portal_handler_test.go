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

	celutil "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/soql"
)

func buildPortalHandlerTestCache(portals []metadata.Portal) *metadata.MetadataCache {
	portalLoader := &portalCacheLoaderForView{portals: portals}
	cache := metadata.NewMetadataCache(portalLoader)
	if err := cache.Load(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to load test cache: %v", err))
	}
	return cache
}

type portalCacheLoaderForView struct {
	portals []metadata.Portal
}

func (l *portalCacheLoaderForView) LoadAllObjects(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadAllFields(_ context.Context) ([]metadata.FieldDefinition, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadRelationships(_ context.Context) ([]metadata.RelationshipInfo, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadAllValidationRules(_ context.Context) ([]metadata.ValidationRule, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadAllFunctions(_ context.Context) ([]metadata.Function, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadAllPortals(_ context.Context) ([]metadata.Portal, error) {
	return l.portals, nil
}
func (l *portalCacheLoaderForView) LoadAllProcedures(_ context.Context) ([]metadata.Procedure, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadAllAutomationRules(_ context.Context) ([]metadata.AutomationRule, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadAllLayouts(_ context.Context) ([]metadata.Layout, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) LoadAllSharedLayouts(_ context.Context) ([]metadata.SharedLayout, error) {
	return nil, nil
}
func (l *portalCacheLoaderForView) RefreshMaterializedView(_ context.Context) error {
	return nil
}

func setupViewRouter(t *testing.T, cache *metadata.MetadataCache, soqlSvc soql.QueryService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	h := NewPortalHandler(cache, soqlSvc, nil, nil)
	h.RegisterRoutes(api)
	return r
}

func setupViewRouterWithCEL(t *testing.T, cache *metadata.MetadataCache, soqlSvc soql.QueryService, celCache *celutil.ProgramCache) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	h := NewPortalHandler(cache, soqlSvc, nil, celCache)
	h.RegisterRoutes(api)
	return r
}

func TestPortalHandler_GetByAPIName(t *testing.T) {
	t.Parallel()

	testPortal := metadata.Portal{
		ID:      uuid.New(),
		APIName: "sales_dashboard",
		Label:   "Sales Dashboard",
		Config: metadata.PortalConfig{
			Read: metadata.PortalReadConfig{
				Fields: []metadata.PortalViewField{{Name: "name"}, {Name: "amount"}},
			},
		},
	}

	tests := []struct {
		name       string
		apiName    string
		portals    []metadata.Portal
		wantStatus int
		wantLabel  string
	}{
		{
			name:       "returns OV config when found",
			apiName:    "sales_dashboard",
			portals:    []metadata.Portal{testPortal},
			wantStatus: http.StatusOK,
			wantLabel:  "Sales Dashboard",
		},
		{
			name:       "returns 404 when not found",
			apiName:    "nonexistent",
			portals:    []metadata.Portal{testPortal},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 404 when no views at all",
			apiName:    "anything",
			portals:    nil,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := buildPortalHandlerTestCache(tt.portals)
			r := setupViewRouter(t, cache, nil)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/portal/"+tt.apiName, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Data metadata.Portal `json:"data"`
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
	executeFn  func(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error)
	describeFn func(ctx context.Context, query string) (*soql.DescribeResult, error)
}

func (m *mockSOQLService) Execute(ctx context.Context, query string, params *soql.QueryParams) (*soql.QueryResult, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, query, params)
	}
	return &soql.QueryResult{}, nil
}

func (m *mockSOQLService) Describe(ctx context.Context, query string) (*soql.DescribeResult, error) {
	if m.describeFn != nil {
		return m.describeFn(ctx, query)
	}
	return &soql.DescribeResult{}, nil
}

func TestPortalHandler_ExecuteQuery(t *testing.T) {
	t.Parallel()

	testPortal := metadata.Portal{
		ID:      uuid.New(),
		APIName: "account_view",
		Label:   "Account View",
		Config: metadata.PortalConfig{
			Read: metadata.PortalReadConfig{
				Fields: []metadata.PortalViewField{{Name: "name"}},
				Queries: []metadata.PortalQuery{
					{Name: "main", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :id"},
					{Name: "contacts", SOQL: "SELECT Id, Name FROM Contact WHERE AccountId = :id"},
				},
			},
		},
	}

	tests := []struct {
		name          string
		portalAPIName string
		queryName     string
		queryStr      string
		portals       []metadata.Portal
		setupSOQL     func(m *mockSOQLService)
		wantStatus    int
	}{
		{
			name:          "executes scalar query successfully",
			portalAPIName: "account_view",
			queryName:     "main",
			queryStr:      "id=abc-123",
			portals:       []metadata.Portal{testPortal},
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
			name:          "executes list query",
			portalAPIName: "account_view",
			queryName:     "contacts",
			queryStr:      "id=abc-123&per_page=10",
			portals:       []metadata.Portal{testPortal},
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
			name:          "returns 404 for unknown OV",
			portalAPIName: "nonexistent",
			queryName:     "main",
			portals:       []metadata.Portal{testPortal},
			wantStatus:    http.StatusNotFound,
		},
		{
			name:          "returns 404 for unknown query",
			portalAPIName: "account_view",
			queryName:     "nonexistent",
			portals:       []metadata.Portal{testPortal},
			wantStatus:    http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := buildPortalHandlerTestCache(tt.portals)
			soqlSvc := &mockSOQLService{}
			if tt.setupSOQL != nil {
				tt.setupSOQL(soqlSvc)
			}
			r := setupViewRouter(t, cache, soqlSvc)

			url := fmt.Sprintf("/api/v1/portal/%s/query/%s", tt.portalAPIName, tt.queryName)
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

func strPtr(s string) *string { return &s }

func TestPortalHandler_TypedArgs(t *testing.T) {
	t.Parallel()

	portalWithArgs := metadata.Portal{
		ID:      uuid.New(),
		APIName: "typed_view",
		Label:   "Typed View",
		Config: metadata.PortalConfig{
			Args: []metadata.PortalArg{
				{Name: "account_id", Type: "string"},
				{Name: "limit", Type: "int", Default: strPtr("10")},
				{Name: "active", Type: "bool", Default: strPtr("true")},
			},
			Read: metadata.PortalReadConfig{
				Fields: []metadata.PortalViewField{{Name: "name"}},
				Queries: []metadata.PortalQuery{
					{Name: "main", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :account_id"},
					{Name: "cond", SOQL: "SELECT ROW Id FROM Contact", When: "args.active == true"},
				},
			},
		},
	}

	portalNoArgs := metadata.Portal{
		ID:      uuid.New(),
		APIName: "legacy_view",
		Label:   "Legacy View",
		Config: metadata.PortalConfig{
			Read: metadata.PortalReadConfig{
				Fields: []metadata.PortalViewField{{Name: "name"}},
				Queries: []metadata.PortalQuery{
					{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :id"},
				},
			},
		},
	}

	celEnv, err := celutil.PortalEnv()
	require.NoError(t, err)
	celCache := celutil.NewProgramCache(celEnv)

	tests := []struct {
		name          string
		portalAPIName string
		queryName     string
		queryStr      string
		portals       []metadata.Portal
		setupSOQL     func(m *mockSOQLService)
		useCEL        bool
		wantStatus    int
		checkBody     func(t *testing.T, body []byte)
	}{
		{
			name:          "required arg missing returns 400",
			portalAPIName: "typed_view",
			queryName:     "main",
			queryStr:      "",
			portals:       []metadata.Portal{portalWithArgs},
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "int arg with non-numeric value returns 400",
			portalAPIName: "typed_view",
			queryName:     "main",
			queryStr:      "account_id=abc&limit=notanumber",
			portals:       []metadata.Portal{portalWithArgs},
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "optional arg uses default",
			portalAPIName: "typed_view",
			queryName:     "main",
			queryStr:      "account_id=abc-123",
			portals:       []metadata.Portal{portalWithArgs},
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, query string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					assert.Contains(t, query, "'abc-123'")
					return &soql.QueryResult{
						IsRow:   true,
						Records: []map[string]any{{"Id": "abc-123", "Name": "Acme"}},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "when=false skips query (data: null)",
			portalAPIName: "typed_view",
			queryName:     "cond",
			queryStr:      "account_id=abc&active=false",
			portals:       []metadata.Portal{portalWithArgs},
			useCEL:        true,
			wantStatus:    http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Nil(t, resp["data"])
			},
		},
		{
			name:          "when=true executes query",
			portalAPIName: "typed_view",
			queryName:     "cond",
			queryStr:      "account_id=abc&active=true",
			portals:       []metadata.Portal{portalWithArgs},
			useCEL:        true,
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					return &soql.QueryResult{
						IsRow:   true,
						Records: []map[string]any{{"Id": "c1"}},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.NotNil(t, resp["data"])
			},
		},
		{
			name:          "backward compat: no args uses old substitution",
			portalAPIName: "legacy_view",
			queryName:     "main",
			queryStr:      "id=xyz-789",
			portals:       []metadata.Portal{portalNoArgs},
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, query string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					assert.Contains(t, query, "'xyz-789'")
					return &soql.QueryResult{
						IsRow:   true,
						Records: []map[string]any{{"Id": "xyz-789"}},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := buildPortalHandlerTestCache(tt.portals)
			soqlSvc := &mockSOQLService{}
			if tt.setupSOQL != nil {
				tt.setupSOQL(soqlSvc)
			}

			var r *gin.Engine
			if tt.useCEL {
				r = setupViewRouterWithCEL(t, cache, soqlSvc, celCache)
			} else {
				r = setupViewRouter(t, cache, soqlSvc)
			}

			url := fmt.Sprintf("/api/v1/portal/%s/query/%s", tt.portalAPIName, tt.queryName)
			if tt.queryStr != "" {
				url += "?" + tt.queryStr
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestPortalHandler_ArgValidation(t *testing.T) {
	t.Parallel()

	portalWithValidation := metadata.Portal{
		ID:      uuid.New(),
		APIName: "validated_view",
		Label:   "Validated View",
		Config: metadata.PortalConfig{
			Args: []metadata.PortalArg{
				{
					Name:         "limit",
					Type:         "int",
					Default:      strPtr("10"),
					Validation:   "args.limit > 0 && args.limit <= 100",
					ErrorMessage: "Limit must be between 1 and 100",
				},
				{
					Name: "account_id",
					Type: "string",
				},
			},
			Read: metadata.PortalReadConfig{
				Fields: []metadata.PortalViewField{{Name: "name"}},
				Queries: []metadata.PortalQuery{
					{Name: "main", SOQL: "SELECT ROW Id FROM Account WHERE Id = :account_id LIMIT :limit"},
				},
			},
		},
	}

	celEnv, err := celutil.PortalEnv()
	require.NoError(t, err)
	celCache := celutil.NewProgramCache(celEnv)

	tests := []struct {
		name       string
		queryStr   string
		setupSOQL  func(m *mockSOQLService)
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:       "validation passes — request proceeds",
			queryStr:   "account_id=abc&limit=50",
			wantStatus: http.StatusOK,
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					return &soql.QueryResult{IsRow: true, Records: []map[string]any{{"Id": "abc"}}}, nil
				}
			},
		},
		{
			name:       "validation fails — returns 400 with error_message",
			queryStr:   "account_id=abc&limit=0",
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "Limit must be between 1 and 100")
			},
		},
		{
			name:       "validation fails — negative value",
			queryStr:   "account_id=abc&limit=-5",
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "Limit must be between 1 and 100")
			},
		},
		{
			name:       "validation fails — exceeds max",
			queryStr:   "account_id=abc&limit=200",
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "Limit must be between 1 and 100")
			},
		},
		{
			name:       "validation passes with default value",
			queryStr:   "account_id=abc",
			wantStatus: http.StatusOK,
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					return &soql.QueryResult{IsRow: true, Records: []map[string]any{{"Id": "abc"}}}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := buildPortalHandlerTestCache([]metadata.Portal{portalWithValidation})
			soqlSvc := &mockSOQLService{}
			if tt.setupSOQL != nil {
				tt.setupSOQL(soqlSvc)
			}
			r := setupViewRouterWithCEL(t, cache, soqlSvc, celCache)

			url := fmt.Sprintf("/api/v1/portal/validated_view/query/main?%s", tt.queryStr)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestConvertArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		argName string
		argType string
		raw     string
		want    any
		wantErr bool
	}{
		{name: "string", argName: "s", argType: "string", raw: "hello", want: "hello"},
		{name: "int valid", argName: "n", argType: "int", raw: "42", want: int64(42)},
		{name: "int negative", argName: "n", argType: "int", raw: "-5", want: int64(-5)},
		{name: "int invalid", argName: "n", argType: "int", raw: "abc", wantErr: true},
		{name: "float valid", argName: "f", argType: "float", raw: "3.14", want: float64(3.14)},
		{name: "float invalid", argName: "f", argType: "float", raw: "xyz", wantErr: true},
		{name: "bool true", argName: "b", argType: "bool", raw: "true", want: true},
		{name: "bool false", argName: "b", argType: "bool", raw: "false", want: false},
		{name: "bool invalid", argName: "b", argType: "bool", raw: "yes", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := convertArg(tt.argName, tt.argType, tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubstituteTypedArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		soql string
		args map[string]any
		want string
	}{
		{
			name: "string arg quoted with escaping",
			soql: "SELECT Id FROM Account WHERE Name = :name",
			args: map[string]any{"name": "O'Brien"},
			want: "SELECT Id FROM Account WHERE Name = 'O''Brien'",
		},
		{
			name: "int arg unquoted",
			soql: "SELECT Id FROM Account LIMIT :limit",
			args: map[string]any{"limit": int64(10)},
			want: "SELECT Id FROM Account LIMIT 10",
		},
		{
			name: "float arg unquoted",
			soql: "SELECT Id FROM Deal WHERE Amount > :min",
			args: map[string]any{"min": float64(1000.5)},
			want: "SELECT Id FROM Deal WHERE Amount > 1000.5",
		},
		{
			name: "bool arg unquoted",
			soql: "SELECT Id FROM Account WHERE Active = :active",
			args: map[string]any{"active": true},
			want: "SELECT Id FROM Account WHERE Active = true",
		},
		{
			name: "unknown param left as-is",
			soql: "SELECT Id FROM Account WHERE Id = :unknown",
			args: map[string]any{},
			want: "SELECT Id FROM Account WHERE Id = :unknown",
		},
		{
			name: "multiple params",
			soql: "SELECT Id FROM Account WHERE Id = :id AND Status = :status",
			args: map[string]any{"id": "abc", "status": "active"},
			want: "SELECT Id FROM Account WHERE Id = 'abc' AND Status = 'active'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := substituteTypedArgs(tt.soql, tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}
