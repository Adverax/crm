//go:build integration

package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTestDSN = "postgres://crm:crm_secret@localhost:5433/crm_test?sslmode=disable"

// SetupTestPool creates a pgxpool.Pool connected to the test database.
// DSN is taken from the DB_TEST_DSN environment variable (default: localhost:5433/crm_test).
// The pool is closed in t.Cleanup.
func SetupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := defaultTestDSN
	if v := os.Getenv("DB_TEST_DSN"); v != "" {
		dsn = v
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("testutil.SetupTestPool: connect: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("testutil.SetupTestPool: ping: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

// TruncateTables executes TRUNCATE ... CASCADE for the given tables.
// Used at the beginning of each integration test function for isolation.
func TruncateTables(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()

	if len(tables) == 0 {
		return
	}

	sql := fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", "))
	if _, err := pool.Exec(context.Background(), sql); err != nil {
		t.Fatalf("testutil.TruncateTables: %v", err)
	}
}

// BeginTx starts a transaction and registers Rollback in t.Cleanup.
// Use for write operations in repos that require pgx.Tx.
func BeginTx(t *testing.T, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()

	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("testutil.BeginTx: %v", err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})
	return tx
}
