# Tasks: Target-Aware Sync and Status Resolution

**Input**: Design documents from `/specs/003-target-aware-sync-status/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Include focused automated coverage because the plan and quickstart require validation of target resolution, failure modes, and cross-target isolation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare feature scaffolding and migration entry points used by all stories

- [X] T001 Add delivery-target control-plane fields to `internal/model/delivery_target/type.go` and response structs for target-aware metadata exposure
- [X] T002 Create `migrations/20260611000000_alter_delivery_targets_add_control_plane_fields.sql` to add control-plane metadata columns for target-aware resolution
- [X] T003 [P] Add target-aware config/model helpers in `internal/pkg/config/config.go` for safe default-control-plane fallback inputs used by providers

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core target-resolution infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Create target-resolution and provider interfaces in `internal/usecase/environment/init.go` for target-scoped GitOps and provisioner selection
- [X] T005 [P] Implement target-scoped GitOps provider/factory wiring in `internal/repository/gitops/init.go` and supporting files under `internal/repository/gitops/`
- [X] T006 [P] Implement target-scoped provisioner provider/factory wiring in `internal/repository/provisioner/init.go` and supporting files under `internal/repository/provisioner/`
- [X] T007 Implement delivery-target control-plane lookup helpers in `internal/repository/delivery_target/init.go` and `internal/repository/delivery_target/delivery_target.go`
- [X] T008 Wire target-aware dependencies in `cmd/http/main.go` so `internal/usecase/environment` receives target-scoped GitOps/provisioner providers for sync, GitOps status, and directly affected environment status/workload paths
- [X] T009 Add shared target-resolution error types and audit-safe logging helpers in `internal/usecase/environment/init.go` and `internal/usecase/environment/environment.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Sync an environment through its assigned target control plane (Priority: P1) 🎯 MVP

**Goal**: Trigger sync through the control plane resolved from the environment’s assigned delivery target instead of one global client

**Independent Test**: Trigger sync for environments mapped to two different delivery targets and confirm each request uses the correct target control plane or returns a target-specific resolution error

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T010 [P] [US1] Add usecase tests for target-aware sync success and resolution failures in `internal/usecase/environment/environment_test.go`
- [X] T011 [P] [US1] Add provider/repository tests for target-scoped GitOps selection in `internal/repository/gitops/argocd_test.go`

### Implementation for User Story 1

- [X] T012 [P] [US1] Extend delivery-target create/update/read models in `internal/model/delivery_target/type.go` for sync-related control-plane metadata
- [X] T013 [US1] Persist and query control-plane metadata in `internal/repository/delivery_target/delivery_target.go`
- [X] T014 [US1] Implement target resolution flow for sync requests in `internal/usecase/environment/environment.go` using the new provider interfaces
- [X] T015 [US1] Update sync endpoint error mapping and request handling in `internal/handler/http/environment.go` for target-aware sync failures
- [X] T016 [US1] Update target-aware GitOps repository behavior for sync in `internal/repository/gitops/argocd.go` and related provider files under `internal/repository/gitops/`
- [X] T017 [US1] Add audit-safe logging and last-sync persistence for target-aware sync in `internal/usecase/environment/environment.go` and `internal/repository/environment/`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Read GitOps status through the correct target control plane (Priority: P2)

**Goal**: Return GitOps status from the control plane resolved for the environment’s assigned delivery target

**Independent Test**: Read GitOps status for environments on distinct targets and confirm each result or failure comes from the correct resolved target control plane

### Tests for User Story 2 ⚠️

- [X] T018 [P] [US2] Add usecase tests for target-aware GitOps status, application-not-found behavior, and default-control-plane compatibility in `internal/usecase/environment/environment_test.go`
- [X] T019 [P] [US2] Add repository tests for target-scoped status retrieval and single-target fallback behavior in `internal/repository/gitops/argocd_test.go`

### Implementation for User Story 2

- [X] T020 [US2] Implement target-aware GitOps status resolution in `internal/usecase/environment/environment.go`
- [X] T021 [US2] Update `GET /v1/environments/:id/gitops/status` handling for target-specific failures in `internal/handler/http/environment.go`
- [X] T022 [US2] Route Argo status inside environment status responses through the resolved target control plane in `internal/usecase/environment/environment.go`
- [X] T023 [US2] Update target-scoped status retrieval behavior in `internal/repository/gitops/argocd.go` and provider files under `internal/repository/gitops/`
- [X] T024 [US2] Integrate target-aware provisioner selection for environment status/workload reads where control-plane scoping is required in `internal/repository/provisioner/` and `internal/usecase/environment/environment.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Diagnose target resolution failures and preserve audit visibility (Priority: P3)

**Goal**: Make control-plane selection and resolution failures visible to operators without exposing secrets

**Independent Test**: Trigger sync and status against valid and broken target mappings and verify logs, retained records, and responses identify the attempted target selection safely

### Tests for User Story 3 ⚠️

- [X] T025 [P] [US3] Add usecase tests for audit-safe target-resolution outcomes, concurrent cross-target isolation, and no client leakage between targets in `internal/usecase/environment/environment_test.go`
- [X] T026 [P] [US3] Add handler tests for target-aware sync/status failure responses in `internal/handler/http/environment_test.go`

### Implementation for User Story 3

- [X] T027 [P] [US3] Add operator-visible target-resolution payload helpers in `internal/model/environment/type.go` or a new model file under `internal/model/environment/`
- [X] T028 [US3] Record audit-safe target-resolution outcomes in `internal/usecase/environment/environment.go` and existing audit/notification integrations used there
- [X] T029 [US3] Expose control-plane metadata in delivery-target read responses in `internal/model/delivery_target/type.go` and `internal/handler/http/delivery_target.go`
- [X] T030 [US3] Update delivery-target usecase validation for sync/status-ready targets in `internal/usecase/delivery_target/delivery_target.go`
- [X] T031 [US3] Add target-specific failure messaging and secret-safe response handling in `internal/handler/http/environment.go`

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and contract alignment across all user stories

- [X] T032 [P] Update contract and validation details in `specs/003-target-aware-sync-status/contracts/target-aware-sync-status.md` and `specs/003-target-aware-sync-status/quickstart.md` if implementation changes any endpoint/error specifics
- [ ] T033 Run multi-target and single-target validation scenarios from `specs/003-target-aware-sync-status/quickstart.md`, including concurrent requests across two targets, and capture any required implementation follow-ups in code/tests
- [X] T034 [P] Run regression coverage with `go test ./...` and fix any target-aware sync/status regressions in affected files
- [X] T035 [P] Regenerate Swagger or API documentation artifacts affected by `internal/handler/http/environment.go` or delivery-target response changes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Reuses the same provider path as US1 but remains independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Builds on target-resolution behavior from US1/US2 but should be independently testable through failure and audit scenarios

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Model and contract exposure before repository persistence changes
- Repository/provider selection before usecase orchestration
- Usecase orchestration before handler response updates
- Story complete before moving to next priority

### Parallel Opportunities

- Setup tasks marked [P] can run in parallel
- Foundational provider tasks T005 and T006 can run in parallel once T004 establishes shared interfaces
- Tests within each user story marked [P] can run in parallel
- US1 model extension T012 can run in parallel with test tasks after the foundation is ready
- US3 observability payload/model work T027 can run in parallel with handler tests T026

---

## Parallel Example: User Story 1

```bash
# Launch both User Story 1 test tasks together:
Task: "Add usecase tests for target-aware sync success and resolution failures in internal/usecase/environment/environment_test.go"
Task: "Add provider/repository tests for target-scoped GitOps selection in internal/repository/gitops/argocd_test.go"

# After tests exist, launch independent model/provider work together:
Task: "Extend delivery-target create/update/read models in internal/model/delivery_target/type.go for sync-related control-plane metadata"
Task: "Update target-aware GitOps repository behavior for sync in internal/repository/gitops/argocd.go and related provider files under internal/repository/gitops/"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently with two targets and one broken mapping
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 sync path
   - Developer B: User Story 2 status path
   - Developer C: User Story 3 observability and failure reporting
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Verify tests fail before implementing
- Validate multi-target isolation and default-control-plane compatibility before closing the feature
- Avoid vague tasks, same-file conflicts across parallel work, and hidden fallback behavior to an unrelated global control plane
