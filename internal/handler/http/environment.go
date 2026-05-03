package http

import (
	"errors"
	"net/http"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/model/environment"
	_ "github.com/davidsugianto/idp-core/internal/model/workload" // for swagger docs
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	"github.com/gin-gonic/gin"
)

// CreateEnvironment godoc
// @Summary Create a new environment
// @Description Create a new Kubernetes environment with ArgoCD integration
// @Tags environments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body environment.CreateEnvironmentRequest true "Environment request"
// @Success 201 {object} environment.EnvironmentResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments [post]
func (h *Handler) CreateEnvironment(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req environment.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	env, err := h.environmentUseCase.Create(c.Request.Context(), teamID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, environment.ToEnvironmentResponse(env))
}

// ListEnvironments godoc
// @Summary List environments
// @Description List all environments for the caller's team
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Success 200 {array} environment.EnvironmentResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments [get]
func (h *Handler) ListEnvironments(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	envs, err := h.environmentUseCase.List(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]*environment.EnvironmentResponse, len(envs))
	for i, env := range envs {
		responses[i] = environment.ToEnvironmentResponse(&env)
	}

	c.JSON(http.StatusOK, responses)
}

// GetEnvironment godoc
// @Summary Get environment details
// @Description Get details of a specific environment
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Environment ID"
// @Success 200 {object} environment.EnvironmentResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments/{id} [get]
func (h *Handler) GetEnvironment(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing environment id"})
		return
	}

	env, err := h.environmentUseCase.Get(c.Request.Context(), teamID, id)
	if err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, environment.ToEnvironmentResponse(env))
}

// DeleteEnvironment godoc
// @Summary Delete environment
// @Description Tear down an environment and all its resources
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Environment ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments/{id} [delete]
func (h *Handler) DeleteEnvironment(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing environment id"})
		return
	}

	if err := h.environmentUseCase.Delete(c.Request.Context(), teamID, id); err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "environment deleted"})
}

// GetEnvironmentStatus godoc
// @Summary Get environment status
// @Description Get live status of an environment including pod and ArgoCD sync status
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Environment ID"
// @Success 200 {object} environment.EnvironmentStatusResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments/{id}/status [get]
func (h *Handler) GetEnvironmentStatus(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing environment id"})
		return
	}

	status, err := h.environmentUseCase.GetStatus(c.Request.Context(), teamID, id)
	if err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// SyncEnvironment godoc
// @Summary Trigger environment sync
// @Description Trigger a manual sync of the ArgoCD application for an environment
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Environment ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments/{id}/sync [post]
func (h *Handler) SyncEnvironment(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing environment id"})
		return
	}

	if err := h.environmentUseCase.TriggerSync(c.Request.Context(), teamID, id); err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sync triggered"})
}

// GetGitOpsStatus godoc
// @Summary Get GitOps status
// @Description Get the ArgoCD sync and health status for an environment
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Environment ID"
// @Success 200 {object} environment.ArgoStatus
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments/{id}/gitops/status [get]
func (h *Handler) GetGitOpsStatus(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing environment id"})
		return
	}

	status, err := h.environmentUseCase.GetGitOpsStatus(c.Request.Context(), teamID, id)
	if err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetWorkloads godoc
// @Summary Get environment workloads
// @Description Get all workloads (deployments) and their pods in an environment
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Environment ID"
// @Success 200 {object} workload.WorkloadStatusResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments/{id}/workloads [get]
func (h *Handler) GetWorkloads(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing environment id"})
		return
	}

	status, err := h.environmentUseCase.GetWorkloads(c.Request.Context(), teamID, id)
	if err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetWorkloadDetails godoc
// @Summary Get workload details
// @Description Get details of a specific workload in an environment
// @Tags environments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Environment ID"
// @Param name path string true "Workload name"
// @Success 200 {object} workload.WorkloadInfo
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/environments/{id}/workloads/{name} [get]
func (h *Handler) GetWorkloadDetails(c *gin.Context) {
	teamID := middleware.GetTeamID(c)
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing environment id"})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing workload name"})
		return
	}

	details, err := h.environmentUseCase.GetWorkloadDetails(c.Request.Context(), teamID, id, name)
	if err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) || errors.Is(err, envUsecase.ErrWorkloadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}
