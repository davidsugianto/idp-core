package http

import (
	"strconv"
	"time"

	auditlogModel "github.com/davidsugianto/idp-core/internal/model/auditlog"
	"github.com/davidsugianto/go-pkgs/response"
	"github.com/gin-gonic/gin"
)

// ListAuditLogs godoc
// @Summary List audit logs
// @Description List audit logs with filtering and pagination
// @Tags audit-logs
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param team_id query string false "Filter by team ID"
// @Param action query string false "Filter by action"
// @Param resource_type query string false "Filter by resource type"
// @Param resource_id query string false "Filter by resource ID"
// @Param status query string false "Filter by status"
// @Param start_date query string false "Filter by start date (RFC3339)"
// @Param end_date query string false "Filter by end date (RFC3339)"
// @Param limit query int false "Page limit" default(50)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} auditlogModel.AuditLogListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /v1/audit-logs [get]
// @Security ApiKeyAuth
func (h *Handler) ListAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := auditlogModel.AuditLogFilter{
		UserID:       c.Query("user_id"),
		TeamID:       c.Query("team_id"),
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		ResourceID:   c.Query("resource_id"),
		Status:       c.Query("status"),
		Limit:        limit,
		Offset:       offset,
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

	result, err := h.auditLogUseCase.List(c.Request.Context(), filter)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetAuditLog godoc
// @Summary Get audit log by ID
// @Description Get a specific audit log entry
// @Tags audit-logs
// @Produce json
// @Param id path string true "Audit Log ID"
// @Success 200 {object} auditlogModel.AuditLogResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/audit-logs/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetAuditLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, nil)
		return
	}

	result, err := h.auditLogUseCase.Get(c.Request.Context(), id)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}
	if result == nil {
		response.GinNotFound(c, nil)
		return
	}

	response.GinSuccess(c, result)
}