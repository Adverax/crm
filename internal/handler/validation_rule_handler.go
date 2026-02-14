package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// ValidationRuleHandler handles admin CRUD for validation rules.
type ValidationRuleHandler struct {
	service metadata.ValidationRuleService
}

// NewValidationRuleHandler creates a new ValidationRuleHandler.
func NewValidationRuleHandler(service metadata.ValidationRuleService) *ValidationRuleHandler {
	return &ValidationRuleHandler{service: service}
}

// RegisterRoutes registers validation rule routes on the metadata group.
func (h *ValidationRuleHandler) RegisterRoutes(meta *gin.RouterGroup) {
	meta.POST("/objects/:objectId/rules", h.Create)
	meta.GET("/objects/:objectId/rules", h.List)
	meta.GET("/objects/:objectId/rules/:ruleId", h.Get)
	meta.PUT("/objects/:objectId/rules/:ruleId", h.Update)
	meta.DELETE("/objects/:objectId/rules/:ruleId", h.Delete)
}

type createValidationRuleRequest struct {
	ApiName        string  `json:"api_name" binding:"required"`
	Label          string  `json:"label" binding:"required"`
	Expression     string  `json:"expression" binding:"required"`
	ErrorMessage   string  `json:"error_message" binding:"required"`
	ErrorCode      *string `json:"error_code"`
	Severity       *string `json:"severity"`
	WhenExpression *string `json:"when_expression"`
	AppliesTo      *string `json:"applies_to"`
	SortOrder      *int    `json:"sort_order"`
	IsActive       *bool   `json:"is_active"`
	Description    *string `json:"description"`
}

type updateValidationRuleRequest struct {
	Label          string  `json:"label" binding:"required"`
	Expression     string  `json:"expression" binding:"required"`
	ErrorMessage   string  `json:"error_message" binding:"required"`
	ErrorCode      *string `json:"error_code"`
	Severity       *string `json:"severity"`
	WhenExpression *string `json:"when_expression"`
	AppliesTo      *string `json:"applies_to"`
	SortOrder      *int    `json:"sort_order"`
	IsActive       *bool   `json:"is_active"`
	Description    *string `json:"description"`
}

func (h *ValidationRuleHandler) Create(c *gin.Context) {
	objectID, err := uuid.Parse(c.Param("objectId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object ID"))
		return
	}

	var req createValidationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.CreateValidationRuleInput{
		ObjectID:       objectID,
		APIName:        req.ApiName,
		Label:          req.Label,
		Expression:     req.Expression,
		ErrorMessage:   req.ErrorMessage,
		WhenExpression: req.WhenExpression,
		IsActive:       true,
	}

	if req.ErrorCode != nil {
		input.ErrorCode = *req.ErrorCode
	}
	if req.Severity != nil {
		input.Severity = *req.Severity
	}
	if req.AppliesTo != nil {
		input.AppliesTo = *req.AppliesTo
	}
	if req.SortOrder != nil {
		input.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		input.IsActive = *req.IsActive
	}
	if req.Description != nil {
		input.Description = *req.Description
	}

	rule, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

func (h *ValidationRuleHandler) List(c *gin.Context) {
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

func (h *ValidationRuleHandler) Get(c *gin.Context) {
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

func (h *ValidationRuleHandler) Update(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("ruleId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid rule ID"))
		return
	}

	var req updateValidationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateValidationRuleInput{
		Label:          req.Label,
		Expression:     req.Expression,
		ErrorMessage:   req.ErrorMessage,
		WhenExpression: req.WhenExpression,
		IsActive:       true,
		Severity:       "error",
		ErrorCode:      "validation_failed",
		AppliesTo:      "create,update",
	}

	if req.ErrorCode != nil {
		input.ErrorCode = *req.ErrorCode
	}
	if req.Severity != nil {
		input.Severity = *req.Severity
	}
	if req.AppliesTo != nil {
		input.AppliesTo = *req.AppliesTo
	}
	if req.SortOrder != nil {
		input.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		input.IsActive = *req.IsActive
	}
	if req.Description != nil {
		input.Description = *req.Description
	}

	rule, err := h.service.Update(c.Request.Context(), ruleID, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

func (h *ValidationRuleHandler) Delete(c *gin.Context) {
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
