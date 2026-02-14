package dml

import (
	"context"
	"errors"
	"fmt"

	"github.com/adverax/crm/internal/pkg/apperror"
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
		return nil, fmt.Errorf("dmlService.Prepare: %w", mapDMLError(err))
	}
	return compiled, nil
}

// Execute parses, validates, compiles, and executes a DML statement.
func (s *dmlService) Execute(ctx context.Context, statement string) (*Result, error) {
	compiled, err := s.engine.Prepare(ctx, statement)
	if err != nil {
		return nil, fmt.Errorf("dmlService.Execute: %w", mapDMLError(err))
	}

	result, err := s.executor.Execute(ctx, compiled)
	if err != nil {
		return nil, fmt.Errorf("dmlService.Execute: %w", err)
	}

	return result, nil
}

// mapDMLError maps engine errors to application errors.
func mapDMLError(err error) error {
	var ruleErr *engine.RuleValidationError
	if errors.As(err, &ruleErr) {
		if len(ruleErr.Rules) > 0 {
			return apperror.BadRequest(ruleErr.Rules[0].Message)
		}
		return apperror.BadRequest("validation rule failed")
	}

	var defaultErr *engine.DefaultEvalError
	if errors.As(err, &defaultErr) {
		return apperror.Internal("default expression evaluation failed")
	}

	var validationErr *engine.ValidationError
	if errors.As(err, &validationErr) {
		return apperror.BadRequest(validationErr.Message)
	}

	var accessErr *engine.AccessError
	if errors.As(err, &accessErr) {
		return apperror.Forbidden(accessErr.Message)
	}

	return err
}
