package procedure

import (
	"fmt"
	"log/slog"
)

// ExecuteRollback runs all rollback actions in LIFO order (Saga pattern).
func ExecuteRollback(execCtx *ExecutionContext) error {
	var firstErr error
	for i := len(execCtx.RollbackStack) - 1; i >= 0; i-- {
		entry := execCtx.RollbackStack[i]
		if err := entry.Action(); err != nil {
			slog.Error("rollback action failed",
				"step", entry.StepName,
				"error", err,
			)
			if firstErr == nil {
				firstErr = fmt.Errorf("rollback %s: %w", entry.StepName, err)
			}
		}
	}
	return firstErr
}
