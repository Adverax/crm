package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// SharedLayoutHandler handles admin CRUD for shared layouts.
type SharedLayoutHandler struct {
	service metadata.SharedLayoutService
}

// NewSharedLayoutHandler creates a new SharedLayoutHandler.
func NewSharedLayoutHandler(service metadata.SharedLayoutService) *SharedLayoutHandler {
	return &SharedLayoutHandler{service: service}
}

// RegisterRoutes registers shared layout routes on the admin group.
func (h *SharedLayoutHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/shared-layouts", h.Create)
	rg.GET("/shared-layouts", h.List)
	rg.GET("/shared-layouts/:sharedLayoutId", h.Get)
	rg.PUT("/shared-layouts/:sharedLayoutId", h.Update)
	rg.DELETE("/shared-layouts/:sharedLayoutId", h.Delete)
}

type createSharedLayoutRequest struct {
	APIName string          `json:"api_name" binding:"required"`
	Type    string          `json:"type" binding:"required"`
	Label   string          `json:"label"`
	Config  json.RawMessage `json:"config"`
}

type updateSharedLayoutRequest struct {
	Label  string          `json:"label"`
	Config json.RawMessage `json:"config"`
}

func (h *SharedLayoutHandler) Create(c *gin.Context) {
	var req createSharedLayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.CreateSharedLayoutInput{
		APIName: req.APIName,
		Type:    req.Type,
		Label:   req.Label,
		Config:  req.Config,
	}

	sl, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": sl})
}

func (h *SharedLayoutHandler) List(c *gin.Context) {
	layouts, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": layouts})
}

func (h *SharedLayoutHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("sharedLayoutId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid shared layout ID"))
		return
	}

	sl, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sl})
}

func (h *SharedLayoutHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("sharedLayoutId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid shared layout ID"))
		return
	}

	var req updateSharedLayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateSharedLayoutInput{
		Label:  req.Label,
		Config: req.Config,
	}

	sl, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sl})
}

func (h *SharedLayoutHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("sharedLayoutId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid shared layout ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
