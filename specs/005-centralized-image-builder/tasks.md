# Tasks: Centralized Image Builder

**Input**: Design documents from `/specs/005-centralized-image-builder/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No separate test-first tasks are generated because the specification did not explicitly require TDD. Validation is covered through focused implementation verification and the runnable flows in `specs/005-centralized-image-builder/quickstart.md`.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Use the real idp-core repository paths from plan.md
- Backend code normally lives under `cmd/`, `internal/`, `configs/`, `migrations/`, and `deployments/`
- Keep handler/usecase/repository/model work in separate tasks when they touch different files
- Add dedicated test paths only when the feature introduces them

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the feature-facing documentation, domain scaffolding, and migration baseline before story work begins

- [X] T001 Capture central design assumptions and validation prerequisites in `specs/005-centralized-image-builder/quickstart.md`
- [X] T002 Create feature migration scaffolding for build applications, builds, artifacts, security verification, deployment updates, and lifecycle events in `migrations/202606260001_create_build_application_tables.up.sql` and `migrations/202606260001_create_build_application_tables.down.sql`
- [X] T003 [P] Add new domain package scaffolding in `internal/model/build_application/type.go`, `internal/repository/build_application/init.go`, `internal/usecase/build_application/init.go`, and `internal/handler/http/build_application.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core shared infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Implement shared build application, build, artifact, security verification, deployment update, and lifecycle event models in `internal/model/build_application/type.go`
- [X] T005 [P] Implement repository interfaces and persistence methods for build applications and core build records in `internal/repository/build_application/init.go` and `internal/repository/build_application/build_application.go`
- [X] T006 [P] Implement execution-phase persistence for build state transitions, idempotency records, and retry/cancel linkage in `internal/repository/build_application/build_execution.go`
- [X] T007 Implement shared build lifecycle/status, idempotency, and retry validation helpers in `internal/usecase/build_application/init.go`
- [X] T008 Implement team-scoped authorization, audit/event emission helpers, and sanitized error mapping in `internal/usecase/build_application/helpers.go` and `internal/handler/http/build_application.go`
- [X] T009 Wire the new repository and usecase dependencies in `cmd/http/main.go`, `cmd/http/server.go`, and `internal/handler/http/handler.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in priority order

---

## Phase 3: User Story 1 - Register and manage buildable applications (Priority: P1) 🎯 MVP

**Goal**: Let a team register, inspect, update, list, and delete buildable applications using repository and descriptor metadata instead of Dockerfiles

**Independent Test**: Create a build application, fetch it, update mutable metadata, list it in team scope, then delete it while confirming historical records remain intact and no other team’s data is exposed.

### Implementation for User Story 1

- [X] T010 [P] [US1] Add build application request/response models and validation rules in `internal/model/build_application/type.go`
- [X] T011 [US1] Implement build application CRUD usecase methods in `internal/usecase/build_application/build_application.go`
- [X] T012 [US1] Implement build application CRUD repository queries in `internal/repository/build_application/build_application.go`
- [X] T013 [US1] Implement HTTP handlers for `POST/GET/PATCH/DELETE /v1/build-applications` in `internal/handler/http/build_application.go`
- [X] T014 [US1] Register build application routes and dependency plumbing in `cmd/http/server.go`, `cmd/http/main.go`, and `internal/handler/http/handler.go`
- [X] T015 [US1] Align application CRUD contracts and quickstart expectations, including explicit `DELETE /v1/build-applications/{id}` status semantics, in `specs/005-centralized-image-builder/contracts/centralized-image-builder.md` and `specs/005-centralized-image-builder/quickstart.md

**Checkpoint**: User Story 1 should now be fully functional and testable independently

---

## Phase 4: User Story 2 - Trigger and observe asynchronous builds (Priority: P2)

**Goal**: Let teams trigger, inspect, retry, cancel, stream logs for, and review the history of asynchronous builds with stable lifecycle visibility

**Independent Test**: Trigger a build for a registered application, inspect build detail and history, stream logs while it runs, retry a failed build, and cancel an active build while confirming idempotent behavior and ordered lifecycle outcomes.

### Implementation for User Story 2

- [X] T016 [P] [US2] Add build action, build detail, and build history response models in `internal/model/build_application/type.go`
- [X] T017 [US2] Implement build trigger, status, history, retry, and cancel orchestration in `internal/usecase/build_application/build_execution.go`
- [X] T018 [US2] Implement canonical build state persistence, idempotency tracking, and retry linkage in `internal/repository/build_application/build_execution.go`
- [X] T019 [US2] Implement build logs streaming and terminal-summary retrieval in `internal/repository/build_application/log_stream.go` and `internal/usecase/build_application/log_stream.go`
- [X] T020 [US2] Implement HTTP handlers for `POST /v1/build-applications/{id}/builds`, `GET /v1/build-applications/{id}/builds`, `GET /v1/builds/{buildId}`, `POST /v1/builds/{buildId}/retry`, `POST /v1/builds/{buildId}/cancel`, and `GET /v1/builds/{buildId}/logs/stream` in `internal/handler/http/build_application.go`
- [X] T021 [US2] Update route registration and handler dependencies for build execution endpoints in `cmd/http/server.go` and `internal/handler/http/handler.go`

**Checkpoint**: User Stories 1 and 2 should now both work independently

---

## Phase 5: User Story 3 - Enforce security verification and GitOps-first deployment readiness (Priority: P3)

**Goal**: Ensure successful builds only become deployment-ready after SBOM, scan, signing, and deployment-source update outcomes are recorded and policy-compliant

**Independent Test**: Run one compliant build and one policy-blocked build, then confirm artifact publication, security verification, deployment update results, and deployment-ready versus blocked outcomes are attached to the correct build records.

### Implementation for User Story 3

- [X] T022 [P] [US3] Extend artifact, security verification, deployment update, and lifecycle event models in `internal/model/build_application/type.go`
- [X] T023 [US3] Implement post-build security verification and deployment-readiness state transitions in `internal/usecase/build_application/post_build.go`
- [X] T024 [US3] Implement post-build persistence for artifact publication, security verification, deployment updates, and lifecycle events after execution completion in `internal/repository/build_application/post_build.go`
- [X] T025 [US3] Integrate GitOps update and deployment outcome recording through `internal/repository/gitops/` and `internal/usecase/build_application/post_build.go`
- [X] T026 [US3] Expose security verification, artifact, and deployment update fields in build detail/history handlers in `internal/handler/http/build_application.go`
- [X] T027 [US3] Align security-gated deployment contract and failure semantics in `specs/005-centralized-image-builder/contracts/centralized-image-builder.md` and `specs/005-centralized-image-builder/quickstart.md`

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cross-story hardening

- [X] T028 [P] Add focused unit and handler coverage for build application flows in `internal/usecase/build_application/build_application_test.go` and `internal/handler/http/build_application_test.go`
- [X] T029 [P] Add repository-level coverage for idempotency, lifecycle ordering, and persistence rules in `internal/repository/build_application/build_application_test.go` and `internal/repository/build_application/build_execution_test.go`
- [X] T030 Harden observability and notification integration for lifecycle events in `internal/usecase/build_application/helpers.go` and `internal/usecase/notification/init.go`
- [X] T031 Refresh Swagger annotations and API-facing documentation in `internal/handler/http/build_application.go` and `docs/swagger/`
- [X] T032 Run the end-to-end validation scenarios from `specs/005-centralized-image-builder/quickstart.md` and capture follow-up adjustments in `specs/005-centralized-image-builder/quickstart.md`
- [X] T033 Implement explicit runtime-family and registry-type validation rules for supported values in `internal/model/build_application/type.go` and `internal/usecase/build_application/build_application.go`
- [X] T034 Implement and verify lifecycle event taxonomy coverage for application, build, security, and deployment transitions in `internal/usecase/build_application/helpers.go` and `internal/repository/build_application/post_build.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational completion
- **User Story 2 (Phase 4)**: Depends on Foundational completion and builds on the application domain introduced in User Story 1
- **User Story 3 (Phase 5)**: Depends on Foundational completion and builds on canonical build records from User Story 2
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational - MVP slice
- **User Story 2 (P2)**: Depends on the shared application/build domain and should follow User Story 1 because builds require registered applications
- **User Story 3 (P3)**: Depends on build execution records from User Story 2

### Within Each User Story

- Shared models before usecase orchestration
- Usecase logic before handlers and routes
- Handler/route work before contract and quickstart alignment
- Story-specific flows should be runnable before moving to the next priority

### Parallel Opportunities

- T003 can run while migration scaffolding in T002 is being prepared
- T005 and T006 can run in parallel after T004 because they touch different repository files
- T010 and T012 can run in parallel once foundational domain types are stable
- T022 and T024 can run in parallel once User Story 2 has established canonical build execution records
- T028 and T029 can run in parallel during polish because they target different test layers

---

## Parallel Example: Foundational Work

```bash
Task: "Implement repository interfaces and persistence methods in internal/repository/build_application/init.go and internal/repository/build_application/build_application.go"
Task: "Implement execution-phase persistence for build transitions and idempotency in internal/repository/build_application/build_execution.go"
```

---

## Parallel Example: User Story 3

```bash
Task: "Extend artifact, security verification, deployment update, and lifecycle event models in internal/model/build_application/type.go"
Task: "Implement post-build persistence in internal/repository/build_application/post_build.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run the application registration and CRUD scenarios from `specs/005-centralized-image-builder/quickstart.md`

### Incremental Delivery

1. Complete Setup + Foundational
2. Deliver User Story 1 and validate build application CRUD
3. Deliver User Story 2 and validate build execution, history, and log streaming
4. Deliver User Story 3 and validate security verification plus GitOps deployment readiness
5. Finish with cross-cutting tests, docs, and end-to-end quickstart validation

### Parallel Team Strategy

With multiple developers:

1. One developer prepares migrations while another scaffolds the new build domain packages
2. After foundational work, one developer can focus on application CRUD while another prepares build execution persistence and log streaming support
3. User Story 3 can begin once canonical build records are stable from User Story 2

---

## Notes

- All tasks follow the required checklist format with task ID, optional parallel marker, story label where required, and exact file paths
- The initial implementation is backend-first; UI work is intentionally out of scope for this task list
- Security verification and deployment updates must remain distinct from basic build success semantics
- Security verification waiver states are future enhancement only and out of scope for this implementation task list
- Team-scoped authorization, auditability, and sanitized error handling apply across every story
