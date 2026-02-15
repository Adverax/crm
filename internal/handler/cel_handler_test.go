package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	celengine "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/metadata"
)

func setupCELRouter(t *testing.T, cache *metadata.MetadataCache, registry *celengine.FunctionRegistry) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	admin := r.Group("/api/v1/admin")
	h := NewCELHandler(cache, registry)
	h.RegisterRoutes(admin)
	return r
}

func newTestMetadataCache(t *testing.T) *metadata.MetadataCache {
	t.Helper()
	loader := &testCacheLoader{}
	cache := metadata.NewMetadataCache(loader)
	return cache
}

type testCacheLoader struct{}

func (l *testCacheLoader) LoadAllObjects(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return nil, nil
}
func (l *testCacheLoader) LoadAllFields(_ context.Context) ([]metadata.FieldDefinition, error) {
	return nil, nil
}
func (l *testCacheLoader) LoadRelationships(_ context.Context) ([]metadata.RelationshipInfo, error) {
	return nil, nil
}
func (l *testCacheLoader) LoadAllValidationRules(_ context.Context) ([]metadata.ValidationRule, error) {
	return nil, nil
}
func (l *testCacheLoader) LoadAllFunctions(_ context.Context) ([]metadata.Function, error) {
	return nil, nil
}
func (l *testCacheLoader) RefreshMaterializedView(_ context.Context) error {
	return nil
}

func TestCELHandler_Validate(t *testing.T) {
	t.Parallel()

	cache := newTestMetadataCache(t)
	registry, err := celengine.NewFunctionRegistry(nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		wantValid  *bool
	}{
		{
			name: "validates valid expression",
			body: map[string]interface{}{
				"expression": "true",
				"context":    "validation_rule",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(true),
		},
		{
			name: "detects invalid expression",
			body: map[string]interface{}{
				"expression": "invalid !!!! syntax",
				"context":    "validation_rule",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(false),
		},
		{
			name: "validates default_expr context",
			body: map[string]interface{}{
				"expression": `"hello"`,
				"context":    "default_expr",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(true),
		},
		{
			name: "validates function_body context with params",
			body: map[string]interface{}{
				"expression": "x * 2",
				"context":    "function_body",
				"params": []map[string]string{
					{"name": "x", "type": "number"},
				},
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(true),
		},
		{
			name: "validates when_expression context",
			body: map[string]interface{}{
				"expression": "true",
				"context":    "when_expression",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(true),
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for missing expression",
			body: map[string]interface{}{
				"context": "validation_rule",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for missing context",
			body: map[string]interface{}{
				"expression": "true",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := setupCELRouter(t, cache, registry)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/cel/validate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantValid != nil {
				var resp celValidateResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, *tt.wantValid, resp.Valid)
			}
		})
	}
}

func TestCELHandler_Validate_WithFunctions(t *testing.T) {
	t.Parallel()

	cache := newTestMetadataCache(t)
	registry, err := celengine.NewFunctionRegistry([]celengine.FunctionDef{
		{
			Name:       "double",
			Params:     []celengine.ParamDef{{Name: "x", Type: "number"}},
			ReturnType: "number",
			Body:       "x * 2",
		},
	})
	require.NoError(t, err)

	r := setupCELRouter(t, cache, registry)

	body, _ := json.Marshal(map[string]interface{}{
		"expression": "fn.double(42) > 10",
		"context":    "validation_rule",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/cel/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp celValidateResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.NotNil(t, resp.ReturnType)
	assert.Equal(t, "bool", *resp.ReturnType)
}

func TestCELHandler_SetRegistry(t *testing.T) {
	cache := newTestMetadataCache(t)
	h := NewCELHandler(cache, nil)

	// Initially no registry
	registry, err := celengine.NewFunctionRegistry([]celengine.FunctionDef{
		{
			Name:       "greet",
			Params:     []celengine.ParamDef{{Name: "name", Type: "string"}},
			ReturnType: "string",
			Body:       `"Hello, " + name`,
		},
	})
	require.NoError(t, err)

	h.SetRegistry(registry)
	assert.NotNil(t, h.registry)
}

func boolPtr(b bool) *bool {
	return &b
}
