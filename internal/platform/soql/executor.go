package soql

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/security/rls"
	"github.com/adverax/crm/internal/platform/soql/engine"
)

// Executor executes compiled SOQL queries against PostgreSQL.
type Executor struct {
	pool        *pgxpool.Pool
	cache       metadata.MetadataReader
	rlsEnforcer rls.Enforcer
}

// NewExecutor creates a new pgx-based SOQL executor.
func NewExecutor(pool *pgxpool.Pool, cache metadata.MetadataReader, rlsEnforcer rls.Enforcer) *Executor {
	return &Executor{
		pool:        pool,
		cache:       cache,
		rlsEnforcer: rlsEnforcer,
	}
}

// Execute runs a compiled SOQL query and returns the result.
func (e *Executor) Execute(ctx context.Context, compiled *engine.CompiledQuery) (*QueryResult, error) {
	uc, _ := security.UserFromContext(ctx)

	sql := compiled.SQL
	params := compiled.Params

	// Inject RLS WHERE clause if the user context is available.
	if uc.UserID != uuid.Nil && compiled.Shape.Table != "" {
		objectID, err := resolveObjectID(e.cache, compiled.Shape.Object)
		if err == nil {
			rlsClause, rlsParams, rlsErr := e.rlsEnforcer.BuildWhereClause(ctx, uc.UserID, objectID)
			if rlsErr != nil {
				return nil, fmt.Errorf("soqlExecutor.Execute: RLS: %w", rlsErr)
			}
			if rlsClause != "" && rlsClause != "TRUE" {
				sql, params = injectRLSClause(sql, params, rlsClause, rlsParams)
			}
		}
	}

	// Resolve date parameters.
	if len(compiled.DateParams) > 0 {
		resolver := engine.NewDefaultDateResolver()
		queryCopy := *compiled
		queryCopy.SQL = sql
		queryCopy.Params = params
		if err := engine.ResolveDateParams(ctx, &queryCopy, resolver); err != nil {
			return nil, fmt.Errorf("soqlExecutor.Execute: date resolve: %w", err)
		}
		sql = queryCopy.SQL
		params = queryCopy.Params
	}

	rows, err := e.pool.Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("soqlExecutor.Execute: query: %w", err)
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	var records []map[string]any

	for rows.Next() {
		values, scanErr := rows.Values()
		if scanErr != nil {
			return nil, fmt.Errorf("soqlExecutor.Execute: scan: %w", scanErr)
		}
		record := make(map[string]any, len(fieldDescs))
		for i, fd := range fieldDescs {
			record[fd.Name] = values[i]
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("soqlExecutor.Execute: rows: %w", err)
	}

	// Map SQL columns back to SOQL field names using shape.
	mappedRecords := mapRecordsToSOQL(records, compiled.Shape)

	// Enforce single-row constraint for SELECT ROW queries.
	if compiled.IsRow && len(mappedRecords) > 1 {
		return nil, fmt.Errorf("soqlExecutor.Execute: SELECT ROW returned %d records, expected at most 1", len(mappedRecords))
	}

	return &QueryResult{
		TotalSize: len(mappedRecords),
		Done:      true,
		Records:   mappedRecords,
		IsRow:     compiled.IsRow,
	}, nil
}

// mapRecordsToSOQL converts SQL column names to SOQL field names.
func mapRecordsToSOQL(records []map[string]any, shape *engine.ResultShape) []map[string]any {
	if shape == nil || len(shape.Fields) == 0 {
		return records
	}

	colToName := make(map[string]string, len(shape.Fields))
	for _, f := range shape.Fields {
		colToName[f.Column] = f.Name
	}

	result := make([]map[string]any, len(records))
	for i, rec := range records {
		mapped := make(map[string]any, len(rec))
		for col, val := range rec {
			if name, ok := colToName[col]; ok {
				mapped[name] = val
			} else {
				mapped[col] = val
			}
		}
		result[i] = mapped
	}
	return result
}

// injectRLSClause adds the RLS WHERE clause to the compiled SQL.
// It re-numbers the RLS parameters starting after existing params.
func injectRLSClause(sql string, params []any, rlsClause string, rlsParams []any) (string, []any) {
	// Re-number RLS parameter placeholders ($1, $2, ...) to start after existing params.
	offset := len(params)
	rewrittenClause := rlsClause
	for i := len(rlsParams); i >= 1; i-- {
		old := fmt.Sprintf("$%d", i)
		new := fmt.Sprintf("$%d", i+offset)
		rewrittenClause = strings.ReplaceAll(rewrittenClause, old, new)
	}

	// Find where to inject: after WHERE keyword or before ORDER BY/LIMIT.
	upperSQL := strings.ToUpper(sql)
	if idx := strings.Index(upperSQL, "WHERE "); idx >= 0 {
		// Existing WHERE — append with AND.
		insertPoint := idx + len("WHERE ")
		sql = sql[:insertPoint] + rewrittenClause + " AND " + sql[insertPoint:]
	} else {
		// No WHERE — inject before ORDER BY, GROUP BY, LIMIT, or at end.
		insertBefore := len(sql)
		for _, kw := range []string{"ORDER BY", "GROUP BY", "HAVING", "LIMIT", "OFFSET", "FOR UPDATE"} {
			if pos := strings.Index(upperSQL, kw); pos >= 0 && pos < insertBefore {
				insertBefore = pos
			}
		}
		sql = sql[:insertBefore] + "WHERE " + rewrittenClause + " " + sql[insertBefore:]
	}

	params = append(params, rlsParams...)
	return sql, params
}
