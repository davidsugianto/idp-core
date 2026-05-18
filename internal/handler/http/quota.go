package http

import (
	"github.com/davidsugianto/go-pkgs/response"
	quotaModel "github.com/davidsugianto/idp-core/internal/model/resourcequota"
	"github.com/gin-gonic/gin"
)

// ListResourceQuotas godoc
// @Summary List resource quotas
// @Description List all resource quotas with optional filters
// @Tags quota
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Param environment_id query string false "Filter by environment ID"
// @Param namespace query string false "Filter by namespace"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} quotaModel.ResourceQuotaListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas [get]
// @Security ApiKeyAuth
func (h *Handler) ListResourceQuotas(c *gin.Context) {
	var req quotaModel.ListResourceQuotasRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.quotaUseCase.ListQuotas(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetResourceQuota godoc
// @Summary Get a resource quota
// @Description Get detailed information about a specific quota
// @Tags quota
// @Produce json
// @Param id path string true "Quota ID"
// @Success 200 {object} quotaModel.ResourceQuotaResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetResourceQuota(c *gin.Context) {
	id := c.Param("id")

	result, err := h.quotaUseCase.GetQuota(c.Request.Context(), id)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetResourceQuotaByNamespace godoc
// @Summary Get quota by namespace
// @Description Get resource quota for a specific namespace
// @Tags quota
// @Produce json
// @Param namespace path string true "Namespace"
// @Success 200 {object} quotaModel.ResourceQuotaResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas/namespace/{namespace} [get]
// @Security ApiKeyAuth
func (h *Handler) GetResourceQuotaByNamespace(c *gin.Context) {
	namespace := c.Param("namespace")

	result, err := h.quotaUseCase.GetQuotaByNamespace(c.Request.Context(), namespace)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateResourceQuota godoc
// @Summary Create a resource quota
// @Description Create a new resource quota for a namespace
// @Tags quota
// @Accept json
// @Produce json
// @Param body body quotaModel.CreateResourceQuotaRequest true "Quota configuration"
// @Success 200 {object} quotaModel.ResourceQuotaResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas [post]
// @Security ApiKeyAuth
func (h *Handler) CreateResourceQuota(c *gin.Context) {
	var req quotaModel.CreateResourceQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.quotaUseCase.CreateQuota(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateResourceQuota godoc
// @Summary Update a resource quota
// @Description Update an existing resource quota
// @Tags quota
// @Accept json
// @Produce json
// @Param id path string true "Quota ID"
// @Param body body quotaModel.UpdateResourceQuotaRequest true "Quota updates"
// @Success 200 {object} quotaModel.ResourceQuotaResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateResourceQuota(c *gin.Context) {
	id := c.Param("id")

	var req quotaModel.UpdateResourceQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.quotaUseCase.UpdateQuota(c.Request.Context(), id, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// DeleteResourceQuota godoc
// @Summary Delete a resource quota
// @Description Delete a resource quota
// @Tags quota
// @Produce json
// @Param id path string true "Quota ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteResourceQuota(c *gin.Context) {
	id := c.Param("id")

	if err := h.quotaUseCase.DeleteQuota(c.Request.Context(), id); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "quota deleted"})
}

// GetNamespaceUsage godoc
// @Summary Get namespace resource usage
// @Description Get current resource usage for a namespace
// @Tags quota
// @Produce json
// @Param namespace path string true "Namespace"
// @Success 200 {object} quotaModel.UsageResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas/namespace/{namespace}/usage [get]
// @Security ApiKeyAuth
func (h *Handler) GetNamespaceUsage(c *gin.Context) {
	namespace := c.Param("namespace")

	result, err := h.quotaUseCase.GetUsage(c.Request.Context(), namespace)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// RefreshNamespaceUsage godoc
// @Summary Refresh namespace usage
// @Description Refresh cached resource usage for a namespace
// @Tags quota
// @Produce json
// @Param namespace path string true "Namespace"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas/namespace/{namespace}/usage/refresh [post]
// @Security ApiKeyAuth
func (h *Handler) RefreshNamespaceUsage(c *gin.Context) {
	namespace := c.Param("namespace")

	if err := h.quotaUseCase.RefreshUsage(c.Request.Context(), namespace); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "usage refreshed"})
}

// CheckQuota godoc
// @Summary Check quota
// @Description Check if a resource request would exceed quota limits
// @Tags quota
// @Accept json
// @Produce json
// @Param body body quotaModel.QuotaCheckRequest true "Resource check request"
// @Success 200 {object} quotaModel.QuotaCheckResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/quotas/check [post]
// @Security ApiKeyAuth
func (h *Handler) CheckQuota(c *gin.Context) {
	var req quotaModel.QuotaCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.quotaUseCase.CheckQuota(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}
