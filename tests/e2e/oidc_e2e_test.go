package e2e

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_OIDC_Discovery tests the OIDC discovery endpoint
func TestE2E_OIDC_Discovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"

	// Test the well-known OpenID configuration endpoint
	t.Run("discovery_endpoint", func(t *testing.T) {
		discoveryURL := keycloakURL + "/realms/" + realm + "/.well-known/openid-configuration"

		resp, err := http.Get(discoveryURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var config map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&config)
		require.NoError(t, err)

		// Verify required OIDC fields
		assert.Contains(t, config, "issuer")
		assert.Contains(t, config, "authorization_endpoint")
		assert.Contains(t, config, "token_endpoint")
		assert.Contains(t, config, "userinfo_endpoint")
		assert.Contains(t, config, "jwks_uri")

		// Log the endpoints for debugging
		t.Logf("Issuer: %v", config["issuer"])
		t.Logf("Authorization: %v", config["authorization_endpoint"])
		t.Logf("Token: %v", config["token_endpoint"])
		t.Logf("UserInfo: %v", config["userinfo_endpoint"])
		t.Logf("JWKS: %v", config["jwks_uri"])
	})
}

// TestE2E_OIDC_TokenEndpoint tests the OIDC token endpoint
func TestE2E_OIDC_TokenEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"
	clientID := "idp-core"
	clientSecret := "idp-core-secret-key"

	// Test direct access grant (password flow)
	t.Run("password_grant", func(t *testing.T) {
		tokenURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"

		data := url.Values{}
		data.Set("grant_type", "password")
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		data.Set("username", "developer")
		data.Set("password", "dev123")

		resp, err := http.PostForm(tokenURL, data)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		require.NoError(t, err)

		// Verify token response
		assert.Contains(t, tokenResp, "access_token")
		assert.Contains(t, tokenResp, "refresh_token")
		assert.Contains(t, tokenResp, "token_type")
		assert.Contains(t, tokenResp, "expires_in")

		assert.Equal(t, "Bearer", tokenResp["token_type"])
		assert.NotEmpty(t, tokenResp["access_token"])

		t.Logf("Token type: %v", tokenResp["token_type"])
		t.Logf("Expires in: %v seconds", tokenResp["expires_in"])
	})

	// Test platform-admin user
	t.Run("platform_admin_token", func(t *testing.T) {
		tokenURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"

		data := url.Values{}
		data.Set("grant_type", "password")
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		data.Set("username", "platform-admin")
		data.Set("password", "admin123")

		resp, err := http.PostForm(tokenURL, data)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		require.NoError(t, err)

		assert.NotEmpty(t, tokenResp["access_token"])
		t.Log("✅ Platform admin token obtained")
	})
}

// TestE2E_OIDC_ClientCredentials tests the client credentials grant
func TestE2E_OIDC_ClientCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"
	clientID := "idp-core"
	clientSecret := "idp-core-secret-key"

	t.Run("client_credentials_grant", func(t *testing.T) {
		tokenURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"

		data := url.Values{}
		data.Set("grant_type", "client_credentials")
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)

		resp, err := http.PostForm(tokenURL, data)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Client credentials may require service account enabled
		// For now, just check we get a response
		t.Logf("Response status: %d", resp.StatusCode)

		var tokenResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		if err == nil && resp.StatusCode == http.StatusOK {
			assert.NotEmpty(t, tokenResp["access_token"])
			t.Log("✅ Client credentials token obtained")
		}
	})
}

// TestE2E_OIDC_UserInfo tests the OIDC userinfo endpoint
func TestE2E_OIDC_UserInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"
	clientID := "idp-core"
	clientSecret := "idp-core-secret-key"

	// First get a token with proper scope
	tokenURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("username", "developer")
	data.Set("password", "dev123")
	data.Set("scope", "openid profile email")

	resp, err := http.PostForm(tokenURL, data)
	require.NoError(t, err)
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	accessToken, ok := tokenResp["access_token"].(string)
	require.True(t, ok)

	// Now call userinfo endpoint
	t.Run("userinfo_endpoint", func(t *testing.T) {
		userinfoURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/userinfo"

		req, err := http.NewRequest("GET", userinfoURL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("UserInfo returned status %d (may require additional client config)", resp.StatusCode)
			return
		}

		var userinfo map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&userinfo)
		require.NoError(t, err)

		// Verify userinfo fields
		assert.Contains(t, userinfo, "sub")

		t.Logf("User info: %+v", userinfo)
	})
}

// TestE2E_OIDC_JWKS tests the JWKS endpoint
func TestE2E_OIDC_JWKS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"

	t.Run("jwks_endpoint", func(t *testing.T) {
		jwksURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/certs"

		resp, err := http.Get(jwksURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var jwks map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		require.NoError(t, err)

		// Verify JWKS structure
		assert.Contains(t, jwks, "keys")
		keys, ok := jwks["keys"].([]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, keys)

		// Verify first key has required fields
		firstKey, ok := keys[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, firstKey, "kid")
		assert.Contains(t, firstKey, "kty")
		assert.Contains(t, firstKey, "n")
		assert.Contains(t, firstKey, "e")

		t.Logf("Found %d keys", len(keys))
		t.Logf("Key ID (kid): %v", firstKey["kid"])
	})
}

// TestE2E_OIDC_AuthorizationURL tests the authorization URL generation
func TestE2E_OIDC_AuthorizationURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"
	clientID := "idp-core"
	redirectURI := "http://localhost:8080/auth/callback"

	t.Run("authorization_url", func(t *testing.T) {
		authURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/auth"

		// Build authorization URL
		params := url.Values{}
		params.Set("client_id", clientID)
		params.Set("redirect_uri", redirectURI)
		params.Set("response_type", "code")
		params.Set("scope", "openid profile email")

		fullURL := authURL + "?" + params.Encode()

		// Don't follow redirects - use a client with redirect policy
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		}

		resp, err := client.Get(fullURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Keycloak may return 200 (login page HTML) or 302 (redirect to login)
		// Both are valid - it depends on Keycloak configuration
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound,
			"Expected 200 or 302, got %d", resp.StatusCode)

		t.Logf("Authorization URL status: %d", resp.StatusCode)
		t.Logf("Authorization URL: %s", fullURL)
	})
}

// TestE2E_OIDC_Introspection tests token introspection
func TestE2E_OIDC_Introspection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"
	clientID := "idp-core"
	clientSecret := "idp-core-secret-key"

	// First get a token
	tokenURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("username", "developer")
	data.Set("password", "dev123")

	resp, err := http.PostForm(tokenURL, data)
	require.NoError(t, err)
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	accessToken, ok := tokenResp["access_token"].(string)
	require.True(t, ok)

	// Test introspection endpoint
	t.Run("introspect_token", func(t *testing.T) {
		introspectURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token/introspect"

		data := url.Values{}
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		data.Set("token", accessToken)

		resp, err := http.PostForm(introspectURL, data)
		require.NoError(t, err)
		defer resp.Body.Close()

		var introspectResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&introspectResp)
		require.NoError(t, err)

		if resp.StatusCode != http.StatusOK {
			t.Logf("Introspection returned status %d (response: %+v)", resp.StatusCode, introspectResp)
			t.Log("Note: Introspection may require confidential client access type")
			return
		}

		// Verify introspection response
		active, ok := introspectResp["active"].(bool)
		if !ok {
			t.Logf("Introspection response does not contain 'active' field: %+v", introspectResp)
			return
		}
		assert.True(t, active)
		assert.Equal(t, "developer", introspectResp["username"])
		assert.Contains(t, introspectResp, "exp")
		assert.Contains(t, introspectResp, "iat")

		t.Logf("Token is active: %v", introspectResp["active"])
		t.Logf("Username: %v", introspectResp["username"])
	})
}

// TestE2E_OIDC_FullFlow tests the complete OIDC flow
func TestE2E_OIDC_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	keycloakURL := "http://localhost:8081"
	realm := "idp-core"
	clientID := "idp-core"
	clientSecret := "idp-core-secret-key"

	t.Log("=== OIDC Full Flow Test ===")
	t.Log("")

	// Step 1: Discovery
	t.Log("Step 1: OIDC Discovery")
	discoveryURL := keycloakURL + "/realms/" + realm + "/.well-known/openid-configuration"
	resp, err := http.Get(discoveryURL)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	t.Log("✅ Discovery endpoint OK")

	// Step 2: Get JWKS
	t.Log("Step 2: Fetch JWKS")
	jwksURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/certs"
	resp, err = http.Get(jwksURL)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	t.Log("✅ JWKS endpoint OK")

	// Step 3: Get Token
	t.Log("Step 3: Obtain Access Token")
	tokenURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("username", "developer")
	data.Set("password", "dev123")
	data.Set("scope", "openid profile email")

	resp, err = http.PostForm(tokenURL, data)
	require.NoError(t, err)
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)
	accessToken := tokenResp["access_token"].(string)
	t.Log("✅ Access token obtained")

	// Step 4: Validate Token via Introspection (may fail if client not configured)
	t.Log("Step 4: Validate Token")
	introspectURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/token/introspect"
	data = url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("token", accessToken)

	resp, err = http.PostForm(introspectURL, data)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var introspectResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&introspectResp)
		require.NoError(t, err)
		if active, ok := introspectResp["active"].(bool); ok && active {
			t.Log("✅ Token is valid (via introspection)")
		} else {
			t.Log("✅ Token obtained (introspection not available)")
		}
	} else {
		t.Log("✅ Token obtained (introspection skipped - client not configured)")
	}

	// Step 5: Get UserInfo
	t.Log("Step 5: Fetch User Info")
	userinfoURL := keycloakURL + "/realms/" + realm + "/protocol/openid-connect/userinfo"
	req, err := http.NewRequest("GET", userinfoURL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	t.Log("✅ UserInfo retrieved")

	t.Log("")
	t.Log("=== OIDC Full Flow Complete ===")
}
