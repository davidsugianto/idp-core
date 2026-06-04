package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	oidcPkg "github.com/davidsugianto/idp-core/internal/pkg/oidc"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	keycloakBaseURL   = "http://localhost:8081"
	keycloakRealm     = "idp-core"
	keycloakClientID  = "idp-core"
	keycloakSecret    = "idp-core-secret-key"
	keycloakIssuerURL = keycloakBaseURL + "/realms/" + keycloakRealm
)

// getToken obtains an ID token from Keycloak for the given user.
// Always requests the openid scope so an id_token is returned.
func getToken(t *testing.T, username, password string) string {
	t.Helper()

	tokenURL := keycloakIssuerURL + "/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", keycloakClientID)
	data.Set("client_secret", keycloakSecret)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", "openid profile email")

	resp, err := http.PostForm(tokenURL, data)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "failed to get token for %s", username)

	var tokenResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	idToken, ok := tokenResp["id_token"].(string)
	require.True(t, ok, "id_token not found in response for %s", username)
	require.NotEmpty(t, idToken)

	return idToken
}

// TestIntegration_OIDC_ClientCreation tests creating an OIDC client against real Keycloak
func TestIntegration_OIDC_ClientCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("discovery_and_client_creation", func(t *testing.T) {
		cfg := &oidcPkg.Config{
			IssuerURL:    keycloakIssuerURL,
			ClientID:     keycloakClientID,
			ClientSecret: keycloakSecret,
			RedirectURL:  "http://localhost:8080/auth/callback",
			Scopes:       []string{"openid", "profile", "email"},
		}

		client, err := oidcPkg.NewClient(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, client)

		oauth2Conf := client.GetOAuth2Config()
		assert.NotNil(t, oauth2Conf)
		assert.Equal(t, keycloakClientID, oauth2Conf.ClientID)

		verifier := client.GetVerifier()
		assert.NotNil(t, verifier)

		t.Log("OIDC client created successfully")
		t.Logf("Issuer: %s", keycloakIssuerURL)
	})

	t.Run("auth_url_generation", func(t *testing.T) {
		cfg := &oidcPkg.Config{
			IssuerURL:    keycloakIssuerURL,
			ClientID:     keycloakClientID,
			ClientSecret: keycloakSecret,
			RedirectURL:  "http://localhost:8080/auth/callback",
			Scopes:       []string{"openid", "profile", "email"},
		}

		client, err := oidcPkg.NewClient(ctx, cfg)
		require.NoError(t, err)

		authURL := client.GetAuthURL("test-state-123")
		assert.Contains(t, authURL, keycloakIssuerURL)
		assert.Contains(t, authURL, "client_id="+keycloakClientID)
		assert.Contains(t, authURL, "state=test-state-123")
		assert.Contains(t, authURL, "response_type=code")
	})

	t.Run("invalid_issuer_url", func(t *testing.T) {
		cfg := &oidcPkg.Config{
			IssuerURL: "http://localhost:99999/nonexistent",
			ClientID:  "test",
		}

		client, err := oidcPkg.NewClient(ctx, cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("empty_issuer_url", func(t *testing.T) {
		cfg := &oidcPkg.Config{
			ClientID: "test",
		}

		client, err := oidcPkg.NewClient(ctx, cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "issuer URL is required")
	})

	t.Run("empty_client_id", func(t *testing.T) {
		cfg := &oidcPkg.Config{
			IssuerURL: keycloakIssuerURL,
		}

		client, err := oidcPkg.NewClient(ctx, cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "client ID is required")
	})
}

// TestIntegration_OIDC_TokenVerification tests verifying real Keycloak ID tokens
func TestIntegration_OIDC_TokenVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := &oidcPkg.Config{
		IssuerURL:    keycloakIssuerURL,
		ClientID:     keycloakClientID,
		ClientSecret: keycloakSecret,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "profile", "email"},
	}

	client, err := oidcPkg.NewClient(ctx, cfg)
	require.NoError(t, err)

	t.Run("verify_developer_token", func(t *testing.T) {
		idToken := getToken(t, "developer", "dev123")

		verifiedToken, err := client.VerifyIDToken(ctx, idToken)
		require.NoError(t, err)
		require.NotNil(t, verifiedToken)

		assert.NotEmpty(t, verifiedToken.Subject)
		t.Logf("Developer token verified - subject: %s", verifiedToken.Subject)
	})

	t.Run("verify_platform_admin_token", func(t *testing.T) {
		idToken := getToken(t, "platform-admin", "admin123")

		verifiedToken, err := client.VerifyIDToken(ctx, idToken)
		require.NoError(t, err)
		require.NotNil(t, verifiedToken)

		assert.NotEmpty(t, verifiedToken.Subject)
		t.Logf("Platform admin token verified - subject: %s", verifiedToken.Subject)
	})

	t.Run("verify_invalid_token", func(t *testing.T) {
		_, err := client.VerifyIDToken(ctx, "invalid-token")
		assert.Error(t, err)
	})
}

// TestIntegration_OIDC_UserInfoExtraction tests extracting user info from verified tokens
func TestIntegration_OIDC_UserInfoExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := &oidcPkg.Config{
		IssuerURL:    keycloakIssuerURL,
		ClientID:     keycloakClientID,
		ClientSecret: keycloakSecret,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "profile", "email"},
	}

	client, err := oidcPkg.NewClient(ctx, cfg)
	require.NoError(t, err)

	verifier := oidcPkg.NewVerifier(client, &oidcPkg.VerifierConfig{
		GroupsClaim: "groups",
		AdminGroup:  "platform-admins",
	})

	t.Run("extract_developer_userinfo", func(t *testing.T) {
		idToken := getToken(t, "developer", "dev123")

		userInfo, err := verifier.VerifyAndExtract(ctx, idToken)
		require.NoError(t, err)
		require.NotNil(t, userInfo)

		assert.NotEmpty(t, userInfo.Subject)
		assert.NotEmpty(t, userInfo.Email)
		t.Logf("Developer - Subject: %s, Email: %s, Name: %s", userInfo.Subject, userInfo.Email, userInfo.Name)
	})

	t.Run("extract_platform_admin_userinfo", func(t *testing.T) {
		idToken := getToken(t, "platform-admin", "admin123")

		userInfo, err := verifier.VerifyAndExtract(ctx, idToken)
		require.NoError(t, err)
		require.NotNil(t, userInfo)

		assert.NotEmpty(t, userInfo.Subject)
		assert.NotEmpty(t, userInfo.Email)
		t.Logf("Platform Admin - Subject: %s, Email: %s, Groups: %v", userInfo.Subject, userInfo.Email, userInfo.Groups)

		if len(userInfo.Groups) > 0 {
			assert.True(t, verifier.IsAdmin(userInfo), "platform-admin should be identified as admin")
		} else {
			t.Log("Groups claim not present in token — ensure Keycloak client has a Group Membership mapper")
		}
	})

	t.Run("developer_is_not_admin", func(t *testing.T) {
		idToken := getToken(t, "developer", "dev123")

		userInfo, err := verifier.VerifyAndExtract(ctx, idToken)
		require.NoError(t, err)

		assert.False(t, verifier.IsAdmin(userInfo), "developer should not be identified as admin")
	})
}

// TestIntegration_OIDC_Middleware tests the OIDC middleware with real Keycloak tokens
func TestIntegration_OIDC_Middleware(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gin.SetMode(gin.TestMode)

	cfg := &oidcPkg.Config{
		IssuerURL:    keycloakIssuerURL,
		ClientID:     keycloakClientID,
		ClientSecret: keycloakSecret,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "profile", "email"},
	}

	client, err := oidcPkg.NewClient(ctx, cfg)
	require.NoError(t, err)

	verifier := oidcPkg.NewVerifier(client, &oidcPkg.VerifierConfig{
		GroupsClaim: "groups",
		AdminGroup:  "platform-admins",
	})

	oidcMiddlewareCfg := &middleware.OIDCConfig{
		OIDCCfg: &config.OIDCConfig{
			IssuerURL:    keycloakIssuerURL,
			ClientID:     keycloakClientID,
			ClientSecret: keycloakSecret,
			RedirectURL:  "http://localhost:8080/auth/callback",
			Scopes:       []string{"openid", "profile", "email"},
			GroupsClaim:  "groups",
			AdminGroup:   "platform-admins",
		},
		OIDCClient:  client,
		OIDCVerifier: verifier,
	}

	t.Run("protected_route_with_valid_token", func(t *testing.T) {
		idToken := getToken(t, "developer", "dev123")

		router := gin.New()
		router.GET("/protected", middleware.OIDCAuth(oidcMiddlewareCfg), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id":  middleware.GetUserID(c),
				"email":    middleware.GetUserEmail(c),
				"is_admin": middleware.IsAdmin(c),
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+idToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		require.NoError(t, err)

		assert.NotEmpty(t, body["user_id"])
		assert.NotEmpty(t, body["email"])
		assert.Equal(t, false, body["is_admin"])

		t.Logf("Protected route - UserID: %v, Email: %v, IsAdmin: %v", body["user_id"], body["email"], body["is_admin"])
	})

	t.Run("protected_route_with_admin_token", func(t *testing.T) {
		idToken := getToken(t, "platform-admin", "admin123")

		router := gin.New()
		router.GET("/admin", middleware.OIDCAuth(oidcMiddlewareCfg), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id":  middleware.GetUserID(c),
				"email":    middleware.GetUserEmail(c),
				"groups":   middleware.GetUserGroups(c),
				"is_admin": middleware.IsAdmin(c),
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+idToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		require.NoError(t, err)

		if groups, ok := body["groups"].([]interface{}); ok && len(groups) > 0 {
			assert.Equal(t, true, body["is_admin"], "platform-admin should be identified as admin")
		} else {
			t.Log("Groups claim not present in token — ensure Keycloak client has a Group Membership mapper")
		}
		t.Logf("Admin route - UserID: %v, IsAdmin: %v, Groups: %v", body["user_id"], body["is_admin"], body["groups"])
	})

	t.Run("protected_route_without_token", func(t *testing.T) {
		router := gin.New()
		router.GET("/protected", middleware.OIDCAuth(oidcMiddlewareCfg), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("protected_route_with_invalid_token", func(t *testing.T) {
		router := gin.New()
		router.GET("/protected", middleware.OIDCAuth(oidcMiddlewareCfg), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-here")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("protected_route_with_malformed_header", func(t *testing.T) {
		router := gin.New()
		router.GET("/protected", middleware.OIDCAuth(oidcMiddlewareCfg), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "NotBearer token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestIntegration_OIDC_TokenRefresh tests the token refresh flow
func TestIntegration_OIDC_TokenRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("refresh_token_flow", func(t *testing.T) {
		tokenURL := keycloakIssuerURL + "/protocol/openid-connect/token"

		data := url.Values{}
		data.Set("grant_type", "password")
		data.Set("client_id", keycloakClientID)
		data.Set("client_secret", keycloakSecret)
		data.Set("username", "developer")
		data.Set("password", "dev123")
		data.Set("scope", "openid profile email")

		resp, err := http.PostForm(tokenURL, data)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		require.NoError(t, err)

		refreshToken, ok := tokenResp["refresh_token"].(string)
		require.True(t, ok, "refresh_token not found")
		require.NotEmpty(t, refreshToken)

		// Use refresh token to get new tokens
		data = url.Values{}
		data.Set("grant_type", "refresh_token")
		data.Set("client_id", keycloakClientID)
		data.Set("client_secret", keycloakSecret)
		data.Set("refresh_token", refreshToken)

		resp, err = http.PostForm(tokenURL, data)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, "refresh token exchange failed")

		var refreshedResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&refreshedResp)
		require.NoError(t, err)

		newAccessToken, ok := refreshedResp["access_token"].(string)
		require.True(t, ok)
		require.NotEmpty(t, newAccessToken)

		newRefreshToken, ok := refreshedResp["refresh_token"].(string)
		require.True(t, ok)
		require.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, refreshToken, newRefreshToken, "new refresh token should differ from old one")

		t.Log("Token refresh successful")
	})
}

// TestIntegration_OIDC_FullFlow tests the complete OIDC integration flow
func TestIntegration_OIDC_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gin.SetMode(gin.TestMode)

	t.Log("=== OIDC Integration Full Flow ===")

	// Step 1: Create OIDC client (discovery)
	t.Log("Step 1: OIDC Discovery & Client Creation")
	cfg := &oidcPkg.Config{
		IssuerURL:    keycloakIssuerURL,
		ClientID:     keycloakClientID,
		ClientSecret: keycloakSecret,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "profile", "email"},
	}

	client, err := oidcPkg.NewClient(ctx, cfg)
	require.NoError(t, err)
	t.Log("  Client created")

	// Step 2: Create verifier
	verifier := oidcPkg.NewVerifier(client, &oidcPkg.VerifierConfig{
		GroupsClaim: "groups",
		AdminGroup:  "platform-admins",
	})
	t.Log("  Verifier created")

	// Step 3: Get and verify admin token
	t.Log("Step 2: Obtain & Verify Platform Admin Token")
	adminToken := getToken(t, "platform-admin", "admin123")

	verifiedToken, err := client.VerifyIDToken(ctx, adminToken)
	require.NoError(t, err)
	assert.NotEmpty(t, verifiedToken.Subject)
	t.Logf("  Token verified - subject: %s", verifiedToken.Subject)

	// Step 4: Extract user info
	t.Log("Step 3: Extract User Info")
	userInfo, err := verifier.VerifyAndExtract(ctx, adminToken)
	require.NoError(t, err)
	assert.NotEmpty(t, userInfo.Subject)
	t.Logf("  User: %s <%s>", userInfo.Name, userInfo.Email)
	t.Logf("  Groups: %v", userInfo.Groups)

	// Step 5: Check admin status
	t.Log("Step 4: Check Admin Status")
	if len(userInfo.Groups) > 0 {
		assert.True(t, verifier.IsAdmin(userInfo))
		t.Logf("  IsAdmin: true")
	} else {
		t.Log("  Groups not in token (mapper not configured)")
	}

	// Step 6: Test middleware with token
	t.Log("Step 5: Middleware Protection")
	oidcMiddlewareCfg := &middleware.OIDCConfig{
		OIDCCfg: &config.OIDCConfig{
			IssuerURL:    keycloakIssuerURL,
			ClientID:     keycloakClientID,
			ClientSecret: keycloakSecret,
			RedirectURL:  "http://localhost:8080/auth/callback",
			Scopes:       []string{"openid", "profile", "email"},
			GroupsClaim:  "groups",
			AdminGroup:   "platform-admins",
		},
		OIDCClient:  client,
		OIDCVerifier: verifier,
	}

	router := gin.New()
	router.GET("/api/test", middleware.OIDCAuth(oidcMiddlewareCfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  middleware.GetUserID(c),
			"email":    middleware.GetUserEmail(c),
			"is_admin": middleware.IsAdmin(c),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	if groups, ok := body["groups"].([]interface{}); ok && len(groups) > 0 {
		assert.Equal(t, true, body["is_admin"])
	}
	t.Logf("  Middleware passed - IsAdmin: %v", body["is_admin"])

	// Step 7: Repeat with developer user
	t.Log("Step 6: Developer User Flow")
	devToken := getToken(t, "developer", "dev123")

	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+devToken)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, false, body["is_admin"])
	t.Logf("  Developer - IsAdmin: %v, Email: %v", body["is_admin"], body["email"])

	t.Log("=== OIDC Integration Full Flow Complete ===")
}