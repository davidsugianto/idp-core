package http

import (
	"errors"
	"net/http"

	"github.com/davidsugianto/go-pkgs/response"
	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	templateUsecase "github.com/davidsugianto/idp-core/internal/usecase/template"
	"github.com/gin-gonic/gin"
)

// ListTemplates godoc
// @Summary List templates
// @Description List all templates with optional filters
// @Tags template
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Param category query string false "Filter by category"
// @Param visibility query string false "Filter by visibility"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name or description"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} templateModel.TemplateListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates [get]
// @Security ApiKeyAuth
func (h *Handler) ListTemplates(c *gin.Context) {
	var req templateModel.ListTemplatesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.List(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateTemplate godoc
// @Summary Create a template
// @Description Create a new infrastructure template
// @Tags template
// @Accept json
// @Produce json
// @Param body body templateModel.CreateTemplateRequest true "Template configuration"
// @Success 201 {object} templateModel.TemplateResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates [post]
// @Security ApiKeyAuth
func (h *Handler) CreateTemplate(c *gin.Context) {
	var req templateModel.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateExists) {
			response.GinError(c, http.StatusConflict, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrTemplateNameRequired) || errors.Is(err, templateUsecase.ErrInvalidVisibility) {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, result)
}

// GetTemplate godoc
// @Summary Get a template
// @Description Get detailed information about a specific template
// @Tags template
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} templateModel.TemplateResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	result, err := h.templateUseCase.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateTemplate godoc
// @Summary Update a template
// @Description Update an existing template
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param body body templateModel.UpdateTemplateRequest true "Template updates"
// @Success 200 {object} templateModel.TemplateResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var req templateModel.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateNotFound) {
			response.GinNotFound(c, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrInvalidVisibility) || errors.Is(err, templateUsecase.ErrInvalidStatus) {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// DeleteTemplate godoc
// @Summary Delete a template
// @Description Soft delete a template
// @Tags template
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := h.templateUseCase.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "template deleted"})
}

// ListTemplateVersions godoc
// @Summary List template versions
// @Description List all versions for a template
// @Tags template
// @Produce json
// @Param id path string true "Template ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} templateModel.TemplateVersionListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id}/versions [get]
// @Security ApiKeyAuth
func (h *Handler) ListTemplateVersions(c *gin.Context) {
	templateID := c.Param("id")

	var req templateModel.ListTemplateVersionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.ListVersions(c.Request.Context(), templateID, &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateTemplateVersion godoc
// @Summary Create a template version
// @Description Create a new version for a template
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param body body templateModel.CreateTemplateVersionRequest true "Template version configuration"
// @Success 201 {object} templateModel.TemplateVersionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id}/versions [post]
// @Security ApiKeyAuth
func (h *Handler) CreateTemplateVersion(c *gin.Context) {
	templateID := c.Param("id")

	var req templateModel.CreateTemplateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.CreateVersion(c.Request.Context(), templateID, &req)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateNotFound) {
			response.GinNotFound(c, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrTemplateVersionExists) {
			response.GinError(c, http.StatusConflict, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrTemplateVersionRequired) || errors.Is(err, templateUsecase.ErrInvalidStatus) {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, result)
}

// GetTemplateVersion godoc
// @Summary Get a template version
// @Description Get detailed information about a specific template version
// @Tags template
// @Produce json
// @Param id path string true "Template ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} templateModel.TemplateVersionResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id}/versions/{versionId} [get]
// @Security ApiKeyAuth
func (h *Handler) GetTemplateVersion(c *gin.Context) {
	versionID := c.Param("versionId")

	result, err := h.templateUseCase.GetVersion(c.Request.Context(), versionID)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateVersionNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateTemplateVersion godoc
// @Summary Update a template version
// @Description Update an existing template version
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param versionId path string true "Version ID"
// @Param body body templateModel.UpdateTemplateVersionRequest true "Template version updates"
// @Success 200 {object} templateModel.TemplateVersionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id}/versions/{versionId} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateTemplateVersion(c *gin.Context) {
	versionID := c.Param("versionId")

	var req templateModel.UpdateTemplateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.UpdateVersion(c.Request.Context(), versionID, &req)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateVersionNotFound) {
			response.GinNotFound(c, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrInvalidStatus) {
			response.GinBadRequest(c, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrInvalidLifecycleTransition) {
			response.GinError(c, http.StatusConflict, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// ReplaceTemplateParameters godoc
// @Summary Replace template parameter definitions
// @Description Replace the ordered parameter definition set for a template version
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param versionId path string true "Version ID"
// @Param body body templateModel.ReplaceTemplateParametersRequest true "Template parameter replacement payload"
// @Success 200 {array} templateModel.TemplateParameter
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id}/versions/{versionId}/parameters [put]
// @Security ApiKeyAuth
func (h *Handler) ReplaceTemplateParameters(c *gin.Context) {
	templateID := c.Param("id")
	versionID := c.Param("versionId")

	var req templateModel.ReplaceTemplateParametersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.ReplaceParameters(c.Request.Context(), templateID, versionID, &req)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateNotFound) || errors.Is(err, templateUsecase.ErrTemplateVersionNotFound) {
			response.GinNotFound(c, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrTemplateParameterRequired) || errors.Is(err, templateUsecase.ErrTemplateParameterDuplicate) {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// ReplaceTemplateResources godoc
// @Summary Replace template resource definitions
// @Description Replace the ordered resource definition set for a template version
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param versionId path string true "Version ID"
// @Param body body templateModel.ReplaceTemplateResourcesRequest true "Template resource replacement payload"
// @Success 200 {array} templateModel.TemplateResource
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id}/versions/{versionId}/resources [put]
// @Security ApiKeyAuth
func (h *Handler) ReplaceTemplateResources(c *gin.Context) {
	templateID := c.Param("id")
	versionID := c.Param("versionId")

	var req templateModel.ReplaceTemplateResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.ReplaceResources(c.Request.Context(), templateID, versionID, &req)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateNotFound) || errors.Is(err, templateUsecase.ErrTemplateVersionNotFound) {
			response.GinNotFound(c, err)
			return
		}
		if errors.Is(err, templateUsecase.ErrTemplateResourceRequired) {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// ValidateTemplateVersion godoc
// @Summary Validate template version inputs
// @Description Validate prospective template inputs before environment creation
// @Tags template
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param versionId path string true "Version ID"
// @Param body body templateModel.ValidateTemplateVersionRequest true "Template input validation payload"
// @Success 200 {object} templateModel.ValidateTemplateVersionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/templates/{id}/versions/{versionId}/validate [post]
// @Security ApiKeyAuth
func (h *Handler) ValidateTemplateVersion(c *gin.Context) {
	templateID := c.Param("id")
	versionID := c.Param("versionId")

	var req templateModel.ValidateTemplateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.templateUseCase.ValidateVersionInputs(c.Request.Context(), templateID, versionID, &req)
	if err != nil {
		if errors.Is(err, templateUsecase.ErrTemplateNotFound) || errors.Is(err, templateUsecase.ErrTemplateVersionNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}
