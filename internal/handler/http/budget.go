package http

import (
	"github.com/davidsugianto/go-pkgs/response"
	budgetModel "github.com/davidsugianto/idp-core/internal/model/budget"
	"github.com/gin-gonic/gin"
)

// ListBudgets godoc
// @Summary List budgets
// @Description List all budgets for a team
// @Tags budgets
// @Produce json
// @Param team_id query string true "Team ID"
// @Success 200 {object} budgetModel.BudgetListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/budgets [get]
// @Security ApiKeyAuth
func (h *Handler) ListBudgets(c *gin.Context) {
	teamID := c.Query("team_id")
	if teamID == "" {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.budgetUseCase.List(c.Request.Context(), teamID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CreateBudget godoc
// @Summary Create budget
// @Description Create a new budget with alert thresholds
// @Tags budgets
// @Accept json
// @Produce json
// @Param budget body budgetModel.CreateBudgetRequest true "Budget details"
// @Success 200 {object} budgetModel.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/budgets [post]
// @Security ApiKeyAuth
func (h *Handler) CreateBudget(c *gin.Context) {
	var req budgetModel.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.budgetUseCase.Create(c.Request.Context(), req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetBudget godoc
// @Summary Get budget
// @Description Get a budget by ID
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} budgetModel.BudgetResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/budgets/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetBudget(c *gin.Context) {
	id := c.Param("id")

	result, err := h.budgetUseCase.Get(c.Request.Context(), id)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// UpdateBudget godoc
// @Summary Update budget
// @Description Partially update a budget
// @Tags budgets
// @Accept json
// @Produce json
// @Param id path string true "Budget ID"
// @Param budget body budgetModel.UpdateBudgetRequest true "Budget fields to update"
// @Success 200 {object} budgetModel.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/budgets/{id} [patch]
// @Security ApiKeyAuth
func (h *Handler) UpdateBudget(c *gin.Context) {
	id := c.Param("id")

	var req budgetModel.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.budgetUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// DeleteBudget godoc
// @Summary Delete budget
// @Description Delete a budget by ID
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/budgets/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteBudget(c *gin.Context) {
	id := c.Param("id")

	if err := h.budgetUseCase.Delete(c.Request.Context(), id); err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "budget deleted"})
}

// ListBudgetAlerts godoc
// @Summary Get budget alerts
// @Description Get alert history for a budget
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {array} budgetModel.BudgetAlertResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/budgets/{id}/alerts [get]
// @Security ApiKeyAuth
func (h *Handler) ListBudgetAlerts(c *gin.Context) {
	id := c.Param("id")

	result, err := h.budgetUseCase.GetAlerts(c.Request.Context(), id)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}