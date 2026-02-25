package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// AutomationRuleHandler handles admin CRUD for automation rules (ADR-0031).
type AutomationRuleHandler struct {
	service metadata.AutomationRuleService
}

// NewAutomationRuleHandler creates a new AutomationRuleHandler.
func NewAutomationRuleHandler(service metadata.AutomationRuleService) *AutomationRuleHandler {
	return &AutomationRuleHandler{service: service}
}

// RegisterRoutes registers automation rule routes on the admin group.
func (h *AutomationRuleHandler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.GET("/metadata/objects/:objectId/automation-rules", h.List)
	admin.POST("/metadata/objects/:objectId/automation-rules", h.Create)
	admin.GET("/metadata/automation-rules/:ruleId", h.Get)
	admin.PUT("/metadata/automation-rules/:ruleId", h.Update)
	admin.DELETE("/metadata/automation-rules/:ruleId", h.Delete)
}

type createAutomationRuleRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	EventType     string  `json:"event_type" binding:"required"`
	Condition     *string `json:"condition"`
	ProcedureCode string  `json:"procedure_code" binding:"required"`
	ExecutionMode string  `json:"execution_mode"`
	SortOrder     *int    `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
}

type updateAutomationRuleRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	EventType     string  `json:"event_type" binding:"required"`
	Condition     *string `json:"condition"`
	ProcedureCode string  `json:"procedure_code" binding:"required"`
	ExecutionMode string  `json:"execution_mode"`
	SortOrder     *int    `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
}

func (h *AutomationRuleHandler) Create(c *gin.Context) {
	objectID, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	var req createAutomationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	executionMode := req.ExecutionMode
	if executionMode == "" {
		executionMode = "per_record"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	input := metadata.CreateAutomationRuleInput{
		ObjectID:      objectID,
		Name:          req.Name,
		Description:   req.Description,
		EventType:     req.EventType,
		Condition:     req.Condition,
		ProcedureCode: req.ProcedureCode,
		ExecutionMode: executionMode,
		SortOrder:     sortOrder,
		IsActive:      isActive,
	}

	rule, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

func (h *AutomationRuleHandler) List(c *gin.Context) {
	objectID, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	rules, err := h.service.ListByObjectID(c.Request.Context(), objectID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

func (h *AutomationRuleHandler) Get(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("ruleId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid rule ID"))
		return
	}

	rule, err := h.service.GetByID(c.Request.Context(), ruleID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

func (h *AutomationRuleHandler) Update(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("ruleId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid rule ID"))
		return
	}

	var req updateAutomationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	executionMode := req.ExecutionMode
	if executionMode == "" {
		executionMode = "per_record"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	input := metadata.UpdateAutomationRuleInput{
		Name:          req.Name,
		Description:   req.Description,
		EventType:     req.EventType,
		Condition:     req.Condition,
		ProcedureCode: req.ProcedureCode,
		ExecutionMode: executionMode,
		SortOrder:     sortOrder,
		IsActive:      isActive,
	}

	rule, err := h.service.Update(c.Request.Context(), ruleID, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

func (h *AutomationRuleHandler) Delete(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("ruleId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid rule ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), ruleID); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
