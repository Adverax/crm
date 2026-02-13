package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/middleware"
	"github.com/adverax/crm/internal/pkg/apperror"
)

// Handler handles auth HTTP endpoints.
type Handler struct {
	service     Service
	rateLimiter *RateLimiter
}

// NewHandler creates a new auth Handler.
func NewHandler(service Service, rateLimiter *RateLimiter) *Handler {
	return &Handler{
		service:     service,
		rateLimiter: rateLimiter,
	}
}

// RegisterPublicRoutes registers unauthenticated auth routes.
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/forgot-password", h.ForgotPassword)
	auth.POST("/reset-password", h.ResetPassword)
}

// RegisterProtectedRoutes registers authenticated auth routes.
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/logout", h.Logout)
	auth.GET("/me", h.Me)
}

func (h *Handler) Login(c *gin.Context) {
	if !h.rateLimiter.Allow(c.ClientIP()) {
		apperror.Respond(c, apperror.BadRequest("too many login attempts, try again later"))
		return
	}

	var req LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	pair, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pair})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	pair, err := h.service.Refresh(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pair})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	if !h.rateLimiter.Allow(c.ClientIP()) {
		apperror.Respond(c, apperror.BadRequest("too many requests, try again later"))
		return
	}

	var req ForgotPasswordInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	if err := h.service.ForgotPassword(c.Request.Context(), req); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password has been reset"})
}

func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		apperror.Respond(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	info, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": info})
}
