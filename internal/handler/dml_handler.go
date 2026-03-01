package handler

import (
	"context"
	"errors"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
)

// DMLHandler handles DML statement validation and test execution requests.
type DMLHandler struct {
	engine   *engine.Engine
	pool     *pgxpool.Pool
	metadata metadata.MetadataReader
}

// NewDMLHandler creates a new DMLHandler.
func NewDMLHandler(eng *engine.Engine, pool *pgxpool.Pool, meta metadata.MetadataReader) *DMLHandler {
	return &DMLHandler{engine: eng, pool: pool, metadata: meta}
}

// RegisterRoutes registers DML routes on the admin group.
func (h *DMLHandler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.POST("/dml/validate", h.Validate)
	admin.POST("/dml/test", h.TestExecute)
	admin.GET("/dml/objects", h.ListObjects)
	admin.GET("/dml/objects/:objectName/fields", h.ListFields)
}

type dmlValidateRequest struct {
	Statement string `json:"statement" binding:"required"`
}

type dmlValidateResponse struct {
	Valid     bool               `json:"valid"`
	Operation *string            `json:"operation,omitempty"`
	Object    *string            `json:"object,omitempty"`
	Fields    []string           `json:"fields,omitempty"`
	SQL       *string            `json:"sql,omitempty"`
	Errors    []dmlValidateError `json:"errors,omitempty"`
}

type dmlValidateError struct {
	Message string  `json:"message"`
	Line    *int    `json:"line,omitempty"`
	Column  *int    `json:"column,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// Validate handles POST /api/v1/admin/dml/validate.
func (h *DMLHandler) Validate(c *gin.Context) {
	var req dmlValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Stage 1: Parse
	ast, err := h.engine.Parse(req.Statement)
	if err != nil {
		c.JSON(http.StatusOK, h.buildErrorResponse(err))
		return
	}

	// Stage 2: Validate against metadata
	validated, err := h.engine.Validate(c.Request.Context(), ast)
	if err != nil {
		c.JSON(http.StatusOK, h.buildErrorResponse(err))
		return
	}

	// Stage 3: Compile to SQL
	compiled, err := h.engine.Compile(validated)
	if err != nil {
		c.JSON(http.StatusOK, h.buildErrorResponse(err))
		return
	}

	operation := validated.Operation.String()
	object := ast.GetObject()
	fields := h.extractFieldNames(validated)

	c.JSON(http.StatusOK, dmlValidateResponse{
		Valid:     true,
		Operation: &operation,
		Object:    &object,
		Fields:    fields,
		SQL:       &compiled.SQL,
	})
}

type dmlTestRequest struct {
	Statement string `json:"statement" binding:"required"`
}

type dmlTestResponse struct {
	Operation    string  `json:"operation"`
	Object       string  `json:"object"`
	RowsAffected int64   `json:"rows_affected"`
	RolledBack   bool    `json:"rolled_back"`
	Error        *string `json:"error,omitempty"`
}

// TestExecute handles POST /api/v1/admin/dml/test.
// Executes a DML statement in a rolled-back transaction (design-time admin tool).
func (h *DMLHandler) TestExecute(c *gin.Context) {
	var req dmlTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Parse + Validate + Compile (using clean context, no OLS/FLS)
	compiled, err := h.engine.Prepare(context.Background(), req.Statement)
	if err != nil {
		errMsg := err.Error()
		c.JSON(http.StatusOK, dmlTestResponse{
			Error:      &errMsg,
			RolledBack: true,
		})
		return
	}

	// Execute in a rolled-back transaction
	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		errMsg := "failed to begin transaction"
		c.JSON(http.StatusOK, dmlTestResponse{
			Error:      &errMsg,
			RolledBack: true,
		})
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	executor := engine.NewDefaultExecutor(tx)
	result, err := executor.Execute(c.Request.Context(), compiled)
	if err != nil {
		errMsg := err.Error()
		c.JSON(http.StatusOK, dmlTestResponse{
			Operation:  compiled.Operation.String(),
			Object:     compiled.Object,
			Error:      &errMsg,
			RolledBack: true,
		})
		return
	}

	// Always rollback — this is a test/dry-run
	c.JSON(http.StatusOK, dmlTestResponse{
		Operation:    compiled.Operation.String(),
		Object:       compiled.Object,
		RowsAffected: result.RowsAffected,
		RolledBack:   true,
	})
}

// ListObjects handles GET /api/v1/admin/dml/objects.
// Returns all writable objects without OLS filtering (design-time admin tool).
func (h *DMLHandler) ListObjects(c *gin.Context) {
	names := h.metadata.ListObjectAPINames()
	sort.Strings(names)

	result := make([]dmlObjectItem, 0, len(names))
	for _, name := range names {
		obj, ok := h.metadata.GetObjectByAPIName(name)
		if !ok {
			continue
		}
		result = append(result, dmlObjectItem{
			APIName: obj.APIName,
			Label:   obj.Label,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

type dmlObjectItem struct {
	APIName string `json:"api_name"`
	Label   string `json:"label"`
}

type dmlFieldItem struct {
	APIName   string `json:"api_name"`
	Label     string `json:"label"`
	FieldType string `json:"field_type"`
}

// ListFields handles GET /api/v1/admin/dml/objects/:objectName/fields.
// Returns all writable fields for an object without FLS filtering (design-time admin tool).
func (h *DMLHandler) ListFields(c *gin.Context) {
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
	result := make([]dmlFieldItem, 0, len(fields))
	for _, f := range fields {
		if f.IsSystemField {
			continue
		}
		result = append(result, dmlFieldItem{
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

func (h *DMLHandler) extractFieldNames(validated *engine.ValidatedDML) []string {
	switch validated.Operation {
	case engine.OperationInsert, engine.OperationUpsert:
		fields := make([]string, 0, len(validated.Fields))
		for _, f := range validated.Fields {
			fields = append(fields, f.Name)
		}
		sort.Strings(fields)
		return fields
	case engine.OperationUpdate:
		fields := make([]string, 0, len(validated.Assignments))
		for _, a := range validated.Assignments {
			fields = append(fields, a.Field.Name)
		}
		sort.Strings(fields)
		return fields
	default:
		return nil
	}
}

func (h *DMLHandler) buildErrorResponse(err error) dmlValidateResponse {
	var errs []dmlValidateError

	var parseErr *engine.ParseError
	var validationErr *engine.ValidationError
	var accessErr *engine.AccessError
	var limitErr *engine.LimitError

	switch {
	case errors.As(err, &parseErr):
		ve := dmlValidateError{Message: parseErr.Message}
		if parseErr.Pos.Line > 0 {
			ve.Line = &parseErr.Pos.Line
			ve.Column = &parseErr.Pos.Column
		}
		code := "ParseError"
		ve.Code = &code
		errs = append(errs, ve)

	case errors.As(err, &validationErr):
		ve := dmlValidateError{Message: validationErr.Message}
		if validationErr.Pos.Line > 0 {
			ve.Line = &validationErr.Pos.Line
			ve.Column = &validationErr.Pos.Column
		}
		code := validationErr.Code.String()
		ve.Code = &code
		errs = append(errs, ve)

	case errors.As(err, &accessErr):
		ve := dmlValidateError{Message: accessErr.Message}
		code := "AccessError"
		ve.Code = &code
		errs = append(errs, ve)

	case errors.As(err, &limitErr):
		ve := dmlValidateError{Message: limitErr.Message}
		code := "LimitError"
		ve.Code = &code
		errs = append(errs, ve)

	default:
		errs = append(errs, dmlValidateError{Message: err.Error()})
	}

	return dmlValidateResponse{
		Valid:  false,
		Errors: errs,
	}
}
