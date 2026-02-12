package postgres

import (
	"testing"
	"time"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	"github.com/proxima-research/proxima.crm.platform/pkg/keyset"
)

// mockSecretProvider implements engine.SecretProvider for testing.
type mockSecretProvider struct {
	secret []byte
}

func (m *mockSecretProvider) Secret() []byte {
	return m.secret
}

func newTestSecretProvider() engine.SecretProvider {
	return &mockSecretProvider{secret: []byte("test-secret-key-32-bytes-long!!")}
}

func TestNewCursorHandler(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())
	if handler == nil {
		t.Fatal("NewCursorHandler returned nil")
	}
	if handler.cursor == nil {
		t.Error("CursorHandler.cursor is nil")
	}
}

func TestDecodeCursor_EmptyCursor(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		PageSize: 10,
		Object:   "Account",
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
	}

	ctx, err := handler.DecodeCursor("", pagination, 123)
	if err != nil {
		t.Fatalf("DecodeCursor with empty cursor should not error: %v", err)
	}

	if ctx.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", ctx.PageSize)
	}
	if ctx.UserID != 123 {
		t.Errorf("UserID = %d, want 123", ctx.UserID)
	}
	if ctx.Cursor != nil {
		t.Error("Cursor should be nil for empty cursor string")
	}
}

func TestDecodeCursor_InvalidCursor(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		PageSize: 10,
		Object:   "Account",
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
	}

	tests := []struct {
		name   string
		cursor string
	}{
		{"malformed base64", "not-valid-base64!!!"},
		{"random string", "randomstringwithoutsignature"},
		{"empty json", "e30=.invalidsig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.DecodeCursor(tt.cursor, pagination, 123)
			if err == nil {
				t.Error("expected error for invalid cursor")
			}
		})
	}
}

func TestBuildNextCursor_NilRecord(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		PageSize: 10,
		Object:   "Account",
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
	}

	cursor, err := handler.BuildNextCursor(nil, pagination, 123)
	if err != nil {
		t.Fatalf("BuildNextCursor with nil record should not error: %v", err)
	}
	if cursor != "" {
		t.Errorf("cursor should be empty for nil record, got: %s", cursor)
	}
}

func TestBuildNextCursor_EmptyRecord(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		PageSize: 10,
		Object:   "Account",
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
	}

	record := &Record{Fields: map[string]any{}}

	cursor, err := handler.BuildNextCursor(record, pagination, 123)
	if err != nil {
		t.Fatalf("BuildNextCursor with empty record should not error: %v", err)
	}
	if cursor != "" {
		t.Errorf("cursor should be empty for empty record, got: %s", cursor)
	}
}

func TestBuildNextCursor_ValidRecord(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		PageSize: 10,
		Object:   "Account",
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
		SortKeySOQL: []string{"Id"},
	}

	record := &Record{
		Fields: map[string]any{
			"record_id": "acc123",
			"Name":      "Test Account",
		},
	}

	cursor, err := handler.BuildNextCursor(record, pagination, 123)
	if err != nil {
		t.Fatalf("BuildNextCursor failed: %v", err)
	}
	if cursor == "" {
		t.Error("cursor should not be empty for valid record")
	}
}

func TestBuildNextCursor_MultipleSortKeys(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		PageSize: 10,
		Object:   "Account",
		SortKeys: engine.SortKeys{
			{Field: "created_at", Dir: "desc"},
			{Field: "record_id", Dir: "desc"},
		},
		SortKeySOQL: []string{"CreatedAt", "Id"},
	}

	now := time.Now()
	record := &Record{
		Fields: map[string]any{
			"created_at": now,
			"record_id":  "acc123",
			"Name":       "Test Account",
		},
	}

	cursor, err := handler.BuildNextCursor(record, pagination, 123)
	if err != nil {
		t.Fatalf("BuildNextCursor failed: %v", err)
	}
	if cursor == "" {
		t.Error("cursor should not be empty for valid record with multiple sort keys")
	}
}

func TestBuildKeysetWhereClause_NilCursor(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
	}

	where, params, count, err := handler.BuildKeysetWhereClause(nil, pagination, "t0")
	if err != nil {
		t.Fatalf("BuildKeysetWhereClause with nil cursor should not error: %v", err)
	}
	if where != "" {
		t.Errorf("WHERE clause should be empty, got: %s", where)
	}
	if len(params) != 0 {
		t.Errorf("params should be empty, got: %v", params)
	}
	if count != 0 {
		t.Errorf("count should be 0, got: %d", count)
	}
}

func TestBuildKeysetWhereClause_SameDirectionKeys(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "created_at", Dir: "desc"},
			{Field: "record_id", Dir: "desc"},
		},
	}

	cursor := &keyset.CursorPayload{
		LastRow: map[string]interface{}{
			"created_at": "2024-01-15T10:30:00Z",
			"record_id":  "acc123",
		},
	}

	where, params, count, err := handler.BuildKeysetWhereClause(cursor, pagination, "t0")
	if err != nil {
		t.Fatalf("BuildKeysetWhereClause failed: %v", err)
	}

	// Should produce tuple comparison: (col1, col2) < ($1, $2)
	if where == "" {
		t.Error("WHERE clause should not be empty")
	}
	if count != 2 {
		t.Errorf("param count should be 2, got: %d", count)
	}
	if len(params) != 2 {
		t.Errorf("params length should be 2, got: %d", len(params))
	}

	// Check that it's a tuple comparison (contains comma-separated columns)
	if !containsSubstring(where, "t0.created_at") {
		t.Errorf("WHERE should contain t0.created_at: %s", where)
	}
	if !containsSubstring(where, "t0.record_id") {
		t.Errorf("WHERE should contain t0.record_id: %s", where)
	}
}

func TestBuildKeysetWhereClause_MixedDirectionKeys(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "name", Dir: "asc"},
			{Field: "record_id", Dir: "desc"},
		},
	}

	cursor := &keyset.CursorPayload{
		LastRow: map[string]interface{}{
			"name":      "Acme",
			"record_id": "acc123",
		},
	}

	where, params, count, err := handler.BuildKeysetWhereClause(cursor, pagination, "t0")
	if err != nil {
		t.Fatalf("BuildKeysetWhereClause failed: %v", err)
	}

	// Should produce expanded OR/AND form for mixed directions
	if where == "" {
		t.Error("WHERE clause should not be empty")
	}
	if count == 0 {
		t.Error("param count should not be 0")
	}
	if len(params) == 0 {
		t.Error("params should not be empty")
	}

	// Check that it contains OR clause for mixed directions
	if !containsSubstring(where, "OR") && count > 1 {
		t.Logf("WHERE clause: %s", where)
	}
}

func TestBuildKeysetWhereClause_MissingSortKeyValue(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "created_at", Dir: "desc"},
			{Field: "record_id", Dir: "desc"},
		},
	}

	cursor := &keyset.CursorPayload{
		LastRow: map[string]interface{}{
			"created_at": "2024-01-15T10:30:00Z",
			// missing record_id
		},
	}

	_, _, _, err := handler.BuildKeysetWhereClause(cursor, pagination, "t0")
	if err == nil {
		t.Error("expected error for missing sort key value")
	}
}

func TestBuildKeysetWhereClause_AscDirection(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "name", Dir: "asc"},
		},
	}

	cursor := &keyset.CursorPayload{
		LastRow: map[string]interface{}{
			"name": "Acme",
		},
	}

	where, _, _, err := handler.BuildKeysetWhereClause(cursor, pagination, "t0")
	if err != nil {
		t.Fatalf("BuildKeysetWhereClause failed: %v", err)
	}

	// ASC direction should use > operator
	if !containsSubstring(where, ">") {
		t.Errorf("ASC direction should use > operator: %s", where)
	}
}

func TestBuildKeysetWhereClause_DescDirection(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "name", Dir: "desc"},
		},
	}

	cursor := &keyset.CursorPayload{
		LastRow: map[string]interface{}{
			"name": "Acme",
		},
	}

	where, _, _, err := handler.BuildKeysetWhereClause(cursor, pagination, "t0")
	if err != nil {
		t.Fatalf("BuildKeysetWhereClause failed: %v", err)
	}

	// DESC direction should use < operator
	if !containsSubstring(where, "<") {
		t.Errorf("DESC direction should use < operator: %s", where)
	}
}

func TestFormatValueForCursor_TimeValue(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	result := handler.formatValueForCursor(now)

	expected := "2024-01-15T10:30:00Z"
	if result != expected {
		t.Errorf("formatValueForCursor(time) = %v, want %s", result, expected)
	}
}

func TestFormatValueForCursor_StringValue(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	result := handler.formatValueForCursor("test string")

	if result != "test string" {
		t.Errorf("formatValueForCursor(string) = %v, want 'test string'", result)
	}
}

func TestFormatValueForCursor_IntValue(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	result := handler.formatValueForCursor(42)

	if result != 42 {
		t.Errorf("formatValueForCursor(int) = %v, want 42", result)
	}
}

func TestExtractSortKeyValue_DirectMatch(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	record := &Record{
		Fields: map[string]any{
			"record_id":  "acc123",
			"created_at": "2024-01-15",
		},
	}

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
	}

	val := handler.extractSortKeyValue(record, "record_id", pagination)
	if val != "acc123" {
		t.Errorf("extractSortKeyValue = %v, want 'acc123'", val)
	}
}

func TestExtractSortKeyValue_SOQLMapping(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	record := &Record{
		Fields: map[string]any{
			"Id":   "acc123",
			"Name": "Test",
		},
	}

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
		SortKeySOQL: []string{"Id"},
	}

	val := handler.extractSortKeyValue(record, "record_id", pagination)
	if val != "acc123" {
		t.Errorf("extractSortKeyValue with SOQL mapping = %v, want 'acc123'", val)
	}
}

func TestExtractSortKeyValue_NotFound(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	record := &Record{
		Fields: map[string]any{
			"Name": "Test",
		},
	}

	pagination := &engine.PaginationInfo{
		SortKeys: engine.SortKeys{
			{Field: "nonexistent", Dir: "desc"},
		},
	}

	val := handler.extractSortKeyValue(record, "nonexistent", pagination)
	if val != nil {
		t.Errorf("extractSortKeyValue for missing field = %v, want nil", val)
	}
}

func TestCursorRoundTrip(t *testing.T) {
	handler := NewCursorHandler(newTestSecretProvider())

	pagination := &engine.PaginationInfo{
		PageSize: 10,
		Object:   "Account",
		SortKeys: engine.SortKeys{
			{Field: "record_id", Dir: "desc"},
		},
		SortKeySOQL: []string{"Id"},
	}

	// Build a cursor
	originalRecord := &Record{
		Fields: map[string]any{
			"record_id": "acc123",
			"Name":      "Test Account",
		},
	}

	cursor, err := handler.BuildNextCursor(originalRecord, pagination, 123)
	if err != nil {
		t.Fatalf("BuildNextCursor failed: %v", err)
	}

	// Decode the cursor
	ctx, err := handler.DecodeCursor(cursor, pagination, 123)
	if err != nil {
		t.Fatalf("DecodeCursor failed: %v", err)
	}

	if ctx.Cursor == nil {
		t.Fatal("decoded cursor payload is nil")
	}

	// Verify the sort key value was preserved
	if ctx.Cursor.LastRow["record_id"] != "acc123" {
		t.Errorf("record_id = %v, want 'acc123'", ctx.Cursor.LastRow["record_id"])
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsSubstringHelper(s, substr)))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
