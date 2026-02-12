// Package executor provides execution of compiled SOQL queries against a database.
package postgres

import (
	"context"
	"database/sql"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
)

// DB is a minimal interface for database operations.
// It is compatible with *sql.DB and *sql.Tx.
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// ExecuteParams contains parameters for query execution.
type ExecuteParams struct {
	// Cursor is the pagination cursor (empty for first page).
	Cursor string

	// PageSize is the number of records per page.
	PageSize int

	// UserID is the user ID for cursor FID generation.
	UserID int64
}

// Executor executes compiled SOQL queries and returns results.
type Executor interface {
	// Execute executes a compiled query and returns the result.
	Execute(ctx context.Context, query *engine.CompiledQuery) (*QueryResult, error)

	// ExecuteWithParams executes a query with pagination parameters.
	ExecuteWithParams(ctx context.Context, query *engine.CompiledQuery, params *ExecuteParams) (*QueryResult, error)

	// ExecuteWithDB executes a query using the provided database connection.
	// Useful for executing within a transaction.
	ExecuteWithDB(ctx context.Context, db DB, query *engine.CompiledQuery) (*QueryResult, error)
}

// QueryResult represents the result of a SOQL query execution.
type QueryResult struct {
	// Records contains the query results.
	Records []Record

	// TotalSize is the total number of records matching the query.
	// May be greater than len(Records) if LIMIT was applied.
	TotalSize int

	// Done indicates whether all matching records have been returned.
	// False if there are more records available (pagination).
	Done bool

	// NextCursor is the cursor for fetching the next page.
	// Empty if Done is true.
	NextCursor string
}

// Record represents a single record in the query result.
type Record struct {
	// Type is the SOQL object type (e.g., "Account", "Contact").
	Type string

	// Fields contains the field values keyed by field name/alias.
	Fields map[string]any

	// Relationships contains nested records from Parent-to-Child subqueries.
	// Key is the relationship name (e.g., "Contacts").
	Relationships map[string][]Record
}

// Get returns a field value by name.
// Returns nil if the field doesn't exist.
func (r *Record) Get(name string) any {
	if r.Fields == nil {
		return nil
	}
	return r.Fields[name]
}

// GetString returns a field value as string.
// Returns empty string if the field doesn't exist or is not a string.
func (r *Record) GetString(name string) string {
	v := r.Get(name)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// GetInt returns a field value as int64.
// Returns 0 if the field doesn't exist or is not numeric.
func (r *Record) GetInt(name string) int64 {
	v := r.Get(name)
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	}
	return 0
}

// GetFloat returns a field value as float64.
// Returns 0 if the field doesn't exist or is not numeric.
func (r *Record) GetFloat(name string) float64 {
	v := r.Get(name)
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	}
	return 0
}

// GetBool returns a field value as bool.
// Returns false if the field doesn't exist or is not a bool.
func (r *Record) GetBool(name string) bool {
	v := r.Get(name)
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// GetRelationship returns nested records for a relationship.
// Returns nil if the relationship doesn't exist.
func (r *Record) GetRelationship(name string) []Record {
	if r.Relationships == nil {
		return nil
	}
	return r.Relationships[name]
}

// IsNull returns true if the field value is nil/NULL.
func (r *Record) IsNull(name string) bool {
	if r.Fields == nil {
		return true
	}
	_, exists := r.Fields[name]
	if !exists {
		return true
	}
	return r.Fields[name] == nil
}
