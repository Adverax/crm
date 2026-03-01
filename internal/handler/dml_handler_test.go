package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmlengine "github.com/adverax/crm/internal/platform/dml/engine"
)

func newTestDMLMetadata() *dmlengine.StaticMetadataProvider {
	return dmlengine.NewStaticMetadataProvider(map[string]*dmlengine.ObjectMeta{
		"Account": dmlengine.NewObjectMeta("Account", "obj_account").
			Schema("public").
			PrimaryKey("id").
			RequiredField("Name", "name", dmlengine.FieldTypeString).
			Field("Industry", "industry", dmlengine.FieldTypeString).
			ReadOnlyField("Id", "id", dmlengine.FieldTypeID).
			Build(),
	})
}

func setupDMLRouter(t *testing.T, engineMeta dmlengine.MetadataProvider) *gin.Engine {
	t.Helper()
	return setupDMLRouterFull(t, engineMeta, newTestCRMMetadata())
}

func setupDMLRouterFull(t *testing.T, engineMeta dmlengine.MetadataProvider, crmMeta *testMetadataReader) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(contractValidationMiddleware(t))
	admin := r.Group("/api/v1/admin")
	eng := dmlengine.NewEngine(
		dmlengine.WithMetadata(engineMeta),
	)
	h := NewDMLHandler(eng, nil, crmMeta)
	h.RegisterRoutes(admin)
	return r
}

func TestDMLHandler_Validate(t *testing.T) {
	t.Parallel()

	meta := newTestDMLMetadata()

	tests := []struct {
		name          string
		body          interface{}
		wantStatus    int
		wantValid     *bool
		wantOperation *string
		wantObject    *string
	}{
		{
			name: "validates valid INSERT statement",
			body: map[string]interface{}{
				"statement": "INSERT INTO Account (Name) VALUES ('Acme')",
			},
			wantStatus:    http.StatusOK,
			wantValid:     boolPtr(true),
			wantOperation: stringPtr("INSERT"),
			wantObject:    stringPtr("Account"),
		},
		{
			name: "validates valid UPDATE statement",
			body: map[string]interface{}{
				"statement": "UPDATE Account SET Name = 'Acme' WHERE Id = '123'",
			},
			wantStatus:    http.StatusOK,
			wantValid:     boolPtr(true),
			wantOperation: stringPtr("UPDATE"),
			wantObject:    stringPtr("Account"),
		},
		{
			name: "validates valid DELETE statement",
			body: map[string]interface{}{
				"statement": "DELETE FROM Account WHERE Id = '123'",
			},
			wantStatus:    http.StatusOK,
			wantValid:     boolPtr(true),
			wantOperation: stringPtr("DELETE"),
			wantObject:    stringPtr("Account"),
		},
		{
			name: "detects syntax error",
			body: map[string]interface{}{
				"statement": "INSRT INTO Account (Name) VALUES ('Acme')",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(false),
		},
		{
			name: "detects unknown object",
			body: map[string]interface{}{
				"statement": "INSERT INTO UnknownObj (Name) VALUES ('x')",
			},
			wantStatus: http.StatusOK,
			wantValid:  boolPtr(false),
		},
		{
			name: "detects unknown field",
			body: map[string]interface{}{
				"statement": "INSERT INTO Account (NonExistent) VALUES ('x')",
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
			name: "returns 400 for empty statement",
			body: map[string]interface{}{
				"statement": "",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for missing statement field",
			body: map[string]interface{}{
				"other": "value",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := setupDMLRouter(t, meta)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/dml/validate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantValid != nil {
				var resp dmlValidateResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, *tt.wantValid, resp.Valid)

				if tt.wantObject != nil {
					require.NotNil(t, resp.Object)
					assert.Equal(t, *tt.wantObject, *resp.Object)
				}

				if tt.wantOperation != nil {
					require.NotNil(t, resp.Operation)
					assert.Equal(t, *tt.wantOperation, *resp.Operation)
				}

				if !resp.Valid {
					assert.NotEmpty(t, resp.Errors, "invalid response should have errors")
				}

				if resp.Valid {
					assert.NotNil(t, resp.SQL, "valid response should have SQL")
				}
			}
		})
	}
}

func TestDMLHandler_Validate_ReturnsFields(t *testing.T) {
	t.Parallel()

	meta := newTestDMLMetadata()
	r := setupDMLRouter(t, meta)

	body, _ := json.Marshal(map[string]interface{}{
		"statement": "INSERT INTO Account (Name, Industry) VALUES ('Acme', 'Tech')",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/dml/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dmlValidateResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.NotNil(t, resp.Object)
	assert.Equal(t, "Account", *resp.Object)
	assert.NotEmpty(t, resp.Fields)
	assert.Contains(t, resp.Fields, "Name")
	assert.Contains(t, resp.Fields, "Industry")
}

func TestDMLHandler_ListObjects(t *testing.T) {
	t.Parallel()

	meta := newTestDMLMetadata()
	r := setupDMLRouter(t, meta)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/dml/objects", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []dmlObjectItem `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Data)
	assert.Equal(t, "Account", resp.Data[0].APIName)
	assert.Equal(t, "Accounts", resp.Data[0].Label)
}

func TestDMLHandler_ListFields(t *testing.T) {
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

			meta := newTestDMLMetadata()
			r := setupDMLRouter(t, meta)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/dml/objects/"+tt.objectName+"/fields", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Data []dmlFieldItem `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Len(t, resp.Data, tt.wantCount)
			}
		})
	}
}
