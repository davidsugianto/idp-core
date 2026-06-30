package http

import (
	"errors"
	"strconv"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	buildApplicationUsecase "github.com/davidsugianto/idp-core/internal/usecase/build_application"
	"github.com/gin-gonic/gin"
)

// CreateBuildApplication godoc
// @Summary Create build application
// @Description Register a buildable application for a team using repository metadata and descriptor path
// @Tags build-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body buildApplicationModel.CreateBuildApplicationRequest true "Build application create request"
// @Success 201 {object} buildApplicationModel.BuildApplicationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/build-applications [post]
func (h *Handler) CreateBuildApplication(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	actorID := middleware.GetUserID(c)
	if teamID == "" || actorID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req buildApplicationModel.CreateBuildApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	app, err := h.buildApplicationUseCase.CreateApplication(c.Request.Context(), teamID, actorID, &req)
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinCreated(c, app)
}

// ListBuildApplications godoc
// @Summary List build applications
// @Description List build applications for the caller's team
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param name query string false "Filter by name"
// @Param runtime_family query string false "Filter by runtime family"
// @Param status query string false "Filter by application status"
// @Param deployment_automation_enabled query bool false "Filter by deployment automation toggle"
// @Param limit query int false "Page size" default(20)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} buildApplicationModel.BuildApplicationListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/build-applications [get]
func (h *Handler) ListBuildApplications(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req buildApplicationModel.ListBuildApplicationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	apps, err := h.buildApplicationUseCase.ListApplications(c.Request.Context(), teamID, &req)
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, apps)
}

// GetBuildApplication godoc
// @Summary Get build application
// @Description Get details for a build application in team scope
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Build application ID"
// @Success 200 {object} buildApplicationModel.BuildApplicationResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/build-applications/{id} [get]
func (h *Handler) GetBuildApplication(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	app, err := h.buildApplicationUseCase.GetApplication(c.Request.Context(), teamID, c.Param("id"))
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, app)
}

// UpdateBuildApplication godoc
// @Summary Update build application
// @Description Update mutable fields on a build application in team scope
// @Tags build-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Build application ID"
// @Param request body buildApplicationModel.UpdateBuildApplicationRequest true "Build application update request"
// @Success 200 {object} buildApplicationModel.BuildApplicationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/build-applications/{id} [patch]
func (h *Handler) UpdateBuildApplication(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	actorID := middleware.GetUserID(c)
	if teamID == "" || actorID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req buildApplicationModel.UpdateBuildApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	app, err := h.buildApplicationUseCase.UpdateApplication(c.Request.Context(), teamID, actorID, c.Param("id"), &req)
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, app)
}

// DeleteBuildApplication godoc
// @Summary Delete build application
// @Description Soft delete a build application in team scope
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Build application ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/build-applications/{id} [delete]
func (h *Handler) DeleteBuildApplication(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	actorID := middleware.GetUserID(c)
	if teamID == "" || actorID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	if err := h.buildApplicationUseCase.DeleteApplication(c.Request.Context(), teamID, actorID, c.Param("id")); err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "build application deleted"})
}

// TriggerBuild godoc
// @Summary Trigger build
// @Description Trigger an asynchronous build for a build application
// @Tags build-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Build application ID"
// @Param request body buildApplicationModel.TriggerBuildRequest true "Build trigger request"
// @Success 201 {object} buildApplicationModel.BuildActionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/build-applications/{id}/builds [post]
func (h *Handler) TriggerBuild(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	actorID := middleware.GetUserID(c)
	if teamID == "" || actorID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req buildApplicationModel.TriggerBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	result, err := h.buildApplicationUseCase.TriggerBuild(c.Request.Context(), teamID, actorID, c.Param("id"), &req)
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinCreated(c, result)
}

// ListBuilds godoc
// @Summary List build history
// @Description List build history for a build application
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Build application ID"
// @Param limit query int false "Page size" default(20)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} buildApplicationModel.BuildHistoryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/build-applications/{id}/builds [get]
func (h *Handler) ListBuilds(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.buildApplicationUseCase.ListBuilds(c.Request.Context(), teamID, c.Param("id"), limit, offset)
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetBuild godoc
// @Summary Get build detail
// @Description Get details for a build in team scope
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param buildId path string true "Build ID"
// @Success 200 {object} buildApplicationModel.BuildResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/builds/{buildId} [get]
func (h *Handler) GetBuild(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	result, err := h.buildApplicationUseCase.GetBuild(c.Request.Context(), teamID, c.Param("buildId"))
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// RetryBuild godoc
// @Summary Retry build
// @Description Retry a failed or canceled build
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param buildId path string true "Build ID"
// @Success 200 {object} buildApplicationModel.BuildActionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/builds/{buildId}/retry [post]
func (h *Handler) RetryBuild(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	actorID := middleware.GetUserID(c)
	if teamID == "" || actorID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	result, err := h.buildApplicationUseCase.RetryBuild(c.Request.Context(), teamID, actorID, c.Param("buildId"))
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// CancelBuild godoc
// @Summary Cancel build
// @Description Cancel an active build
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param buildId path string true "Build ID"
// @Success 200 {object} buildApplicationModel.BuildActionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/builds/{buildId}/cancel [post]
func (h *Handler) CancelBuild(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	actorID := middleware.GetUserID(c)
	if teamID == "" || actorID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	result, err := h.buildApplicationUseCase.CancelBuild(c.Request.Context(), teamID, actorID, c.Param("buildId"))
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// StreamBuildLogs godoc
// @Summary Stream build logs
// @Description Stream build logs incrementally with optional cursor and page size
// @Tags build-applications
// @Produce json
// @Security BearerAuth
// @Param buildId path string true "Build ID"
// @Param after_sequence query int false "Last seen log sequence" default(0)
// @Param limit query int false "Maximum log lines" default(200)
// @Success 200 {object} buildApplicationModel.BuildLogStreamResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/builds/{buildId}/logs/stream [get]
func (h *Handler) StreamBuildLogs(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	afterSequence, _ := strconv.ParseInt(c.DefaultQuery("after_sequence", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))

	result, err := h.buildApplicationUseCase.StreamBuildLogs(c.Request.Context(), teamID, c.Param("buildId"), afterSequence, limit)
	if err != nil {
		h.handleBuildApplicationError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

func (h *Handler) handleBuildApplicationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, buildApplicationUsecase.ErrApplicationNotFound), errors.Is(err, buildApplicationUsecase.ErrBuildNotFound):
		response.GinNotFound(c, err)
	case errors.Is(err, buildApplicationUsecase.ErrApplicationAlreadyExists),
		errors.Is(err, buildApplicationUsecase.ErrBuildNotCancelable),
		errors.Is(err, buildApplicationUsecase.ErrBuildRetryNotAllowed),
		errors.Is(err, buildApplicationUsecase.ErrBuildIdempotencyConflict),
		errors.Is(err, buildApplicationUsecase.ErrInvalidRuntimeFamily),
		errors.Is(err, buildApplicationUsecase.ErrInvalidRegistryType),
		errors.Is(err, buildApplicationUsecase.ErrInvalidApplicationStatus),
		errors.Is(err, buildApplicationUsecase.ErrInvalidBuildStatus),
		errors.Is(err, buildApplicationUsecase.ErrInvalidBuildTriggerType),
		errors.Is(err, buildApplicationUsecase.ErrInvalidLifecycleEventType),
		errors.Is(err, buildApplicationUsecase.ErrInvalidDeploymentUpdateState),
		errors.Is(err, buildApplicationUsecase.ErrInvalidSecurityStatus),
		errors.Is(err, buildApplicationUsecase.ErrUnauthorizedTeamScope):
		response.GinBadRequest(c, err)
	default:
		response.GinInternalServerError(c, err)
	}
}
