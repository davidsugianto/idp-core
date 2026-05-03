package http

import (
	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authConfig *config.AuthConfig
}

func NewAuthHandler(cfg *config.AuthConfig) *AuthHandler {
	return &AuthHandler{authConfig: cfg}
}

type LoginRequest struct {
	UserID string `json:"user_id" binding:"required"`
	TeamID string `json:"team_id" binding:"required"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	token, err := middleware.GenerateToken(h.authConfig, req.UserID, req.TeamID)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, gin.H{
		"token": token,
		"type":  "Bearer",
	})
}
