# Implementation Plan: OIDC Issuer Consistency Fix

**Branch**: `[001-phase-3-platform]` | **Date**: 2026-06-05 | **Spec**: [`specs/001-phase-3-platform/spec.md`](spec.md)

**Input**: Feature specification from `/specs/001-phase-3-platform/spec.md` and planning input describing an OIDC login bug where follow-up API calls fail with `invalid_grant` due to token issuer mismatch.

## Summary

Stabilize the authentication foundation for the Phase 3 platform by making OIDC issuer handling consistent across login, token refresh, and authenticated API access. The implementation will align provider discovery, refresh-token exchange, and token verification around one canonical issuer configuration, remove the current behavior that masks issuer mismatches during login, and add validation/tests so internal-only Keycloak URLs cannot silently diverge from the browser-visible issuer used by clients.

## Technical Context

**Language/Version**: Go 1.25+

**Primary Dependencies**: Gin, `github.com/coreos/go-oidc/v3/oidc`, `golang.org/x/oauth2`, Viper-based config loading

**Storage**: PostgreSQL for local user persistence; OIDC provider metadata and tokens are external

**Testing**: `go test ./...` with focused unit tests in `internal/pkg/oidc` and handler/middleware tests for auth flows

**Target Platform**: Linux server / containerized backend API

**Project Type**: Web service

**Performance Goals**: Preserve current login and authenticated request latency; token verification and refresh should not add additional network round-trips beyond normal provider discovery and token exchange

**Constraints**:
- Use the existing clean architecture and dependency injection pattern from `CLAUDE.md`
- Keep OIDC login, callback, refresh, and middleware behavior compatible with the existing frontend flow
- Do not rely on environment-specific issuer mismatches being accepted implicitly
- Return clear authentication failures when configured issuer data is inconsistent

**Scale/Scope**: One backend auth subsystem serving all signed-in platform users; immediate scope is OIDC login, token refresh, and bearer-token validation paths

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution file at `.specify/memory/constitution.md` is still a placeholder template and does not define enforceable MUST/SHOULD rules. No project-specific constitutional gates can be evaluated yet.

## Project Structure

### Documentation (this feature)

```text
specs/001-phase-3-platform/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── oidc-auth.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/http/
├── main.go
└── server.go

configs/
└── config.development.yaml

internal/
├── handler/http/
│   ├── init.go
│   ├── oidc_auth.go
│   └── middleware/oidc.go
└── pkg/
    ├── config/config.go
    └── oidc/
        ├── client.go
        ├── client_test.go
        ├── verifier.go
        └── verifier_test.go
```

**Structure Decision**: This work stays inside the existing backend API. The fix is concentrated in configuration loading, OIDC client/verifier construction, auth handlers, middleware, and associated tests.

## Phase 0: Research

- Confirm the runtime source of truth for OIDC issuer configuration across YAML and environment overrides
- Document how login currently succeeds even when refresh later fails
- Decide whether the canonical issuer should be the browser-visible Keycloak URL and how to fail fast if it is not consistent
- Verify whether logout URL derivation must follow the same canonical issuer rule

## Phase 1: Design & Contracts

- Model the auth/session objects affected by issuer consistency: provider config, ID token claims, refresh token lifecycle, local JWT session
- Define backend auth contract expectations for login, callback, refresh, and bearer-token validation when issuer configuration is valid vs invalid
- Provide an operator quickstart for validating the corrected flow in development
- Update agent context to reference this implementation plan

## Phase 2: Implementation Preview

- Introduce a single canonical issuer configuration strategy for OIDC provider discovery, refresh, and verification
- Remove or constrain issuer-check skipping so login does not mask invalid configuration
- Add startup/config validation and tests covering mismatched issuer scenarios
- Verify protected API requests succeed after login and after token refresh when configuration is correct

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
