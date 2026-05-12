package oidc

import (
	"context"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestNewClient_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing issuer URL",
			config:  &Config{ClientID: "test-client"},
			wantErr: true,
			errMsg:  "issuer URL is required",
		},
		{
			name:    "missing client ID",
			config:  &Config{IssuerURL: "https://example.com"},
			wantErr: true,
			errMsg:  "client ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_GetOAuth2Config(t *testing.T) {
	cfg := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/callback",
		Scopes:       []string{"openid", "profile"},
	}

	client := &Client{
		oauth2Conf: cfg,
	}

	result := client.GetOAuth2Config()
	assert.NotNil(t, result)
	assert.Equal(t, "test-client", result.ClientID)
	assert.Equal(t, "test-secret", result.ClientSecret)
	assert.Equal(t, "http://localhost/callback", result.RedirectURL)
	assert.Equal(t, []string{"openid", "profile"}, result.Scopes)
}

func TestClient_GetVerifier(t *testing.T) {
	client := &Client{
		verifier: oidc.NewVerifier("https://example.com", nil, &oidc.Config{ClientID: "test"}),
	}

	result := client.GetVerifier()
	assert.NotNil(t, result)
}

func TestClient_GetAuthURL(t *testing.T) {
	cfg := &oauth2.Config{
		ClientID:    "test-client",
		RedirectURL: "http://localhost/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL: "https://example.com/auth",
		},
		Scopes: []string{"openid"},
	}

	client := &Client{
		oauth2Conf: cfg,
	}

	url := client.GetAuthURL("test-state")
	assert.Contains(t, url, "https://example.com/auth")
	assert.Contains(t, url, "client_id=test-client")
	assert.Contains(t, url, "state=test-state")
	assert.Contains(t, url, "redirect_uri=http%3A%2F%2Flocalhost%2Fcallback")
}
