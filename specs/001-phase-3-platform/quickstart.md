# Quickstart: Validate OIDC Issuer Consistency Fix

## Goal

Verify that login, protected API access, and token refresh all succeed when the backend uses one canonical OIDC issuer, and that misconfiguration fails early with a clear auth error.

## Prerequisites

- Keycloak realm `idp-core` is running and reachable from both the backend container and the browser through the same canonical issuer URL
- Backend configuration points `oidc.issuer_url` (or `OIDC_ISSUER_URL`) to that canonical issuer
- OIDC client redirect URI matches the backend callback endpoint
- Test user exists in Keycloak and has any required groups/roles

## Reference Artifacts

- Plan: `specs/001-phase-3-platform/plan.md`
- Research: `specs/001-phase-3-platform/research.md`
- Data model: `specs/001-phase-3-platform/data-model.md`
- Contract: `specs/001-phase-3-platform/contracts/oidc-auth.md`

## Setup

1. Confirm the configured issuer value in runtime config and environment overrides.
2. Start the backend API.
3. Ensure the frontend or API client uses the same login flow as production/development usage.

## Validation Scenario 1: Fresh Login

1. Open the OIDC login flow.
2. Authenticate with Keycloak.
3. Confirm the callback succeeds.
4. Confirm the backend issues the local auth cookie/JWT and, if applicable, stores the refresh token cookie.

**Expected outcome**:
- Login completes without a delayed issuer-related failure.
- No `invalid_grant` or issuer mismatch error appears immediately after callback.

## Validation Scenario 2: Protected API After Login

1. After successful login, call a protected backend API using the issued credentials.
2. Repeat against at least one endpoint that uses the OIDC auth middleware.

**Expected outcome**:
- The API returns a success response for an authorized user.
- The backend does not reject the token with issuer mismatch.

## Validation Scenario 3: Refresh Flow

1. Trigger `POST /auth/oidc/refresh` using the stored refresh token or let the frontend invoke its normal refresh behavior.
2. Use the newly issued credentials to call a protected API again.

**Expected outcome**:
- Refresh returns success and new tokens when the issuer is configured correctly.
- Subsequent protected API calls also succeed.

## Validation Scenario 4: Misconfigured Issuer Guardrail

1. Intentionally point the backend to an issuer that differs from the one used by the browser-visible Keycloak host.
2. Attempt the login and refresh flow.

**Expected outcome**:
- The auth flow fails clearly during validation or callback/refresh.
- The system does not allow a misleading “login succeeded” state followed by a later unexplained 401.

## Test Commands

```bash
go test ./internal/pkg/oidc/...
go test ./internal/handler/http/...
go test ./...
```

## Success Criteria

- Login, refresh, and protected API access all work with one canonical issuer configuration.
- Issuer mismatch is surfaced as a controlled auth/configuration failure.
- Automated tests cover the issuer validation path and refresh behavior assumptions.
