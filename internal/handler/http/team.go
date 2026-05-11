package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/model/team"
	teamUsecase "github.com/davidsugianto/idp-core/internal/usecase/team"
	"github.com/gin-gonic/gin"
)

// CreateTeam godoc
// @Summary Create a new team
// @Description Create a new team
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body team.CreateTeamRequest true "Team request"
// @Success 201 {object} team.TeamResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams [post]
func (h *Handler) CreateTeam(c *gin.Context) {
	var req team.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	t, err := h.teamUseCase.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, teamUsecase.ErrTeamAlreadyExists) {
			response.GinError(c, http.StatusConflict, err)
			return
		}
		if errors.Is(err, teamUsecase.ErrInvalidSlug) {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, team.ToTeamResponse(t))
}

// ListTeams godoc
// @Summary List teams
// @Description List all teams with pagination
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} team.TeamListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams [get]
func (h *Handler) ListTeams(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.teamUseCase.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// GetTeam godoc
// @Summary Get team details
// @Description Get details of a specific team
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Success 200 {object} team.TeamResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams/{id} [get]
func (h *Handler) GetTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, errors.New("missing team id"))
		return
	}

	t, err := h.teamUseCase.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, teamUsecase.ErrTeamNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, team.ToTeamResponse(t))
}

// UpdateTeam godoc
// @Summary Update team
// @Description Update a team's details
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param request body team.UpdateTeamRequest true "Team update request"
// @Success 200 {object} team.TeamResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams/{id} [patch]
func (h *Handler) UpdateTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, errors.New("missing team id"))
		return
	}

	var req team.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	t, err := h.teamUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, teamUsecase.ErrTeamNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, team.ToTeamResponse(t))
}

// DeleteTeam godoc
// @Summary Delete team
// @Description Soft delete a team
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams/{id} [delete]
func (h *Handler) DeleteTeam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, errors.New("missing team id"))
		return
	}

	if err := h.teamUseCase.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, teamUsecase.ErrTeamNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "team deleted"})
}

// ListTeamMembers godoc
// @Summary List team members
// @Description List all members of a team
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Success 200 {object} team.TeamWithMembersResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams/{id}/members [get]
func (h *Handler) ListTeamMembers(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.GinBadRequest(c, errors.New("missing team id"))
		return
	}

	result, err := h.teamUseCase.ListMembers(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, teamUsecase.ErrTeamNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// AddTeamMember godoc
// @Summary Add team member
// @Description Add a member to a team
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param request body team.AddTeamMemberRequest true "Add member request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams/{id}/members [post]
func (h *Handler) AddTeamMember(c *gin.Context) {
	teamID := c.Param("id")
	if teamID == "" {
		response.GinBadRequest(c, errors.New("missing team id"))
		return
	}

	var req team.AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	member, err := h.teamUseCase.AddMember(c.Request.Context(), teamID, req)
	if err != nil {
		if errors.Is(err, teamUsecase.ErrTeamNotFound) {
			response.GinNotFound(c, err)
			return
		}
		if errors.Is(err, teamUsecase.ErrMemberAlreadyExists) {
			response.GinError(c, http.StatusConflict, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinCreated(c, gin.H{
		"id":      member.ID,
		"team_id": member.TeamID,
		"user_id": member.UserID,
		"role":    member.Role,
	})
}

// UpdateTeamMember godoc
// @Summary Update team member role
// @Description Update a team member's role
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Param request body team.UpdateTeamMemberRequest true "Update member request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams/{id}/members/{userId} [patch]
func (h *Handler) UpdateTeamMember(c *gin.Context) {
	teamID := c.Param("id")
	userID := c.Param("userId")
	if teamID == "" || userID == "" {
		response.GinBadRequest(c, errors.New("missing team id or user id"))
		return
	}

	var req team.UpdateTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	if err := h.teamUseCase.UpdateMember(c.Request.Context(), teamID, userID, req); err != nil {
		if errors.Is(err, teamUsecase.ErrTeamNotFound) || errors.Is(err, teamUsecase.ErrMemberNotFound) {
			response.GinNotFound(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "member role updated"})
}

// RemoveTeamMember godoc
// @Summary Remove team member
// @Description Remove a member from a team
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/teams/{id}/members/{userId} [delete]
func (h *Handler) RemoveTeamMember(c *gin.Context) {
	teamID := c.Param("id")
	userID := c.Param("userId")
	if teamID == "" || userID == "" {
		response.GinBadRequest(c, errors.New("missing team id or user id"))
		return
	}

	if err := h.teamUseCase.RemoveMember(c.Request.Context(), teamID, userID); err != nil {
		if errors.Is(err, teamUsecase.ErrTeamNotFound) || errors.Is(err, teamUsecase.ErrMemberNotFound) {
			response.GinNotFound(c, err)
			return
		}
		if errors.Is(err, teamUsecase.ErrCannotRemoveOwner) {
			response.GinBadRequest(c, err)
			return
		}
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{"message": "member removed"})
}
