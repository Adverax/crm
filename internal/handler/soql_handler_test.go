package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/platform/metadata"
	soqlengine "github.com/adverax/crm/internal/platform/soql/engine"
)

// testMetadataProvider provides test metadata for SOQL engine.
type testMetadataProvider struct {
	objects map[string]*soqlengine.ObjectMeta
}

func (p *testMetadataProvider) GetObject(_ context.Context, name string) (*soqlengine.ObjectMeta, error) {
	obj, ok := p.objects[name]
	if !ok {
		return nil, nil
	}
	return obj, nil
}

func (p *testMetadataProvider) ListObjects(_ context.Context) ([]string, error) {
	names := make([]string, 0, len(p.objects))
	for name := range p.objects {
		names = append(names, name)
	}
	return names, nil
}

func newTestSOQLMetadata() *testMetadataProvider {
	return &testMetadataProvider{
		objects: map[string]*soqlengine.ObjectMeta{
			"Account": {
				Name:       "Account",
				SchemeName: "public",
				TableName:  "obj_account",
				Fields: map[string]*soqlengine.FieldMeta{
					"Id": {
						Name:       "Id",
						Column:     "id",
						Type:       soqlengine.FieldTypeID,
						Filterable: true,
						Sortable:   true,
						Groupable:  true,
					},
					"Name": {
						Name:       "Name",
						Column:     "name",
						Type:       soqlengine.FieldTypeString,
						Filterable: true,
						Sortable:   true,
						Groupable:  true,
					},
					"Industry": {
						Name:       "Industry",
						Column:     "industry",
						Type:       soqlengine.FieldTypeString,
						Filterable: true,
						Sortable:   true,
						Groupable:  true,
					},
				},
			},
		},
	}
}

// testMetadataReader implements metadata.MetadataReader for SOQL handler tests.
type testMetadataReader struct {
	objects map[string]metadata.ObjectDefinition
	fields  map[uuid.UUID][]metadata.FieldDefinition
}

func (r *testMetadataReader) GetObjectByID(id uuid.UUID) (metadata.ObjectDefinition, bool) {
	for _, obj := range r.objects {
		if obj.ID == id {
			return obj, true
		}
	}
	return metadata.ObjectDefinition{}, false
}

func (r *testMetadataReader) GetObjectByAPIName(name string) (metadata.ObjectDefinition, bool) {
	obj, ok := r.objects[name]
	return obj, ok
}

func (r *testMetadataReader) ListObjectAPINames() []string {
	names := make([]string, 0, len(r.objects))
	for name := range r.objects {
		names = append(names, name)
	}
	return names
}

func (r *testMetadataReader) GetFieldByID(_ uuid.UUID) (metadata.FieldDefinition, bool) {
	return metadata.FieldDefinition{}, false
}

func (r *testMetadataReader) GetFieldsByObjectID(objectID uuid.UUID) []metadata.FieldDefinition {
	return r.fields[objectID]
}

func (r *testMetadataReader) GetForwardRelationships(_ uuid.UUID) []metadata.RelationshipInfo {
	return nil
}
func (r *testMetadataReader) GetReverseRelationships(_ uuid.UUID) []metadata.RelationshipInfo {
	return nil
}
func (r *testMetadataReader) GetValidationRules(_ uuid.UUID) []metadata.ValidationRule {
	return nil
}
func (r *testMetadataReader) GetFunctions() []metadata.Function { return nil }
func (r *testMetadataReader) GetFunctionByName(_ string) (metadata.Function, bool) {
	return metadata.Function{}, false
}
func (r *testMetadataReader) GetObjectViewByAPIName(_ string) (metadata.ObjectView, bool) {
	return metadata.ObjectView{}, false
}
func (r *testMetadataReader) GetProcedureByCode(_ string) (metadata.Procedure, bool) {
	return metadata.Procedure{}, false
}
func (r *testMetadataReader) GetProcedures() []metadata.Procedure { return nil }
func (r *testMetadataReader) GetAutomationRules(_ uuid.UUID) []metadata.AutomationRule {
	return nil
}
func (r *testMetadataReader) GetLayoutsForOV(_ uuid.UUID) []metadata.Layout { return nil }
func (r *testMetadataReader) GetSharedLayoutByAPIName(_ string) (metadata.SharedLayout, bool) {
	return metadata.SharedLayout{}, false
}

var testAccountID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func newTestCRMMetadata() *testMetadataReader {
	return &testMetadataReader{
		objects: map[string]metadata.ObjectDefinition{
			"Account": {
				ID:          testAccountID,
				APIName:     "Account",
				Label:       "Accounts",
				IsQueryable: true,
			},
		},
		fields: map[uuid.UUID][]metadata.FieldDefinition{
			testAccountID: {
				{APIName: "Name", Label: "Name", FieldType: "text"},
				{APIName: "Industry", Label: "Industry", FieldType: "text"},
				{APIName: "Id", Label: "ID", FieldType: "text", IsSystemField: true},
			},
		},
	}
}

func setupSOQLRouter(t *testing.T, engineMeta soqlengine.MetadataProvider) *gin.Engine {
	t.Helper()
	return setupSOQLRouterFull(t, engineMeta, newTestCRMMetadata())
}

func setupSOQLRouterFull(t *testing.T, engineMeta soqlengine.MetadataProvider, crmMeta metadata.MetadataReader) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	admin := r.Group("/api/v1/admin")
	engine := soqlengine.NewEngine(
		soqlengine.WithMetadata(engineMeta),
	)
	h := NewSOQLHandler(engine, nil, crmMeta)
	h.RegisterRoutes(admin)
	return r
}

func TestSOQLHandler_Validate(t *testing.T) {
	t.Parallel()

	metadata := newTestSOQLMetadata()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		wantValid  *bool
		wantObject *string
	}{
		{
			name: "validates valid SELECT query",
			body: map[string]interface{}{
				"query": "SELECT Id, Name FROM Account",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(true),
			wantObject: stringPtr("Account"),
		},
		{
			name: "validates query with WHERE clause",
			body: map[string]interface{}{
				"query": "SELECT Id, Name FROM Account WHERE Name = 'Test'",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(true),
			wantObject: stringPtr("Account"),
		},
		{
			name: "validates query with ORDER BY",
			body: map[string]interface{}{
				"query": "SELECT Id, Name FROM Account ORDER BY Name ASC",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(true),
			wantObject: stringPtr("Account"),
		},
		{
			name: "detects syntax error",
			body: map[string]interface{}{
				"query": "SELEC Id FROM Account",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(false),
		},
		{
			name: "detects unknown object",
			body: map[string]interface{}{
				"query": "SELECT Id FROM UnknownObject",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(false),
		},
		{
			name: "detects unknown field",
			body: map[string]interface{}{
				"query": "SELECT NonExistent FROM Account",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(false),
		},
		{
			name:       "returns 400 for missing body",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for empty query",
			body: map[string]interface{}{
				"query": "",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for missing query field",
			body: map[string]interface{}{
				"other": "value",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := setupSOQLRouter(t, metadata)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/soql/validate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantValid != nil {
				var resp soqlValidateResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, *tt.wantValid, resp.Valid)

				if tt.wantObject != nil {
					require.NotNil(t, resp.Object)
					assert.Equal(t, *tt.wantObject, *resp.Object)
				}

				if !resp.Valid {
					assert.NotEmpty(t, resp.Errors, "invalid response should have errors")
				}
			}
		})
	}
}

func TestSOQLHandler_Validate_ReturnsFields(t *testing.T) {
	t.Parallel()

	metadata := newTestSOQLMetadata()
	r := setupSOQLRouter(t, metadata)

	body, _ := json.Marshal(map[string]interface{}{
		"query": "SELECT Id, Name FROM Account",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/soql/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp soqlValidateResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.NotNil(t, resp.Object)
	assert.Equal(t, "Account", *resp.Object)
	assert.NotEmpty(t, resp.Fields)
}

func TestSOQLHandler_Validate_SyntaxErrorHasPosition(t *testing.T) {
	t.Parallel()

	metadata := newTestSOQLMetadata()
	r := setupSOQLRouter(t, metadata)

	body, _ := json.Marshal(map[string]interface{}{
		"query": "SELECT FROM Account",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/soql/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp soqlValidateResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.NotEmpty(t, resp.Errors)
}

func TestSOQLHandler_ListObjects(t *testing.T) {
	t.Parallel()

	metadata := newTestSOQLMetadata()
	r := setupSOQLRouter(t, metadata)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/soql/objects", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []soqlObjectItem `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Data)
	assert.Equal(t, "Account", resp.Data[0].APIName)
	assert.Equal(t, "Accounts", resp.Data[0].Label)
}

func TestSOQLHandler_ListFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		objectName string
		wantStatus int
		wantCount  int
	}{
		{
			name:       "returns fields for existing object",
			objectName: "Account",
			wantStatus: http.StatusOK,
			wantCount:  2, // Name, Industry (Id is system field, excluded)
		},
		{
			name:       "returns 404 for unknown object",
			objectName: "Unknown",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			metadata := newTestSOQLMetadata()
			r := setupSOQLRouter(t, metadata)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/soql/objects/"+tt.objectName+"/fields", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Data []soqlFieldItem `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Len(t, resp.Data, tt.wantCount)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
