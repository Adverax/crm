package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
)

// ObjectViewHandler handles admin CRUD for object views.
type ObjectViewHandler struct {
	service metadata.ObjectViewService
}

// NewObjectViewHandler creates a new ObjectViewHandler.
func NewObjectViewHandler(service metadata.ObjectViewService) *ObjectViewHandler {
	return &ObjectViewHandler{service: service}
}

// RegisterRoutes registers object view routes on the admin group.
func (h *ObjectViewHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/object-views", h.Create)
	rg.GET("/object-views", h.List)
	rg.GET("/object-views/:viewId", h.Get)
	rg.PUT("/object-views/:viewId", h.Update)
	rg.DELETE("/object-views/:viewId", h.Delete)
}

type createObjectViewRequest struct {
	ObjectID    string            `json:"object_id" binding:"required"`
	ProfileID   *string           `json:"profile_id"`
	APIName     string            `json:"api_name" binding:"required"`
	Label       string            `json:"label" binding:"required"`
	Description *string           `json:"description"`
	IsDefault   bool              `json:"is_default"`
	Config      metadata.OVConfig `json:"config"`
}

type updateObjectViewRequest struct {
	Label       string            `json:"label" binding:"required"`
	Description *string           `json:"description"`
	IsDefault   bool              `json:"is_default"`
	Config      metadata.OVConfig `json:"config"`
}

func (h *ObjectViewHandler) Create(c *gin.Context) {
	var req createObjectViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	objectID, err := uuid.Parse(req.ObjectID)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid object_id"))
		return
	}

	input := metadata.CreateObjectViewInput{
		ObjectID:  objectID,
		APIName:   req.APIName,
		Label:     req.Label,
		IsDefault: req.IsDefault,
		Config:    req.Config,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.ProfileID != nil {
		pid, err := uuid.Parse(*req.ProfileID)
		if err != nil {
			apperror.Respond(c, apperror.BadRequest("invalid profile_id"))
			return
		}
		input.ProfileID = &pid
	}

	ov, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ov})
}

func (h *ObjectViewHandler) List(c *gin.Context) {
	objectIDStr := c.Query("object_id")
	if objectIDStr != "" {
		objectID, err := uuid.Parse(objectIDStr)
		if err != nil {
			apperror.Respond(c, apperror.BadRequest("invalid object_id"))
			return
		}
		views, err := h.service.ListByObjectID(c.Request.Context(), objectID)
		if err != nil {
			apperror.Respond(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": views})
		return
	}

	views, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": views})
}

func (h *ObjectViewHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("viewId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid view ID"))
		return
	}

	ov, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ov})
}

func (h *ObjectViewHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("viewId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid view ID"))
		return
	}

	var req updateObjectViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	input := metadata.UpdateObjectViewInput{
		Label:     req.Label,
		IsDefault: req.IsDefault,
		Config:    req.Config,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}

	ov, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ov})
}

func (h *ObjectViewHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("viewId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid view ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
