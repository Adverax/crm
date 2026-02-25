package procedure

import (
	"context"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// WaitCommandExecutor is a stub for wait.* commands.
type WaitCommandExecutor struct{}

// NewWaitCommandExecutor creates a stub WaitCommandExecutor.
func NewWaitCommandExecutor() *WaitCommandExecutor {
	return &WaitCommandExecutor{}
}

func (e *WaitCommandExecutor) Category() string { return "wait" }

func (e *WaitCommandExecutor) Execute(_ context.Context, cmd metadata.CommandDef, _ *ExecutionContext) (any, error) {
	return nil, apperror.BadRequest("wait.* commands are not yet implemented (command: " + cmd.Type + ")")
}
