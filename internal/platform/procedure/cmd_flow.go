package procedure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// FlowCommandExecutor handles flow.if, flow.match, and flow.call.
type FlowCommandExecutor struct {
	engine   *Engine
	resolver *ExpressionResolver
}

// NewFlowCommandExecutor creates a new FlowCommandExecutor.
func NewFlowCommandExecutor(engine *Engine, resolver *ExpressionResolver) *FlowCommandExecutor {
	return &FlowCommandExecutor{
		engine:   engine,
		resolver: resolver,
	}
}

func (e *FlowCommandExecutor) Category() string { return "flow" }

func (e *FlowCommandExecutor) Execute(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	parts := strings.SplitN(cmd.Type, ".", 2)
	if len(parts) != 2 {
		return nil, apperror.BadRequest("invalid flow command type")
	}

	switch parts[1] {
	case "if":
		return e.executeIf(ctx, cmd, execCtx)
	case "match":
		return e.executeMatch(ctx, cmd, execCtx)
	case "call":
		return e.executeCall(ctx, cmd, execCtx)
	case "try":
		return e.executeTry(ctx, cmd, execCtx)
	default:
		return nil, apperror.BadRequest(fmt.Sprintf("unknown flow command: %s", cmd.Type))
	}
}

func (e *FlowCommandExecutor) executeIf(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Condition == "" {
		return nil, apperror.BadRequest("flow.if requires 'condition'")
	}

	result, err := e.resolver.ResolveBool(cmd.Condition, execCtx)
	if err != nil {
		return nil, fmt.Errorf("flow.if: %w", err)
	}

	if result {
		if len(cmd.Then) > 0 {
			if err := e.engine.ExecuteSubCommands(ctx, cmd.Then, execCtx); err != nil {
				return nil, err
			}
		}
	} else {
		if len(cmd.Else) > 0 {
			if err := e.engine.ExecuteSubCommands(ctx, cmd.Else, execCtx); err != nil {
				return nil, err
			}
		}
	}

	return map[string]any{"branch": result}, nil
}

func (e *FlowCommandExecutor) executeMatch(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Expression == "" {
		return nil, apperror.BadRequest("flow.match requires 'expression'")
	}

	value, err := e.resolver.ResolveAny(cmd.Expression, execCtx)
	if err != nil {
		return nil, fmt.Errorf("flow.match: %w", err)
	}

	valueStr := fmt.Sprintf("%v", value)

	if cmd.Cases != nil {
		if branch, ok := cmd.Cases[valueStr]; ok {
			if err := e.engine.ExecuteSubCommands(ctx, branch, execCtx); err != nil {
				return nil, err
			}
			return map[string]any{"matched": valueStr}, nil
		}
	}

	// Default branch
	if len(cmd.Default) > 0 {
		if err := e.engine.ExecuteSubCommands(ctx, cmd.Default, execCtx); err != nil {
			return nil, err
		}
		return map[string]any{"matched": "default"}, nil
	}

	return map[string]any{"matched": ""}, nil
}

func (e *FlowCommandExecutor) executeCall(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Procedure == "" {
		return nil, apperror.BadRequest("flow.call requires 'procedure'")
	}

	// Check call depth
	if len(execCtx.CallStack) >= MaxCallDepth {
		return nil, apperror.BadRequest(fmt.Sprintf("max call depth exceeded (%d)", MaxCallDepth))
	}

	// Check for circular calls
	for _, caller := range execCtx.CallStack {
		if caller == cmd.Procedure {
			return nil, apperror.BadRequest(fmt.Sprintf("circular call detected: %s", cmd.Procedure))
		}
	}

	// Resolve input
	var callInput map[string]any
	if cmd.Input != nil {
		resolved, err := e.resolver.ResolveMap(cmd.Input, execCtx)
		if err != nil {
			return nil, fmt.Errorf("flow.call: resolve input: %w", err)
		}
		callInput = resolved
	}

	if execCtx.DryRun {
		return map[string]any{"called": cmd.Procedure}, nil
	}

	// Get published definition
	def, err := e.engine.procSvc.GetPublishedDefinition(ctx, cmd.Procedure)
	if err != nil {
		return nil, fmt.Errorf("flow.call: %w", err)
	}

	// Create a child execution context sharing limits but with its own scope
	childCtx := &ExecutionContext{
		Vars: map[string]any{
			"input": callInput,
			"now":   execCtx.Vars["now"],
		},
		CallStack:    append(append([]string{}, execCtx.CallStack...), cmd.Procedure),
		CommandCount: execCtx.CommandCount,
		HTTPCount:    execCtx.HTTPCount,
		NotifyCount:  execCtx.NotifyCount,
		DryRun:       execCtx.DryRun,
		Deadline:     execCtx.Deadline,
		Trace:        execCtx.Trace,
	}

	if err := e.engine.executeCommands(ctx, def.Commands, childCtx); err != nil {
		return nil, fmt.Errorf("flow.call(%s): %w", cmd.Procedure, err)
	}

	// Propagate counters back
	execCtx.CommandCount = childCtx.CommandCount
	execCtx.HTTPCount = childCtx.HTTPCount
	execCtx.NotifyCount = childCtx.NotifyCount
	execCtx.Warnings = append(execCtx.Warnings, childCtx.Warnings...)
	execCtx.Trace = childCtx.Trace

	// Build result from called procedure's result map
	resultMap := make(map[string]any)
	if def.Result != nil {
		for k, v := range def.Result {
			resolved, resolveErr := e.engine.resolver.ResolveAny(v, childCtx)
			if resolveErr != nil {
				resultMap[k] = nil
			} else {
				resultMap[k] = resolved
			}
		}
	}

	return resultMap, nil
}

func (e *FlowCommandExecutor) executeTry(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if len(cmd.Try) == 0 {
		return nil, apperror.BadRequest("flow.try requires 'try' commands")
	}

	tryErr := e.engine.ExecuteSubCommands(ctx, cmd.Try, execCtx)

	if tryErr == nil {
		return map[string]any{"caught": false}, nil
	}

	if len(cmd.Catch) == 0 {
		return nil, tryErr
	}

	errorInfo := map[string]any{
		"code":    extractErrorCode(tryErr),
		"message": tryErr.Error(),
	}
	prevError := execCtx.Vars["error"]
	execCtx.Vars["error"] = errorInfo

	catchErr := e.engine.ExecuteSubCommands(ctx, cmd.Catch, execCtx)

	if prevError != nil {
		execCtx.Vars["error"] = prevError
	} else {
		delete(execCtx.Vars, "error")
	}

	if catchErr != nil {
		return nil, catchErr
	}

	return map[string]any{
		"caught":        true,
		"error_code":    errorInfo["code"],
		"error_message": errorInfo["message"],
	}, nil
}

func extractErrorCode(err error) string {
	var execErr *ExecutionError
	if errors.As(err, &execErr) {
		return execErr.Code
	}
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return string(appErr.Code)
	}
	return "unknown"
}
