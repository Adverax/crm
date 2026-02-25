package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/credential"
)

// CredentialHandler handles admin CRUD for named credentials.
type CredentialHandler struct {
	service credential.Service
}

// NewCredentialHandler creates a new CredentialHandler.
func NewCredentialHandler(service credential.Service) *CredentialHandler {
	return &CredentialHandler{service: service}
}

// RegisterRoutes registers credential routes on the admin group.
func (h *CredentialHandler) RegisterRoutes(admin *gin.RouterGroup) {
	g := admin.Group("/credentials")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:credentialId", h.Get)
	g.PUT("/:credentialId", h.Update)
	g.DELETE("/:credentialId", h.Delete)
	g.POST("/:credentialId/test", h.TestConnection)
	g.GET("/:credentialId/usage", h.GetUsageLog)
	g.POST("/:credentialId/deactivate", h.Deactivate)
	g.POST("/:credentialId/activate", h.Activate)
}

type createCredentialRequest struct {
	Code        string                    `json:"code" binding:"required"`
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	Type        credential.CredentialType `json:"type" binding:"required"`
	BaseURL     string                    `json:"base_url" binding:"required"`
	AuthData    json.RawMessage           `json:"auth_data" binding:"required"`
}

type updateCredentialRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	BaseURL     string          `json:"base_url" binding:"required"`
	AuthData    json.RawMessage `json:"auth_data,omitempty"`
}

func (h *CredentialHandler) Create(c *gin.Context) {
	var req createCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	result, err := h.service.Create(c.Request.Context(), credential.CreateCredentialInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		BaseURL:     req.BaseURL,
		AuthData:    []byte(req.AuthData),
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *CredentialHandler) List(c *gin.Context) {
	credentials, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": credentials})
}

func (h *CredentialHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("credentialId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid credential ID"))
		return
	}

	cred, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cred})
}

func (h *CredentialHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("credentialId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid credential ID"))
		return
	}

	var req updateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	var authData []byte
	if req.AuthData != nil {
		authData = []byte(req.AuthData)
	}

	result, err := h.service.Update(c.Request.Context(), id, credential.UpdateCredentialInput{
		Name:        req.Name,
		Description: req.Description,
		BaseURL:     req.BaseURL,
		AuthData:    authData,
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *CredentialHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("credentialId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid credential ID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CredentialHandler) TestConnection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("credentialId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid credential ID"))
		return
	}

	if err := h.service.TestConnection(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"success": true}})
}

func (h *CredentialHandler) GetUsageLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("credentialId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid credential ID"))
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	entries, err := h.service.GetUsageLog(c.Request.Context(), id, limit)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entries})
}

func (h *CredentialHandler) Deactivate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("credentialId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid credential ID"))
		return
	}

	if err := h.service.Deactivate(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"success": true}})
}

func (h *CredentialHandler) Activate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("credentialId"))
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid credential ID"))
		return
	}

	if err := h.service.Activate(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"success": true}})
}
