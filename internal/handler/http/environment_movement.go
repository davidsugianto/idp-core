package http

import (
	"errors"
	"net/http"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	environmentMovementUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment_movement"
	"github.com/gin-gonic/gin"
)

type environmentMovementResponse = environmentMovementModel.EnvironmentMovement

type createEnvironmentMovementRequest struct {
	DestinationTargetID string `json:"destination_target_id" binding:"required"`
}

type updateEnvironmentMovementStatusRequest struct {
	Status          string `json:"status" binding:"required"`
	ProgressPercent int    `json:"progress_percent"`
	Message         string `json:"message"`
}

// CreateEnvironmentMovement godoc
// @Summary Create an environment movement request
// @Description Request a move of an environment to another delivery target
// @Tags environment-movements
// @Accept json
// @Produce json
// @Param id path string true "Environment ID"
// @Param body body createEnvironmentMovementRequest true "Environment movement request payload"
// @Success 201 {object} environmentMovementModel.EnvironmentMovement
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/environments/{id}/movements [post]
// @Security ApiKeyAuth
func (h *Handler) CreateEnvironmentMovement(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	userID := middleware.GetUserID(c)
	if teamID == "" || userID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req createEnvironmentMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.environmentMovementUseCase.Create(c.Request.Context(), teamID, userID, c.Param("id"), req.DestinationTargetID)
	if err != nil {
		switch {
		case errors.Is(err, environmentMovementUsecase.ErrEnvironmentNotFound),
			errors.Is(err, environmentMovementUsecase.ErrDeliveryTargetNotFound),
			errors.Is(err, environmentMovementUsecase.ErrMovementNotFound):
			response.GinNotFound(c, err)
		case errors.Is(err, environmentMovementUsecase.ErrInvalidDestinationTarget),
			errors.Is(err, environmentMovementUsecase.ErrMovementTargetConflict),
			errors.Is(err, environmentMovementUsecase.ErrInvalidMovementStatus):
			response.GinError(c, http.StatusConflict, err)
		default:
			response.GinInternalServerError(c, err)
		}
		return
	}

	response.GinCreated(c, result)
}

// ListEnvironmentMovements godoc
// @Summary List environment movements
// @Description List the historical movement requests for an environment
// @Tags environment-movements
// @Produce json
// @Param id path string true "Environment ID"
// @Success 200 {array} environmentMovementModel.EnvironmentMovement
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/environments/{id}/movements [get]
// @Security ApiKeyAuth
func (h *Handler) ListEnvironmentMovements(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	result, err := h.environmentMovementUseCase.ListByEnvironment(c.Request.Context(), teamID, c.Param("id"))
	if err != nil {
		if errors.Is(err, environmentMovementUsecase.ErrEnvironmentNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetEnvironmentMovement godoc
// @Summary Get an environment movement
// @Description Get the latest state for a specific environment movement request
// @Tags environment-movements
// @Produce json
// @Param id path string true "Environment ID"
// @Param movementId path string true "Movement ID"
// @Success 200 {object} environmentMovementModel.EnvironmentMovement
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/environments/{id}/movements/{movementId} [get]
// @Security ApiKeyAuth
func (h *Handler) GetEnvironmentMovement(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	result, err := h.environmentMovementUseCase.Get(c.Request.Context(), teamID, c.Param("id"), c.Param("movementId"))
	if err != nil {
		if errors.Is(err, environmentMovementUsecase.ErrEnvironmentNotFound) || errors.Is(err, environmentMovementUsecase.ErrMovementNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

func (h *Handler) UpdateEnvironmentMovementStatus(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req updateEnvironmentMovementStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	if req.ProgressPercent < 0 || req.ProgressPercent > 100 {
		response.GinBadRequest(c, errors.New("progress_percent must be between 0 and 100"))
		return
	}

	result, err := h.environmentMovementUseCase.UpdateStatus(c.Request.Context(), teamID, c.Param("id"), c.Param("movementId"), req.Status, req.ProgressPercent, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, environmentMovementUsecase.ErrEnvironmentNotFound),
			errors.Is(err, environmentMovementUsecase.ErrMovementNotFound):
			response.GinNotFound(c, err)
		case errors.Is(err, environmentMovementUsecase.ErrInvalidMovementStatus),
			errors.Is(err, environmentMovementUsecase.ErrInvalidDestinationTarget),
			errors.Is(err, environmentMovementUsecase.ErrMovementTargetConflict):
			response.GinBadRequest(c, err)
		default:
			response.GinInternalServerError(c, err)
		}
		return
	}

	response.GinSuccess(c, result)
}
