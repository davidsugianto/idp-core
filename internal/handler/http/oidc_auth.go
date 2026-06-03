package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	userModel "github.com/davidsugianto/idp-core/internal/model/user"
	oidcPkg "github.com/davidsugianto/idp-core/internal/pkg/oidc"
	"github.com/gin-gonic/gin"
)

var (
	errMissingState        = errors.New("missing state parameter")
	errStateMismatch       = errors.New("state parameter mismatch")
	errMissingCode         = errors.New("missing authorization code")
	errNoIDToken           = errors.New("no id_token in response")
	errNoEmail             = errors.New("no email in user info")
	errMissingRefreshToken = errors.New("refresh token is required")
	errUserNotFound        = errors.New("user not found")
)

type OIDCRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type OIDCLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	IsAdmin      bool   `json:"is_admin"`
}

// OIDCLogin godoc
// @Summary Initiate OIDC login
// @Description Redirects to the OIDC provider for authentication
// @Tags auth
// @Produce json
// @Success 200 {object} OIDCLoginResponse
// @Router /auth/oidc/login [get]
func (h *Handler) OIDCLogin(c *gin.Context) {
	state := generateState()

	// Set state cookie (HttpOnly, 10 min TTL)
	c.SetCookie("oidc_state", state, 600, "/", "", false, true)

	authURL := h.oidcClient.GetAuthURL(state)

	if c.GetHeader("Accept") == "application/json" {
		response.GinSuccess(c, OIDCLoginResponse{AuthURL: authURL})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// OIDCCallback godoc
// @Summary OIDC callback
// @Description Handles the OIDC provider callback, exchanges code for tokens, and issues a JWT
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "OAuth2 state"
// @Success 200 {object} OIDCTokenResponse
// @Router /auth/oidc/callback [get]
func (h *Handler) OIDCCallback(c *gin.Context) {
	stateCookie, err := c.Cookie("oidc_state")
	if err != nil || stateCookie == "" {
		response.GinBadRequest(c, errMissingState)
		return
	}

	stateParam := c.Query("state")
	if stateParam == "" || stateParam != stateCookie {
		response.GinBadRequest(c, errStateMismatch)
		return
	}

	// Clear state cookie
	c.SetCookie("oidc_state", "", -1, "/", "", false, true)

	code := c.Query("code")
	if code == "" {
		response.GinBadRequest(c, errMissingCode)
		return
	}

	ctx := c.Request.Context()

	// Exchange authorization code for tokens
	oauth2Token, err := h.oidcClient.Exchange(ctx, code)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	// Get the raw ID token for user info extraction
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		response.GinInternalServerError(c, errNoIDToken)
		return
	}

	// Verify and extract user info
	userInfo, err := h.oidcVerifier.VerifyAndExtract(ctx, rawIDToken)
	if err != nil {
		response.GinUnauthorized(c, err)
		return
	}

	if userInfo.Email == "" {
		response.GinInternalServerError(c, errNoEmail)
		return
	}

	// Upsert user in local database
	dbUser, err := h.upsertOIDCUser(ctx, userInfo)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	isAdmin := h.oidcVerifier.IsAdmin(userInfo)

	// Issue JWT
	jwt, err := middleware.GenerateToken(
		h.authConfig,
		dbUser.ID,
		"",
		userInfo.Email,
		isAdmin,
	)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	// Set JWT as HttpOnly cookie (24h)
	c.SetCookie("auth_token", jwt, int((24 * time.Hour).Seconds()), "/", "", false, true)

	// Set refresh token as HttpOnly cookie (7 days)
	if oauth2Token.RefreshToken != "" {
		c.SetCookie("refresh_token", oauth2Token.RefreshToken, int((7 * 24 * time.Hour).Seconds()), "/", "", false, true)
	}

	expiresIn := int64(24 * time.Hour.Seconds())
	if !oauth2Token.Expiry.IsZero() {
		expiresIn = int64(time.Until(oauth2Token.Expiry).Seconds())
	}

	if c.GetHeader("Accept") == "application/json" {
		response.GinSuccess(c, OIDCTokenResponse{
			AccessToken:  jwt,
			RefreshToken: oauth2Token.RefreshToken,
			ExpiresIn:    expiresIn,
			TokenType:    "Bearer",
			UserID:       dbUser.ID,
			Email:        userInfo.Email,
			IsAdmin:      isAdmin,
		})
		return
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = "/"
	}
	c.Redirect(http.StatusFound, redirectURI)
}

// OIDCRefresh godoc
// @Summary Refresh OIDC token
// @Description Exchanges a refresh token for new tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body OIDCRefreshRequest true "Refresh token"
// @Success 200 {object} OIDCTokenResponse
// @Router /auth/oidc/refresh [post]
func (h *Handler) OIDCRefresh(c *gin.Context) {
	var req OIDCRefreshRequest

	// Try JSON body, fallback to cookie
	if err := c.ShouldBindJSON(&req); err != nil {
		if cookie, cookieErr := c.Cookie("refresh_token"); cookieErr == nil && cookie != "" {
			req.RefreshToken = cookie
		}
	}

	if req.RefreshToken == "" {
		response.GinBadRequest(c, errMissingRefreshToken)
		return
	}

	ctx := c.Request.Context()

	// Exchange refresh token for new tokens via OIDC provider
	newToken, err := h.oidcClient.Refresh(ctx, req.RefreshToken)
	if err != nil {
		response.GinUnauthorized(c, err)
		return
	}

	// Extract user info from new ID token
	rawIDToken, ok := newToken.Extra("id_token").(string)
	if ok && rawIDToken != "" {
		userInfo, err := h.oidcVerifier.VerifyAndExtract(ctx, rawIDToken)
		if err == nil && userInfo != nil {
			dbUser, err := h.upsertOIDCUser(ctx, userInfo)
			if err == nil && dbUser != nil {
				isAdmin := h.oidcVerifier.IsAdmin(userInfo)

				jwt, err := middleware.GenerateToken(
					h.authConfig,
					dbUser.ID,
					"",
					userInfo.Email,
					isAdmin,
				)
				if err == nil {
					c.SetCookie("auth_token", jwt, int((24 * time.Hour).Seconds()), "/", "", false, true)
					if newToken.RefreshToken != "" {
						c.SetCookie("refresh_token", newToken.RefreshToken, int((7 * 24 * time.Hour).Seconds()), "/", "", false, true)
					}

					expiresIn := int64(24 * time.Hour.Seconds())
					if !newToken.Expiry.IsZero() {
						expiresIn = int64(time.Until(newToken.Expiry).Seconds())
					}

					response.GinSuccess(c, OIDCTokenResponse{
						AccessToken:  jwt,
						RefreshToken: newToken.RefreshToken,
						ExpiresIn:    expiresIn,
						TokenType:    "Bearer",
						UserID:       dbUser.ID,
						Email:        userInfo.Email,
						IsAdmin:      isAdmin,
					})
					return
				}
			}
		}
	}

	// If we couldn't get user info from the new token, just return the raw tokens
	if newToken.RefreshToken != "" {
		c.SetCookie("refresh_token", newToken.RefreshToken, int((7 * 24 * time.Hour).Seconds()), "/", "", false, true)
	}

	expiresIn := int64(24 * time.Hour.Seconds())
	if !newToken.Expiry.IsZero() {
		expiresIn = int64(time.Until(newToken.Expiry).Seconds())
	}

	response.GinSuccess(c, gin.H{
		"access_token":  newToken.AccessToken,
		"refresh_token": newToken.RefreshToken,
		"expires_in":    expiresIn,
		"token_type":    "Bearer",
	})
}

// OIDCLogout godoc
// @Summary Logout from OIDC session
// @Description Clears auth cookies and redirects to OIDC provider end session endpoint
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/oidc/logout [post]
func (h *Handler) OIDCLogout(c *gin.Context) {
	// Clear auth cookies
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("oidc_state", "", -1, "/", "", false, true)

	if h.oidcEndSessionURL != "" {
		c.Redirect(http.StatusFound, h.oidcEndSessionURL)
		return
	}

	response.GinSuccess(c, gin.H{"message": "logged out"})
}

// upsertOIDCUser finds or creates a local user from OIDC user info
func (h *Handler) upsertOIDCUser(ctx context.Context, userInfo *oidcPkg.UserInfo) (*userModel.User, error) {
	// Check for existing user by email
	existing, err := h.userUseCase.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Create new user
	name := userInfo.Name
	if name == "" {
		name = userInfo.Email
	}
	createReq := userModel.CreateUserRequest{
		Email:    userInfo.Email,
		Name:     name,
		Provider: "oidc",
	}
	newUser, err := h.userUseCase.Create(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}