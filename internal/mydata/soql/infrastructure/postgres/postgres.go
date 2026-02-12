package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
)

// PostgresExecutor executes compiled SOQL queries against PostgreSQL.
type PostgresExecutor struct {
	db            DB
	dateResolver  engine.DateResolver
	cursorHandler *CursorHandler
}

// NewPostgresExecutor creates a new PostgreSQL executor.
func NewPostgresExecutor(db DB, secret engine.SecretProvider) *PostgresExecutor {
	return &PostgresExecutor{
		db:            db,
		dateResolver:  engine.NewDefaultDateResolver(),
		cursorHandler: NewCursorHandler(secret),
	}
}

// NewPostgresExecutorWithResolver creates a PostgreSQL executor with a custom date resolver.
func NewPostgresExecutorWithResolver(db DB, secret engine.SecretProvider, resolver engine.DateResolver) *PostgresExecutor {
	if resolver == nil {
		resolver = engine.NewDefaultDateResolver()
	}
	return &PostgresExecutor{
		db:            db,
		dateResolver:  resolver,
		cursorHandler: NewCursorHandler(secret),
	}
}

// Execute implements Executor.
func (e *PostgresExecutor) Execute(ctx context.Context, query *engine.CompiledQuery) (*QueryResult, error) {
	return e.ExecuteWithDB(ctx, e.db, query)
}

// ExecuteWithParams implements Executor with keyset pagination support.
func (e *PostgresExecutor) ExecuteWithParams(ctx context.Context, query *engine.CompiledQuery, params *ExecuteParams) (*QueryResult, error) {
	return e.executeWithParamsAndDB(ctx, e.db, query, params)
}

// executeWithParamsAndDB executes a query with pagination using the provided database connection.
func (e *PostgresExecutor) executeWithParamsAndDB(ctx context.Context, db DB, query *engine.CompiledQuery, params *ExecuteParams) (*QueryResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if params == nil {
		params = &ExecuteParams{}
	}

	// Determine page size
	pageSize := params.PageSize
	if pageSize <= 0 && query.Pagination != nil {
		pageSize = query.Pagination.PageSize
	}
	if pageSize <= 0 {
		pageSize = 100 // Default page size
	}

	// Decode cursor if provided
	var paginationCtx *PaginationContext
	var err error
	if query.Pagination != nil {
		paginationCtx, err = e.cursorHandler.DecodeCursor(params.Cursor, query.Pagination, params.UserID)
		if err != nil {
			return nil, engine.NewValidationError(engine.ErrCodeInvalidPagination, err.Error())
		}
	}

	// Build modified SQL with keyset pagination
	modifiedSQL, modifiedParams, err := e.buildPaginatedSQL(query, paginationCtx, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to build paginated query: %w", err)
	}

	// Resolve date parameters
	if len(query.DateParams) > 0 {
		if err := e.resolveDateParams(ctx, query, modifiedParams); err != nil {
			return nil, fmt.Errorf("failed to resolve date parameters: %w", err)
		}
	}

	// Execute query
	rows, err := db.QueryContext(ctx, modifiedSQL, modifiedParams...)
	if err != nil {
		return nil, engine.NewExecutionErrorWithSQL("query execution failed", modifiedSQL, err)
	}
	defer rows.Close()

	// Parse results
	records, err := e.parseRows(rows, query.Shape)
	if err != nil {
		return nil, engine.NewExecutionError("failed to parse results", err)
	}

	// Determine if there are more records (we fetched pageSize+1)
	done := len(records) <= pageSize
	if !done {
		// Remove the extra record
		records = records[:pageSize]
	}

	// Build next cursor if there are more records
	var nextCursor string
	if !done && len(records) > 0 && query.Pagination != nil {
		lastRecord := &records[len(records)-1]
		nextCursor, err = e.cursorHandler.BuildNextCursor(lastRecord, query.Pagination, params.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to build next cursor: %w", err)
		}
	}

	return &QueryResult{
		Records:    records,
		TotalSize:  len(records),
		Done:       done,
		NextCursor: nextCursor,
	}, nil
}

// buildPaginatedSQL modifies the SQL to include keyset pagination.
func (e *PostgresExecutor) buildPaginatedSQL(query *engine.CompiledQuery, paginationCtx *PaginationContext, pageSize int) (string, []any, error) {
	sql := query.SQL
	params := make([]any, len(query.Params))
	copy(params, query.Params)

	// If no pagination context or no cursor, just add LIMIT
	if paginationCtx == nil || paginationCtx.Cursor == nil {
		sql = e.addOrReplaceLimitInSQL(sql, pageSize+1)
		return sql, params, nil
	}

	// Extract main table alias from query
	mainAlias := e.extractMainAlias(sql)

	// Build keyset WHERE clause
	keysetWhere, keysetParams, _, err := e.cursorHandler.BuildKeysetWhereClause(
		paginationCtx.Cursor,
		query.Pagination,
		mainAlias,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build keyset clause: %w", err)
	}

	if keysetWhere != "" {
		// Inject keyset WHERE clause into the SQL
		sql, params = e.injectKeysetWhere(sql, params, keysetWhere, keysetParams)
	}

	// Update LIMIT to pageSize+1
	sql = e.addOrReplaceLimitInSQL(sql, pageSize+1)

	return sql, params, nil
}

// extractMainAlias extracts the main table alias from the SQL.
// Looks for patterns like "FROM schema.table t0" or "FROM schema.table AS t0".
func (e *PostgresExecutor) extractMainAlias(sql string) string {
	// Pattern: FROM schema.table alias or FROM schema.table AS alias
	re := regexp.MustCompile(`(?i)FROM\s+\S+\s+(?:AS\s+)?(\w+)`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) >= 2 {
		return matches[1]
	}
	return "t0" // Default alias
}

// injectKeysetWhere injects the keyset WHERE clause into the SQL.
// It handles both queries with existing WHERE and without.
func (e *PostgresExecutor) injectKeysetWhere(sql string, params []any, keysetWhere string, keysetParams []any) (string, []any) {
	// Renumber keyset params to continue after existing params
	offset := len(params)
	renumberedWhere := keysetWhere
	for i := len(keysetParams); i >= 1; i-- {
		oldPlaceholder := fmt.Sprintf("$%d", i)
		newPlaceholder := fmt.Sprintf("$%d", i+offset)
		renumberedWhere = strings.ReplaceAll(renumberedWhere, oldPlaceholder, newPlaceholder)
	}

	// Append keyset params
	params = append(params, keysetParams...)

	// Find WHERE clause position
	whereRe := regexp.MustCompile(`(?i)\bWHERE\b`)
	whereLoc := whereRe.FindStringIndex(sql)

	if whereLoc != nil {
		// Has existing WHERE - add keyset condition with AND
		// Find the end of WHERE clause (before ORDER BY, GROUP BY, LIMIT, etc.)
		clauseEndRe := regexp.MustCompile(`(?i)\b(ORDER\s+BY|GROUP\s+BY|HAVING|LIMIT|OFFSET|$)`)
		afterWhere := sql[whereLoc[1]:]
		clauseEndLoc := clauseEndRe.FindStringIndex(afterWhere)

		insertPos := whereLoc[1] + clauseEndLoc[0]
		sql = sql[:insertPos] + " AND " + renumberedWhere + sql[insertPos:]
	} else {
		// No WHERE - find position before ORDER BY, GROUP BY, etc.
		clauseRe := regexp.MustCompile(`(?i)\b(ORDER\s+BY|GROUP\s+BY|HAVING|LIMIT|OFFSET)`)
		clauseLoc := clauseRe.FindStringIndex(sql)

		if clauseLoc != nil {
			// Insert WHERE before the clause
			sql = sql[:clauseLoc[0]] + "WHERE " + renumberedWhere + " " + sql[clauseLoc[0]:]
		} else {
			// No ORDER BY etc. - append at end
			sql = sql + " WHERE " + renumberedWhere
		}
	}

	return sql, params
}

// addOrReplaceLimitInSQL adds or replaces LIMIT clause in the SQL.
func (e *PostgresExecutor) addOrReplaceLimitInSQL(sql string, limit int) string {
	// Check if LIMIT exists
	limitRe := regexp.MustCompile(`(?i)\bLIMIT\s+\d+`)
	if limitRe.MatchString(sql) {
		// Replace existing LIMIT
		return limitRe.ReplaceAllString(sql, fmt.Sprintf("LIMIT %d", limit))
	}

	// No LIMIT - add it at the end (before any trailing semicolon)
	sql = strings.TrimRight(sql, "; \t\n")
	return sql + fmt.Sprintf(" LIMIT %d", limit)
}

// ExecuteWithDB implements Executor.
func (e *PostgresExecutor) ExecuteWithDB(ctx context.Context, db DB, query *engine.CompiledQuery) (*QueryResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// Resolve date parameters
	params := make([]any, len(query.Params))
	copy(params, query.Params)

	if len(query.DateParams) > 0 {
		if err := e.resolveDateParams(ctx, query, params); err != nil {
			return nil, fmt.Errorf("failed to resolve date parameters: %w", err)
		}
	}

	// Execute query
	rows, err := db.QueryContext(ctx, query.SQL, params...)
	if err != nil {
		return nil, engine.NewExecutionErrorWithSQL("query execution failed", query.SQL, err)
	}
	defer rows.Close()

	// Parse results
	records, err := e.parseRows(rows, query.Shape)
	if err != nil {
		return nil, engine.NewExecutionError("failed to parse results", err)
	}

	return &QueryResult{
		Records:   records,
		TotalSize: len(records),
		Done:      true,
	}, nil
}

// resolveDateParams resolves date literal parameters.
func (e *PostgresExecutor) resolveDateParams(ctx context.Context, query *engine.CompiledQuery, params []any) error {
	for _, dp := range query.DateParams {
		var resolvedValue time.Time
		var err error

		if dp.Static != nil {
			resolvedValue, err = e.dateResolver.ResolveStatic(ctx, *dp.Static)
		} else if dp.Dynamic != nil {
			resolvedValue, err = e.dateResolver.ResolveDynamic(ctx, dp.Dynamic)
		} else {
			continue
		}

		if err != nil {
			return err
		}

		// ParamIndex is 1-based
		if dp.ParamIndex > 0 && dp.ParamIndex <= len(params) {
			params[dp.ParamIndex-1] = resolvedValue
		}
	}

	return nil
}

// parseRows parses SQL rows into Records.
func (e *PostgresExecutor) parseRows(rows *sql.Rows, shape *engine.ResultShape) ([]Record, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	var records []Record

	for rows.Next() {
		// Create scan destinations
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i, ct := range columnTypes {
			values[i] = e.createScanDest(ct, shape, i)
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Build record
		record := Record{
			Type:          shape.Object,
			Fields:        make(map[string]any),
			Relationships: make(map[string][]Record),
		}

		for i, col := range columns {
			value := values[i]

			// Find field name from shape
			fieldName := col
			if i < len(shape.Fields) {
				fieldName = shape.Fields[i].Name
			}

			// Handle JSON fields (subqueries)
			if e.isJSONColumn(columnTypes[i]) {
				records, err := e.parseJSONRelationship(value, shape, fieldName)
				if err != nil {
					return nil, fmt.Errorf("failed to parse JSON relationship %s: %w", fieldName, err)
				}
				if records != nil {
					// Find relationship name
					relName := fieldName
					for _, rel := range shape.Relationships {
						if rel.Name == fieldName {
							relName = rel.Name
							break
						}
					}
					record.Relationships[relName] = records
				}
			} else {
				// Regular field
				record.Fields[fieldName] = e.convertValue(value)
			}
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return records, nil
}

// createScanDest creates appropriate scan destination based on column type.
func (e *PostgresExecutor) createScanDest(ct *sql.ColumnType, shape *engine.ResultShape, index int) any {
	// Check if this is a JSON column (subquery result)
	typeName := ct.DatabaseTypeName()
	if typeName == "JSON" || typeName == "JSONB" {
		return new([]byte)
	}

	// For other types, use interface{}
	return new(any)
}

// isJSONColumn checks if a column type is JSON/JSONB.
func (e *PostgresExecutor) isJSONColumn(ct *sql.ColumnType) bool {
	typeName := ct.DatabaseTypeName()
	return typeName == "JSON" || typeName == "JSONB"
}

// parseJSONRelationship parses a JSON column into nested records.
func (e *PostgresExecutor) parseJSONRelationship(value any, shape *engine.ResultShape, fieldName string) ([]Record, error) {
	var jsonData []byte

	switch v := value.(type) {
	case *[]byte:
		if v == nil || *v == nil {
			return nil, nil
		}
		jsonData = *v
	case []byte:
		if v == nil {
			return nil, nil
		}
		jsonData = v
	case *any:
		if v == nil || *v == nil {
			return nil, nil
		}
		if b, ok := (*v).([]byte); ok {
			jsonData = b
		} else {
			return nil, nil
		}
	default:
		return nil, nil
	}

	if len(jsonData) == 0 {
		return nil, nil
	}

	// Parse JSON array
	var items []map[string]any
	if err := json.Unmarshal(jsonData, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Find relationship shape
	var relShape *engine.ResultShape
	for _, rel := range shape.Relationships {
		if rel.Name == fieldName {
			relShape = rel.Shape
			break
		}
	}

	// Convert to records
	records := make([]Record, 0, len(items))
	for _, item := range items {
		record := Record{
			Fields:        make(map[string]any),
			Relationships: make(map[string][]Record),
		}

		if relShape != nil {
			record.Type = relShape.Object
		}

		for k, v := range item {
			record.Fields[k] = v
		}

		records = append(records, record)
	}

	return records, nil
}

// convertValue converts database values to Go types.
func (e *PostgresExecutor) convertValue(value any) any {
	if value == nil {
		return nil
	}

	// Dereference pointers
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	value = rv.Interface()

	// Handle specific types
	switch v := value.(type) {
	case time.Time:
		return v
	case []byte:
		// Try to convert to string if it looks like text
		return string(v)
	case int64, int32, int, float64, float32, bool, string:
		return v
	default:
		return v
	}
}

// SetDateResolver sets a custom date resolver.
func (e *PostgresExecutor) SetDateResolver(resolver engine.DateResolver) {
	if resolver != nil {
		e.dateResolver = resolver
	}
}
