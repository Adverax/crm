package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/service"
)

// RecordHandler handles generic CRUD operations for any object.
type RecordHandler struct {
	recordService service.RecordService
}

// NewRecordHandler creates a new RecordHandler.
func NewRecordHandler(recordService service.RecordService) *RecordHandler {
	return &RecordHandler{recordService: recordService}
}

// RegisterRoutes registers record CRUD routes on the given group.
func (h *RecordHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/records/:objectName", h.ListRecords)
	rg.GET("/records/:objectName/:recordId", h.GetRecord)
	rg.POST("/records/:objectName", h.CreateRecord)
	rg.PUT("/records/:objectName/:recordId", h.UpdateRecord)
	rg.DELETE("/records/:objectName/:recordId", h.DeleteRecord)
}

// ListRecords handles GET /api/v1/records/:objectName
func (h *RecordHandler) ListRecords(c *gin.Context) {
	objectName := c.Param("objectName")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	result, err := h.recordService.List(c.Request.Context(), objectName, service.ListParams{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetRecord handles GET /api/v1/records/:objectName/:recordId
func (h *RecordHandler) GetRecord(c *gin.Context) {
	objectName := c.Param("objectName")
	recordID := c.Param("recordId")

	record, err := h.recordService.GetByID(c.Request.Context(), objectName, recordID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": record})
}

// CreateRecord handles POST /api/v1/records/:objectName
func (h *RecordHandler) CreateRecord(c *gin.Context) {
	objectName := c.Param("objectName")

	var fields map[string]any
	if err := c.ShouldBindJSON(&fields); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	result, err := h.recordService.Create(c.Request.Context(), objectName, fields)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// UpdateRecord handles PUT /api/v1/records/:objectName/:recordId
func (h *RecordHandler) UpdateRecord(c *gin.Context) {
	objectName := c.Param("objectName")
	recordID := c.Param("recordId")

	var fields map[string]any
	if err := c.ShouldBindJSON(&fields); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	if err := h.recordService.Update(c.Request.Context(), objectName, recordID, fields); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"success": true}})
}

// DeleteRecord handles DELETE /api/v1/records/:objectName/:recordId
func (h *RecordHandler) DeleteRecord(c *gin.Context) {
	objectName := c.Param("objectName")
	recordID := c.Param("recordId")

	if err := h.recordService.Delete(c.Request.Context(), objectName, recordID); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
