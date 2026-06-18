# Tasks: Environment Status Accuracy

**Input**: Design documents from `/specs/004-fix-environment-status/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No separate test-file tasks are generated because the specification did not explicitly require TDD. Validation is covered through implementation verification and the runnable flows in `specs/004-fix-environment-status/quickstart.md`.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Use the real idp-core repository paths from plan.md
- Backend code lives under `cmd/`, `internal/`, `configs/`, and `deployments/`
- Keep handler/usecase/repository/model work in separate tasks when they touch different files

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Align feature-facing validation and contract references before implementation

- [x] T001 Capture implementation-time validation assumptions for populated and empty namespace scenarios in `specs/004-fix-environment-status/quickstart.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core environment operational semantics that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Add explicit workload/status unavailable error semantics in `internal/usecase/environment/init.go` and `internal/usecase/environment/environment.go`
- [x] T003 [P] Verify and, if needed, correct delivery-target-aware provisioner resolution for `/v1/environments/{id}/status`, `/v1/environments/{id}/workloads`, and `/v1/environments/{id}/workloads/{name}` in `internal/usecase/environment/environment.go`
- [x] T004 [P] Verify team-scoped environment lookup and unchanged authz/error semantics for affected environment endpoints in `internal/usecase/environment/environment.go` and `internal/handler/http/environment.go`
- [x] T005 [P] Add environment status summary helper structures for authoritative namespace-derived counts in `internal/model/environment/type.go`
- [x] T006 [P] Add workload response constructors that preserve environment and namespace context for empty namespaces in `internal/model/workload/type.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in priority order

---

## Phase 3: User Story 1 - View accurate environment status (Priority: P1) 🎯 MVP

**Goal**: Make `/v1/environments/{id}/status` return trustworthy pod and deployment summaries for the resolved environment namespace

**Independent Test**: Request `/v1/environments/{id}/status` for one environment with workloads and one empty namespace, then confirm the summary values match live namespace state and only return zeros for a truly empty namespace.

### Implementation for User Story 1

- [x] T007 [US1] Implement authoritative namespace summary derivation and cache-miss fallback in `internal/usecase/environment/environment.go`
- [x] T008 [US1] Update environment status response shaping in `internal/model/environment/type.go`
- [x] T009 [US1] Align `/v1/environments/{id}/status` error handling and contract annotations in `internal/handler/http/environment.go`

**Checkpoint**: User Story 1 should now return accurate environment status independently of the workload list endpoints

---

## Phase 4: User Story 2 - Browse environment workloads (Priority: P2)

**Goal**: Make `/v1/environments/{id}/workloads` always return the requested environment and namespace context with accurate workload and pod summaries

**Independent Test**: Request `/v1/environments/{id}/workloads` for one populated namespace and one empty namespace, then confirm `environment_id`, `namespace`, summary totals, and `workloads` contents are correct in both cases.

### Implementation for User Story 2

- [x] T010 [US2] Implement consistent workload list assembly from resolved namespace deployments and pods in `internal/usecase/environment/environment.go`
- [x] T011 [US2] Update workload summary and empty-list response shaping in `internal/model/workload/type.go`
- [x] T012 [US2] Align `/v1/environments/{id}/workloads` error handling and contract annotations in `internal/handler/http/environment.go`

**Checkpoint**: User Stories 1 and 2 should now both return consistent namespace-scoped operational data

---

## Phase 5: User Story 3 - Inspect an individual workload (Priority: P3)

**Goal**: Make `/v1/environments/{id}/workloads/{name}` resolve workload details from the same target-aware namespace context as the workload list and related environment operational endpoints

**Independent Test**: Request `/v1/environments/{id}/workloads/{name}` for a workload returned by the list endpoint and for a nonexistent workload name, then confirm the detail response matches the list data and missing workloads return not found.

### Implementation for User Story 3

- [x] T013 [US3] Align workload detail resolution and missing-workload handling in `internal/usecase/environment/environment.go`
- [x] T014 [US3] Update workload detail response contract and not-found handling in `internal/handler/http/environment.go`
- [x] T015 [US3] Review `/v1/environments/{id}/gitops/status`, `/v1/environments/{id}/logs/stream`, and `/v1/environments/{id}/events/stream` for namespace/target-resolution and error-semantic consistency in `internal/usecase/environment/environment.go` and `internal/handler/http/environment.go`

**Checkpoint**: All three user stories should now be independently functional and consistent with each other

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation alignment across all stories

- [x] T016 [P] Refresh feature documentation to match the final endpoint behavior in `specs/004-fix-environment-status/contracts/environment-operational-status.md` and `specs/004-fix-environment-status/quickstart.md`
- [x] T017 Run the end-to-end validation scenarios from `specs/004-fix-environment-status/quickstart.md` and capture any needed follow-up adjustments in `specs/004-fix-environment-status/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational completion
- **User Story 2 (Phase 4)**: Depends on Foundational completion and should build on the namespace summary semantics established in User Story 1
- **User Story 3 (Phase 5)**: Depends on Foundational completion and should build on the workload list semantics established in User Story 2
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start immediately after Foundational - MVP slice
- **User Story 2 (P2)**: Depends on the shared response semantics from Foundational and should follow User Story 1 for consistency
- **User Story 3 (P3)**: Depends on the workload list behavior from User Story 2

### Within Each User Story

- Shared usecase semantics before handler error/contract updates
- Model response shaping before final validation
- Story-specific endpoint behavior before cross-endpoint review

### Parallel Opportunities

- T005 and T006 can run in parallel after T002 through T004 because they touch different model files
- T016 can run after all implementation tasks while validation prep for T017 is happening

---

## Parallel Example: Foundational Work

```bash
Task: "Add environment status summary helper structures in internal/model/environment/type.go"
Task: "Add workload response constructors in internal/model/workload/type.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run the `/status` quickstart scenarios for populated and empty namespaces

### Incremental Delivery

1. Complete Setup + Foundational
2. Deliver User Story 1 and validate `/status`
3. Deliver User Story 2 and validate `/workloads`
4. Deliver User Story 3 and validate `/workloads/{name}` plus adjacent endpoint consistency
5. Finish with documentation and end-to-end validation

### Parallel Team Strategy

With multiple developers:

1. One developer completes T002 while another prepares T001
2. After T002 through T004, split T005 and T006 across two developers
3. Finish user stories in priority order to avoid conflicting edits in `internal/usecase/environment/environment.go`

---

## Notes

- All tasks follow the required checklist format with task ID, optional parallel marker, story label where required, and exact file paths
- The feature centers on existing environment endpoints, so most implementation work converges in `internal/usecase/environment/environment.go` and `internal/handler/http/environment.go`
- Empty namespace success behavior must stay distinct from target-resolution or workload-read failures
- Delivery-target-aware workload resolution must remain consistent with the target-aware sync/status work already planned in this repository
