package dml

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/dml/engine"
)

// PostExecuteHook is called after a successful DML execute (Stage 8: Automation).
type PostExecuteHook interface {
	AfterDMLExecute(ctx context.Context, compiled *engine.CompiledDML, result *engine.Result) error
}

// TxExecutor is an executor that supports transaction-scoped variants.
type TxExecutor interface {
	engine.Executor
	WithTx(tx engine.DB) engine.Executor
}

// DMLService executes DML statements with full security enforcement.
type DMLService interface {
	Execute(ctx context.Context, statement string) (*Result, error)
	ExecuteBatch(ctx context.Context, statements []string) ([]*Result, error)
	Prepare(ctx context.Context, statement string) (*engine.CompiledDML, error)
	SetPostExecuteHook(hook PostExecuteHook)
}

type dmlService struct {
	pool         *pgxpool.Pool
	engine       *engine.Engine
	executor     TxExecutor
	postExecHook PostExecuteHook
}

// NewDMLService creates a new DMLService.
func NewDMLService(pool *pgxpool.Pool, eng *engine.Engine, executor TxExecutor) DMLService {
	return &dmlService{
		pool:     pool,
		engine:   eng,
		executor: executor,
	}
}

// SetPostExecuteHook sets the post-execute hook (automation rules).
func (s *dmlService) SetPostExecuteHook(hook PostExecuteHook) {
	s.postExecHook = hook
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
// After successful execution, fires Stage 8 (post-execute / automation rules).
func (s *dmlService) Execute(ctx context.Context, statement string) (*Result, error) {
	compiled, err := s.engine.Prepare(ctx, statement)
	if err != nil {
		return nil, fmt.Errorf("dmlService.Execute: %w", mapDMLError(err))
	}

	result, err := s.executor.Execute(ctx, compiled)
	if err != nil {
		return nil, fmt.Errorf("dmlService.Execute: %w", err)
	}

	// Stage 8: Post-execute hook (automation rules)
	if s.postExecHook != nil {
		if hookErr := s.postExecHook.AfterDMLExecute(ctx, compiled, result); hookErr != nil {
			return nil, fmt.Errorf("dmlService.Execute: post-execute: %w", hookErr)
		}
	}

	return result, nil
}

// ExecuteBatch prepares, validates, and executes multiple DML statements in a single transaction.
// Post-execute hooks fire for each result after the transaction commits.
func (s *dmlService) ExecuteBatch(ctx context.Context, statements []string) ([]*Result, error) {
	// Phase 1: Prepare all statements (validate before execute)
	compiled := make([]*engine.CompiledDML, len(statements))
	for i, stmt := range statements {
		c, err := s.engine.Prepare(ctx, stmt)
		if err != nil {
			return nil, fmt.Errorf("dmlService.ExecuteBatch: statement[%d]: %w", i, mapDMLError(err))
		}
		compiled[i] = c
	}

	// Phase 2: Execute all in a single transaction
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("dmlService.ExecuteBatch: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txExec := s.executor.WithTx(tx)

	results := make([]*Result, len(compiled))
	for i, c := range compiled {
		r, execErr := txExec.Execute(ctx, c)
		if execErr != nil {
			return nil, fmt.Errorf("dmlService.ExecuteBatch: statement[%d]: %w", i, execErr)
		}
		results[i] = r
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("dmlService.ExecuteBatch: commit: %w", err)
	}

	// Phase 3: Fire post-execute hooks after successful commit
	if s.postExecHook != nil {
		for i, c := range compiled {
			if hookErr := s.postExecHook.AfterDMLExecute(ctx, c, results[i]); hookErr != nil {
				return nil, fmt.Errorf("dmlService.ExecuteBatch: post-execute[%d]: %w", i, hookErr)
			}
		}
	}

	return results, nil
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
