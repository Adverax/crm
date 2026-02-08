package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

// SecurityHandler handles admin CRUD for IAM/security resources.
type SecurityHandler struct {
	roleService       security.UserRoleService
	psService         security.PermissionSetService
	profileService    security.ProfileService
	userService       security.UserService
	permissionService security.PermissionService
}

// NewSecurityHandler creates a new SecurityHandler.
func NewSecurityHandler(
	roleService security.UserRoleService,
	psService security.PermissionSetService,
	profileService security.ProfileService,
	userService security.UserService,
	permissionService security.PermissionService,
) *SecurityHandler {
	return &SecurityHandler{
		roleService:       roleService,
		psService:         psService,
		profileService:    profileService,
		userService:       userService,
		permissionService: permissionService,
	}
}

// RegisterRoutes registers security admin routes on the given router group.
func (h *SecurityHandler) RegisterRoutes(rg *gin.RouterGroup) {
	sec := rg.Group("/security")

	sec.POST("/roles", h.CreateRole)
	sec.GET("/roles", h.ListRoles)
	sec.GET("/roles/:id", h.GetRole)
	sec.PUT("/roles/:id", h.UpdateRole)
	sec.DELETE("/roles/:id", h.DeleteRole)

	sec.POST("/permission-sets", h.CreatePermissionSet)
	sec.GET("/permission-sets", h.ListPermissionSets)
	sec.GET("/permission-sets/:id", h.GetPermissionSet)
	sec.PUT("/permission-sets/:id", h.UpdatePermissionSet)
	sec.DELETE("/permission-sets/:id", h.DeletePermissionSet)

	sec.POST("/profiles", h.CreateProfile)
	sec.GET("/profiles", h.ListProfiles)
	sec.GET("/profiles/:id", h.GetProfile)
	sec.PUT("/profiles/:id", h.UpdateProfile)
	sec.DELETE("/profiles/:id", h.DeleteProfile)

	sec.POST("/users", h.CreateUser)
	sec.GET("/users", h.ListUsers)
	sec.GET("/users/:id", h.GetUser)
	sec.PUT("/users/:id", h.UpdateUser)
	sec.DELETE("/users/:id", h.DeleteUser)

	sec.POST("/users/:id/permission-sets", h.AssignPermissionSet)
	sec.DELETE("/users/:id/permission-sets/:psId", h.RevokePermissionSet)
	sec.GET("/users/:id/permission-sets", h.ListUserPermissionSets)

	sec.PUT("/permission-sets/:id/object-permissions", h.SetObjectPermission)
	sec.GET("/permission-sets/:id/object-permissions", h.ListObjectPermissions)
	sec.DELETE("/permission-sets/:id/object-permissions/:objectId", h.RemoveObjectPermission)

	sec.PUT("/permission-sets/:id/field-permissions", h.SetFieldPermission)
	sec.GET("/permission-sets/:id/field-permissions", h.ListFieldPermissions)
	sec.DELETE("/permission-sets/:id/field-permissions/:fieldId", h.RemoveFieldPermission)
}

// --- Roles ---

func (h *SecurityHandler) CreateRole(c *gin.Context) {
	var req security.CreateUserRoleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	role, err := h.roleService.Create(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": role})
}

func (h *SecurityHandler) ListRoles(c *gin.Context) {
	page, perPage := parsePagination(c)
	roles, total, err := h.roleService.List(c.Request.Context(), page, perPage)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       roles,
		"pagination": paginationMeta(page, perPage, total),
	})
}

func (h *SecurityHandler) GetRole(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	role, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": role})
}

func (h *SecurityHandler) UpdateRole(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req security.UpdateUserRoleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	role, err := h.roleService.Update(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": role})
}

func (h *SecurityHandler) DeleteRole(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.roleService.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Permission Sets ---

func (h *SecurityHandler) CreatePermissionSet(c *gin.Context) {
	var req security.CreatePermissionSetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	ps, err := h.psService.Create(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": ps})
}

func (h *SecurityHandler) ListPermissionSets(c *gin.Context) {
	page, perPage := parsePagination(c)
	sets, total, err := h.psService.List(c.Request.Context(), page, perPage)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       sets,
		"pagination": paginationMeta(page, perPage, total),
	})
}

func (h *SecurityHandler) GetPermissionSet(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	ps, err := h.psService.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ps})
}

func (h *SecurityHandler) UpdatePermissionSet(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req security.UpdatePermissionSetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	ps, err := h.psService.Update(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ps})
}

func (h *SecurityHandler) DeletePermissionSet(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.psService.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Profiles ---

func (h *SecurityHandler) CreateProfile(c *gin.Context) {
	var req security.CreateProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	profile, err := h.profileService.Create(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": profile})
}

func (h *SecurityHandler) ListProfiles(c *gin.Context) {
	page, perPage := parsePagination(c)
	profiles, total, err := h.profileService.List(c.Request.Context(), page, perPage)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       profiles,
		"pagination": paginationMeta(page, perPage, total),
	})
}

func (h *SecurityHandler) GetProfile(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	profile, err := h.profileService.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profile})
}

func (h *SecurityHandler) UpdateProfile(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req security.UpdateProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	profile, err := h.profileService.Update(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profile})
}

func (h *SecurityHandler) DeleteProfile(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.profileService.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Users ---

func (h *SecurityHandler) CreateUser(c *gin.Context) {
	var req security.CreateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	user, err := h.userService.Create(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *SecurityHandler) ListUsers(c *gin.Context) {
	page, perPage := parsePagination(c)
	users, total, err := h.userService.List(c.Request.Context(), page, perPage)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       users,
		"pagination": paginationMeta(page, perPage, total),
	})
}

func (h *SecurityHandler) GetUser(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *SecurityHandler) UpdateUser(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req security.UpdateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	user, err := h.userService.Update(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *SecurityHandler) DeleteUser(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- User Permission Sets ---

type assignPSRequest struct {
	PermissionSetID uuid.UUID `json:"permission_set_id"`
}

func (h *SecurityHandler) AssignPermissionSet(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req assignPSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	if err := h.userService.AssignPermissionSet(c.Request.Context(), id, req.PermissionSetID); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SecurityHandler) RevokePermissionSet(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	psID, err := parseUUID(c, "psId")
	if err != nil {
		return
	}
	if err := h.userService.RevokePermissionSet(c.Request.Context(), id, psID); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SecurityHandler) ListUserPermissionSets(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	assignments, err := h.userService.ListPermissionSets(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

// --- Object Permissions ---

func (h *SecurityHandler) SetObjectPermission(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req security.SetObjectPermissionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	op, err := h.permissionService.SetObjectPermission(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": op})
}

func (h *SecurityHandler) ListObjectPermissions(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	perms, err := h.permissionService.ListObjectPermissions(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

func (h *SecurityHandler) RemoveObjectPermission(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	objectID, err := parseUUID(c, "objectId")
	if err != nil {
		return
	}
	if err := h.permissionService.RemoveObjectPermission(c.Request.Context(), id, objectID); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Field Permissions ---

func (h *SecurityHandler) SetFieldPermission(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req security.SetFieldPermissionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid request body"))
		return
	}
	fp, err := h.permissionService.SetFieldPermission(c.Request.Context(), id, req)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": fp})
}

func (h *SecurityHandler) ListFieldPermissions(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	perms, err := h.permissionService.ListFieldPermissions(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

func (h *SecurityHandler) RemoveFieldPermission(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	fieldID, err := parseUUID(c, "fieldId")
	if err != nil {
		return
	}
	if err := h.permissionService.RemoveFieldPermission(c.Request.Context(), id, fieldID); err != nil {
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Helpers ---

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("invalid UUID: "+param))
		return uuid.Nil, err
	}
	return id, nil
}

func parsePagination(c *gin.Context) (int32, int32) {
	var page, perPage int32 = 1, 20
	if v := c.Query("page"); v != "" {
		if p := parseInt32(v); p > 0 {
			page = p
		}
	}
	if v := c.Query("per_page"); v != "" {
		if p := parseInt32(v); p > 0 {
			perPage = p
		}
	}
	return page, perPage
}

func parseInt32(s string) int32 {
	var n int32
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int32(c-'0')
	}
	return n
}

func paginationMeta(page, perPage int32, total int64) gin.H {
	totalPages := (total + int64(perPage) - 1) / int64(perPage)
	return gin.H{
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	}
}
