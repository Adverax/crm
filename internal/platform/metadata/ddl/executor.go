package ddl

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Executor executes DDL statements within a transaction.
type Executor struct{}

// NewExecutor creates a new DDL executor.
func NewExecutor() *Executor {
	return &Executor{}
}

// ExecInTx executes a list of DDL statements within the given transaction.
func (e *Executor) ExecInTx(ctx context.Context, tx pgx.Tx, statements []string) error {
	for i, stmt := range statements {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("ddl.ExecInTx: statement %d (%s): %w", i, truncate(stmt, 80), err)
		}
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
