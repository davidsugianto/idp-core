package http

import (
	"strconv"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	roleUsecase "github.com/davidsugianto/idp-core/internal/usecase/role"
	"github.com/davidsugianto/go-pkgs/response"
	"github.com/gin-gonic/gin"
)

// ListRoles godoc
// @Summary List roles
// @Description Get a paginated list of roles
// @Tags roles
// @Produce json
// @Param limit query int false "Page limit" default(20)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} roleModel.RoleListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/roles [get]
// @Security ApiKeyAuth
func (h *Handler) ListRoles(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.roleUseCase.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetRole godoc
// @Summary Get role by ID
// @Description Get a specific role by ID
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} roleModel.RoleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/roles/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.roleUseCase.Get(c.Request.Context(), id)
	if err != nil {
		if err == roleUsecase.ErrRoleNotFound {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, roleModel.ToRoleResponse(result))
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with permissions
// @Tags roles
// @Accept json
// @Produce json
// @Param role body roleModel.CreateRoleRequest true "Role creation request"
// @Success 201 {object} roleModel.RoleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/roles [post]
// @Security ApiKeyAuth
func (h *Handler) CreateRole(c *gin.Context) {
	var req roleModel.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.roleUseCase.Create(c.Request.Context(), req)
	if err != nil {
		if err == roleUsecase.ErrRoleAlreadyExists {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, roleModel.ToRoleResponse(result))
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update role details and permissions
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body roleModel.UpdateRoleRequest true "Role update request"
// @Success 200 {object} roleModel.RoleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/roles/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, nil)
		return
	}

	var req roleModel.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.roleUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == roleUsecase.ErrRoleNotFound {
			response.GinNotFound(c, err)
			return
		}
		if err == roleUsecase.ErrRoleAlreadyExists {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, roleModel.ToRoleResponse(result))
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Soft delete a role
// @Tags roles
// @Param id path string true "Role ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/roles/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, nil)
		return
	}

	err := h.roleUseCase.Delete(c.Request.Context(), id)
	if err != nil {
		if err == roleUsecase.ErrRoleNotFound {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	c.Status(204)
}

// AssignRole godoc
// @Summary Assign a role to a user
// @Description Assign a role to a user with optional team context
// @Tags roles
// @Accept json
// @Produce json
// @Param assignment body roleModel.AssignRoleRequest true "Role assignment request"
// @Success 201 {object} roleModel.UserRoleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/roles/assign [post]
// @Security ApiKeyAuth
func (h *Handler) AssignRole(c *gin.Context) {
	var req roleModel.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	// Get the current user who is granting the role
	grantedBy := c.GetString("user_id")

	result, err := h.roleUseCase.AssignRole(c.Request.Context(), req, grantedBy)
	if err != nil {
		if err == roleUsecase.ErrRoleNotFound {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, roleModel.ToUserRoleResponse(result))
}

// RevokeRole godoc
// @Summary Revoke a role from a user
// @Description Revoke a role from a user
// @Tags roles
// @Accept json
// @Produce json
// @Param revocation body roleModel.RevokeRoleRequest true "Role revocation request"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/roles/revoke [post]
// @Security ApiKeyAuth
func (h *Handler) RevokeRole(c *gin.Context) {
	var req roleModel.RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	err := h.roleUseCase.RevokeRole(c.Request.Context(), req)
	if err != nil {
		if err == roleUsecase.ErrUserRoleNotFound {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	c.Status(204)
}

// GetUserRoles godoc
// @Summary Get user's roles
// @Description Get all roles assigned to a user
// @Tags roles
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} roleModel.UserRoleListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/users/{id}/roles [get]
// @Security ApiKeyAuth
func (h *Handler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.roleUseCase.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetUserTeamRoles godoc
// @Summary Get user's roles in a team
// @Description Get all roles assigned to a user for a specific team
// @Tags roles
// @Produce json
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Success 200 {object} roleModel.UserRoleListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/teams/{id}/members/{userId}/roles [get]
// @Security ApiKeyAuth
func (h *Handler) GetUserTeamRoles(c *gin.Context) {
	teamID := c.Param("id")
	userID := c.Param("userId")

	if teamID == "" || userID == "" {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.roleUseCase.GetUserRolesByTeam(c.Request.Context(), userID, teamID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}
