package security

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TxBeginner abstracts the ability to begin a transaction.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func withTx(ctx context.Context, beginner TxBeginner, fn func(tx pgx.Tx) error) error {
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed (%v) after error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
