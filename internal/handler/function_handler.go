package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// FunctionHandler handles admin CRUD for custom functions.
type FunctionHandler struct {
	service metadata.FunctionService
}

// NewFunctionHandler creates a new FunctionHandler.
func NewFunctionHandler(service metadata.FunctionService) *FunctionHandler {
	return &FunctionHandler{service: service}
}

// RegisterRoutes registers function routes on the admin metadata group.
func (h *FunctionHandler) RegisterRoutes(meta *gin.RouterGroup) {
	meta.POST("/functions", h.Create)
	meta.GET("/functions", h.List)
	meta.GET("/functions/:functionId", h.Get)
	meta.PUT("/functions/:functionId", h.Update)
	meta.DELETE("/functions/:functionId", h.Delete)
}

type createFunctionRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description *string                  `json:"description"`
	Params      []metadata.FunctionParam `json:"params"`
	ReturnType  *string                  `json:"return_type"`
	Body        string                   `json:"body" binding:"required"`
}

type updateFunctionRequest struct {
	Description *string                  `json:"description"`
	Params      []metadata.FunctionParam `json:"params"`
	ReturnType  *string                  `json:"return_type"`
	Body        string                   `json:"body" binding:"required"`
}

func (h *FunctionHandler) Create(c *gin.Context) {
	var req createFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.CreateFunctionInput{
		Name:   req.Name,
		Body:   req.Body,
		Params: req.Params,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.ReturnType != nil {
		input.ReturnType = *req.ReturnType
	}

	fn, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": fn})
}

func (h *FunctionHandler) List(c *gin.Context) {
	functions, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": functions})
}

func (h *FunctionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("functionId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid function ID"))
		return
	}

	fn, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fn})
}

func (h *FunctionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("functionId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid function ID"))
		return
	}

	var req updateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateFunctionInput{
		Body:   req.Body,
		Params: req.Params,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.ReturnType != nil {
		input.ReturnType = *req.ReturnType
	}

	fn, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fn})
}

func (h *FunctionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("functionId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid function ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
