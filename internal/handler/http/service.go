package http

import (
	"github.com/davidsugianto/go-pkgs/response"
	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	"github.com/gin-gonic/gin"
)

// ListServices godoc
// @Summary List services
// @Description List all services with optional filters
// @Tags service
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Param visibility query string false "Filter by visibility"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name or description"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} serviceModel.ServiceListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services [get]
// @Security ApiKeyAuth
func (h *Handler) ListServices(c *gin.Context) {
	var req serviceModel.ListServicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.List(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateService godoc
// @Summary Register a service
// @Description Register a new service in the catalog
// @Tags service
// @Accept json
// @Produce json
// @Param body body serviceModel.CreateServiceRequest true "Service configuration"
// @Success 200 {object} serviceModel.ServiceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services [post]
// @Security ApiKeyAuth
func (h *Handler) CreateService(c *gin.Context) {
	var req serviceModel.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.Register(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetService godoc
// @Summary Get a service
// @Description Get detailed information about a specific service
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} serviceModel.ServiceResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetService(c *gin.Context) {
	id := c.Param("id")

	result, err := h.serviceUseCase.Get(c.Request.Context(), id)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateService godoc
// @Summary Update a service
// @Description Update an existing service
// @Tags service
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param body body serviceModel.UpdateServiceRequest true "Service updates"
// @Success 200 {object} serviceModel.ServiceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateService(c *gin.Context) {
	id := c.Param("id")

	var req serviceModel.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// DeleteService godoc
// @Summary Deregister a service
// @Description Deregister (soft delete) a service
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteService(c *gin.Context) {
	id := c.Param("id")

	if err := h.serviceUseCase.Deregister(c.Request.Context(), id); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "service deregistered"})
}

// DiscoverServices godoc
// @Summary Discover services
// @Description Search for services by name or description
// @Tags service
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} serviceModel.ServiceListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/discover [get]
// @Security ApiKeyAuth
func (h *Handler) DiscoverServices(c *gin.Context) {
	query := c.Query("q")
	endpointType := c.Query("type")

	if endpointType != "" {
		// Discover by endpoint type
		result, err := h.serviceUseCase.DiscoverByType(c.Request.Context(), endpointType)
		if err != nil {
			response.GinInternalServerError(c, err)
			return
		}
		response.GinSuccess(c, result)
		return
	}

	// Discover by search query
	result, err := h.serviceUseCase.Discover(c.Request.Context(), query)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// ListServiceVersions godoc
// @Summary List service versions
// @Description List all versions for a service
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} versionModel.ServiceVersionListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions [get]
// @Security ApiKeyAuth
func (h *Handler) ListServiceVersions(c *gin.Context) {
	serviceID := c.Param("id")

	var req versionModel.ListServiceVersionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.ListVersions(c.Request.Context(), serviceID, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateServiceVersion godoc
// @Summary Create a service version
// @Description Create a new version for a service
// @Tags service
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param body body versionModel.CreateServiceVersionRequest true "Version configuration"
// @Success 200 {object} versionModel.ServiceVersionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions [post]
// @Security ApiKeyAuth
func (h *Handler) CreateServiceVersion(c *gin.Context) {
	serviceID := c.Param("id")

	var req versionModel.CreateServiceVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.RegisterVersion(c.Request.Context(), serviceID, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetServiceVersion godoc
// @Summary Get a service version
// @Description Get detailed information about a specific version
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} versionModel.ServiceVersionResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions/{versionId} [get]
// @Security ApiKeyAuth
func (h *Handler) GetServiceVersion(c *gin.Context) {
	versionID := c.Param("versionId")

	result, err := h.serviceUseCase.GetVersion(c.Request.Context(), versionID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateServiceVersion godoc
// @Summary Update a service version
// @Description Update an existing version
// @Tags service
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param versionId path string true "Version ID"
// @Param body body versionModel.UpdateServiceVersionRequest true "Version updates"
// @Success 200 {object} versionModel.ServiceVersionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions/{versionId} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateServiceVersion(c *gin.Context) {
	versionID := c.Param("versionId")

	var req versionModel.UpdateServiceVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.UpdateVersion(c.Request.Context(), versionID, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// ListServiceEndpoints godoc
// @Summary List service endpoints
// @Description List all endpoints for a version
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} endpointModel.ServiceEndpointListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions/{versionId}/endpoints [get]
// @Security ApiKeyAuth
func (h *Handler) ListServiceEndpoints(c *gin.Context) {
	versionID := c.Param("versionId")

	result, err := h.serviceUseCase.ListEndpoints(c.Request.Context(), versionID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateServiceEndpoint godoc
// @Summary Add a service endpoint
// @Description Add a new endpoint for a version
// @Tags service
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param versionId path string true "Version ID"
// @Param body body endpointModel.CreateServiceEndpointRequest true "Endpoint configuration"
// @Success 200 {object} endpointModel.ServiceEndpointResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions/{versionId}/endpoints [post]
// @Security ApiKeyAuth
func (h *Handler) CreateServiceEndpoint(c *gin.Context) {
	versionID := c.Param("versionId")

	var req endpointModel.CreateServiceEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.AddEndpoint(c.Request.Context(), versionID, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetServiceEndpoint godoc
// @Summary Get a service endpoint
// @Description Get detailed information about a specific endpoint
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Param versionId path string true "Version ID"
// @Param endpointId path string true "Endpoint ID"
// @Success 200 {object} endpointModel.ServiceEndpointResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions/{versionId}/endpoints/{endpointId} [get]
// @Security ApiKeyAuth
func (h *Handler) GetServiceEndpoint(c *gin.Context) {
	endpointID := c.Param("endpointId")

	result, err := h.serviceUseCase.GetEndpoint(c.Request.Context(), endpointID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateServiceEndpoint godoc
// @Summary Update a service endpoint
// @Description Update an existing endpoint
// @Tags service
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param versionId path string true "Version ID"
// @Param endpointId path string true "Endpoint ID"
// @Param body body endpointModel.UpdateServiceEndpointRequest true "Endpoint updates"
// @Success 200 {object} endpointModel.ServiceEndpointResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions/{versionId}/endpoints/{endpointId} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateServiceEndpoint(c *gin.Context) {
	endpointID := c.Param("endpointId")

	var req endpointModel.UpdateServiceEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.serviceUseCase.UpdateEndpoint(c.Request.Context(), endpointID, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// DeleteServiceEndpoint godoc
// @Summary Delete a service endpoint
// @Description Delete an endpoint
// @Tags service
// @Produce json
// @Param id path string true "Service ID"
// @Param versionId path string true "Version ID"
// @Param endpointId path string true "Endpoint ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/services/{id}/versions/{versionId}/endpoints/{endpointId} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteServiceEndpoint(c *gin.Context) {
	endpointID := c.Param("endpointId")

	if err := h.serviceUseCase.RemoveEndpoint(c.Request.Context(), endpointID); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "endpoint removed"})
}
