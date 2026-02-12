package dml

import (
	"context"
	"fmt"

	"github.com/adverax/crm/internal/platform/dml/engine"
)

// DMLService executes DML statements with full security enforcement.
type DMLService interface {
	Execute(ctx context.Context, statement string) (*Result, error)
	Prepare(ctx context.Context, statement string) (*engine.CompiledDML, error)
}

type dmlService struct {
	engine   *engine.Engine
	executor engine.Executor
}

// NewDMLService creates a new DMLService.
func NewDMLService(eng *engine.Engine, executor engine.Executor) DMLService {
	return &dmlService{
		engine:   eng,
		executor: executor,
	}
}

// Prepare parses, validates, and compiles a DML statement without executing.
func (s *dmlService) Prepare(ctx context.Context, statement string) (*engine.CompiledDML, error) {
	compiled, err := s.engine.Prepare(ctx, statement)
	if err != nil {
		return nil, fmt.Errorf("dmlService.Prepare: %w", err)
	}
	return compiled, nil
}

// Execute parses, validates, compiles, and executes a DML statement.
func (s *dmlService) Execute(ctx context.Context, statement string) (*Result, error) {
	compiled, err := s.engine.Prepare(ctx, statement)
	if err != nil {
		return nil, fmt.Errorf("dmlService.Execute: %w", err)
	}

	result, err := s.executor.Execute(ctx, compiled)
	if err != nil {
		return nil, fmt.Errorf("dmlService.Execute: %w", err)
	}

	return result, nil
}
