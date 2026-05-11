package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/model/user"
	userUsecase "github.com/davidsugianto/idp-core/internal/usecase/user"
	"github.com/gin-gonic/gin"
)

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body user.CreateUserRequest true "User request"
// @Success 201 {object} user.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	u, err := h.userUseCase.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, userUsecase.ErrUserAlreadyExists) {
			response.GinError(c, http.StatusConflict, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, user.ToUserResponse(u))
}

// ListUsers godoc
// @Summary List users
// @Description List all users with pagination
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} user.UserListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.userUseCase.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetUser godoc
// @Summary Get user details
// @Description Get details of a specific user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} user.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, errors.New("missing user id"))
		return
	}

	u, err := h.userUseCase.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, userUsecase.ErrUserNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, user.ToUserResponse(u))
}

// UpdateUser godoc
// @Summary Update user
// @Description Update a user's details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body user.UpdateUserRequest true "User update request"
// @Success 200 {object} user.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/users/{id} [patch]
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, errors.New("missing user id"))
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	u, err := h.userUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, userUsecase.ErrUserNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, user.ToUserResponse(u))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete a user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, errors.New("missing user id"))
		return
	}

	if err := h.userUseCase.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, userUsecase.ErrUserNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "user deleted"})
}

// UpdateUserStatus godoc
// @Summary Update user status
// @Description Update a user's status (active/disabled)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param status query string true "Status (active/disabled)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/users/{id}/status [put]
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")

	if id == "" {
		response.GinBadRequest(c, errors.New("missing user id"))
		return
	}

	if status == "" {
		response.GinBadRequest(c, errors.New("missing status parameter"))
		return
	}

	if status != "active" && status != "disabled" {
		response.GinBadRequest(c, errors.New("invalid status. must be 'active' or 'disabled'"))
		return
	}

	if err := h.userUseCase.UpdateStatus(c.Request.Context(), id, status); err != nil {
		if errors.Is(err, userUsecase.ErrUserNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "user status updated"})
}
