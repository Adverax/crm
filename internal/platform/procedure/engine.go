package procedure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adverax/crm/internal/pkg/apperror"
	celengine "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/metadata"
)

// ProcedureExecutor is the public facade for executing procedures.
type ProcedureExecutor interface {
	Execute(ctx context.Context, code string, input map[string]any) (*ExecutionResult, error)
	DryRun(ctx context.Context, def *metadata.ProcedureDefinition, input map[string]any) (*ExecutionResult, error)
	ExecuteDefinition(ctx context.Context, def *metadata.ProcedureDefinition, input map[string]any, opts ...ExecOption) (*ExecutionResult, error)
}

// Engine orchestrates command execution within a procedure.
type Engine struct {
	resolver *ExpressionResolver
	registry map[string]CommandExecutor // category → executor
	procSvc  metadata.ProcedureService
}

// NewEngine creates a new procedure Engine.
func NewEngine(
	celCache *celengine.ProgramCache,
	procSvc metadata.ProcedureService,
	executors ...CommandExecutor,
) *Engine {
	registry := make(map[string]CommandExecutor, len(executors))
	for _, exec := range executors {
		registry[exec.Category()] = exec
	}

	return &Engine{
		resolver: NewExpressionResolver(celCache),
		registry: registry,
		procSvc:  procSvc,
	}
}

// Execute runs a published procedure by its code.
func (e *Engine) Execute(ctx context.Context, code string, input map[string]any) (*ExecutionResult, error) {
	def, err := e.procSvc.GetPublishedDefinition(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("engine.Execute: %w", err)
	}
	return e.ExecuteDefinition(ctx, def, input)
}

// DryRun executes a definition in dry-run mode (no side effects).
func (e *Engine) DryRun(ctx context.Context, def *metadata.ProcedureDefinition, input map[string]any) (*ExecutionResult, error) {
	return e.ExecuteDefinition(ctx, def, input, WithDryRun())
}

// ExecuteDefinition runs a procedure definition with the given input.
func (e *Engine) ExecuteDefinition(ctx context.Context, def *metadata.ProcedureDefinition, input map[string]any, opts ...ExecOption) (*ExecutionResult, error) {
	options := &execOptions{}
	for _, opt := range opts {
		opt(options)
	}

	deadline := time.Now().Add(MaxExecutionTimeout)
	execCtx := NewExecutionContext(input, options.dryRun, deadline)

	err := e.executeCommands(ctx, def.Commands, execCtx)
	if err != nil {
		// Attempt rollback on error
		if rbErr := ExecuteRollback(execCtx); rbErr != nil {
			return nil, fmt.Errorf("engine.ExecuteDefinition: execution failed: %w; rollback also failed: %s", err, rbErr)
		}
		return nil, fmt.Errorf("engine.ExecuteDefinition: %w", err)
	}

	// Build result from result map
	resultMap := make(map[string]any)
	if def.Result != nil {
		for k, v := range def.Result {
			resolved, resolveErr := e.resolver.ResolveAny(v, execCtx)
			if resolveErr != nil {
				resultMap[k] = nil
			} else {
				resultMap[k] = resolved
			}
		}
	}

	return &ExecutionResult{
		Success:  true,
		Result:   resultMap,
		Warnings: execCtx.Warnings,
		Trace:    execCtx.Trace,
	}, nil
}

// executeCommands runs a list of commands sequentially.
func (e *Engine) executeCommands(ctx context.Context, cmds []metadata.CommandDef, execCtx *ExecutionContext) error {
	for i, cmd := range cmds {
		if err := e.checkLimits(execCtx); err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		if time.Now().After(execCtx.Deadline) {
			return apperror.BadRequest("procedure execution timeout exceeded")
		}

		stepName := cmd.As
		if stepName == "" {
			stepName = fmt.Sprintf("step_%d", i)
		}

		// Evaluate "when" condition
		if cmd.When != "" {
			shouldRun, err := e.resolver.ResolveBool(cmd.When, execCtx)
			if err != nil {
				return fmt.Errorf("step %s: when condition: %w", stepName, err)
			}
			if !shouldRun {
				execCtx.Trace = append(execCtx.Trace, TraceEntry{
					Step:   stepName,
					Type:   cmd.Type,
					Status: "skipped",
				})
				continue
			}
		}

		start := time.Now()
		var result any
		var err error
		if cmd.Retry != nil && cmd.Retry.MaxAttempts > 1 {
			result, err = e.executeWithRetry(ctx, cmd, execCtx)
		} else {
			result, err = e.executeCommand(ctx, cmd, execCtx)
		}
		duration := time.Since(start).Milliseconds()

		if err != nil {
			if cmd.Optional {
				execCtx.Warnings = append(execCtx.Warnings, ExecutionWarning{
					Command: stepName,
					Message: err.Error(),
				})
				execCtx.Trace = append(execCtx.Trace, TraceEntry{
					Step:     stepName,
					Type:     cmd.Type,
					Status:   "warning",
					Duration: duration,
					Error:    err.Error(),
				})
				continue
			}
			execCtx.Trace = append(execCtx.Trace, TraceEntry{
				Step:     stepName,
				Type:     cmd.Type,
				Status:   "error",
				Duration: duration,
				Error:    err.Error(),
			})
			return fmt.Errorf("step %s (%s): %w", stepName, cmd.Type, err)
		}

		// Store result in context vars
		if cmd.As != "" && result != nil {
			execCtx.Vars[cmd.As] = result
		}

		// Register rollback commands in Saga stack (LIFO) if defined
		if len(cmd.Rollback) > 0 {
			rollbackCmds := make([]metadata.CommandDef, len(cmd.Rollback))
			copy(rollbackCmds, cmd.Rollback)
			rollbackStep := stepName
			execCtx.RollbackStack = append(execCtx.RollbackStack, RollbackEntry{
				StepName: rollbackStep,
				Action: func() error {
					return e.executeCommands(ctx, rollbackCmds, execCtx)
				},
			})
		}

		execCtx.CommandCount++
		execCtx.Trace = append(execCtx.Trace, TraceEntry{
			Step:     stepName,
			Type:     cmd.Type,
			Status:   "ok",
			Duration: duration,
		})
	}
	return nil
}

// executeCommand dispatches a single command to the appropriate executor.
func (e *Engine) executeCommand(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	parts := strings.SplitN(cmd.Type, ".", 2)
	if len(parts) != 2 {
		return nil, apperror.BadRequest(fmt.Sprintf("invalid command type format: %s", cmd.Type))
	}

	category := parts[0]
	executor, ok := e.registry[category]
	if !ok {
		return nil, apperror.BadRequest(fmt.Sprintf("unknown command category: %s", category))
	}

	return executor.Execute(ctx, cmd, execCtx)
}

// RegisterExecutor adds a command executor to the engine's registry.
// Used to break circular dependencies (e.g., flow executor needs engine reference).
func (e *Engine) RegisterExecutor(exec CommandExecutor) {
	e.registry[exec.Category()] = exec
}

// Resolver returns the engine's expression resolver (used by sub-executors).
func (e *Engine) Resolver() *ExpressionResolver {
	return e.resolver
}

// ExecuteSubCommands runs nested commands (used by flow executor).
func (e *Engine) ExecuteSubCommands(ctx context.Context, cmds []metadata.CommandDef, execCtx *ExecutionContext) error {
	return e.executeCommands(ctx, cmds, execCtx)
}

func (e *Engine) executeWithRetry(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	attempts := cmd.Retry.MaxAttempts
	delayMs := cmd.Retry.DelayMs
	mult := cmd.Retry.BackoffMult
	if mult < 1 {
		mult = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		result, err := e.executeCommand(ctx, cmd, execCtx)
		if err == nil {
			return result, nil
		}
		lastErr = err

		stepName := cmd.As
		if stepName == "" {
			stepName = cmd.Type
		}

		execCtx.Trace = append(execCtx.Trace, TraceEntry{
			Step:   fmt.Sprintf("%s_retry_%d", stepName, attempt),
			Type:   cmd.Type,
			Status: "retry",
			Error:  err.Error(),
		})

		if attempt < attempts {
			sleepDuration := time.Duration(delayMs) * time.Millisecond
			if time.Now().Add(sleepDuration).After(execCtx.Deadline) {
				return nil, fmt.Errorf("retry: deadline would be exceeded after delay")
			}
			time.Sleep(sleepDuration)
			delayMs *= mult
		}
	}
	return nil, lastErr
}

func (e *Engine) checkLimits(execCtx *ExecutionContext) error {
	if execCtx.CommandCount >= MaxCommands {
		return apperror.BadRequest(fmt.Sprintf("max command count exceeded (%d)", MaxCommands))
	}
	if execCtx.HTTPCount > MaxHTTPCalls {
		return apperror.BadRequest(fmt.Sprintf("max HTTP calls exceeded (%d)", MaxHTTPCalls))
	}
	if execCtx.NotifyCount > MaxNotifications {
		return apperror.BadRequest(fmt.Sprintf("max notifications exceeded (%d)", MaxNotifications))
	}
	return nil
}
