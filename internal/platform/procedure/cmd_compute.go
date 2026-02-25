package procedure

import (
	"context"
	"fmt"
	"strings"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// ComputeCommandExecutor handles compute.transform, compute.validate, and compute.fail.
type ComputeCommandExecutor struct {
	resolver *ExpressionResolver
}

// NewComputeCommandExecutor creates a new ComputeCommandExecutor.
func NewComputeCommandExecutor(resolver *ExpressionResolver) *ComputeCommandExecutor {
	return &ComputeCommandExecutor{resolver: resolver}
}

func (e *ComputeCommandExecutor) Category() string { return "compute" }

func (e *ComputeCommandExecutor) Execute(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	parts := strings.SplitN(cmd.Type, ".", 2)
	if len(parts) != 2 {
		return nil, apperror.BadRequest("invalid compute command type")
	}

	switch parts[1] {
	case "transform":
		return e.executeTransform(ctx, cmd, execCtx)
	case "validate":
		return e.executeValidate(ctx, cmd, execCtx)
	case "fail":
		return e.executeFail(cmd)
	default:
		return nil, apperror.BadRequest(fmt.Sprintf("unknown compute command: %s", cmd.Type))
	}
}

func (e *ComputeCommandExecutor) executeTransform(_ context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Value == nil {
		return nil, apperror.BadRequest("compute.transform requires 'value'")
	}

	result, err := e.resolver.ResolveMap(cmd.Value, execCtx)
	if err != nil {
		return nil, fmt.Errorf("compute.transform: %w", err)
	}

	return result, nil
}

func (e *ComputeCommandExecutor) executeValidate(_ context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Condition == "" {
		return nil, apperror.BadRequest("compute.validate requires 'condition'")
	}

	valid, err := e.resolver.ResolveBool(cmd.Condition, execCtx)
	if err != nil {
		return nil, fmt.Errorf("compute.validate: %w", err)
	}

	if !valid {
		code := cmd.Code
		if code == "" {
			code = "VALIDATION_FAILED"
		}
		msg := cmd.Message
		if msg == "" {
			msg = "validation failed"
		}
		return nil, &ExecutionError{Code: code, Message: msg}
	}

	return map[string]any{"valid": true}, nil
}

func (e *ComputeCommandExecutor) executeFail(cmd metadata.CommandDef) (any, error) {
	code := cmd.Code
	if code == "" {
		code = "PROCEDURE_FAILED"
	}
	msg := cmd.Message
	if msg == "" {
		msg = "procedure failed"
	}
	return nil, &ExecutionError{Code: code, Message: msg}
}
