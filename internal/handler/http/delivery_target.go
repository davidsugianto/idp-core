package http

import (
	"errors"
	"net/http"

	"github.com/davidsugianto/go-pkgs/response"
	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	deliveryTargetUsecase "github.com/davidsugianto/idp-core/internal/usecase/delivery_target"
	"github.com/gin-gonic/gin"
)

// ListDeliveryTargets godoc
// @Summary List delivery targets
// @Description List delivery targets visible to the caller with optional placement filters
// @Tags delivery-targets
// @Produce json
// @Param team_id query string false "Filter by team scope"
// @Param purpose query string false "Filter by target purpose"
// @Param availability_state query string false "Filter by availability state"
// @Param health_state query string false "Filter by health state"
// @Param search query string false "Search by name, display name, or cluster"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} deliveryTargetModel.DeliveryTargetListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/delivery-targets [get]
// @Security ApiKeyAuth
func (h *Handler) ListDeliveryTargets(c *gin.Context) {
	var req deliveryTargetModel.ListDeliveryTargetsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.deliveryTargetUseCase.List(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateDeliveryTarget godoc
// @Summary Create a delivery target
// @Description Register a placement target with cluster metadata, availability, and capacity details
// @Tags delivery-targets
// @Accept json
// @Produce json
// @Param body body deliveryTargetModel.CreateDeliveryTargetRequest true "Delivery target payload"
// @Success 201 {object} deliveryTargetModel.DeliveryTargetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/delivery-targets [post]
// @Security ApiKeyAuth
func (h *Handler) CreateDeliveryTarget(c *gin.Context) {
	var req deliveryTargetModel.CreateDeliveryTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.deliveryTargetUseCase.Create(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetExists):
			response.GinError(c, http.StatusConflict, err)
		case errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetNameRequired),
			errors.Is(err, deliveryTargetUsecase.ErrClusterNameRequired),
			errors.Is(err, deliveryTargetUsecase.ErrInvalidAvailabilityState),
			errors.Is(err, deliveryTargetUsecase.ErrInvalidHealthState),
			errors.Is(err, deliveryTargetUsecase.ErrIncompleteControlPlaneMetadata):
			response.GinBadRequest(c, err)
		default:
			response.GinInternalServerError(c, err)
		}
		return
	}

	response.GinCreated(c, result)
}

// GetDeliveryTarget godoc
// @Summary Get a delivery target
// @Description Get the details of a specific delivery target
// @Tags delivery-targets
// @Produce json
// @Param id path string true "Delivery target ID"
// @Success 200 {object} deliveryTargetModel.DeliveryTargetResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/delivery-targets/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetDeliveryTarget(c *gin.Context) {
	result, err := h.deliveryTargetUseCase.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateDeliveryTarget godoc
// @Summary Update a delivery target
// @Description Update cluster metadata, capacity, or availability of a delivery target
// @Tags delivery-targets
// @Accept json
// @Produce json
// @Param id path string true "Delivery target ID"
// @Param body body deliveryTargetModel.UpdateDeliveryTargetRequest true "Delivery target update payload"
// @Success 200 {object} deliveryTargetModel.DeliveryTargetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/delivery-targets/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateDeliveryTarget(c *gin.Context) {
	var req deliveryTargetModel.UpdateDeliveryTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.deliveryTargetUseCase.Update(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		switch {
		case errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetNotFound):
			response.GinNotFound(c, err)
		case errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetExists):
			response.GinError(c, http.StatusConflict, err)
		case errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetNameRequired),
			errors.Is(err, deliveryTargetUsecase.ErrClusterNameRequired),
			errors.Is(err, deliveryTargetUsecase.ErrInvalidAvailabilityState),
			errors.Is(err, deliveryTargetUsecase.ErrInvalidHealthState),
			errors.Is(err, deliveryTargetUsecase.ErrIncompleteControlPlaneMetadata):
			response.GinBadRequest(c, err)
		default:
			response.GinInternalServerError(c, err)
		}
		return
	}

	response.GinSuccess(c, result)
}

// DeleteDeliveryTarget godoc
// @Summary Delete a delivery target
// @Description Delete a delivery target when it is no longer referenced by environments or active movements
// @Tags delivery-targets
// @Produce json
// @Param id path string true "Delivery target ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/delivery-targets/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteDeliveryTarget(c *gin.Context) {
	if err := h.deliveryTargetUseCase.Delete(c.Request.Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetNotFound):
			response.GinNotFound(c, err)
		case errors.Is(err, deliveryTargetUsecase.ErrDeliveryTargetInUse):
			response.GinError(c, http.StatusConflict, err)
		default:
			response.GinInternalServerError(c, err)
		}
		return
	}

	response.GinSuccess(c, gin.H{"message": "delivery target deleted"})
}
