package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Client represents an OIDC client
type Client struct {
	provider   *oidc.Provider
	verifier   *oidc.IDTokenVerifier
	oauth2Conf *oauth2.Config
	config     *Config
}

// Config holds OIDC client configuration
type Config struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// NewClient creates a new OIDC client
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg.IssuerURL == "" {
		return nil, fmt.Errorf("issuer URL is required")
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	oauth2Conf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	return &Client{
		provider:   provider,
		verifier:   verifier,
		oauth2Conf: oauth2Conf,
		config:     cfg,
	}, nil
}

// GetAuthURL returns the OAuth2 authorization URL
func (c *Client) GetAuthURL(state string) string {
	return c.oauth2Conf.AuthCodeURL(state)
}

// Exchange exchanges an authorization code for tokens
func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.oauth2Conf.Exchange(ctx, code)
}

// VerifyIDToken verifies an ID token and returns the claims
func (c *Client) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return c.verifier.Verify(ctx, rawIDToken)
}

// GetVerifier returns the ID token verifier
func (c *Client) GetVerifier() *oidc.IDTokenVerifier {
	return c.verifier
}

// GetOAuth2Config returns the OAuth2 configuration
func (c *Client) GetOAuth2Config() *oauth2.Config {
	return c.oauth2Conf
}
