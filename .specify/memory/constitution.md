<!--
Sync Impact Report
Version change: 1.0.0 → 1.0.1
Modified principles:
- I. Self-Service Platform Value → clarified that Phase 1 MVP and Phase 2 Enhancement are completed baseline capabilities
- III. Security and Tenant Isolation → clarified that shipped auth, RBAC, API keys, and audit logging are inherited constraints
- V. Incremental Phase Delivery → clarified that active Spec Kit work starts from the completed Phase 1/2 baseline and targets the next roadmap phase
Added sections:
- None
Removed sections:
- None
Templates requiring updates:
- ✅ updated .specify/templates/plan-template.md
- ✅ updated .specify/templates/spec-template.md
- ✅ updated .specify/templates/tasks-template.md
- ⚠ pending .specify/templates/commands/*.md (directory not present in this repository)
Follow-up TODOs:
- None
-->
# idp-core Constitution

## Core Principles

### I. Self-Service Platform Value
Every feature MUST strengthen the platform promise described in the PRD: engineering teams can
provision, manage, and observe environments without direct cluster access. Phase 1 MVP and Phase 2
Enhancement capabilities are the established product baseline, so new specifications, plans, and
implementation tasks MUST build on those completed API, auth, FinOps, rightsizing, and service
catalog workflows rather than redefining them. Features that add internal complexity without clear
self-service or governance value MUST be rejected.

Rationale: idp-core exists to reduce environment delivery time while abstracting Kubernetes and
GitOps complexity behind an auditable API, and the completed first two phases already define the
baseline platform behaviors new work inherits.

### II. Clean Architecture Boundaries
All implementation MUST preserve the repository's dependency flow: handler → usecase →
repository → model. Handlers MUST own transport concerns only, usecases MUST own business rules
and orchestration, repositories MUST own persistence and external integrations, and models MUST
remain domain-focused. New features MUST follow the established directory pattern under
`internal/model`, `internal/repository`, `internal/usecase`, and `internal/handler/http`, with
wiring performed in the application entry points.

Rationale: the system already depends on strict separation to keep Kubernetes, ArgoCD, auth, and
platform logic testable and maintainable as Phase 3 expands backend scope.

### III. Security and Tenant Isolation
All feature work MUST enforce existing authentication, authorization, and team-scoped access
rules by default. Because Phase 2 is complete, OIDC, JWT/API key auth, RBAC, and audit logging are
mandatory inherited constraints for all new APIs, background jobs, live updates, seed data, and
operational workflows. APIs, background jobs, live updates, and seed data MUST preserve tenant
boundaries, avoid privilege escalation, and retain audit-relevant records for security-sensitive
operations. Secrets, tokens, and credentials MUST never be logged or embedded in generated
artifacts.

Rationale: Phase 2 established RBAC, OIDC, API keys, and audit logging as foundational platform
capabilities, so Phase 3 work must extend those controls rather than bypass them.

### IV. Contracted and Observable Operations
Externally visible behavior MUST be explicit and testable through stable contracts. Every feature
that changes API behavior, long-running operations, or live status delivery MUST define request
and response expectations, failure modes, and operator-visible signals. Plans and tasks MUST
include validation for correctness, operational visibility, and historical traceability whenever a
feature affects provisioning, sync, movement, or notifications.

Rationale: idp-core is an API product that coordinates Kubernetes and GitOps workflows; operators
need predictable contracts and timely visibility to trust automation.

### V. Incremental Phase Delivery
Work MUST be sliced into independently valuable increments aligned to prioritized user stories.
Phase 1 MVP and Phase 2 Enhancement are complete, so active Spec Kit planning MUST treat those
phases as delivered baseline behavior and focus new specifications on the next roadmap increment
unless an explicit maintenance task says otherwise. The first increment in any new feature series
MUST deliver a usable backend capability on its own, and later increments MUST build on the same
models and contracts without invalidating earlier behavior. Plans and tasks MUST show how each
story can be validated independently before broader rollout.

Rationale: the roadmap is phased, and Spec Kit is used to structure delivery so Phase 3 can ship
backend foundations before UI or future advanced capabilities.

## Engineering Standards

- Specs MUST stay focused on user value, business rules, access boundaries, and measurable
  outcomes; they MUST NOT prescribe implementation details unless the constitution explicitly
  requires a repo-specific constraint.
- Plans MUST document the real repository paths they will touch, the affected contracts or data
  models, and the validation steps needed to prove the feature works end to end.
- Tasks MUST reference exact files, keep foundational work separate from story delivery, and
  include testing or validation work whenever contracts, authorization, or long-running operations
  change.
- Changes that affect audits, notifications, template history, cluster targeting, or environment
  lifecycle state MUST preserve historical records rather than overwrite prior state.
- Documentation and generated guidance MUST stay consistent with the active PRD phase and current
  repository structure.

## Delivery Workflow

1. Constitution compliance MUST be checked before research begins and again after design is
   drafted.
2. Specifications MUST describe prioritized user stories, acceptance scenarios, edge cases,
   functional requirements, key entities, success criteria, and assumptions.
3. Implementation plans MUST resolve technical unknowns, document contracts and data models, and
   map the work onto the actual Go service structure used by this repository.
4. Task lists MUST organize work by user story so each increment can be implemented, tested, and
   demonstrated independently.
5. Before completion, changes affecting APIs or UI-visible operational behavior MUST be validated
   with appropriate tests and, when relevant, runnable flows that exercise the user-facing path.

## Governance

This constitution is the authoritative source for delivery standards in this repository.
Conflicting instructions in ad hoc notes, templates, or generated artifacts MUST be brought into
alignment with this document.

Amendments MUST update this constitution, include a short impact summary for affected templates or
workflow documents, and use semantic versioning for governance changes:
- MAJOR: removes or redefines a principle or governance rule in a backward-incompatible way.
- MINOR: adds a principle, section, or materially stronger requirement.
- PATCH: clarifies wording without changing the underlying obligations.

Every specification, implementation plan, task list, and code review MUST verify constitution
compliance. If a necessary change cannot satisfy a principle, the deviation MUST be documented in
Complexity Tracking or an equivalent justification section before implementation proceeds.

**Version**: 1.0.1 | **Ratified**: 2026-06-08 | **Last Amended**: 2026-06-08
