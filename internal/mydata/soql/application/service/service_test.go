package soqlService

import (
	"context"
	"errors"
	"testing"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	soqlModel "github.com/proxima-research/proxima.crm.platform/internal/data/soql/domain"
	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/infrastructure/postgres"
)

// mockPostgresExecutor is a configurable mock implementation of postgres.Executor.
type mockPostgresExecutor struct {
	result *postgres.QueryResult
	err    error

	// Captured for verification
	capturedQuery *engine.CompiledQuery
}

func (m *mockPostgresExecutor) Execute(_ context.Context, query *engine.CompiledQuery) (*postgres.QueryResult, error) {
	m.capturedQuery = query
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *mockPostgresExecutor) ExecuteWithParams(_ context.Context, query *engine.CompiledQuery, _ *postgres.ExecuteParams) (*postgres.QueryResult, error) {
	return m.Execute(context.Background(), query)
}

func (m *mockPostgresExecutor) ExecuteWithDB(_ context.Context, _ postgres.DB, query *engine.CompiledQuery) (*postgres.QueryResult, error) {
	return m.Execute(context.Background(), query)
}

func createTestEngine() *engine.Engine {
	// Create a minimal metadata provider with Account object for testing
	objects := map[string]*engine.ObjectMeta{
		"Account": engine.NewObjectMeta("Account", "", "accounts").
			Field("Id", "id", engine.FieldTypeID).
			Field("Name", "name", engine.FieldTypeString).
			Field("Industry", "industry", engine.FieldTypeString).
			Relationship("Contacts", "Contact", "AccountId", "Id").
			Build(),
		"Contact": engine.NewObjectMeta("Contact", "", "contacts").
			Field("Id", "id", engine.FieldTypeID).
			Field("FirstName", "first_name", engine.FieldTypeString).
			Field("AccountId", "account_id", engine.FieldTypeID).
			Lookup("Account", "AccountId", "Account", "Id").
			Build(),
	}
	metadata := engine.NewStaticMetadataProvider(objects)
	return engine.NewEngine(engine.WithMetadata(metadata))
}

func TestQueryService_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_ReturnsResult", func(t *testing.T) {
		executor := &mockPostgresExecutor{
			result: &postgres.QueryResult{
				TotalSize: 10,
				Done:      true,
				Records: []postgres.Record{
					{Fields: map[string]any{"Id": "001", "Name": "Test"}},
				},
			},
		}
		service := NewQueryService(createTestEngine(), executor)

		result, err := service.Execute(ctx, "SELECT Id, Name FROM Account", "", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.TotalSize != 10 {
			t.Errorf("TotalSize = %d, want 10", result.TotalSize)
		}
		if result.Done != true {
			t.Errorf("Done = %v, want true", result.Done)
		}
		if len(result.Records) != 1 {
			t.Errorf("Records count = %d, want 1", len(result.Records))
		}
	})

	t.Run("ExecutorError_ReturnsError", func(t *testing.T) {
		expectedErr := errors.New("query execution failed")
		executor := &mockPostgresExecutor{err: expectedErr}
		service := NewQueryService(createTestEngine(), executor)

		_, err := service.Execute(ctx, "SELECT Id FROM Account", "", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("EmptyQuery_ReturnsEmptyResult", func(t *testing.T) {
		executor := &mockPostgresExecutor{result: &postgres.QueryResult{}}
		service := NewQueryService(createTestEngine(), executor)

		result, err := service.Execute(ctx, "", "", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.TotalSize != 0 {
			t.Errorf("TotalSize = %d, want 0", result.TotalSize)
		}
		if !result.Done {
			t.Errorf("Done = %v, want true", result.Done)
		}
		if len(result.Records) != 0 {
			t.Errorf("Records count = %d, want 0", len(result.Records))
		}
	})

	t.Run("NilParams_UsesDefaultPageSize", func(t *testing.T) {
		executor := &mockPostgresExecutor{result: &postgres.QueryResult{}}
		service := NewQueryService(createTestEngine(), executor)

		_, err := service.Execute(ctx, "SELECT Id FROM Account", "", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Service normalizes params internally - we just verify it doesn't error
	})

	t.Run("ZeroPageSize_UsesDefaultPageSize", func(t *testing.T) {
		executor := &mockPostgresExecutor{result: &postgres.QueryResult{}}
		service := NewQueryService(createTestEngine(), executor)

		params := &soqlModel.QueryParams{PageSize: 0}
		_, err := service.Execute(ctx, "SELECT Id FROM Account", "", params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("NegativePageSize_UsesDefaultPageSize", func(t *testing.T) {
		executor := &mockPostgresExecutor{result: &postgres.QueryResult{}}
		service := NewQueryService(createTestEngine(), executor)

		params := &soqlModel.QueryParams{PageSize: -10}
		_, err := service.Execute(ctx, "SELECT Id FROM Account", "", params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ExceedMaxPageSize_CapsToMaxPageSize", func(t *testing.T) {
		executor := &mockPostgresExecutor{result: &postgres.QueryResult{}}
		service := NewQueryService(createTestEngine(), executor)

		params := &soqlModel.QueryParams{PageSize: 10000}
		_, err := service.Execute(ctx, "SELECT Id FROM Account", "", params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ValidPageSize_Preserved", func(t *testing.T) {
		executor := &mockPostgresExecutor{result: &postgres.QueryResult{}}
		service := NewQueryService(createTestEngine(), executor)

		params := &soqlModel.QueryParams{PageSize: 50}
		_, err := service.Execute(ctx, "SELECT Id FROM Account", "", params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("InvalidQuery_ReturnsInvalidQueryError", func(t *testing.T) {
		executor := &mockPostgresExecutor{result: &postgres.QueryResult{}}
		service := NewQueryService(createTestEngine(), executor)

		_, err := service.Execute(ctx, "INVALID QUERY SYNTAX", "", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, soqlModel.ErrInvalidQuery) {
			t.Errorf("err = %v, want ErrInvalidQuery", err)
		}
	})

	t.Run("RecordsWithRelationships_ConvertedCorrectly", func(t *testing.T) {
		executor := &mockPostgresExecutor{
			result: &postgres.QueryResult{
				TotalSize: 1,
				Done:      true,
				Records: []postgres.Record{
					{
						Fields: map[string]any{"Id": "001", "Name": "Account1"},
						Relationships: map[string][]postgres.Record{
							"Contacts": {
								{Fields: map[string]any{"Id": "002", "FirstName": "John"}},
								{Fields: map[string]any{"Id": "003", "FirstName": "Jane"}},
							},
						},
					},
				},
			},
		}
		service := NewQueryService(createTestEngine(), executor)

		result, err := service.Execute(ctx, "SELECT Id, Name, (SELECT Id, FirstName FROM Contacts) FROM Account", "", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Records) != 1 {
			t.Fatalf("Records count = %d, want 1", len(result.Records))
		}

		record := result.Records[0]
		contacts, ok := record["Contacts"].([]map[string]any)
		if !ok {
			t.Fatalf("Contacts is not []map[string]any")
		}
		if len(contacts) != 2 {
			t.Errorf("Contacts count = %d, want 2", len(contacts))
		}
	})
}

func TestConvertResult(t *testing.T) {
	t.Run("NilResult_ReturnsEmptyResult", func(t *testing.T) {
		result := convertResult(nil)
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.TotalSize != 0 {
			t.Errorf("TotalSize = %d, want 0", result.TotalSize)
		}
		if !result.Done {
			t.Errorf("Done = %v, want true", result.Done)
		}
		if len(result.Records) != 0 {
			t.Errorf("Records count = %d, want 0", len(result.Records))
		}
	})
}

func TestConvertRecord(t *testing.T) {
	t.Run("NilRecord_ReturnsNil", func(t *testing.T) {
		result := convertRecord(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("WithFields_CopiesFields", func(t *testing.T) {
		rec := &postgres.Record{
			Fields: map[string]any{
				"Id":   "001",
				"Name": "Test",
			},
		}
		result := convertRecord(rec)
		if result["Id"] != "001" {
			t.Errorf("Id = %v, want 001", result["Id"])
		}
		if result["Name"] != "Test" {
			t.Errorf("Name = %v, want Test", result["Name"])
		}
	})
}

func TestMapError(t *testing.T) {
	t.Run("NilError_ReturnsNil", func(t *testing.T) {
		err := mapError(nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("UnknownError_PassesThrough", func(t *testing.T) {
		originalErr := errors.New("unknown error")
		err := mapError(originalErr)
		if err != originalErr {
			t.Errorf("expected original error to pass through")
		}
	})
}
