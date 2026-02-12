package soqlHttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	platformApi "github.com/proxima-research/proxima.crm.platform/api/openapi/platform"
	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/domain"
)

// mockQueryService is a configurable mock implementation of service.QueryService.
type mockQueryService struct {
	result *soqlModel.QueryResult
	err    error

	// Captured arguments for verification
	capturedQuery  string
	capturedCursor string
	capturedParams *soqlModel.QueryParams
}

func (m *mockQueryService) Execute(_ context.Context, query string, cursor string, params *soqlModel.QueryParams) (*soqlModel.QueryResult, error) {
	m.capturedQuery = query
	m.capturedCursor = cursor
	m.capturedParams = params
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestSOQLApi_ExecuteSOQLQuery_POST(t *testing.T) {
	t.Run("Success_ReturnsQueryResult", func(t *testing.T) {
		mockService := &mockQueryService{
			result: &soqlModel.QueryResult{
				TotalSize: 2,
				Done:      true,
				Records: []map[string]any{
					{"Id": "001", "Name": "Acme"},
					{"Id": "002", "Name": "Beta"},
				},
			},
		}

		api := New(mockService)

		reqBody := platformApi.SOQLQueryRequest{
			Query: "SELECT Id, Name FROM Account",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/data/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		api.ExecuteSOQLQuery(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var response platformApi.SOQLQueryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.TotalSize != 2 {
			t.Errorf("TotalSize = %d, want 2", response.TotalSize)
		}
		if !response.Done {
			t.Error("Done = false, want true")
		}
		if len(response.Records) != 2 {
			t.Errorf("Records count = %d, want 2", len(response.Records))
		}
	})

	t.Run("WithCursor_PassesCursorToService", func(t *testing.T) {
		mockService := &mockQueryService{
			result: &soqlModel.QueryResult{Done: true, Records: []map[string]any{}},
		}

		api := New(mockService)

		cursor := "eyJ2IjoxfQ=="
		reqBody := platformApi.SOQLQueryRequest{
			Query:  "SELECT Id FROM Account",
			Cursor: &cursor,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/data/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		api.ExecuteSOQLQuery(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if mockService.capturedCursor != cursor {
			t.Errorf("cursor = %q, want %q", mockService.capturedCursor, cursor)
		}
	})

	t.Run("InvalidQueryError_Returns400", func(t *testing.T) {
		mockService := &mockQueryService{err: soqlModel.ErrInvalidQuery}
		api := New(mockService)

		reqBody := platformApi.SOQLQueryRequest{Query: "INVALID QUERY"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/data/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		api.ExecuteSOQLQuery(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("SemanticError_Returns422", func(t *testing.T) {
		mockService := &mockQueryService{err: soqlModel.ErrSemanticError}
		api := New(mockService)

		reqBody := platformApi.SOQLQueryRequest{Query: "SELECT Unknown FROM Account"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/data/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		api.ExecuteSOQLQuery(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
		}
	})

	t.Run("InvalidCursorError_Returns400", func(t *testing.T) {
		mockService := &mockQueryService{err: soqlModel.ErrInvalidCursor}
		api := New(mockService)

		cursor := "invalid-cursor"
		reqBody := platformApi.SOQLQueryRequest{
			Query:  "SELECT Id FROM Account",
			Cursor: &cursor,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/data/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		api.ExecuteSOQLQuery(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("InternalError_Returns500", func(t *testing.T) {
		mockService := &mockQueryService{err: errors.New("database connection failed")}
		api := New(mockService)

		reqBody := platformApi.SOQLQueryRequest{Query: "SELECT Id FROM Account"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/data/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		api.ExecuteSOQLQuery(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("NextCursor_IncludedInResponse", func(t *testing.T) {
		nextCursor := "eyJ2IjoxLCJvYiI6W119"
		mockService := &mockQueryService{
			result: &soqlModel.QueryResult{
				TotalSize:  100,
				Done:       false,
				Records:    []map[string]any{{"Id": "001"}},
				NextCursor: nextCursor,
			},
		}

		api := New(mockService)

		reqBody := platformApi.SOQLQueryRequest{Query: "SELECT Id FROM Account"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/data/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		api.ExecuteSOQLQuery(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var response platformApi.SOQLQueryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Done {
			t.Error("Done = true, want false")
		}
		if response.NextCursor == nil || *response.NextCursor != nextCursor {
			t.Errorf("NextCursor = %v, want %q", response.NextCursor, nextCursor)
		}
	})
}

func TestSOQLApi_ExecuteSOQLQueryGet(t *testing.T) {
	t.Run("Success_ReturnsQueryResult", func(t *testing.T) {
		mockService := &mockQueryService{
			result: &soqlModel.QueryResult{
				TotalSize: 1,
				Done:      true,
				Records:   []map[string]any{{"Id": "001"}},
			},
		}

		api := New(mockService)

		req := httptest.NewRequest(http.MethodGet, "/data/query?q=SELECT+Id+FROM+Account", nil)
		rec := httptest.NewRecorder()

		params := platformApi.ExecuteSOQLQueryGetParams{
			Q: "SELECT Id FROM Account",
		}
		api.ExecuteSOQLQueryGet(rec, req, params)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		if mockService.capturedQuery != "SELECT Id FROM Account" {
			t.Errorf("query = %q, want %q", mockService.capturedQuery, "SELECT Id FROM Account")
		}
	})

	t.Run("WithCursor_PassesCursorToService", func(t *testing.T) {
		mockService := &mockQueryService{
			result: &soqlModel.QueryResult{Done: true, Records: []map[string]any{}},
		}

		api := New(mockService)
		cursor := "eyJ2IjoxfQ=="

		req := httptest.NewRequest(http.MethodGet, "/data/query?q=SELECT+Id+FROM+Account&cursor="+cursor, nil)
		rec := httptest.NewRecorder()

		params := platformApi.ExecuteSOQLQueryGetParams{
			Q:      "SELECT Id FROM Account",
			Cursor: &cursor,
		}
		api.ExecuteSOQLQueryGet(rec, req, params)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if mockService.capturedCursor != cursor {
			t.Errorf("cursor = %q, want %q", mockService.capturedCursor, cursor)
		}
	})

	t.Run("InvalidQueryError_Returns400", func(t *testing.T) {
		mockService := &mockQueryService{err: soqlModel.ErrInvalidQuery}
		api := New(mockService)

		req := httptest.NewRequest(http.MethodGet, "/data/query?q=INVALID", nil)
		rec := httptest.NewRecorder()

		params := platformApi.ExecuteSOQLQueryGetParams{Q: "INVALID"}
		api.ExecuteSOQLQueryGet(rec, req, params)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("QueryTooComplexError_Returns400", func(t *testing.T) {
		mockService := &mockQueryService{err: soqlModel.ErrQueryTooComplex}
		api := New(mockService)

		req := httptest.NewRequest(http.MethodGet, "/data/query?q=SELECT+...", nil)
		rec := httptest.NewRecorder()

		params := platformApi.ExecuteSOQLQueryGetParams{Q: "SELECT ..."}
		api.ExecuteSOQLQueryGet(rec, req, params)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestMapResultToResponse(t *testing.T) {
	t.Run("EmptyRecords_ReturnsEmptyArray", func(t *testing.T) {
		result := &soqlModel.QueryResult{
			TotalSize: 0,
			Done:      true,
			Records:   []map[string]any{},
		}

		response := mapResultToResponse(result)

		if response.TotalSize != 0 {
			t.Errorf("TotalSize = %d, want 0", response.TotalSize)
		}
		if len(response.Records) != 0 {
			t.Errorf("Records count = %d, want 0", len(response.Records))
		}
		if response.NextCursor != nil {
			t.Errorf("NextCursor = %v, want nil", response.NextCursor)
		}
	})

	t.Run("WithRecords_MapsCorrectly", func(t *testing.T) {
		result := &soqlModel.QueryResult{
			TotalSize: 2,
			Done:      true,
			Records: []map[string]any{
				{"Id": "001", "Name": "Test1"},
				{"Id": "002", "Name": "Test2"},
			},
		}

		response := mapResultToResponse(result)

		if len(response.Records) != 2 {
			t.Fatalf("Records count = %d, want 2", len(response.Records))
		}
		if response.Records[0]["Id"] != "001" {
			t.Errorf("Records[0].Id = %v, want '001'", response.Records[0]["Id"])
		}
		if response.Records[1]["Name"] != "Test2" {
			t.Errorf("Records[1].Name = %v, want 'Test2'", response.Records[1]["Name"])
		}
	})

	t.Run("WithNextCursor_IncludesInResponse", func(t *testing.T) {
		cursor := "nextpage123"
		result := &soqlModel.QueryResult{
			TotalSize:  100,
			Done:       false,
			Records:    []map[string]any{},
			NextCursor: cursor,
		}

		response := mapResultToResponse(result)

		if response.NextCursor == nil {
			t.Fatal("NextCursor is nil, want value")
		}
		if *response.NextCursor != cursor {
			t.Errorf("NextCursor = %q, want %q", *response.NextCursor, cursor)
		}
	})

	t.Run("EmptyNextCursor_IsNil", func(t *testing.T) {
		result := &soqlModel.QueryResult{
			TotalSize:  10,
			Done:       true,
			Records:    []map[string]any{},
			NextCursor: "",
		}

		response := mapResultToResponse(result)

		if response.NextCursor != nil {
			t.Errorf("NextCursor = %v, want nil", response.NextCursor)
		}
	})
}

func TestMapErrorToSOQLError(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedCode int
		expectedType string
	}{
		{"InvalidQuery", soqlModel.ErrInvalidQuery, http.StatusBadRequest, "ParseError"},
		{"SemanticError", soqlModel.ErrSemanticError, http.StatusUnprocessableEntity, "ValidationError"},
		{"QueryTooComplex", soqlModel.ErrQueryTooComplex, http.StatusBadRequest, "LimitError"},
		{"InvalidCursor", soqlModel.ErrInvalidCursor, http.StatusBadRequest, "ParseError"},
		{"UnknownError", errors.New("unknown"), http.StatusInternalServerError, "ExecutionError"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, resp := mapErrorToSOQLError(tc.err)
			if code != tc.expectedCode {
				t.Errorf("code = %d, want %d", code, tc.expectedCode)
			}
			if string(resp.ErrorType) != tc.expectedType {
				t.Errorf("errorType = %s, want %s", resp.ErrorType, tc.expectedType)
			}
		})
	}
}
