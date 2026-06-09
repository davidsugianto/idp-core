package oidc

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
	IssuerURL          string
	DiscoveryURL       string
	ClientID           string
	ClientSecret       string
	RedirectURL        string
	Scopes             []string
	InsecureIssuerURLs []string
}

// NewClient creates a new OIDC client
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg.IssuerURL == "" {
		return nil, fmt.Errorf("issuer URL is required")
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	discoveryURL := cfg.DiscoveryURL
	if discoveryURL == "" {
		discoveryURL = cfg.IssuerURL
	}

	providerCtx := ctx
	if discoveryURL != cfg.IssuerURL {
		if !containsURL(cfg.InsecureIssuerURLs, cfg.IssuerURL) {
			return nil, fmt.Errorf("issuer URL %q must be listed in insecure issuer URLs when discovery URL %q differs", cfg.IssuerURL, discoveryURL)
		}
		providerCtx = oidc.InsecureIssuerURLContext(ctx, cfg.IssuerURL)
	}

	provider, err := oidc.NewProvider(providerCtx, discoveryURL)
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

// Refresh exchanges a refresh token for a new token set
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	return c.oauth2Conf.TokenSource(ctx, token).Token()
}

func containsURL(urls []string, target string) bool {
	targetURL, err := normalizeURL(target)
	if err != nil {
		return false
	}

	for _, candidate := range urls {
		candidateURL, err := normalizeURL(candidate)
		if err != nil {
			continue
		}
		if candidateURL == targetURL {
			return true
		}
	}

	return false
}

func normalizeURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Fragment = ""

	return strings.TrimRight(parsed.String(), "/"), nil
}
