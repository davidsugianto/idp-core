package http

import (
	"strconv"
	"time"

	"github.com/davidsugianto/go-pkgs/response"
	costModel "github.com/davidsugianto/idp-core/internal/model/cost"
	"github.com/gin-gonic/gin"
)

// ListCosts godoc
// @Summary List cost records
// @Description List cost records with filtering and pagination
// @Tags costs
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Param environment_id query string false "Filter by environment ID"
// @Param namespace query string false "Filter by namespace"
// @Param start_date query string false "Filter by start date (RFC3339)"
// @Param end_date query string false "Filter by end date (RFC3339)"
// @Param limit query int false "Page limit" default(50)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} costModel.CostListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/costs [get]
// @Security ApiKeyAuth
func (h *Handler) ListCosts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := costModel.CostFilter{
		TeamID:        c.Query("team_id"),
		EnvironmentID: c.Query("environment_id"),
		Namespace:     c.Query("namespace"),
		Limit:         limit,
		Offset:        offset,
	}

	if startStr := c.Query("start_date"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			filter.StartDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			filter.EndDate = &t
		}
	}

	result, err := h.costUseCase.List(c.Request.Context(), filter)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetTeamCostsHandler godoc
// @Summary Get team cost records
// @Description Get cost records for a specific team within a time range
// @Tags costs
// @Produce json
// @Param team_id query string true "Team ID"
// @Param namespace query string false "Filter by namespace"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Success 200 {object} costModel.CostListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/costs/team [get]
// @Security ApiKeyAuth
func (h *Handler) GetTeamCosts(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		response.GinBadRequest(c, nil)
		return
	}

	namespace := c.Query("namespace")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	result, err := h.costUseCase.GetTeamCosts(c.Request.Context(), teamID, namespace, startDate, endDate)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}