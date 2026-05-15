package http

import (
	"github.com/davidsugianto/go-pkgs/response"
	rightsizingModel "github.com/davidsugianto/idp-core/internal/model/rightsizing"
	"github.com/gin-gonic/gin"
)

// ListRightsizingRecommendations godoc
// @Summary List rightsizing recommendations
// @Description List all rightsizing recommendations with optional filters
// @Tags rightsizing
// @Produce json
// @Param namespace query string false "Filter by namespace"
// @Param status query string false "Filter by status (pending, applied, dismissed, failed)"
// @Param team_id query string false "Filter by team ID"
// @Param workload_type query string false "Filter by workload type (Deployment, StatefulSet)"
// @Param recommendation_type query string false "Filter by recommendation type (scale_down, scale_up, optimal)"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} rightsizingModel.RecommendationListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/rightsizing/recommendations [get]
// @Security ApiKeyAuth
func (h *Handler) ListRightsizingRecommendations(c *gin.Context) {
	var req rightsizingModel.ListRecommendationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.rightsizingUseCase.ListRecommendations(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetRightsizingRecommendation godoc
// @Summary Get a rightsizing recommendation
// @Description Get detailed information about a specific recommendation
// @Tags rightsizing
// @Produce json
// @Param id path string true "Recommendation ID"
// @Success 200 {object} rightsizingModel.RecommendationResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/rightsizing/recommendations/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetRightsizingRecommendation(c *gin.Context) {
	id := c.Param("id")

	result, err := h.rightsizingUseCase.GetRecommendation(c.Request.Context(), id)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// ApplyRightsizingRecommendation godoc
// @Summary Apply a rightsizing recommendation
// @Description Apply the recommended resource changes to the Kubernetes workload
// @Tags rightsizing
// @Produce json
// @Param id path string true "Recommendation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/rightsizing/recommendations/{id}/apply [post]
// @Security ApiKeyAuth
func (h *Handler) ApplyRightsizingRecommendation(c *gin.Context) {
	id := c.Param("id")

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.GinBadRequest(c, nil)
		return
	}

	if err := h.rightsizingUseCase.ApplyRecommendation(c.Request.Context(), id, userID.(string)); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "recommendation applied successfully"})
}

// RollbackRightsizingRecommendation godoc
// @Summary Rollback a rightsizing recommendation
// @Description Restore the workload to its previous resource configuration
// @Tags rightsizing
// @Produce json
// @Param id path string true "Recommendation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/rightsizing/recommendations/{id}/rollback [post]
// @Security ApiKeyAuth
func (h *Handler) RollbackRightsizingRecommendation(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		response.GinBadRequest(c, nil)
		return
	}

	if err := h.rightsizingUseCase.RollbackRecommendation(c.Request.Context(), id, userID.(string)); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "recommendation rolled back successfully"})
}

// DismissRightsizingRecommendation godoc
// @Summary Dismiss a rightsizing recommendation
// @Description Dismiss a pending recommendation that should not be applied
// @Tags rightsizing
// @Accept json
// @Produce json
// @Param id path string true "Recommendation ID"
// @Param body body rightsizingModel.DismissRecommendationRequest false "Dismissal reason"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/rightsizing/recommendations/{id}/dismiss [post]
// @Security ApiKeyAuth
func (h *Handler) DismissRightsizingRecommendation(c *gin.Context) {
	id := c.Param("id")

	var req rightsizingModel.DismissRecommendationRequest
	_ = c.ShouldBindJSON(&req) // Reason is optional

	if err := h.rightsizingUseCase.DismissRecommendation(c.Request.Context(), id, req.Reason); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "recommendation dismissed"})
}
