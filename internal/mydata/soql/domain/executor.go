package soqlModel

import (
	"context"
	"errors"
)

// QueryExecutor executes SOQL queries and returns results.
type QueryExecutor interface {
	// Execute runs a SOQL query and returns the result.
	// The query parameter contains the parsed query and optional cursor.
	// The params parameter contains execution parameters like page size and user context.
	Execute(ctx context.Context, query *Query, params *QueryParams) (*QueryResult, error)
}

// Common errors returned by QueryExecutor.
var (
	// ErrInvalidQuery is returned when the SOQL query syntax is invalid.
	ErrInvalidQuery = errors.New("invalid SOQL query syntax")

	// ErrSemanticError is returned when the query has semantic errors
	// (e.g., unknown field, invalid object).
	ErrSemanticError = errors.New("SOQL semantic error")

	// ErrQueryTooComplex is returned when the query exceeds complexity limits.
	ErrQueryTooComplex = errors.New("query too complex")

	// ErrInvalidCursor is returned when the pagination cursor is invalid.
	ErrInvalidCursor = errors.New("invalid pagination cursor")
)
