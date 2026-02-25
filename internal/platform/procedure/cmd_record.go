package procedure

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/dml"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/soql"
)

// RecordCommandExecutor handles record.create, record.update, record.delete,
// record.get, and record.query commands.
type RecordCommandExecutor struct {
	dmlSvc   dml.DMLService
	querySvc soql.QueryService
	resolver *ExpressionResolver
}

// NewRecordCommandExecutor creates a new RecordCommandExecutor.
func NewRecordCommandExecutor(dmlSvc dml.DMLService, querySvc soql.QueryService, resolver *ExpressionResolver) *RecordCommandExecutor {
	return &RecordCommandExecutor{
		dmlSvc:   dmlSvc,
		querySvc: querySvc,
		resolver: resolver,
	}
}

func (e *RecordCommandExecutor) Category() string { return "record" }

func (e *RecordCommandExecutor) Execute(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	parts := strings.SplitN(cmd.Type, ".", 2)
	if len(parts) != 2 {
		return nil, apperror.BadRequest("invalid record command type")
	}

	switch parts[1] {
	case "create":
		return e.executeCreate(ctx, cmd, execCtx)
	case "update":
		return e.executeUpdate(ctx, cmd, execCtx)
	case "delete":
		return e.executeDelete(ctx, cmd, execCtx)
	case "get":
		return e.executeGet(ctx, cmd, execCtx)
	case "query":
		return e.executeQuery(ctx, cmd, execCtx)
	default:
		return nil, apperror.BadRequest(fmt.Sprintf("unknown record command: %s", cmd.Type))
	}
}

func (e *RecordCommandExecutor) executeCreate(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Object == "" {
		return nil, apperror.BadRequest("record.create requires 'object'")
	}

	data, err := e.resolver.ResolveMap(cmd.Data, execCtx)
	if err != nil {
		return nil, fmt.Errorf("record.create: resolve data: %w", err)
	}

	if execCtx.DryRun {
		fakeID := uuid.New().String()
		return map[string]any{"id": fakeID}, nil
	}

	// Build DML INSERT statement
	fields := make([]string, 0, len(data))
	values := make([]string, 0, len(data))
	for k, v := range data {
		fields = append(fields, k)
		values = append(values, formatDMLValue(v))
	}

	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		cmd.Object,
		strings.Join(fields, ", "),
		strings.Join(values, ", "),
	)

	result, err := e.dmlSvc.Execute(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("record.create: %w", err)
	}

	response := map[string]any{
		"rows_affected": result.RowsAffected,
	}
	if len(result.InsertedIds) > 0 {
		response["id"] = result.InsertedIds[0]
	}

	return response, nil
}

func (e *RecordCommandExecutor) executeUpdate(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Object == "" {
		return nil, apperror.BadRequest("record.update requires 'object'")
	}
	if cmd.ID == "" {
		return nil, apperror.BadRequest("record.update requires 'id'")
	}

	recordID, err := e.resolver.ResolveString(cmd.ID, execCtx)
	if err != nil {
		return nil, fmt.Errorf("record.update: resolve id: %w", err)
	}

	data, err := e.resolver.ResolveMap(cmd.Data, execCtx)
	if err != nil {
		return nil, fmt.Errorf("record.update: resolve data: %w", err)
	}

	if execCtx.DryRun {
		return map[string]any{"id": recordID, "rows_affected": 1}, nil
	}

	sets := make([]string, 0, len(data))
	for k, v := range data {
		sets = append(sets, fmt.Sprintf("%s = %s", k, formatDMLValue(v)))
	}

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE id = '%s'",
		cmd.Object,
		strings.Join(sets, ", "),
		recordID,
	)

	result, err := e.dmlSvc.Execute(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("record.update: %w", err)
	}

	return map[string]any{
		"id":            recordID,
		"rows_affected": result.RowsAffected,
	}, nil
}

func (e *RecordCommandExecutor) executeDelete(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Object == "" {
		return nil, apperror.BadRequest("record.delete requires 'object'")
	}
	if cmd.ID == "" {
		return nil, apperror.BadRequest("record.delete requires 'id'")
	}

	recordID, err := e.resolver.ResolveString(cmd.ID, execCtx)
	if err != nil {
		return nil, fmt.Errorf("record.delete: resolve id: %w", err)
	}

	if execCtx.DryRun {
		return map[string]any{"id": recordID, "rows_affected": 1}, nil
	}

	stmt := fmt.Sprintf("DELETE FROM %s WHERE id = '%s'", cmd.Object, recordID)
	result, err := e.dmlSvc.Execute(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("record.delete: %w", err)
	}

	return map[string]any{
		"id":            recordID,
		"rows_affected": result.RowsAffected,
	}, nil
}

func (e *RecordCommandExecutor) executeGet(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Object == "" {
		return nil, apperror.BadRequest("record.get requires 'object'")
	}
	if cmd.ID == "" {
		return nil, apperror.BadRequest("record.get requires 'id'")
	}

	recordID, err := e.resolver.ResolveString(cmd.ID, execCtx)
	if err != nil {
		return nil, fmt.Errorf("record.get: resolve id: %w", err)
	}

	if execCtx.DryRun {
		return map[string]any{"id": recordID}, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE Id = '%s'", cmd.Object, recordID)
	result, err := e.querySvc.Execute(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("record.get: %w", err)
	}

	if len(result.Records) == 0 {
		return nil, apperror.NotFound(cmd.Object, recordID)
	}

	return result.Records[0], nil
}

func (e *RecordCommandExecutor) executeQuery(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if cmd.Query == "" {
		return nil, apperror.BadRequest("record.query requires 'query'")
	}

	query, err := e.resolver.ResolveString(cmd.Query, execCtx)
	if err != nil {
		return nil, fmt.Errorf("record.query: resolve query: %w", err)
	}

	if execCtx.DryRun {
		return []map[string]any{}, nil
	}

	result, err := e.querySvc.Execute(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("record.query: %w", err)
	}

	return result.Records, nil
}

// formatDMLValue formats a Go value for use in a DML statement.
func formatDMLValue(v any) string {
	switch val := v.(type) {
	case string:
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", val)
	}
}
