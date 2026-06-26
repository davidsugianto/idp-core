# Research: Centralized Image Builder

## Decision 1: Model Centralized Image Builder as a dedicated application-build domain rather than extending the existing service catalog directly

**Decision**: Introduce a dedicated domain for buildable applications, build attempts, artifacts, security verification, deployment updates, and lifecycle events instead of folding this feature into the existing service catalog objects.

**Rationale**: The existing service catalog models describe discoverable services, versions, endpoints, dependencies, and environment deployments. Centralized image building adds source-repository configuration, asynchronous build orchestration, runtime detection, security evidence, signing state, registry publication, GitOps update tracking, and retry/cancel semantics that would overload the meaning of the current service entities.

**Alternatives considered**:
- Reuse the existing service and version entities as the primary build objects. Rejected because build orchestration would become entangled with catalog metadata and make the feature harder to reason about independently.
- Create a thin wrapper around environments only. Rejected because the feature begins before environment deployment and must remain valid even when no environment deployment exists yet.

## Decision 2: Keep long-running build orchestration asynchronous and event-driven while exposing a stable API contract

**Decision**: Accept build trigger, retry, and cancel requests synchronously, persist canonical build intent immediately, and drive execution through asynchronous workers and ordered lifecycle events.

**Rationale**: The feature must support HA, horizontal scaling, retries, log streaming, and idempotent APIs. Those requirements fit a queued, event-driven orchestration model better than request-bound execution, and they align with the repository constitution requirement for explicit contracted and observable long-running operations.

**Alternatives considered**:
- Execute builds directly in the request path. Rejected because it would not meet reliability and scaling requirements and would make cancellation, retries, and restarts harder to manage.
- Hide build internals behind a single polling endpoint only. Rejected because users need explicit lifecycle visibility, history, and operator-observable events.

## Decision 3: Treat the build application descriptor as the authoritative user-supplied contract for source, runtime, and deployment intent

**Decision**: The platform will require a repository reference plus an application descriptor as the minimum user input. The descriptor becomes the canonical description of build intent, registry policy selection, and deployment automation settings.

**Rationale**: The user explicitly wants a Dockerfile-free workflow where developers provide only a Git repository and `application.yaml`. A descriptor-centered contract gives the platform one consistent source of truth without requiring repository-specific Docker build knowledge.

**Alternatives considered**:
- Infer everything from source layout alone. Rejected because deployment intent, registry selection, and optional builder preferences still need an explicit user-controlled contract.
- Require users to provide both descriptor and multiple platform-side forms. Rejected because it weakens onboarding simplicity.

## Decision 4: Separate build production from deployment readiness using explicit post-build security gates

**Decision**: A build may succeed at image creation before it becomes deployment-ready. SBOM generation, security scanning, signing, and policy evaluation must each be recorded distinctly, and deployment-source updates only occur after the required security gates pass.

**Rationale**: The user requires SBOM generation, Trivy scanning, Cosign signing, and GitOps-first deployment. Modeling security verification explicitly prevents the system from conflating artifact creation with compliance approval.

**Alternatives considered**:
- Collapse security outcomes into a single build status. Rejected because operators need to distinguish build failure from policy-blocked promotion.
- Allow deployment updates before verification finishes. Rejected because it would violate the stated security expectations.

## Decision 5: Use platform-governed builder profiles to represent supported runtimes, builder choices, stacks, and registry publication policy

**Decision**: Represent supported runtime families and optional custom builder/buildpack/stack combinations through approved builder profiles that can be assigned to applications.

**Rationale**: The feature must support automatic runtime detection plus optional approved customizations. A policy-driven profile model preserves central governance while still allowing controlled customization.

**Alternatives considered**:
- Let every application define arbitrary builder and buildpack choices. Rejected because it would reintroduce inconsistency and platform drift.
- Support only one global builder with no overrides. Rejected because the requested scope explicitly includes custom builder, buildpack, and stack support.

## Decision 6: Track GitOps deployment updates as first-class records linked to builds

**Decision**: Record each deployment-source update attempt separately from the build and artifact records, including requested image reference, target repository location, resulting revision reference, and rollout-readiness outcome.

**Rationale**: GitOps-first deployment introduces a second long-running workflow after artifact publication. Operators need to know whether failure occurred during build, security verification, or deployment-source update.

**Alternatives considered**:
- Store only the latest deployment result on the build record. Rejected because it would hide repeated attempts and reduce auditability.
- Update deployment state out of band without persistent records. Rejected because the constitution requires observable and traceable operations.

## Decision 7: Preserve strict team isolation even when external integrations are centrally managed

**Decision**: Application records, builds, logs, security evidence, and deployment updates remain team-scoped, while external integration credentials and allowed destinations remain centrally governed and referenced indirectly.

**Rationale**: The repository constitution requires tenant isolation and safe secret handling. Teams need self-service workflows, but cross-system credentials must not be copied into application-owned objects or exposed through API responses.

**Alternatives considered**:
- Allow each application to embed raw source, registry, or signing credentials. Rejected because it weakens security boundaries.
- Use one global build namespace and global visibility with no team partitioning. Rejected because it breaks tenant isolation requirements.
