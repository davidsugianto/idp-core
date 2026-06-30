# Feature Specification: Centralized Image Builder

**Feature Branch**: `005-centralized-image-builder`

**Created**: 2026-06-26

**Status**: Draft

**Input**: User description: "Create a detailed specification for a new feature called \"Centralized Image Builder\".

Problem

Application teams currently maintain Dockerfiles.

This causes:

- inconsistent images
- duplicated Dockerfiles
- CVE patching burden
- slow onboarding
- runtime inconsistencies

Goal

Provide Buildpacks as a Service.

Developers should only provide:

- Git repository
- application.yaml

No Dockerfile is required.

The backend should orchestrate:

Developer
→ Git Repository
→ IDP Backend
→ kpack
→ Cloud Native Buildpacks
→ Registry
→ Trivy Scan
→ SBOM Generation
→ Cosign Signing
→ GitOps Repository
→ ArgoCD
→ Kubernetes

Functional Requirements

Applications

- Create
- Update
- Delete
- List
- Detail

Builds

- Trigger Build
- Retry Build
- Cancel Build
- View Build Status
- Stream Build Logs
- List Build History

Builder

Support

- Go
- Java
- Node
- Python
- .NET

Automatically detect runtime using Buildpacks.

Support custom ClusterBuilder.

Support custom Buildpacks.

Support custom Stack.

Registry

Support

- Harbor
- GHCR
- ECR
- GCR

Security

Every image must

- Generate SBOM
- Scan using Trivy
- Sign using Cosign

Deployment

GitOps-first

Update Git repository

ArgoCD deploys workloads.

Non-functional Requirements

- HA
- Horizontal scaling
- Async builds
- Retry mechanism
- Event-driven architecture
- Idempotent APIs

Output

Include

- User Stories
- Acceptance Criteria
- Database schema
- API contracts
- Event model
- State machine
- Kubernetes CRDs
- Sequence diagrams
- Mermaid diagrams
- Risks
- Future enhancements"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Register and manage buildable applications (Priority: P1)

A platform user can register an application once with its source repository, delivery settings, and build configuration file so the platform can own image creation without requiring a Dockerfile.

**Why this priority**: Without a managed application record, there is no self-service entry point for standardizing image creation or onboarding teams to the new workflow.

**Independent Test**: Can be fully tested by creating an application with a repository URL and application descriptor, retrieving it, updating its metadata, listing it among team applications, and deleting it while confirming the platform preserves audit visibility and prevents cross-team access.

**Acceptance Scenarios**:

1. **Given** an authenticated platform user with access to a team, **When** the user creates an application by submitting a source repository reference and application descriptor, **Then** the system stores a buildable application record and returns its initial configuration and lifecycle status.
2. **Given** an existing application, **When** the user requests the application detail or list views, **Then** the system returns the stored source, deployment, security, and ownership metadata needed to understand how the application will be built and deployed.
3. **Given** an existing application, **When** the user updates supported mutable settings such as source revision policy, builder preference, registry destination, or deployment target, **Then** the system records the change, preserves historical traceability, and makes the new settings available for future builds.
4. **Given** an application that should no longer be managed, **When** the user deletes it, **Then** the system stops offering new build actions for that application and records the deletion outcome without exposing another team’s data.

---

### User Story 2 - Orchestrate builds and observe build progress (Priority: P2)

A platform user can trigger, retry, cancel, inspect, and monitor builds so application teams can adopt standardized image creation while still seeing progress, history, and failure reasons.

**Why this priority**: The feature only delivers value if teams can reliably produce deployable images and understand what happened during each build attempt.

**Independent Test**: Can be fully tested by triggering a build for a registered application, viewing status transitions, streaming logs while it runs, listing historical builds, retrying a failed build, and canceling an in-flight build while verifying idempotent behavior and operator-visible events.

**Acceptance Scenarios**:

1. **Given** a registered application with valid build inputs, **When** the user triggers a build, **Then** the system creates a new build attempt, assigns it a trackable status, and begins asynchronous processing without blocking the API request.
2. **Given** a running build, **When** the user views build status or subscribes to build logs, **Then** the system provides current lifecycle state, timestamps, and streamable progress information until the build reaches a terminal outcome.
3. **Given** a failed build, **When** the user retries it, **Then** the system creates a new build attempt linked to the same application and records that the new attempt was initiated from a prior failure.
4. **Given** a running build, **When** the user cancels it, **Then** the system stops the attempt if possible, marks it with a terminal cancellation outcome, and records the reason in the build history.
5. **Given** an application with prior build activity, **When** the user lists build history, **Then** the system returns the ordered attempts with outcomes, artifact references, and security verification results needed for auditing and troubleshooting.

---

### User Story 3 - Apply enterprise build, registry, security, and deployment policies (Priority: P3)

A platform operator can define supported runtimes, builder variants, registry destinations, security requirements, and GitOps deployment behavior so teams get consistent images and compliant release automation by default.

**Why this priority**: Centralization is not complete unless the platform can enforce supported runtimes, registry choices, required security checks, and GitOps-first rollout behavior across teams.

**Independent Test**: Can be fully tested by configuring an application to use supported runtime detection, builder and stack selections, registry destinations, required security outputs, and GitOps deployment settings, then confirming that successful builds publish compliant artifacts and update deployment records for automated rollout.

**Acceptance Scenarios**:

1. **Given** a registered application, **When** a user selects the default automated runtime detection flow, **Then** the system chooses a supported runtime family and produces a standardized image without requiring a Dockerfile.
2. **Given** a platform policy that allows custom builder options, **When** a user selects an approved builder, custom buildpack set, or stack, **Then** the system records and applies that choice for future builds while rejecting unsupported options.
3. **Given** a successful build, **When** the platform completes artifact processing, **Then** the resulting image includes an SBOM, a security scan result, a signing result, and a published image reference in a supported registry.
4. **Given** a successful and compliant build, **When** deployment automation is enabled for the application, **Then** the system updates the team’s deployment source of truth and exposes a deployment-ready status for downstream rollout systems.

---

### Edge Cases

- What happens when the source repository is reachable but does not contain the required application descriptor or contains an unsupported runtime?
- How does the system handle a build retry request for a build that has already succeeded, been canceled, or already has another retry in progress?
- How does the system handle source, registry, security, or deployment credentials that are missing, expired, revoked, or not authorized for the requesting team?
- What happens when security verification fails after image creation but before deployment metadata is updated?
- How does the system behave when the deployment source of truth cannot be updated after a build artifact is already published?
- What happens when an application references a builder, buildpack set, stack, or registry destination that was previously allowed but has since been deprecated or disabled?
- How does the system prevent duplicate build attempts when the same trigger request is submitted repeatedly or retried by clients?
- How does the system preserve auditability and operator visibility when asynchronous build processing is interrupted by worker restarts or regional failover?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow authenticated team-scoped users to create an application record for centralized image building by providing a source repository reference and an application descriptor reference without requiring a Dockerfile.
- **FR-002**: System MUST allow authorized users to list, inspect, update, and delete application records within their team scope.
- **FR-003**: System MUST preserve historical records for application changes, build attempts, deployment updates, and security outcomes so operators can audit what changed and when.
- **FR-004**: System MUST allow users to trigger a new build for a registered application through an asynchronous API that returns a trackable build identifier.
- **FR-005**: System MUST allow users to retry a prior build attempt while preserving linkage between the new attempt and the original attempt.
- **FR-006**: System MUST allow users to cancel an in-progress build and expose a terminal cancellation outcome when the stop request is accepted.
- **FR-007**: System MUST expose current build status, timestamps, artifact references, and terminal outcomes for an individual build attempt.
- **FR-008**: System MUST allow users to stream build logs for an authorized build attempt while it is running and retain access to summarized failure reasons after completion.
- **FR-009**: System MUST allow users to list historical build attempts for an application in reverse chronological order.
- **FR-010**: System MUST support centralized image creation for applications built from the supported runtime families: Go, Java, Node, Python, and .NET.
- **FR-011**: System MUST automatically detect the application runtime from the submitted source and descriptor when the user does not explicitly override an approved builder option.
- **FR-012**: System MUST support approved custom builder, buildpack, and stack selections on an application when those options are made available by platform policy.
- **FR-013**: System MUST support publishing compliant build artifacts to approved registry destinations, including Harbor, GHCR, ECR, and GCR.
- **FR-014**: System MUST generate an SBOM for every successful image build and make the SBOM result discoverable from the build record.
- **FR-015**: System MUST perform a security scan for every successful image build and record whether the image passed or failed required security gates.
- **FR-016**: System MUST sign every successful image build and record the signing outcome and artifact reference in the build record.
- **FR-017**: System MUST prevent deployment-source updates for builds that do not satisfy required security verification outcomes.
- **FR-018**: System MUST support GitOps-first deployment workflows by updating the application’s deployment source of truth after a successful and compliant build when deployment automation is enabled.
- **FR-019**: System MUST expose whether a build has reached a deployment-ready outcome, a blocked security outcome, or a failed deployment-update outcome.
- **FR-020**: System MUST enforce existing authentication, authorization, and team-scoped access rules for application management, build operations, log streaming, security metadata, and deployment updates.
- **FR-021**: System MUST provide idempotent handling for client-submitted build trigger, retry, and cancel requests so repeated submissions do not create conflicting lifecycle outcomes.
- **FR-022**: System MUST continue asynchronous build processing across worker restarts or equivalent service interruptions without losing the canonical build state.
- **FR-023**: System MUST emit operator-visible lifecycle events for application changes, build state transitions, security verification outcomes, and deployment-update outcomes.
- **FR-024**: System MUST expose explicit failure reasons when repository access, runtime detection, builder selection, registry publishing, security verification, or deployment-source updates cannot be completed.
- **FR-025**: System MUST support horizontally scalable processing for build orchestration without allowing two workers to own the same build transition at the same time.

### Key Entities *(include if feature involves data)*

- **Application**: A team-owned service definition that identifies the source repository, application descriptor, supported runtime expectations, deployment preferences, registry destination, allowed builder options, and current management status.
- **Build**: A single build attempt for an application that records who initiated it, what source revision it used, its lifecycle state, timestamps, retry ancestry, and terminal outcome.
- **Build Artifact**: The published image output and associated metadata for a build attempt, including image reference, digest, registry destination, and deployment-readiness status.
- **Security Verification**: The collection of compliance evidence produced for a build artifact, including SBOM availability, scan outcome, signing outcome, policy-gate result, and timestamps.
- **Builder Profile**: A platform-governed selection of runtime detection behavior, builder choice, buildpack extensions, stack choice, and registry publishing rules that may be assigned to an application.
- **Deployment Update**: The record of how a successful build changed the deployment source of truth, including requested change, resulting revision reference, rollout readiness, and failure reason if the update did not complete.
- **Lifecycle Event**: A timestamped, ordered event describing a meaningful change to an application, build, security verification, or deployment update for auditability and troubleshooting.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A team can register a new application for centralized image building in under 10 minutes using only a repository reference and an application descriptor.
- **SC-002**: At least 95% of successful build requests reach a terminal outcome without manual operator intervention.
- **SC-003**: 100% of deployable images produced through this feature have discoverable SBOM, security scan, and signing results attached to the build record.
- **SC-004**: Teams can determine the current status or terminal failure reason for any build attempt within 1 minute of opening the build detail view.
- **SC-005**: Repeated trigger, retry, or cancel requests for the same build intent do not produce conflicting lifecycle outcomes in 100% of validated cases.
- **SC-006**: At least 90% of compliant successful builds that have deployment automation enabled update the deployment source of truth without manual editing.
- **SC-007**: Platform operators can support at least five independent application teams building concurrently without one team seeing another team’s application, build, or security data.

## Assumptions

- Existing authentication, authorization, team-scoped access control, and audit logging capabilities in idp-core will be reused for this feature.
- The first release targets backend APIs and operator workflows; separate end-user UI work can be layered on after the backend contracts are available.
- Application teams can provide or reference an application descriptor that contains the minimum metadata required for source retrieval, runtime detection, and deployment intent.
- Approved builder options, registries, security policies, and deployment targets are managed by the platform rather than created ad hoc by every application team.
- Deployment automation updates a team-owned source of truth and does not directly modify live workloads outside the established GitOps workflow.
- The platform can distinguish transient processing failures from terminal build failures and will surface both to users.
- Existing phase-complete platform capabilities for service management, environments, delivery targets, and notifications remain the baseline context for this feature.

## Data Model Considerations

- The planning phase MUST define persistent storage for applications, build attempts, build artifacts, security verification results, deployment updates, and lifecycle events.
- The data model MUST support one-to-many relationships from application to builds and from build to artifact, security verification, deployment update, and lifecycle events.
- The data model MUST support retry ancestry and idempotency tracking so repeated client requests can be reconciled safely.
- The data model MUST preserve immutable historical records for completed build attempts and security outcomes.

## API and Contract Expectations

- The planning phase MUST define stable API contracts for application CRUD, build actions, build status retrieval, log streaming, and build history retrieval.
- Contracts MUST distinguish synchronous request acceptance from asynchronous build completion.
- Contracts MUST expose explicit terminal outcomes for success, failure, cancellation, policy block, and deployment-update failure.
- Contracts MUST define team-scoped authorization behavior and sanitized failure responses for repository, registry, security, and deployment errors.

## Event Model Expectations

- The system MUST produce ordered lifecycle events for application creation, application update, build queued, build running, build canceled, build failed, build succeeded, security verification completed, deployment update started, deployment update succeeded, and deployment update failed.
- Events MUST be attributable to an application and, when relevant, a specific build attempt.
- Events MUST be suitable for asynchronous processing, notifications, and operator troubleshooting.

## State Machine Expectations

- An application lifecycle MUST support active, suspended, deleting, and deleted states.
- A build lifecycle MUST support pending, queued, running, canceling, canceled, failed, succeeded, blocked, and deployment-ready states.
- Security verification MUST support pending, passed, failed, and waived states when a waiver mechanism is later introduced.
- Deployment update tracking MUST support pending, in-progress, succeeded, and failed states.

## External Resource Expectations

- The planning phase MUST define any external build or deployment resources needed to represent application build intent, builder configuration, build execution, and artifact status.
- Those resources MUST support team-scoped ownership, lifecycle observability, and reconciliation between requested and observed build state.

## Sequence and Diagram Expectations

- The planning phase MUST document the primary happy path from application registration through build completion and deployment-source update.
- The planning phase MUST document failure paths for repository access errors, security verification failures, registry publish failures, and deployment update failures.
- Mermaid diagrams SHOULD be used in design artifacts to describe state transitions and end-to-end flow once the implementation plan is produced.

## Risks

- Runtime auto-detection may produce ambiguous results for repositories containing multiple buildable services or unsupported layouts.
- Centralized security gates may delay delivery if policy exceptions and failure triage workflows are not defined clearly.
- Registry, repository, and deployment credentials create cross-system dependency risks that can fail a build after source retrieval succeeds.
- Event-driven and horizontally scaled processing increases the risk of duplicate or out-of-order transitions if idempotency and ownership rules are not designed carefully.
- Deployment-source updates can diverge from live rollout state unless downstream rollout observability is linked back to the build record.

## Future Enhancements

- Support promotion workflows that reuse previously verified images across environments without rebuilding.
- Support richer policy controls for team-specific security gates, signing requirements, and deployment approval steps.
- Support monorepo-aware application discovery and build coordination.
- Support build cost visibility, quota controls, and chargeback reporting.
- Support waiver workflows for time-bound policy exceptions with explicit approvals and expiry.