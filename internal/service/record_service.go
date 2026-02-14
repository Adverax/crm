package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/dml"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
	"github.com/adverax/crm/internal/platform/soql"
)

const (
	defaultPerPage = 20
	maxPerPage     = 100
)

// RecordService provides generic CRUD operations for any metadata-defined object.
type RecordService interface {
	List(ctx context.Context, objectName string, params ListParams) (*RecordListResult, error)
	GetByID(ctx context.Context, objectName string, recordID string) (map[string]any, error)
	Create(ctx context.Context, objectName string, fields map[string]any) (*CreateResult, error)
	Update(ctx context.Context, objectName string, recordID string, fields map[string]any) error
	Delete(ctx context.Context, objectName string, recordID string) error
}

type recordService struct {
	cache       *metadata.MetadataCache
	soqlService soql.QueryService
	dmlService  dml.DMLService
}

// NewRecordService creates a new RecordService.
func NewRecordService(
	cache *metadata.MetadataCache,
	soqlService soql.QueryService,
	dmlService dml.DMLService,
) RecordService {
	return &recordService{
		cache:       cache,
		soqlService: soqlService,
		dmlService:  dmlService,
	}
}

func (s *recordService) List(ctx context.Context, objectName string, params ListParams) (*RecordListResult, error) {
	objDef, ok := s.cache.GetObjectByAPIName(objectName)
	if !ok {
		return nil, fmt.Errorf("recordService.List: %w", apperror.NotFound("object", objectName))
	}

	perPage := normalizePerPage(params.PerPage)
	page := params.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	fieldNames := s.buildSelectFields(objDef)

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT() FROM %s", objectName)
	countResult, err := s.soqlService.Execute(ctx, countQuery, &soql.QueryParams{PageSize: 1})
	if err != nil {
		return nil, fmt.Errorf("recordService.List: count: %w", err)
	}
	total := countResult.TotalSize

	// Data query
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT %d OFFSET %d",
		strings.Join(fieldNames, ", "), objectName, perPage, offset)
	result, err := s.soqlService.Execute(ctx, query, &soql.QueryParams{PageSize: perPage})
	if err != nil {
		return nil, fmt.Errorf("recordService.List: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return &RecordListResult{
		Data: result.Records,
		Pagination: PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *recordService) GetByID(ctx context.Context, objectName string, recordID string) (map[string]any, error) {
	if err := validateUUID(recordID); err != nil {
		return nil, fmt.Errorf("recordService.GetByID: %w", err)
	}

	objDef, ok := s.cache.GetObjectByAPIName(objectName)
	if !ok {
		return nil, fmt.Errorf("recordService.GetByID: %w", apperror.NotFound("object", objectName))
	}

	fieldNames := s.buildSelectFields(objDef)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE Id = '%s'",
		strings.Join(fieldNames, ", "), objectName, recordID)

	result, err := s.soqlService.Execute(ctx, query, &soql.QueryParams{PageSize: 1})
	if err != nil {
		return nil, fmt.Errorf("recordService.GetByID: %w", err)
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("recordService.GetByID: %w", apperror.NotFound("record", recordID))
	}

	return result.Records[0], nil
}

func (s *recordService) Create(ctx context.Context, objectName string, fields map[string]any) (*CreateResult, error) {
	objDef, ok := s.cache.GetObjectByAPIName(objectName)
	if !ok {
		return nil, fmt.Errorf("recordService.Create: %w", apperror.NotFound("object", objectName))
	}

	if !objDef.IsCreateable {
		return nil, fmt.Errorf("recordService.Create: %w", apperror.Forbidden("object is not createable"))
	}

	uc, ok := security.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("recordService.Create: %w", apperror.Unauthorized("user context required"))
	}

	// Inject system fields
	if _, ok := fields["OwnerId"]; !ok {
		fields["OwnerId"] = uc.UserID.String()
	}
	fields["CreatedById"] = uc.UserID.String()
	fields["UpdatedById"] = uc.UserID.String()

	// Inject static defaults from metadata
	s.injectDefaults(objDef, fields)

	fieldNames, fieldValues := buildFieldLists(fields)

	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		objectName,
		strings.Join(fieldNames, ", "),
		strings.Join(fieldValues, ", "))

	result, err := s.dmlService.Execute(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("recordService.Create: %w", err)
	}

	var id string
	if len(result.InsertedIds) > 0 {
		id = result.InsertedIds[0]
	}

	return &CreateResult{ID: id}, nil
}

func (s *recordService) Update(ctx context.Context, objectName string, recordID string, fields map[string]any) error {
	if err := validateUUID(recordID); err != nil {
		return fmt.Errorf("recordService.Update: %w", err)
	}

	objDef, ok := s.cache.GetObjectByAPIName(objectName)
	if !ok {
		return fmt.Errorf("recordService.Update: %w", apperror.NotFound("object", objectName))
	}

	if !objDef.IsUpdateable {
		return fmt.Errorf("recordService.Update: %w", apperror.Forbidden("object is not updateable"))
	}

	uc, ok := security.UserFromContext(ctx)
	if !ok {
		return fmt.Errorf("recordService.Update: %w", apperror.Unauthorized("user context required"))
	}

	fields["UpdatedById"] = uc.UserID.String()

	setClauses := buildSetClauses(fields)

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE Id = '%s'",
		objectName, strings.Join(setClauses, ", "), recordID)

	_, err := s.dmlService.Execute(ctx, stmt)
	if err != nil {
		return fmt.Errorf("recordService.Update: %w", err)
	}

	return nil
}

func (s *recordService) Delete(ctx context.Context, objectName string, recordID string) error {
	if err := validateUUID(recordID); err != nil {
		return fmt.Errorf("recordService.Delete: %w", err)
	}

	objDef, ok := s.cache.GetObjectByAPIName(objectName)
	if !ok {
		return fmt.Errorf("recordService.Delete: %w", apperror.NotFound("object", objectName))
	}

	if !objDef.IsDeleteable {
		return fmt.Errorf("recordService.Delete: %w", apperror.Forbidden("object is not deleteable"))
	}

	stmt := fmt.Sprintf("DELETE FROM %s WHERE Id = '%s'", objectName, recordID)

	_, err := s.dmlService.Execute(ctx, stmt)
	if err != nil {
		return fmt.Errorf("recordService.Delete: %w", err)
	}

	return nil
}

// buildSelectFields returns the list of field names for SOQL SELECT.
func (s *recordService) buildSelectFields(objDef metadata.ObjectDefinition) []string {
	// System fields always included
	fieldNames := []string{"Id", "OwnerId", "CreatedAt", "UpdatedAt", "CreatedById", "UpdatedById"}

	fields := s.cache.GetFieldsByObjectID(objDef.ID)
	for _, f := range fields {
		fieldNames = append(fieldNames, f.APIName)
	}

	return fieldNames
}

// injectDefaults fills in static default values for fields not provided in the input.
func (s *recordService) injectDefaults(objDef metadata.ObjectDefinition, fields map[string]any) {
	defs := s.cache.GetFieldsByObjectID(objDef.ID)
	for _, f := range defs {
		if f.Config.DefaultValue == nil {
			continue
		}
		if _, exists := fields[f.APIName]; exists {
			continue
		}
		fields[f.APIName] = *f.Config.DefaultValue
	}
}

func normalizePerPage(perPage int) int {
	if perPage <= 0 {
		return defaultPerPage
	}
	if perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}

func validateUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return apperror.BadRequest(fmt.Sprintf("invalid UUID: %s", id))
	}
	return nil
}

func buildFieldLists(fields map[string]any) ([]string, []string) {
	names := make([]string, 0, len(fields))
	for k := range fields {
		names = append(names, k)
	}
	sort.Strings(names)

	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, formatDMLValue(fields[name]))
	}

	return names, values
}

func buildSetClauses(fields map[string]any) []string {
	names := make([]string, 0, len(fields))
	for k := range fields {
		names = append(names, k)
	}
	sort.Strings(names)

	clauses := make([]string, 0, len(names))
	for _, name := range names {
		clauses = append(clauses, fmt.Sprintf("%s = %s", name, formatDMLValue(fields[name])))
	}

	return clauses
}

func formatDMLValue(value any) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case float32:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}
