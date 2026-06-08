# Contract: OIDC Authentication Consistency

## Scope

This contract defines the expected backend behavior for OIDC login, callback, refresh, and bearer-token validation after the issuer consistency fix.

## Configuration Contract

When OIDC authentication is enabled:

- The backend must use one canonical issuer URL for provider discovery, token verification, refresh-token exchange, and logout URL derivation.
- If the configured issuer cannot produce verifiable tokens for the configured client, the backend must fail the auth flow rather than allowing login to appear successful.

## Endpoint Behavior Contract

### `GET /auth/oidc/login`

- Returns or redirects to an authorization URL derived from the canonical issuer's discovered authorization endpoint.
- Persists login state and the validated frontend redirect destination.

### `GET /auth/oidc/callback`

- Exchanges the authorization code against the token endpoint derived from the canonical issuer.
- Requires an ID token in the token response.
- Verifies the ID token against the configured client and canonical issuer.
- Creates or reuses the local user and issues the local session JWT only after successful verification.
- Fails the request if issuer/client verification fails.

### `POST /auth/oidc/refresh`

- Accepts a refresh token from request body or cookie.
- Exchanges the refresh token against the token endpoint derived from the same canonical issuer.
- Returns `401 Unauthorized` when the refresh token is invalid for that issuer or client.
- Reissues the local session JWT only when a new ID token can be successfully verified.

### Protected API routes using OIDC bearer tokens

- Must validate bearer tokens using the same canonical issuer and client configuration used by login and refresh.
- Must reject tokens whose issuer does not match the configured canonical issuer.

## Error Contract

- Missing refresh token → `400 Bad Request`
- Invalid or mismatched issuer/client token → `401 Unauthorized`
- Misconfigured provider/client setup surfaced during login/callback/refresh → explicit auth failure, not silent fallback

## Operational Contract

- Development, container, and local-browser environments must agree on the externally visible issuer used for Keycloak/OIDC flows.
- Environment overrides must not silently point the backend to a different issuer than the one minting browser-obtained refresh tokens.
