package handler

import (
	"context"
	"errors"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/soql"
	soqlengine "github.com/adverax/crm/internal/platform/soql/engine"
)

// SOQLHandler handles SOQL query validation and test execution requests.
type SOQLHandler struct {
	engine       *soqlengine.Engine
	queryService soql.QueryService
	metadata     metadata.MetadataReader
}

// NewSOQLHandler creates a new SOQLHandler.
func NewSOQLHandler(engine *soqlengine.Engine, queryService soql.QueryService, meta metadata.MetadataReader) *SOQLHandler {
	return &SOQLHandler{engine: engine, queryService: queryService, metadata: meta}
}

// RegisterRoutes registers SOQL routes on the admin group.
func (h *SOQLHandler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.POST("/soql/validate", h.Validate)
	admin.POST("/soql/test", h.TestQuery)
	admin.GET("/soql/objects", h.ListObjects)
	admin.GET("/soql/objects/:objectName/fields", h.ListFields)
}

type soqlValidateRequest struct {
	Query string `json:"query" binding:"required"`
}

type soqlValidateResponse struct {
	Valid  bool                `json:"valid"`
	Object *string             `json:"object,omitempty"`
	Fields []string            `json:"fields,omitempty"`
	Errors []soqlValidateError `json:"errors,omitempty"`
}

type soqlValidateError struct {
	Message string  `json:"message"`
	Line    *int    `json:"line,omitempty"`
	Column  *int    `json:"column,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// Validate handles POST /api/v1/admin/soql/validate.
func (h *SOQLHandler) Validate(c *gin.Context) {
	var req soqlValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Step 1: Parse
	ast, err := h.engine.Parse(req.Query)
	if err != nil {
		c.JSON(http.StatusOK, h.buildErrorResponse(err))
		return
	}

	// Step 2: Validate against metadata
	validated, err := h.engine.Validate(c.Request.Context(), ast)
	if err != nil {
		c.JSON(http.StatusOK, h.buildErrorResponse(err))
		return
	}

	// Success: extract object and field names
	object := validated.RootObject.Name
	fields := h.extractFieldNames(validated)

	c.JSON(http.StatusOK, soqlValidateResponse{
		Valid:  true,
		Object: &object,
		Fields: fields,
	})
}

type soqlTestRequest struct {
	Query    string `json:"query" binding:"required"`
	PageSize int    `json:"pageSize"`
}

// TestQuery handles POST /api/v1/admin/soql/test.
// Executes a SOQL query without security enforcement (design-time admin tool).
func (h *SOQLHandler) TestQuery(c *gin.Context) {
	var req soqlTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.PageSize <= 0 {
		req.PageSize = 5
	}
	if req.PageSize > 20 {
		req.PageSize = 20
	}

	params := &soql.QueryParams{PageSize: req.PageSize}
	// Use a clean context without user identity to bypass OLS/FLS/RLS (design-time admin tool).
	result, err := h.queryService.Execute(context.Background(), req.Query, params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"totalSize": 0,
			"records":   []any{},
			"error":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalSize": result.TotalSize,
		"records":   result.Records,
	})
}

type soqlObjectItem struct {
	APIName string `json:"api_name"`
	Label   string `json:"label"`
}

type soqlFieldItem struct {
	APIName   string `json:"api_name"`
	Label     string `json:"label"`
	FieldType string `json:"field_type"`
}

// ListObjects handles GET /api/v1/admin/soql/objects.
// Returns all queryable objects without OLS filtering (design-time admin tool).
func (h *SOQLHandler) ListObjects(c *gin.Context) {
	names := h.metadata.ListObjectAPINames()
	sort.Strings(names)

	result := make([]soqlObjectItem, 0, len(names))
	for _, name := range names {
		obj, ok := h.metadata.GetObjectByAPIName(name)
		if !ok || !obj.IsQueryable {
			continue
		}
		result = append(result, soqlObjectItem{
			APIName: obj.APIName,
			Label:   obj.Label,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ListFields handles GET /api/v1/admin/soql/objects/:objectName/fields.
// Returns all fields for an object without FLS filtering (design-time admin tool).
func (h *SOQLHandler) ListFields(c *gin.Context) {
	objectName := c.Param("objectName")

	obj, ok := h.metadata.GetObjectByAPIName(objectName)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code":    "NOT_FOUND",
			"message": "object not found: " + objectName,
		}})
		return
	}

	fields := h.metadata.GetFieldsByObjectID(obj.ID)
	result := make([]soqlFieldItem, 0, len(fields))
	for _, f := range fields {
		if f.IsSystemField {
			continue
		}
		result = append(result, soqlFieldItem{
			APIName:   f.APIName,
			Label:     f.Label,
			FieldType: string(f.FieldType),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].APIName < result[j].APIName
	})

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *SOQLHandler) extractFieldNames(validated *soqlengine.ValidatedQuery) []string {
	if len(validated.ResolvedRefs) == 0 {
		return nil
	}

	fields := make([]string, 0, len(validated.ResolvedRefs))
	for path := range validated.ResolvedRefs {
		fields = append(fields, path)
	}
	sort.Strings(fields)
	return fields
}

func (h *SOQLHandler) buildErrorResponse(err error) soqlValidateResponse {
	var errs []soqlValidateError

	var parseErr *soqlengine.ParseError
	var validationErr *soqlengine.ValidationError
	var accessErr *soqlengine.AccessError
	var limitErr *soqlengine.LimitError

	switch {
	case errors.As(err, &parseErr):
		ve := soqlValidateError{Message: parseErr.Message}
		if parseErr.Pos.Line > 0 {
			ve.Line = &parseErr.Pos.Line
			ve.Column = &parseErr.Pos.Column
		}
		code := "ParseError"
		ve.Code = &code
		errs = append(errs, ve)

	case errors.As(err, &validationErr):
		ve := soqlValidateError{Message: validationErr.Message}
		if validationErr.Pos.Line > 0 {
			ve.Line = &validationErr.Pos.Line
			ve.Column = &validationErr.Pos.Column
		}
		code := validationErr.Code.String()
		ve.Code = &code
		errs = append(errs, ve)

	case errors.As(err, &accessErr):
		ve := soqlValidateError{Message: accessErr.Message}
		code := "AccessError"
		ve.Code = &code
		errs = append(errs, ve)

	case errors.As(err, &limitErr):
		ve := soqlValidateError{Message: limitErr.Message}
		code := "LimitError"
		ve.Code = &code
		errs = append(errs, ve)

	default:
		errs = append(errs, soqlValidateError{Message: err.Error()})
	}

	return soqlValidateResponse{
		Valid:  false,
		Errors: errs,
	}
}
