package procedure

import (
	"context"

	"github.com/adverax/crm/internal/platform/metadata"
)

// CommandExecutor executes a specific category of commands (e.g., "record", "compute").
type CommandExecutor interface {
	// Category returns the command category (e.g., "record", "compute", "flow").
	Category() string

	// Execute runs a single command and returns its result.
	Execute(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error)
}
