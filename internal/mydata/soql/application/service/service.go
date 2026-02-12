package soqlService

import (
	"context"
	"errors"
	"fmt"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	soqlModel "github.com/proxima-research/proxima.crm.platform/internal/data/soql/domain"
	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/infrastructure/postgres"
)

// QueryService provides SOQL query execution capabilities.
type QueryService interface {
	// Execute runs a SOQL query and returns the result.
	Execute(ctx context.Context, query string, cursor string, params *soqlModel.QueryParams) (*soqlModel.QueryResult, error)
}

// queryService implements QueryService.
type queryService struct {
	engine   *engine.Engine
	executor postgres.Executor
}

// NewQueryService creates a new QueryService with the given engine and executor.
func NewQueryService(eng *engine.Engine, exec postgres.Executor) QueryService {
	return &queryService{
		engine:   eng,
		executor: exec,
	}
}

// Execute runs a SOQL query and returns the result.
func (s *queryService) Execute(ctx context.Context, queryStr string, cursor string, params *soqlModel.QueryParams) (*soqlModel.QueryResult, error) {
	// Validate and normalize parameters
	if params == nil {
		params = &soqlModel.QueryParams{}
	}

	// Note: PageSize <= 0 means "use LIMIT from SOQL query" (handled by executor)
	// Only cap if explicitly provided and exceeds max
	if params.PageSize > soqlModel.MaxPageSize {
		params.PageSize = soqlModel.MaxPageSize
	}

	// Handle empty query
	if queryStr == "" {
		return &soqlModel.QueryResult{
			TotalSize: 0,
			Done:      true,
			Records:   []map[string]any{},
		}, nil
	}

	// Prepare query (Parse → Validate → Compile)
	compiled, err := s.engine.PrepareAndResolve(ctx, queryStr)
	if err != nil {
		return nil, mapError(err)
	}

	// Build execution params with cursor and pagination
	execParams := &postgres.ExecuteParams{
		Cursor:   cursor,
		PageSize: params.PageSize,
		UserID:   params.UserID,
	}

	// Execute query with pagination support
	result, err := s.executor.ExecuteWithParams(ctx, compiled, execParams)
	if err != nil {
		return nil, mapError(err)
	}

	// Convert result
	return convertResult(result), nil
}

// convertResult converts postgres.QueryResult to domain.QueryResult.
func convertResult(result *postgres.QueryResult) *soqlModel.QueryResult {
	if result == nil {
		return &soqlModel.QueryResult{
			TotalSize: 0,
			Done:      true,
			Records:   []map[string]any{},
		}
	}

	records := make([]map[string]any, 0, len(result.Records))
	for _, rec := range result.Records {
		record := convertRecord(&rec)
		records = append(records, record)
	}

	return &soqlModel.QueryResult{
		TotalSize:  result.TotalSize,
		Done:       result.Done,
		Records:    records,
		NextCursor: result.NextCursor,
	}
}

// convertRecord converts postgres.Record to map[string]any.
func convertRecord(rec *postgres.Record) map[string]any {
	if rec == nil {
		return nil
	}

	result := make(map[string]any, len(rec.Fields)+len(rec.Relationships))

	// Copy fields
	for k, v := range rec.Fields {
		result[k] = v
	}

	// Convert relationships to nested arrays
	for relName, relRecords := range rec.Relationships {
		nested := make([]map[string]any, 0, len(relRecords))
		for i := range relRecords {
			nested = append(nested, convertRecord(&relRecords[i]))
		}
		result[relName] = nested
	}

	return result
}

// mapError maps engine errors to domain errors while preserving the original error.
// The original error is wrapped so that errors.As can still extract detailed info.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	// Check for pagination/cursor errors
	var validationErr *engine.ValidationError
	if engine.IsValidationError(err) {
		if ok := errors.As(err, &validationErr); ok && validationErr.Code == engine.ErrCodeInvalidPagination {
			return fmt.Errorf("%w: %w", soqlModel.ErrInvalidCursor, err)
		}
		return fmt.Errorf("%w: %w", soqlModel.ErrSemanticError, err)
	}

	// Return original error to preserve detailed info (Position, Object, Field)
	// The HTTP handler will use errors.As to extract details
	switch {
	case engine.IsParseError(err):
		return fmt.Errorf("%w: %w", soqlModel.ErrInvalidQuery, err)
	case engine.IsAccessError(err):
		return fmt.Errorf("%w: %w", soqlModel.ErrSemanticError, err)
	case engine.IsLimitError(err):
		return fmt.Errorf("%w: %w", soqlModel.ErrQueryTooComplex, err)
	default:
		return err
	}
}
