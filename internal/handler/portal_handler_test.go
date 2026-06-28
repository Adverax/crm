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
			EntryGate: "main",
			Gates: map[string]metadata.PortalGate{
				"main": {
					Label: "Main",
					Body:  []metadata.GateBodyStep{{Name: "data", Type: "soql", SOQL: "SELECT Id FROM Account"}},
					Layout: &metadata.GateLayout{
						Fields: []metadata.PortalViewField{{Name: "name"}, {Name: "amount"}},
					},
				},
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
			name:       "returns portal config when found",
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
			name:       "returns 404 when no portals at all",
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

func TestPortalHandler_ExecuteGate_GET(t *testing.T) {
	t.Parallel()

	testPortal := metadata.Portal{
		ID:      uuid.New(),
		APIName: "account_view",
		Label:   "Account View",
		Config: metadata.PortalConfig{
			EntryGate: "list",
			Gates: map[string]metadata.PortalGate{
				"list": {
					Label: "Accounts",
					Body: []metadata.GateBodyStep{
						{Name: "accounts", Type: "soql", SOQL: "SELECT Id, Name FROM Account"},
					},
					Layout: &metadata.GateLayout{
						Fields: []metadata.PortalViewField{{Name: "Name"}},
					},
					Outcomes: []metadata.GateOutcome{
						{Name: "detail", Gate: "detail", Label: "View"},
					},
				},
				"detail": {
					Label: "Detail",
					Args:  []metadata.PortalArg{{Name: "id", Type: "string"}},
					Body: []metadata.GateBodyStep{
						{Name: "record", Type: "soql", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :id"},
					},
					Layout: &metadata.GateLayout{
						Fields: []metadata.PortalViewField{{Name: "Name"}},
					},
					Outcomes: []metadata.GateOutcome{
						{Name: "back", Gate: "list"},
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		portalAPIName string
		gateName      string
		queryStr      string
		portals       []metadata.Portal
		setupSOQL     func(m *mockSOQLService)
		wantStatus    int
		checkBody     func(t *testing.T, body []byte)
	}{
		{
			name:          "executes SOQL gate successfully",
			portalAPIName: "account_view",
			gateName:      "list",
			portals:       []metadata.Portal{testPortal},
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					return &soql.QueryResult{
						TotalSize: 2,
						Records:   []map[string]any{{"Id": "1", "Name": "Acme"}, {"Id": "2", "Name": "Corp"}},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "list", data["gate"])
				assert.NotNil(t, data["layout"])
				assert.NotNil(t, data["datasets"])
				assert.NotNil(t, data["outcomes"])
			},
		},
		{
			name:          "executes scalar (ROW) gate successfully",
			portalAPIName: "account_view",
			gateName:      "detail",
			queryStr:      "id=abc-123",
			portals:       []metadata.Portal{testPortal},
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
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				datasets := data["datasets"].(map[string]any)
				record := datasets["record"].(map[string]any)
				assert.Equal(t, "abc-123", record["Id"])
			},
		},
		{
			name:          "returns 404 for unknown portal",
			portalAPIName: "nonexistent",
			gateName:      "list",
			portals:       []metadata.Portal{testPortal},
			wantStatus:    http.StatusNotFound,
		},
		{
			name:          "returns 404 for unknown gate",
			portalAPIName: "account_view",
			gateName:      "nonexistent",
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

			url := fmt.Sprintf("/api/v1/portal/%s/gate/%s", tt.portalAPIName, tt.gateName)
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

func TestPortalHandler_ExecuteGate_POST_NoBody(t *testing.T) {
	t.Parallel()

	actionPortal := metadata.Portal{
		ID:      uuid.New(),
		APIName: "action_portal",
		Label:   "Action Portal",
		Config: metadata.PortalConfig{
			EntryGate: "main",
			Gates: map[string]metadata.PortalGate{
				"main": {
					Label: "Main",
					Body: []metadata.GateBodyStep{
						{Name: "data", Type: "soql", SOQL: "SELECT Id FROM Account"},
					},
					Outcomes: []metadata.GateOutcome{
						{Name: "action", Gate: "action"},
					},
				},
				"action": {
					Label: "Action",
					Body: []metadata.GateBodyStep{
						{Name: "delete", Type: "dml", DML: "DELETE FROM Account WHERE Id = :id"},
					},
					Args: []metadata.PortalArg{{Name: "id", Type: "string"}},
					Outcomes: []metadata.GateOutcome{
						{Name: "back", Gate: "main"},
					},
				},
			},
		},
	}

	cache := buildPortalHandlerTestCache([]metadata.Portal{actionPortal})
	r := setupViewRouter(t, cache, nil)

	// POST without body should fail
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/portal/action_portal/gate/action", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPortalHandler_ExecuteGate_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	portal := metadata.Portal{
		ID:      uuid.New(),
		APIName: "method_test",
		Label:   "Method Test",
		Config: metadata.PortalConfig{
			EntryGate: "read_only",
			Gates: map[string]metadata.PortalGate{
				"read_only": {
					Label: "Read Only",
					Body: []metadata.GateBodyStep{
						{Name: "data", Type: "soql", SOQL: "SELECT Id FROM Account"},
					},
				},
			},
		},
	}

	cache := buildPortalHandlerTestCache([]metadata.Portal{portal})
	r := setupViewRouter(t, cache, &mockSOQLService{})

	// POST to SOQL-only gate should be 405
	body, _ := json.Marshal(map[string]any{"args": map[string]any{}, "data": map[string]any{}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/portal/method_test/gate/read_only", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
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
			EntryGate: "main",
			Gates: map[string]metadata.PortalGate{
				"main": {
					Label: "Main",
					Body: []metadata.GateBodyStep{
						{Name: "record", Type: "soql", SOQL: "SELECT ROW Id, Name FROM Account WHERE Id = :account_id"},
						{Name: "cond", Type: "soql", SOQL: "SELECT ROW Id FROM Contact", When: "args.active == true"},
					},
					Layout: &metadata.GateLayout{
						Fields: []metadata.PortalViewField{{Name: "Name"}},
					},
				},
			},
		},
	}

	celEnv, err := celutil.GateEnv()
	require.NoError(t, err)
	celCache := celutil.NewProgramCache(celEnv)

	tests := []struct {
		name       string
		queryStr   string
		setupSOQL  func(m *mockSOQLService)
		useCEL     bool
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:       "required arg missing returns 400",
			queryStr:   "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "int arg with non-numeric value returns 400",
			queryStr:   "account_id=abc&limit=notanumber",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "optional arg uses default",
			queryStr: "account_id=abc-123",
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, query string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					return &soql.QueryResult{
						IsRow:   true,
						Records: []map[string]any{{"Id": "abc-123", "Name": "Acme"}},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "when=false skips step (dataset is null)",
			queryStr:   "account_id=abc&active=false",
			useCEL:     true,
			wantStatus: http.StatusOK,
			setupSOQL: func(m *mockSOQLService) {
				m.executeFn = func(_ context.Context, _ string, _ *soql.QueryParams) (*soql.QueryResult, error) {
					return &soql.QueryResult{IsRow: true, Records: []map[string]any{{"Id": "abc"}}}, nil
				}
			},
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				datasets := data["datasets"].(map[string]any)
				assert.Nil(t, datasets["cond"])
			},
		},
		{
			name:     "when=true executes step",
			queryStr: "account_id=abc&active=true",
			useCEL:   true,
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
				data := resp["data"].(map[string]any)
				datasets := data["datasets"].(map[string]any)
				assert.NotNil(t, datasets["cond"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := buildPortalHandlerTestCache([]metadata.Portal{portalWithArgs})
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

			url := fmt.Sprintf("/api/v1/portal/typed_view/gate/main?%s", tt.queryStr)
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
				{Name: "limit", Type: "int", Default: strPtr("10")},
				{Name: "account_id", Type: "string"},
			},
			ArgRules: []metadata.PortalArgRule{
				{
					Name:         "limit_range",
					Condition:    "args.limit > 0 && args.limit <= 100",
					ErrorMessage: "Limit must be between 1 and 100",
				},
			},
			EntryGate: "main",
			Gates: map[string]metadata.PortalGate{
				"main": {
					Label: "Main",
					Body: []metadata.GateBodyStep{
						{Name: "record", Type: "soql", SOQL: "SELECT ROW Id FROM Account WHERE Id = :account_id LIMIT :limit"},
					},
				},
			},
		},
	}

	celEnv, err := celutil.GateEnv()
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
		},
		{
			name:       "validation fails — exceeds max",
			queryStr:   "account_id=abc&limit=200",
			wantStatus: http.StatusBadRequest,
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

			url := fmt.Sprintf("/api/v1/portal/validated_view/gate/main?%s", tt.queryStr)
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
		text string
		args map[string]any
		want string
	}{
		{
			name: "string arg quoted with escaping",
			text: "SELECT Id FROM Account WHERE Name = :name",
			args: map[string]any{"name": "O'Brien"},
			want: "SELECT Id FROM Account WHERE Name = 'O''Brien'",
		},
		{
			name: "int arg unquoted",
			text: "SELECT Id FROM Account LIMIT :limit",
			args: map[string]any{"limit": int64(10)},
			want: "SELECT Id FROM Account LIMIT 10",
		},
		{
			name: "float arg unquoted",
			text: "SELECT Id FROM Deal WHERE Amount > :min",
			args: map[string]any{"min": float64(1000.5)},
			want: "SELECT Id FROM Deal WHERE Amount > 1000.5",
		},
		{
			name: "bool arg unquoted",
			text: "SELECT Id FROM Account WHERE Active = :active",
			args: map[string]any{"active": true},
			want: "SELECT Id FROM Account WHERE Active = true",
		},
		{
			name: "unknown param left as-is",
			text: "SELECT Id FROM Account WHERE Id = :unknown",
			args: map[string]any{},
			want: "SELECT Id FROM Account WHERE Id = :unknown",
		},
		{
			name: "multiple params",
			text: "SELECT Id FROM Account WHERE Id = :id AND Status = :status",
			args: map[string]any{"id": "abc", "status": "active"},
			want: "SELECT Id FROM Account WHERE Id = 'abc' AND Status = 'active'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := substituteTypedArgs(tt.text, tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMergeArgDefs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		portalArgs []metadata.PortalArg
		gateArgs   []metadata.PortalArg
		wantNames  []string
	}{
		{
			name:       "empty both",
			portalArgs: nil,
			gateArgs:   nil,
			wantNames:  nil,
		},
		{
			name:       "portal only",
			portalArgs: []metadata.PortalArg{{Name: "a", Type: "string"}},
			gateArgs:   nil,
			wantNames:  []string{"a"},
		},
		{
			name:       "gate only",
			portalArgs: nil,
			gateArgs:   []metadata.PortalArg{{Name: "b", Type: "int"}},
			wantNames:  []string{"b"},
		},
		{
			name:       "merged without collision",
			portalArgs: []metadata.PortalArg{{Name: "a", Type: "string"}},
			gateArgs:   []metadata.PortalArg{{Name: "b", Type: "int"}},
			wantNames:  []string{"a", "b"},
		},
		{
			name:       "gate shadows portal on collision",
			portalArgs: []metadata.PortalArg{{Name: "id", Type: "string"}},
			gateArgs:   []metadata.PortalArg{{Name: "id", Type: "int"}},
			wantNames:  []string{"id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			merged := mergeArgDefs(tt.portalArgs, tt.gateArgs)
			if tt.wantNames == nil {
				assert.Empty(t, merged)
				return
			}
			gotNames := make([]string, len(merged))
			for i, a := range merged {
				gotNames[i] = a.Name
			}
			assert.Equal(t, tt.wantNames, gotNames)
		})
	}
}
