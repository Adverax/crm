package soql

// QueryResult represents the result of executing a SOQL query.
type QueryResult struct {
	// TotalSize is the total number of records matching the query.
	TotalSize int `json:"totalSize"`

	// Done indicates whether all records have been returned.
	Done bool `json:"done"`

	// Records contains the query results as maps.
	Records []map[string]any `json:"records"`

	// NextCursor is the cursor for fetching the next page.
	// Empty if Done is true.
	NextCursor string `json:"nextRecordsUrl,omitempty"`
}

// QueryParams contains parameters for query execution.
type QueryParams struct {
	// PageSize overrides the LIMIT in the query.
	PageSize int
}

const (
	DefaultPageSize = 100
	MaxPageSize     = 2000
	MinPageSize     = 1
)
