package http

import (
	"errors"
	"net/http"

	"github.com/davidsugianto/go-pkgs/response"
	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	deliveryTargetUsecase "github.com/davidsugianto/idp-core/internal/usecase/delivery_target"
	"github.com/gin-gonic/gin"
)

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
			errors.Is(err, deliveryTargetUsecase.ErrInvalidHealthState):
			response.GinBadRequest(c, err)
		default:
			response.GinInternalServerError(c, err)
		}
		return
	}

	response.GinCreated(c, result)
}

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
			errors.Is(err, deliveryTargetUsecase.ErrInvalidHealthState):
			response.GinBadRequest(c, err)
		default:
			response.GinInternalServerError(c, err)
		}
		return
	}

	response.GinSuccess(c, result)
}

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
