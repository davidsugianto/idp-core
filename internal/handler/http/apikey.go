package http

import (
	apikeyModel "github.com/davidsugianto/idp-core/internal/model/apikey"
	apikeyUsecase "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	"github.com/davidsugianto/go-pkgs/response"
	"github.com/gin-gonic/gin"
)

// ListAPIKeys godoc
// @Summary List API keys
// @Description List all API keys, optionally filtered by team
// @Tags api-keys
// @Produce json
// @Param team_id query string false "Team ID"
// @Success 200 {array} apikeyModel.APIKeyResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/api-keys [get]
// @Security ApiKeyAuth
func (h *Handler) ListAPIKeys(c *gin.Context) {
	teamID := c.Query("team_id")

	result, err := h.apiKeyUseCase.List(c.Request.Context(), teamID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetAPIKey godoc
// @Summary Get API key by ID
// @Description Get a specific API key by ID
// @Tags api-keys
// @Produce json
// @Param id path string true "API Key ID"
// @Success 200 {object} apikeyModel.APIKeyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/api-keys/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.apiKeyUseCase.Get(c.Request.Context(), id)
	if err != nil {
		if err == apikeyUsecase.ErrAPIKeyNotFound {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateAPIKey godoc
// @Summary Create a new API key
// @Description Create a new API key. The full key value is only returned once in this response.
// @Tags api-keys
// @Accept json
// @Produce json
// @Param api_key body apikeyModel.CreateAPIKeyRequest true "API key creation request"
// @Success 201 {object} apikeyModel.APIKeyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/api-keys [post]
// @Security ApiKeyAuth
func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req apikeyModel.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	createdBy := ""
	if str, ok := userID.(string); ok {
		createdBy = str
	}

	result, err := h.apiKeyUseCase.Create(c.Request.Context(), createdBy, req)
	if err != nil {
		if err == apikeyUsecase.ErrKeyNameRequired {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, result)
}

// UpdateAPIKey godoc
// @Summary Update an API key
// @Description Update an existing API key's metadata
// @Tags api-keys
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Param api_key body apikeyModel.CreateAPIKeyRequest true "API key update request"
// @Success 200 {object} apikeyModel.APIKeyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/api-keys/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, nil)
		return
	}

	var req apikeyModel.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.apiKeyUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == apikeyUsecase.ErrAPIKeyNotFound {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// DeleteAPIKey godoc
// @Summary Delete an API key
// @Description Delete (revoke) an API key
// @Tags api-keys
// @Produce json
// @Param id path string true "API Key ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/api-keys/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, nil)
		return
	}

	err := h.apiKeyUseCase.Delete(c.Request.Context(), id)
	if err != nil {
		if err == apikeyUsecase.ErrAPIKeyNotFound {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "api key deleted"})
}