package dml

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/rls"
)

// rlsExecutor wraps the default DML executor and injects RLS WHERE clauses
// for UPDATE and DELETE operations. INSERT and UPSERT pass through unchanged.
type rlsExecutor struct {
	inner       engine.Executor
	pool        *pgxpool.Pool
	cache       metadata.MetadataReader
	rlsEnforcer rls.Enforcer
}

// NewRLSExecutor creates a DML executor with RLS enforcement.
func NewRLSExecutor(pool *pgxpool.Pool, cache metadata.MetadataReader, rlsEnforcer rls.Enforcer) TxExecutor {
	return &rlsExecutor{
		inner:       engine.NewDefaultExecutor(pool),
		pool:        pool,
		cache:       cache,
		rlsEnforcer: rlsEnforcer,
	}
}

// Execute implements engine.Executor.
func (e *rlsExecutor) Execute(ctx context.Context, compiled *engine.CompiledDML) (*engine.Result, error) {
	if compiled.Operation == engine.OperationUpdate || compiled.Operation == engine.OperationDelete {
		injected, err := e.injectRLS(ctx, compiled)
		if err != nil {
			return nil, err
		}
		compiled = injected
	}

	return e.inner.Execute(ctx, compiled)
}

func (e *rlsExecutor) injectRLS(ctx context.Context, compiled *engine.CompiledDML) (*engine.CompiledDML, error) {
	uc, ok := security.UserFromContext(ctx)
	if !ok || uc.UserID == uuid.Nil {
		return compiled, nil
	}

	objectID, err := resolveObjectID(e.cache, compiled.Object)
	if err != nil {
		return compiled, nil
	}

	rlsClause, rlsParams, err := e.rlsEnforcer.BuildWhereClause(ctx, uc.UserID, objectID)
	if err != nil {
		return nil, fmt.Errorf("dmlExecutor.injectRLS: %w", err)
	}

	if rlsClause == "" || rlsClause == "TRUE" {
		return compiled, nil
	}

	sql, params := injectDMLRLSClause(compiled.SQL, compiled.Params, rlsClause, rlsParams)

	result := *compiled
	result.SQL = sql
	result.Params = params
	return &result, nil
}

// injectDMLRLSClause adds the RLS WHERE clause to a DML statement (UPDATE/DELETE).
// It re-numbers the RLS parameters starting after existing params.
func injectDMLRLSClause(sql string, params []any, rlsClause string, rlsParams []any) (string, []any) {
	offset := len(params)
	rewrittenClause := rlsClause
	for i := len(rlsParams); i >= 1; i-- {
		old := fmt.Sprintf("$%d", i)
		replacement := fmt.Sprintf("$%d", i+offset)
		rewrittenClause = strings.ReplaceAll(rewrittenClause, old, replacement)
	}

	upperSQL := strings.ToUpper(sql)
	if idx := strings.Index(upperSQL, "WHERE "); idx >= 0 {
		// Existing WHERE — prepend RLS with AND.
		insertPoint := idx + len("WHERE ")
		sql = sql[:insertPoint] + rewrittenClause + " AND " + sql[insertPoint:]
	} else {
		// No WHERE — inject before RETURNING or at end.
		insertBefore := len(sql)
		if pos := strings.Index(upperSQL, "RETURNING"); pos >= 0 && pos < insertBefore {
			insertBefore = pos
		}
		sql = sql[:insertBefore] + "WHERE " + rewrittenClause + " " + sql[insertBefore:]
	}

	params = append(params, rlsParams...)
	return sql, params
}

// WithTx returns a new executor that uses the given transaction for all operations.
// The returned executor applies the same RLS injection logic.
func (e *rlsExecutor) WithTx(tx engine.DB) engine.Executor {
	return &rlsExecutor{
		inner:       engine.NewDefaultExecutor(tx),
		pool:        e.pool,
		cache:       e.cache,
		rlsEnforcer: e.rlsEnforcer,
	}
}

// resolveObjectID resolves an API name to a UUID via the metadata cache.
func resolveObjectID(cache metadata.MetadataReader, apiName string) (uuid.UUID, error) {
	objDef, ok := cache.GetObjectByAPIName(apiName)
	if !ok {
		return uuid.Nil, fmt.Errorf("object %q not found", apiName)
	}
	return objDef.ID, nil
}
