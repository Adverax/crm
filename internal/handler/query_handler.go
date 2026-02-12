package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/dml"
	"github.com/adverax/crm/internal/platform/soql"
)

// QueryHandler handles SOQL and DML API endpoints.
type QueryHandler struct {
	soqlService soql.QueryService
	dmlService  dml.DMLService
}

// NewQueryHandler creates a new QueryHandler.
func NewQueryHandler(soqlService soql.QueryService, dmlService dml.DMLService) *QueryHandler {
	return &QueryHandler{
		soqlService: soqlService,
		dmlService:  dmlService,
	}
}

// RegisterRoutes registers query/data routes on the given group.
func (h *QueryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/query", h.ExecuteQuery)
	rg.POST("/query", h.ExecuteQueryPost)
	rg.POST("/data", h.ExecuteDML)
}

type queryRequest struct {
	Query    string `json:"query" binding:"required"`
	PageSize int    `json:"pageSize"`
}

type dmlRequest struct {
	Statement string `json:"statement" binding:"required"`
}

// ExecuteQuery handles GET /api/v1/query?q=SELECT...
func (h *QueryHandler) ExecuteQuery(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		apperror.Respond(c, apperror.BadRequest("query parameter 'q' is required"))
		return
	}

	params := &soql.QueryParams{}
	result, err := h.soqlService.Execute(c.Request.Context(), q, params)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExecuteQueryPost handles POST /api/v1/query
func (h *QueryHandler) ExecuteQueryPost(c *gin.Context) {
	var req queryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	params := &soql.QueryParams{PageSize: req.PageSize}
	result, err := h.soqlService.Execute(c.Request.Context(), req.Query, params)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExecuteDML handles POST /api/v1/data
func (h *QueryHandler) ExecuteDML(c *gin.Context) {
	var req dmlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	result, err := h.dmlService.Execute(c.Request.Context(), req.Statement)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
