package soqlModel

// Query represents a parsed SOQL query ready for execution.
type Query struct {
	// Raw is the original SOQL query string.
	Raw string

	// Cursor is the pagination cursor for fetching next page.
	Cursor string
}

// QueryResult represents the result of executing a SOQL query.
type QueryResult struct {
	// TotalSize is the total number of records matching the query.
	// May be approximate for large result sets.
	TotalSize int

	// Done indicates whether all records have been returned.
	// If false, use NextCursor to fetch more records.
	Done bool

	// Records contains the query results.
	// Structure depends on selected fields in the query.
	Records []map[string]any

	// NextCursor is the cursor for fetching the next page.
	// Empty if Done is true.
	NextCursor string
}

// QueryParams contains parameters for query execution.
type QueryParams struct {
	// PageSize is the maximum number of records to return per page.
	PageSize int

	// UserID is the ID of the user executing the query (for RLS).
	UserID int64
}

// DefaultPageSize is the default number of records per page.
const DefaultPageSize = 100

// MaxPageSize is the maximum allowed page size.
const MaxPageSize = 2000

// MinPageSize is the minimum allowed page size.
const MinPageSize = 1
