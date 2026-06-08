# Data Model: OIDC Issuer Consistency Fix

## Entity: OIDC Provider Configuration

- **Purpose**: Defines the canonical identity provider settings used by the backend for authorization-code exchange, refresh-token exchange, token verification, and logout.
- **Fields**:
  - `enabled` (boolean): Whether OIDC auth is active
  - `issuer_url` (string): Canonical issuer URL used for discovery and token validation
  - `client_id` (string): OAuth client identifier
  - `client_secret` (string): OAuth client secret
  - `redirect_url` (string): Backend callback URL registered with the provider
  - `scopes` ([]string): Requested OAuth scopes
  - `groups_claim` (string): Claim used to extract group membership
  - `admin_group` (string): Group name that grants platform admin access
- **Validation Rules**:
  - `issuer_url` must be present when OIDC is enabled
  - `client_id` must be present when OIDC is enabled
  - The configured issuer must match the issuer that signs returned tokens
  - `redirect_url` must correspond to the callback endpoint expected by the provider

## Entity: OIDC Token Set

- **Purpose**: Represents the provider-issued tokens returned during login and refresh.
- **Fields**:
  - `id_token` (string): Token used to verify identity claims
  - `access_token` (string): Provider bearer token accepted by OIDC-protected API paths
  - `refresh_token` (string): Token used to obtain a new token set
  - `expires_at` / `expires_in` (timestamp/duration): Token lifetime metadata
  - `token_type` (string): Usually `Bearer`
- **Validation Rules**:
  - `id_token` must be verifiable against the canonical issuer and configured client ID
  - `refresh_token` must only be exchanged against the token endpoint derived from the same canonical issuer
  - Missing refresh tokens are handled as an explicit client error in refresh requests
- **State Transitions**:
  - `issued` → `verified` during callback
  - `verified` → `refreshed` during refresh
  - `verified`/`refreshed` → `rejected` when issuer or client validation fails

## Entity: Authenticated User Claims

- **Purpose**: Identity and authorization data extracted from the verified ID token.
- **Fields**:
  - `sub` (string): Provider subject identifier
  - `email` (string): User email
  - `email_verified` (boolean): Email verification status
  - `name` (string): Display name
  - `given_name` (string): Optional first name
  - `family_name` (string): Optional last name
  - `groups` ([]string): Membership values from the configured groups claim
  - `picture` (string): Optional avatar URL
- **Validation Rules**:
  - `sub` must exist on every verified token
  - `email` is required before upserting a local user
  - `groups` extraction depends on the configured claim name

## Entity: Local Session JWT

- **Purpose**: Backend-issued session token used by frontend/API consumers after a successful OIDC callback or refresh.
- **Fields**:
  - `user_id` (string): Local user identifier
  - `email` (string): Local session email
  - `is_admin` (boolean): Authorization shortcut derived from OIDC groups
  - `exp` (timestamp): Session expiry
- **Validation Rules**:
  - Issued only after provider identity is verified and the local user is found or created
  - Re-issued after a successful refresh when a new ID token can be verified

## Relationships

- One **OIDC Provider Configuration** governs all **OIDC Token Set** exchanges and verifications.
- One verified **OIDC Token Set** yields one **Authenticated User Claims** object.
- One **Authenticated User Claims** object is mapped to one local user record and one **Local Session JWT**.
