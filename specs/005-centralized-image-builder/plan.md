# Implementation Plan: Centralized Image Builder

**Branch**: `[005-centralized-image-builder]` | **Date**: 2026-06-26 | **Spec**: [`specs/005-centralized-image-builder/spec.md`](spec.md)

**Input**: Feature specification from `/specs/005-centralized-image-builder/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add a team-scoped Centralized Image Builder capability that lets platform users register source-driven applications, trigger and observe asynchronous image builds, enforce required post-build security verification, and update GitOps deployment sources without requiring application teams to maintain Dockerfiles.

## Technical Context

**Language/Version**: Go 1.25+

**Primary Dependencies**: Gin, GORM, PostgreSQL, existing JWT/OIDC middleware, existing audit logging patterns, Kubernetes client-go/controller-runtime integration patterns already used in the repository, swaggo/OpenAPI

**Storage**: PostgreSQL for build applications, build attempts, artifacts, security verification records, deployment update records, and lifecycle events; external build/deployment control planes for execution state

**Testing**: `go test ./...`, focused usecase and handler tests for application/build APIs, repository tests for persistence and event ordering, and runnable quickstart validation for async build, security, and deployment flows

**Target Platform**: Linux server / containerized backend API with PostgreSQL plus external Kubernetes-oriented build and deployment control planes

**Project Type**: Web service

**Performance Goals**:
- Build trigger, retry, and cancel APIs acknowledge accepted work without blocking on build completion
- Build status, history, and application detail views remain responsive enough for operator use during concurrent build activity
- Idempotent build action handling prevents contradictory lifecycle outcomes for 100% of validated duplicate request cases
- Required security verification and deployment-update outcomes remain attached to 100% of completed builds

**Constraints**:
- Preserve handler → usecase → repository → model boundaries and existing dependency injection patterns
- Reuse existing authentication, authorization, team-scope, and audit-log behavior
- Keep secrets and raw external credentials out of API models, logs, and persisted user-visible payloads
- Preserve immutable historical records for builds, security verification, deployment updates, and lifecycle events
- Support horizontally scalable, restart-safe async orchestration without allowing two workers to own the same build transition simultaneously
- Keep the initial slice backend-focused and GitOps-first rather than introducing direct workload deployment paths

**Scale/Scope**: Initial release covers team-scoped application registration, asynchronous build lifecycle management, build log streaming, supported runtime families (Go, Java, Node, Python, .NET), approved builder/profile selection, approved registry publication, mandatory SBOM/scan/sign verification, and GitOps deployment-source updates for multiple concurrent teams

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Self-service value is explicit: platform users can register and build applications without Dockerfiles, while operators gain centralized governance, security verification, and deployment-source traceability.
- [x] Clean architecture is preserved: planned changes introduce a dedicated build domain across `internal/model`, `internal/repository`, `internal/usecase`, and `internal/handler/http`, with dependency wiring in `cmd/http/main.go` and routes in `cmd/http/server.go`; external build/deployment interactions stay behind repository/provider interfaces.
- [x] Security and tenant isolation are covered: all application/build operations remain team-scoped, existing authn/authz is reused, security evidence is exposed safely, and raw external credentials stay in centrally governed references instead of user payloads.
- [x] Contracts and observability are covered: the plan defines public API contracts, async lifecycle visibility, log streaming expectations, lifecycle events, security outcomes, deployment-update tracking, and quickstart validation signals.
- [x] Delivery stays incremental: the implementation preview is sliced into independently valuable stories for application registration, build execution/observation, and policy-governed security/deployment behavior.

## Project Structure

### Documentation (this feature)

```text
specs/005-centralized-image-builder/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── centralized-image-builder.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── http/
│   ├── main.go
│   └── server.go
└── cron/

internal/
├── handler/
│   └── http/
│       ├── handler.go
│       └── build_application.go
├── usecase/
│   └── build_application/
├── repository/
│   ├── build_application/
│   └── gitops/
├── model/
│   └── build_application/
├── mocks/
└── pkg/
    └── config/

migrations/
docs/swagger/
specs/005-centralized-image-builder/
```

**Structure Decision**: Introduce a new dedicated build domain under `internal/model/build_application`, `internal/repository/build_application`, `internal/usecase/build_application`, and `internal/handler/http/build_application.go` rather than overloading the existing service catalog domain. Keep transport concerns in handlers, orchestration and state-machine logic in usecases, persistence and external control-plane interactions in repositories/providers, and shared response/request shapes in models. Update `cmd/http/main.go` and `cmd/http/server.go` to wire the new dependencies and routes without disturbing existing service/environment flows.

## Phase 0: Research

1. Confirm the best fit between a new build domain and the existing service catalog so the feature can reuse team/auth/audit patterns without collapsing into the service catalog model.
2. Define the minimum application descriptor contract and which application metadata must be persisted versus derived from source or external systems.
3. Define the canonical async lifecycle and idempotency model for trigger, retry, cancel, and restart-safe processing.
4. Define the separation between build success, security verification success, and deployment-ready success.
5. Identify the public API surface needed for application CRUD, build actions, build detail/history, and log streaming.
6. Identify the external execution/state resources that must be represented in the design without leaking implementation details into the spec.

## Phase 1: Design & Contracts

1. Define the persistent data model for build applications, builder profiles, registry targets, build attempts, artifacts, security verification, deployment updates, and lifecycle events.
2. Define public API contracts for application CRUD, build trigger/retry/cancel, build detail/history, and build log streaming.
3. Define an event model and explicit state machines for application, build, security verification, and deployment-update lifecycles.
4. Define validation scenarios in `quickstart.md` that cover successful build flow, retry/cancel behavior, security-gated deployment blocking, deployment-source updates, idempotency, and team-isolation failure cases.
5. Update `CLAUDE.md` to point to this plan for future implementation context.

## Phase 2: Implementation Preview

### User Story 1 - Register and manage buildable applications (P1)
- Add team-scoped CRUD APIs and persistence for build applications with repository, descriptor, builder-profile, registry-target, and deployment automation metadata.
- Reuse existing auth, audit, and response patterns so application management fits the repository’s established platform workflows.
- Preserve immutable history when applications are updated or deleted.

### User Story 2 - Trigger and observe asynchronous builds (P2)
- Add build trigger, build detail, build history, retry, cancel, and log streaming APIs.
- Implement canonical build state tracking, idempotency handling, lifecycle events, and worker-safe ownership rules in the build usecase/repository layers.
- Expose ordered build progress and terminal outcomes without coupling execution to request lifetimes.

### User Story 3 - Enforce security verification and GitOps-first deployment readiness (P3)
- Add artifact publication, SBOM/scan/sign verification tracking, and policy-gated deployment-update orchestration as post-build phases.
- Prevent deployment-source updates for non-compliant artifacts and make blocked/deployment-failed outcomes explicit in the build contract.
- Attach deployment update results and lifecycle events to the same canonical build record for traceability.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
