package soql

import (
	"context"
	"fmt"

	"github.com/adverax/crm/internal/platform/soql/engine"
)

// QueryService executes SOQL queries with full security enforcement.
type QueryService interface {
	Execute(ctx context.Context, query string, params *QueryParams) (*QueryResult, error)
}

type queryService struct {
	engine   *engine.Engine
	executor *Executor
}

// NewQueryService creates a new QueryService.
func NewQueryService(eng *engine.Engine, executor *Executor) QueryService {
	return &queryService{
		engine:   eng,
		executor: executor,
	}
}

// Execute parses, validates, compiles, and executes a SOQL query.
func (s *queryService) Execute(ctx context.Context, query string, params *QueryParams) (*QueryResult, error) {
	if params == nil {
		params = &QueryParams{PageSize: DefaultPageSize}
	}
	if params.PageSize <= 0 {
		params.PageSize = DefaultPageSize
	}
	if params.PageSize > MaxPageSize {
		params.PageSize = MaxPageSize
	}

	compiled, err := s.engine.PrepareAndResolve(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("queryService.Execute: %w", err)
	}

	result, err := s.executor.Execute(ctx, compiled)
	if err != nil {
		return nil, fmt.Errorf("queryService.Execute: %w", err)
	}

	return result, nil
}
