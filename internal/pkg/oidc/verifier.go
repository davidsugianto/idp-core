package oidc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

// UserInfo represents user information from OIDC provider
type UserInfo struct {
	Subject       string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	GivenName     string   `json:"given_name"`
	FamilyName    string   `json:"family_name"`
	Groups        []string `json:"groups"`
	Picture       string   `json:"picture"`
}

// Verifier provides token verification utilities
type Verifier struct {
	client       *Client
	groupsClaim  string
	adminGroup   string
}

// VerifierConfig holds verifier configuration
type VerifierConfig struct {
	GroupsClaim string
	AdminGroup  string
}

// NewVerifier creates a new verifier
func NewVerifier(client *Client, cfg *VerifierConfig) *Verifier {
	groupsClaim := "groups"
	if cfg != nil && cfg.GroupsClaim != "" {
		groupsClaim = cfg.GroupsClaim
	}

	adminGroup := ""
	if cfg != nil && cfg.AdminGroup != "" {
		adminGroup = cfg.AdminGroup
	}

	return &Verifier{
		client:      client,
		groupsClaim: groupsClaim,
		adminGroup:  adminGroup,
	}
}

// VerifyAndExtract verifies an access token and extracts user info
func (v *Verifier) VerifyAndExtract(ctx context.Context, accessToken string) (*UserInfo, error) {
	token, err := v.client.VerifyIDToken(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	return v.ExtractUserInfo(token)
}

// ExtractUserInfo extracts user information from a verified token
func (v *Verifier) ExtractUserInfo(token *oidc.IDToken) (*UserInfo, error) {
	var claims json.RawMessage
	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	var rawClaims map[string]interface{}
	if err := json.Unmarshal(claims, &rawClaims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	userInfo := &UserInfo{
		Subject: token.Subject,
	}

	// Extract email
	if email, ok := rawClaims["email"].(string); ok {
		userInfo.Email = email
	}
	if emailVerified, ok := rawClaims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}

	// Extract name
	if name, ok := rawClaims["name"].(string); ok {
		userInfo.Name = name
	}
	if givenName, ok := rawClaims["given_name"].(string); ok {
		userInfo.GivenName = givenName
	}
	if familyName, ok := rawClaims["family_name"].(string); ok {
		userInfo.FamilyName = familyName
	}

	// Extract picture
	if picture, ok := rawClaims["picture"].(string); ok {
		userInfo.Picture = picture
	}

	// Extract groups
	userInfo.Groups = v.extractGroups(rawClaims)

	return userInfo, nil
}

// extractGroups extracts groups from claims using the configured claim name
func (v *Verifier) extractGroups(claims map[string]interface{}) []string {
	groupsRaw, ok := claims[v.groupsClaim]
	if !ok {
		return nil
	}

	// Handle array of strings
	if groups, ok := groupsRaw.([]interface{}); ok {
		result := make([]string, 0, len(groups))
		for _, g := range groups {
			if group, ok := g.(string); ok {
				result = append(result, group)
			}
		}
		return result
	}

	// Handle single string
	if group, ok := groupsRaw.(string); ok {
		return []string{group}
	}

	return nil
}

// IsAdmin checks if the user is a platform admin
func (v *Verifier) IsAdmin(userInfo *UserInfo) bool {
	if v.adminGroup == "" {
		return false
	}

	for _, group := range userInfo.Groups {
		if group == v.adminGroup {
			return true
		}
	}

	return false
}

// GetGroupsClaim returns the configured groups claim name
func (v *Verifier) GetGroupsClaim() string {
	return v.groupsClaim
}

// GetAdminGroup returns the configured admin group
func (v *Verifier) GetAdminGroup() string {
	return v.adminGroup
}
