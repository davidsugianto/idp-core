# Tasks: Phase 3 Backend Foundation

**Input**: Design documents from `/specs/002-phase3-backend-foundation/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Include handler/usecase validation for each user story because the specification defines independent test criteria and the constitution requires validation for contract, authorization, and long-running operation changes.

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

**Purpose**: Establish the Phase 3 persistence and bootstrap structure shared by all stories.

- [ ] T001 Create Phase 3 SQL migration files for templates, delivery targets, movements, notifications, and environment foreign keys in `migrations/20260608000000_create_templates_table.sql`, `migrations/20260608000001_create_template_versions_table.sql`, `migrations/20260608000002_create_template_parameters_table.sql`, `migrations/20260608000003_create_template_resources_table.sql`, `migrations/20260608000004_create_template_instances_table.sql`, `migrations/20260608000005_create_delivery_targets_table.sql`, `migrations/20260608000006_create_environment_movements_table.sql`, `migrations/20260608000007_create_notifications_table.sql`, and `migrations/20260608000008_alter_environments_add_phase3_refs.sql`
- [ ] T002 [P] Register Phase 3 model types and non-schema bootstrap wiring in `cmd/http/main.go`
- [ ] T003 [P] Extend Phase 3 permission and role seeds for template, delivery target, movement, notification, and live update actions in `internal/seed/permission.go` and `internal/seed/role.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T004 Create shared Phase 3 models in `internal/model/delivery_target/type.go`, `internal/model/environment_movement/type.go`, `internal/model/notification/type.go`, and `internal/model/live_subscription/type.go`
- [ ] T005 [P] Create repository packages and interfaces for shared Phase 3 persistence in `internal/repository/delivery_target/init.go`, `internal/repository/delivery_target/delivery_target.go`, `internal/repository/environment_movement/init.go`, `internal/repository/environment_movement/environment_movement.go`, `internal/repository/notification/init.go`, and `internal/repository/notification/notification.go`
- [ ] T006 [P] Extend existing repository and usecase interfaces for Phase 3 shared fields and hooks in `internal/repository/environment/init.go`, `internal/repository/environment/environment.go`, `internal/repository/template/init.go`, `internal/repository/template/version.go`, `internal/usecase/environment/init.go`, `internal/usecase/template/init.go`, `internal/usecase/notification/init.go`, and `internal/usecase/live_update/init.go`
- [ ] T007 Wire shared Phase 3 repositories, usecases, and handler dependencies in `cmd/http/main.go`, `cmd/http/server.go`, and `internal/handler/http/init.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel.

---

## Phase 3: User Story 1 - Manage reusable templates (Priority: P1) 🎯 MVP

**Goal**: Complete template management so admins can define versioned templates with parameters/resources, validate input, enforce lifecycle rules, and create environments with preserved template history.

**Independent Test**: Create a template, publish a version, replace parameter/resource definitions, validate good and bad input, verify duplicate-version and invalid-lifecycle conflicts, then create an environment from that version and confirm a historical template-instance record exists.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T008 [P] [US1] Add handler contract coverage for template parameters, resources, validate endpoints, duplicate version conflicts, and lifecycle transition failures in `internal/handler/http/template_test.go`
- [ ] T009 [P] [US1] Add usecase integration coverage for template-backed environment creation and template-instance history preservation in `internal/usecase/template/template_integration_test.go` and `internal/usecase/environment/environment_template_integration_test.go`

### Implementation for User Story 1

- [ ] T010 [P] [US1] Expand template domain request and response models for parameter/resource replacement, lifecycle responses, and template instantiation history in `internal/model/template/type.go`, `internal/model/template_version/type.go`, `internal/model/template_parameter/type.go`, `internal/model/template_resource/type.go`, and `internal/model/template_instance/type.go`
- [ ] T011 [US1] Implement template parameter, resource, and instance persistence methods in `internal/repository/template/init.go`, `internal/repository/template/parameter.go`, `internal/repository/template/resource.go`, and `internal/repository/template/instance.go`
- [ ] T012 [US1] Implement template version publishing, lifecycle transition enforcement, concurrent update handling, parameter/resource replacement, conflict handling, and input validation logic in `internal/usecase/template/init.go` and `internal/usecase/template/template.go`
- [ ] T013 [US1] Extend environment create and read contracts for template version, input payloads, and template-instance history in `internal/model/environment/type.go`
- [ ] T014 [US1] Implement template-backed environment creation and template-instance history recording in `internal/usecase/environment/environment.go`
- [ ] T015 [US1] Add template parameter, resource, validation, and lifecycle handlers in `internal/handler/http/template.go`
- [ ] T016 [US1] Register template validation routes and template-aware environment contract updates in `cmd/http/server.go` and `internal/handler/http/environment.go`
- [ ] T017 [US1] Add template lifecycle authorization, audit, history preservation, and `409` conflict semantics in `internal/handler/http/template.go`, `internal/usecase/template/template.go`, and `internal/usecase/environment/environment.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 4: User Story 2 - Target and manage multiple clusters (Priority: P2)

**Goal**: Introduce delivery targets and movement workflows so users can place environments on approved targets and track relocation progress/history.

**Independent Test**: Register multiple delivery targets, create an environment against an approved target with the existing environment flow, request a move to another target, and verify status/history without relying on live stream delivery.

### Tests for User Story 2 ⚠️

- [ ] T018 [P] [US2] Add handler contract coverage for delivery target and movement endpoints in `internal/handler/http/delivery_target_test.go` and `internal/handler/http/environment_movement_test.go`
- [ ] T019 [P] [US2] Add usecase integration coverage for target-aware environment placement and movement state transitions in `internal/usecase/delivery_target/delivery_target_integration_test.go` and `internal/usecase/environment_movement/environment_movement_integration_test.go`

### Implementation for User Story 2

- [ ] T020 [P] [US2] Extend environment placement fields and delivery target DTOs in `internal/model/environment/type.go` and `internal/model/delivery_target/type.go`
- [ ] T021 [US2] Implement delivery target persistence, filtering, and availability updates in `internal/repository/delivery_target/init.go` and `internal/repository/delivery_target/delivery_target.go`
- [ ] T022 [US2] Implement environment movement persistence and progress history queries in `internal/repository/environment_movement/init.go` and `internal/repository/environment_movement/environment_movement.go`
- [ ] T023 [US2] Implement delivery target business rules and admin CRUD in `internal/usecase/delivery_target/init.go` and `internal/usecase/delivery_target/delivery_target.go`
- [ ] T024 [US2] Implement environment movement orchestration and target-aware placement checks in `internal/usecase/environment_movement/init.go`, `internal/usecase/environment_movement/environment_movement.go`, and `internal/usecase/environment/environment.go`
- [ ] T025 [US2] Implement delivery target and movement handlers in `internal/handler/http/delivery_target.go` and `internal/handler/http/environment_movement.go`
- [ ] T026 [US2] Register delivery target and movement routes and dependency wiring in `cmd/http/main.go`, `cmd/http/server.go`, and `internal/handler/http/init.go`
- [ ] T027 [US2] Preserve delivery target selection and movement audit/history records in `internal/usecase/delivery_target/delivery_target.go`, `internal/usecase/environment_movement/environment_movement.go`, and `internal/repository/environment/environment.go`

**Checkpoint**: At this point, User Stories 1 and 2 should both work independently.

---

## Phase 5: User Story 3 - Receive live operational updates (Priority: P3)

**Goal**: Deliver authenticated backend live-update channels for environment status, progress, logs, and notifications with proper access control and retained notification history.

**Independent Test**: Subscribe to environment events and logs, trigger status changes or movement progress, confirm ordered event delivery for authorized users, verify retained notification history, and confirm streams stop cleanly when unsubscribed, expired, or access is lost.

### Tests for User Story 3 ⚠️

- [ ] T028 [P] [US3] Add handler contract coverage for notification history, SSE stream endpoints, and `401`/`403`/`410` failure modes in `internal/handler/http/live_update_test.go`
- [ ] T029 [P] [US3] Add usecase integration coverage for event streaming, log streaming, notification history retrieval, stream expiry, and access-loss termination in `internal/usecase/live_update/live_update_integration_test.go`

### Implementation for User Story 3

- [ ] T030 [P] [US3] Finalize notification and live subscription models for event payloads, list filters, and ephemeral stream session state in `internal/model/notification/type.go` and `internal/model/live_subscription/type.go`
- [ ] T031 [US3] Implement notification persistence and recent-history queries in `internal/repository/notification/init.go` and `internal/repository/notification/notification.go`
- [ ] T032 [US3] Implement notification history listing and filter handling in `internal/usecase/notification/init.go` and `internal/usecase/notification/notification.go`
- [ ] T033 [US3] Implement SSE event orchestration for status, progress, and notification channels in `internal/usecase/live_update/init.go` and `internal/usecase/live_update/live_update.go`
- [ ] T034 [US3] Implement workload log streaming and informer-backed event sourcing in `internal/usecase/live_update/log_stream.go` and `internal/repository/provisioner/informer.go`
- [ ] T035 [US3] Emit notification and progress events from template, environment, and movement workflows in `internal/usecase/template/template.go`, `internal/usecase/environment/environment.go`, and `internal/usecase/environment_movement/environment_movement.go`
- [ ] T036 [US3] Implement notification list and SSE handlers in `internal/handler/http/live_update.go`
- [ ] T037 [US3] Register notification history and live update routes and dependency wiring in `cmd/http/main.go`, `cmd/http/server.go`, and `internal/handler/http/init.go`
- [ ] T038 [US3] Enforce stream authorization, expiry, unsubscribe, access-loss handling, and `410 Gone` invalidation behavior in `internal/handler/http/live_update.go` and `internal/usecase/live_update/live_update.go`

**Checkpoint**: All user stories should now be independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize contracts, validation, and documentation across all user stories.

- [ ] T039 [P] Update Swagger annotations and request/response examples for Phase 3 endpoints in `internal/handler/http/template.go`, `internal/handler/http/environment.go`, `internal/handler/http/delivery_target.go`, `internal/handler/http/environment_movement.go`, and `internal/handler/http/live_update.go`
- [ ] T040 Run the end-to-end validation scenarios from `specs/002-phase3-backend-foundation/quickstart.md` for template flow, delivery target placement, movement history, notification history, and SSE/log streams
- [ ] T041 [P] Validate Phase 3 performance and failure-mode expectations against `specs/002-phase3-backend-foundation/plan.md` and `specs/002-phase3-backend-foundation/contracts/phase3-backend-api.md` using focused coverage in `internal/handler/http/template_test.go` and `internal/handler/http/live_update_test.go`
- [ ] T042 [P] Run full regression coverage for Phase 3 touched packages and stabilize failures in `internal/handler/http/template_test.go`, `internal/handler/http/delivery_target_test.go`, `internal/handler/http/environment_movement_test.go`, and `internal/handler/http/live_update_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel if multiple engineers are available
  - User stories can also proceed sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Foundational and delivers the MVP template workflow
- **User Story 2 (P2)**: Starts after Foundational and can use the existing environment flow even if template-backed creation is not yet enabled
- **User Story 3 (P3)**: Starts after Foundational and can stream events for existing environments before template or movement polish is complete

### Within Each User Story

- Tests MUST be written and fail before implementation
- Model and contract updates come before repository work
- Repository work comes before usecase orchestration
- Usecase orchestration comes before handler and route wiring
- Authorization, audit, conflict handling, and historical record preservation finish the story before validation

### Parallel Opportunities

- `T002` and `T003` can run in parallel after migration planning begins
- `T005` and `T006` can run in parallel within the foundational phase
- `T008` and `T009` can run in parallel for US1, then `T010` can proceed independently before repository work
- `T018` and `T019` can run in parallel for US2 while US1 is being validated
- `T028` and `T029` can run in parallel for US3 while US2 is underway
- Polish tasks `T039`, `T041`, and `T042` can run in parallel after implementation is complete

---

## Parallel Example: User Story 1

```bash
Task: "T008 [US1] Add handler contract coverage in internal/handler/http/template_test.go"
Task: "T009 [US1] Add usecase integration coverage in internal/usecase/template/template_integration_test.go and internal/usecase/environment/environment_template_integration_test.go"

Task: "T010 [US1] Expand template domain request and response models in internal/model/template/type.go, internal/model/template_version/type.go, internal/model/template_parameter/type.go, internal/model/template_resource/type.go, and internal/model/template_instance/type.go"
Task: "T013 [US1] Extend environment create and read contracts in internal/model/environment/type.go"
```

## Parallel Example: User Story 2

```bash
Task: "T018 [US2] Add handler contract coverage in internal/handler/http/delivery_target_test.go and internal/handler/http/environment_movement_test.go"
Task: "T019 [US2] Add usecase integration coverage in internal/usecase/delivery_target/delivery_target_integration_test.go and internal/usecase/environment_movement/environment_movement_integration_test.go"

Task: "T021 [US2] Implement delivery target persistence in internal/repository/delivery_target/init.go and internal/repository/delivery_target/delivery_target.go"
Task: "T022 [US2] Implement environment movement persistence in internal/repository/environment_movement/init.go and internal/repository/environment_movement/environment_movement.go"
```

## Parallel Example: User Story 3

```bash
Task: "T028 [US3] Add handler contract coverage in internal/handler/http/live_update_test.go"
Task: "T029 [US3] Add usecase integration coverage in internal/usecase/live_update/live_update_integration_test.go"

Task: "T031 [US3] Implement notification persistence in internal/repository/notification/init.go and internal/repository/notification/notification.go"
Task: "T034 [US3] Implement workload log streaming in internal/usecase/live_update/log_stream.go and internal/repository/provisioner/informer.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run the US1 independent test using `internal/handler/http/template_test.go`, `internal/usecase/template/template_integration_test.go`, and `internal/usecase/environment/environment_template_integration_test.go`
5. Demo or ship the backend template workflow before starting cluster-targeting or live-update work

### Incremental Delivery

1. Complete Setup + Foundational → foundation ready
2. Add User Story 1 → validate template workflow independently → demo/deploy MVP
3. Add User Story 2 → validate target placement and movement independently → demo/deploy
4. Add User Story 3 → validate notification history, SSE, and log-stream behavior independently → demo/deploy
5. Finish Phase 6 polish and regression validation

### Parallel Team Strategy

With multiple developers:

1. One engineer completes Setup while another prepares test scaffolding
2. Team completes Foundational tasks together
3. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
4. Rejoin for polish, Swagger updates, and quickstart validation

---

## Notes

- All tasks use the required checklist format with checkbox, task ID, optional `[P]`, required story labels for story phases, and exact file paths
- Tasks touching authorization, contracts, movement state, notifications, or stream failure modes include explicit validation work to satisfy the constitution
- User story tasks are scoped so each story can be tested without requiring Developer Portal UI work in this repository
- MVP scope is User Story 1 after Setup and Foundational phases
