# Research: OIDC Issuer Consistency Fix

## Decision 1: Use one canonical issuer URL for provider discovery, refresh, logout, and token verification

- **Decision**: Treat the configured OIDC issuer URL as the single canonical issuer for every OIDC interaction in the backend.
- **Rationale**: `internal/pkg/oidc/client.go` builds provider discovery, verifier, and OAuth2 endpoints from one issuer value. `internal/handler/http/oidc_auth.go` uses the same client for authorization-code exchange and refresh, and `cmd/http/server.go` derives the logout endpoint from the configured issuer. Using different internal and external hostnames causes refresh-token exchange to fail with `invalid_grant` because the refresh token was minted for a different issuer than the token endpoint the backend later calls.
- **Alternatives considered**:
  - Keep separate internal and external issuer hosts. Rejected because Keycloak ties refresh tokens to the issuer/token endpoint that minted them, making the flow brittle and environment-specific.
  - Continue allowing mismatches only in development. Rejected because the current login flow already hides broken configuration that later fails on protected API access.

## Decision 2: Stop masking issuer mismatch during login

- **Decision**: Remove the current reliance on `SkipIssuerCheck: true` as a compatibility crutch and make issuer consistency visible during login/callback validation.
- **Rationale**: The current verifier in `internal/pkg/oidc/client.go` skips issuer validation, which allows the callback to succeed even when refresh and later API calls fail. That creates a false-success login experience and shifts the failure to a later request where the user receives a 401.
- **Alternatives considered**:
  - Keep `SkipIssuerCheck: true` indefinitely. Rejected because it preserves the exact bug being addressed and weakens token validation guarantees.
  - Skip ID token verification entirely and trust only refresh/access token usage. Rejected because it would reduce security and make user extraction less reliable.

## Decision 3: Add startup and test coverage for issuer-related configuration drift

- **Decision**: Add validation/tests around OIDC configuration so issuer mismatches are caught before or during controlled validation rather than by end users.
- **Rationale**: The relevant failure mode is configuration drift between `configs/config.development.yaml`, environment overrides in `internal/pkg/config/config.go`, and the URL actually used by the browser-visible IdP. Tests should cover verifier construction, refresh behavior assumptions, and auth middleware/handler expectations for a correctly configured issuer.
- **Alternatives considered**:
  - Rely on manual QA only. Rejected because this bug specifically appears after a successful login and is easy to miss without targeted checks.
  - Document the expected issuer without tests. Rejected because the risk is runtime misconfiguration, which documentation alone does not prevent.

## Decision 4: Keep the fix scoped to backend auth contracts and validation

- **Decision**: Limit this change to backend auth configuration, token handling, and validation behavior, without redesigning the wider Phase 3 platform scope.
- **Rationale**: The reported failure is in the backend OIDC subsystem, specifically the interaction among provider discovery, callback verification, refresh exchange, and bearer-token validation. The fastest safe fix is to make these paths consistent and testable.
- **Alternatives considered**:
  - Expand scope into a broader frontend auth redesign. Rejected because the current issue is reproducible and fixable within the backend contract.
  - Introduce a custom translation layer between internal and external issuers. Rejected because it adds avoidable complexity and increases operational risk.
