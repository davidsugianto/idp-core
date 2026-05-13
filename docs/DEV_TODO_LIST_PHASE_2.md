# 📋 idp-core — Development TODO List (Phase 2)

> **Phase**: 2 - Enhancement\
> **Timeline**: Q3 2026 (\~10 weeks)\
> **Status**: 🔄 In Development\
> **Last Updated**: May 2026

***

## 📊 Progress Overview

| Milestone            | Status         | Progress |
| -------------------- | -------------- | -------- |
| M1: Auth & RBAC      | 🔄 In Progress | 85%      |
| M2: FinOps           | 🔄 In Progress | 25%      |
| M3: Rightsizing      | 🔲 Not Started | 0%       |
| M4: Service Catalog  | 🔲 Not Started | 0%       |
| M5: Testing & Polish | 🔲 Not Started | 0%       |

***

## 🗓️ M1: Auth & RBAC (Week 1-3)

> **Implementation**: ✅ Complete (all models, repositories, usecases, handlers, middleware, unit tests)
> **Remaining**: Integration tests (6 suites), platform admin seed script

### Week 1: User & Team Management ✅ COMPLETED

#### Database & Models

- [x] Create migration: `users` table
- [x] Create migration: `teams` table
- [x] Create migration: `team_members` table
- [x] Create model: `internal/model/user/type.go`
- [x] Create model: `internal/model/team/type.go`
- [x] Create model: `internal/model/team_member/type.go` (included in team/type.go)

#### Repository Layer

- [x] Create `internal/repository/user/init.go` (interface + struct)
- [x] Create `internal/repository/user/user.go` (CRUD methods)
- [x] Create `internal/repository/team/init.go`
- [x] Create `internal/repository/team/team.go`
- [x] Create `internal/repository/team/member.go`

#### Usecase Layer

- [x] Create `internal/usecase/user/init.go`
- [x] Create `internal/usecase/user/user.go`
- [x] Create `internal/usecase/team/init.go`
- [x] Create `internal/usecase/team/team.go`

#### Handler Layer

- [x] Create `internal/handler/http/user.go`
- [x] Create `internal/handler/http/team.go`
- [x] Add routes in `cmd/http/server.go`

#### Tests

- [x] Unit tests: user repository (mock created)
- [x] Unit tests: team repository (mock created)
- [x] Unit tests: user usecase
- [x] Unit tests: team usecase
- [ ] Integration tests: user API
- [ ] Integration tests: team API

***

### Week 2: OIDC Integration & RBAC ✅ COMPLETED

#### OIDC Integration

- [x] Add dependency: `github.com/coreos/go-oidc`
- [x] Create `internal/pkg/oidc/client.go`
- [x] Create `internal/pkg/oidc/verifier.go`
- [x] Update config: OIDC settings in `config.go`
- [x] Create auth middleware: `internal/handler/http/middleware/oidc.go`

#### Role & Permission Models

- [x] Create migration: `roles` table
- [x] Create migration: `permissions` table
- [x] Create migration: `role_permissions` table
- [x] Create migration: `user_roles` table
- [x] Create model: `internal/model/role/type.go`
- [x] Create model: `internal/model/permission/type.go`

#### RBAC Repository

- [x] Create `internal/repository/role/init.go`
- [x] Create `internal/repository/role/role.go`
- [x] Create `internal/repository/role/member.go`
- [x] Create `internal/repository/permission/init.go`
- [x] Create `internal/repository/permission/permission.go`

#### RBAC Usecase

- [x] Create `internal/usecase/role/init.go`
- [x] Create `internal/usecase/role/role.go`
- [x] Create RBAC engine: `internal/usecase/auth/rbac.go`

#### RBAC Handler

- [x] Create `internal/handler/http/role.go`
- [x] Update auth middleware with RBAC check
- [x] Add role routes in `cmd/http/server.go`

#### Seed Default Data

- [x] Create seed script for default roles
- [x] Create seed script for default permissions
- [ ] Create seed script for platform admin user

#### Tests

- [x] Unit tests: OIDC client
- [x] Unit tests: role repository (mock created)
- [x] Unit tests: RBAC engine
- [ ] Integration tests: OIDC flow
- [ ] Integration tests: RBAC enforcement

***

### Week 3: API Keys & Audit Logging ✅ COMPLETED

#### API Key Models

- [x] Create migration: `api_keys` table
- [x] Create model: `internal/model/apikey/type.go`

#### API Key Repository

- [x] Create `internal/repository/apikey/init.go`
- [x] Create `internal/repository/apikey/apikey.go`

#### API Key Usecase

- [x] Create `internal/usecase/apikey/init.go`
- [x] Create `internal/usecase/apikey/apikey.go`
- [x] Implement key generation & hashing

#### API Key Handler

- [x] Create `internal/handler/http/apikey.go`
- [x] Add API key authentication middleware
- [x] Add API key routes

#### Audit Logging

- [x] Create migration: `audit_logs` table
- [x] Create model: `internal/model/audit_log/type.go`
- [x] Create `internal/repository/auditlog/init.go`
- [x] Create `internal/repository/auditlog/auditlog.go`
- [x] Create `internal/usecase/auditlog/init.go`
- [x] Create `internal/usecase/auditlog/auditlog.go`
- [x] Create audit middleware for automatic logging
- [x] Create `internal/handler/http/auditlog.go`

#### Tests

- [x] Unit tests: API key generation
- [x] Unit tests: audit log repository
- [ ] Integration tests: API key auth
- [ ] Integration tests: audit log retrieval

***

## 🗓️ M2: FinOps (Week 4-5)

### Week 4: OpenCost Integration & Cost Tracking ✅ COMPLETED

#### Configuration

- [x] Add OpenCost config to `config.go`
- [x] Add Prometheus config to `config.go`

#### OpenCost Client

- [x] Create `internal/pkg/opencost/client.go`
- [x] Create `internal/pkg/opencost/types.go` (AllocationRequest, AllocationResponse, AllocationData)

#### Prometheus Client

- [x] Create `internal/pkg/prometheus/client.go` (stub for future queries)

#### Cost Models

- [x] Create migration: `cost_records` table
- [x] Create model: `internal/model/cost/type.go`

#### Cost Repository

- [x] Create `internal/repository/cost/init.go` (interface + implementation)
- [x] Implement Create, BatchCreate, List, GetByTeamAndPeriod

#### Cost Usecase

- [x] Create `internal/usecase/cost/init.go`
- [x] Create `internal/usecase/cost/cost.go`
- [x] Implement cost sync job (time.Ticker goroutine in main.go)

#### Cost Handler

- [x] Create `internal/handler/http/cost.go`
- [x] Add cost routes (`GET /v1/costs`, `GET /v1/costs/team`)

#### Tests

- [x] Unit tests: OpenCost client (5 tests with httptest.Server)
- [x] Unit tests: cost usecase (11 tests with mocked repo + opencost client)
- [ ] Integration tests: cost API

***

### Week 5: Budget Management & Alerts

#### Budget Models

- [ ] Create migration: `budgets` table
- [ ] Create migration: `budget_alerts` table
- [ ] Create model: `internal/model/budget/type.go`
- [ ] Create model: `internal/model/budget_alert/type.go`

#### Budget Repository

- [ ] Create `internal/repository/budget/init.go`
- [ ] Create `internal/repository/budget/budget.go`

#### Budget Usecase

- [ ] Create `internal/usecase/budget/init.go`
- [ ] Create `internal/usecase/budget/budget.go`
- [ ] Implement budget alert checker (cron)
- [ ] Implement notification sender (email/Slack)

#### Budget Handler

- [ ] Create `internal/handler/http/budget.go`
- [ ] Add budget routes

#### Cost Export

- [ ] Implement CSV export functionality
- [ ] Add export endpoint

#### Tests

- [ ] Unit tests: budget repository
- [ ] Unit tests: budget alert logic
- [ ] Integration tests: budget API
- [ ] Integration tests: alert triggering

***

## 🗓️ M3: Rightsizing (Week 6-7)

### Week 6: Rightsizing Recommendations

#### Configuration

- [ ] Add rightsizing config to `config.go`

#### Rightsizing Models

- [ ] Create migration: `rightsizing_recommendations` table
- [ ] Create model: `internal/model/rightsizing/type.go`

#### Rightsizing Repository

- [ ] Create `internal/repository/rightsizing/init.go`
- [ ] Create `internal/repository/rightsizing/rightsizing.go`

#### Rightsizing Usecase

- [ ] Create `internal/usecase/rightsizing/init.go`
- [ ] Create `internal/usecase/rightsizing/rightsizing.go`
- [ ] Implement usage analyzer (query Prometheus)
- [ ] Implement recommendation generator
- [ ] Implement recommendation scheduler (cron)

#### Rightsizing Handler

- [ ] Create `internal/handler/http/rightsizing.go`
- [ ] Add rightsizing routes

#### K8s Integration

- [ ] Implement apply recommendation (update workload)
- [ ] Handle rollback on failure

#### Tests

- [ ] Unit tests: usage analyzer
- [ ] Unit tests: recommendation generator
- [ ] Integration tests: rightsizing API
- [ ] E2E test: apply recommendation

***

### Week 7: Resource Quotas

#### Resource Quota Models

- [ ] Create migration: `resource_quotas` table
- [ ] Create model: `internal/model/resource_quota/type.go`

#### Resource Quota Repository

- [ ] Create `internal/repository/quota/init.go`
- [ ] Create `internal/repository/quota/quota.go`

#### Resource Quota Usecase

- [ ] Create `internal/usecase/quota/init.go`
- [ ] Create `internal/usecase/quota/quota.go`
- [ ] Implement usage calculator
- [ ] Implement quota enforcement (admission webhook)

#### Resource Quota Handler

- [ ] Create `internal/handler/http/quota.go`
- [ ] Add quota routes

#### Admission Webhook Update

- [ ] Update webhook to check quotas
- [ ] Return quota exceeded error

#### Tests

- [ ] Unit tests: quota repository
- [ ] Unit tests: quota enforcement
- [ ] Integration tests: quota API
- [ ] E2E test: quota enforcement

***

## 🗓️ M4: Service Catalog (Week 8-9)

### Week 8: Service Registration & Discovery

#### Service Models

- [ ] Create migration: `services` table
- [ ] Create migration: `service_versions` table
- [ ] Create migration: `service_endpoints` table
- [ ] Create model: `internal/model/service/type.go`
- [ ] Create model: `internal/model/service_version/type.go`
- [ ] Create model: `internal/model/service_endpoint/type.go`

#### Service Repository

- [ ] Create `internal/repository/service/init.go`
- [ ] Create `internal/repository/service/service.go`
- [ ] Create `internal/repository/service/version.go`

#### Service Usecase

- [ ] Create `internal/usecase/service/init.go`
- [ ] Create `internal/usecase/service/service.go`

#### Service Handler

- [ ] Create `internal/handler/http/service.go`
- [ ] Add service routes

#### Tests

- [ ] Unit tests: service repository
- [ ] Unit tests: service usecase
- [ ] Integration tests: service API

***

### Week 9: Dependencies & Environments

#### Dependency Model

- [ ] Create migration: `service_dependencies` table
- [ ] Create model: `internal/model/service_dependency/type.go`

#### Service Environment Model

- [ ] Create migration: `service_environments` table
- [ ] Create model: `internal/model/service_environment/type.go`

#### Repository Extensions

- [ ] Add dependency methods to service repository
- [ ] Add environment methods to service repository

#### Usecase Extensions

- [ ] Add dependency management methods
- [ ] Add environment deployment tracking

#### Handler Extensions

- [ ] Add dependency routes
- [ ] Add service environment routes

#### Dependency Visualization

- [ ] Create endpoint for dependency graph
- [ ] Format for frontend consumption

#### Tests

- [ ] Unit tests: dependency logic
- [ ] Integration tests: dependency API
- [ ] Integration tests: environment tracking

***

## 🗓️ M5: Testing & Polish (Week 10)

### Integration Testing

- [ ] Create comprehensive integration test suite
- [ ] Test all Phase 2 API endpoints
- [ ] Test OIDC flow end-to-end
- [ ] Test RBAC enforcement across features
- [ ] Test cost sync and budget alerts
- [ ] Test rightsizing recommendations

### E2E Testing

- [ ] Create E2E test scenarios for Phase 2
- [ ] Test full user journey: login → create env → view costs
- [ ] Test admin journey: manage roles → set budgets → view audit logs

### Performance Testing

- [ ] Load test auth endpoints
- [ ] Load test cost queries
- [ ] Identify and fix bottlenecks

### Documentation

- [ ] Update Swagger/OpenAPI specs
- [ ] Update `README.md` with Phase 2 features
- [ ] Update `DEV_GUIDELINE.md` with new patterns
- [ ] Create Phase 2 deployment guide
- [ ] Update `TEST.md` with new test scenarios

### Security Review

- [ ] Review OIDC implementation for vulnerabilities
- [ ] Review RBAC for privilege escalation risks
- [ ] Review API key handling
- [ ] Review audit log for sensitive data

### Final Checks

- [ ] Run all tests: `go test ./...`
- [ ] Run linter: `golangci-lint run`
- [ ] Check test coverage: `go test -cover ./...`
- [ ] Verify Swagger docs render correctly
- [ ] Test Docker build: `docker build -t idp-core:v2 .`

***

## 📦 Dependencies to Add

```go
// Phase 2 dependencies
require (
    // OIDC
    github.com/coreos/go-oidc v1.2.1
    golang.org/x/oauth2 v0.20.0
    
    // Password hashing
    golang.org/x/crypto v0.23.0
    
    // API Key generation
    github.com/google/uuid v1.6.0
    
    // Prometheus client (for queries)
    github.com/prometheus/client_golang v1.19.0
    github.com/prometheus/common v0.53.0
    
    // CSV export
    encoding/csv (stdlib)
    
    // Slack webhook (optional)
    github.com/slack-go/slack v0.13.0
)
```

***

## 🔧 Configuration Updates

Add to `configs/config.yaml`:

```yaml
# Phase 2 configuration
auth:
  oidc:
    enabled: true
    provider: "${OIDC_PROVIDER}"
    issuer_url: "${OIDC_ISSUER_URL}"
    client_id: "${OIDC_CLIENT_ID}"
    client_secret: "${OIDC_CLIENT_SECRET}"
    redirect_url: "${OIDC_REDIRECT_URL}"
    scopes:
      - openid
      - profile
      - email
      - groups
    groups_claim: "groups"
    admin_group: "platform-admins"

finops:
  enabled: true
  opencost:
    base_url: "${OPENCOST_URL:http://opencost-cost-analyzer.opencost.svc.cluster.local:9003}"
    api_key: "${OPENCOST_API_KEY}"
  prometheus:
    url: "${PROMETHEUS_URL:http://prometheus-server.monitoring.svc.cluster.local:80}"
  sync_interval: "5m"

rightsizing:
  enabled: true
  analysis_interval: "1h"
  lookback_days: 7
  recommendation_ttl: "168h"

service_catalog:
  enabled: true
  default_visibility: "team"
```

***

## 📁 File Structure (Phase 2)

```
internal/
├── handler/http/
│   ├── user.go              # ✅ CREATED
│   ├── team.go              # ✅ CREATED
│   ├── role.go              # ✅ CREATED
│   ├── apikey.go            # ✅ CREATED
│   ├── auditlog.go          # ✅ CREATED
│   ├── cost.go              # ✅ CREATED
│   ├── budget.go            # TODO
│   ├── rightsizing.go       # TODO
│   ├── quota.go             # TODO
│   └── service.go           # TODO
│
├── usecase/
│   ├── user/                # ✅ CREATED
│   ├── team/                # ✅ CREATED
│   ├── role/                # ✅ CREATED
│   ├── apikey/              # ✅ CREATED
│   ├── auditlog/            # ✅ CREATED
│   ├── auth/                # ✅ CREATED (RBAC engine)
│   ├── cost/                # ✅ CREATED
│   ├── budget/              # TODO
│   ├── rightsizing/         # TODO
│   ├── quota/               # TODO
│   └── service/             # TODO
│
├── repository/
│   ├── user/                # ✅ CREATED
│   ├── team/                # ✅ CREATED
│   ├── role/                # ✅ CREATED
│   ├── permission/          # ✅ CREATED
│   ├── apikey/              # ✅ CREATED
│   ├── auditlog/            # ✅ CREATED
│   ├── cost/                # ✅ CREATED
│   ├── budget/              # TODO
│   ├── rightsizing/         # TODO
│   ├── quota/               # TODO
│   └── service/             # TODO
│
├── model/
│   ├── user/                # ✅ CREATED
│   ├── team/                # ✅ CREATED
│   ├── role/                # ✅ CREATED
│   ├── permission/          # ✅ CREATED
│   ├── apikey/              # ✅ CREATED
│   ├── auditlog/            # ✅ CREATED
│   ├── cost/                # ✅ CREATED
│   ├── budget/              # TODO
│   ├── rightsizing/         # TODO
│   ├── resource_quota/      # TODO
│   └── service/             # TODO
│
├── pkg/
│   ├── oidc/                # ✅ CREATED
│   ├── opencost/            # ✅ CREATED
│   └── prometheus/          # ✅ CREATED (stub)
│
└── mocks/
    ├── user_repository.go       # ✅ CREATED
    ├── team_repository.go       # ✅ CREATED
    ├── role_repository.go       # ✅ CREATED
    ├── permission_repository.go # ✅ CREATED
    ├── apikey_repository.go     # ✅ CREATED
    ├── auditlog_repository.go   # ✅ CREATED
    ├── cost_repository.go       # ✅ CREATED
    └── opencost_client.go       # ✅ CREATED

migrations/
├── 20260501000000_create_users_table.sql        # ✅ CREATED
├── 20260501000001_create_teams_table.sql        # ✅ CREATED
├── 20260501000002_create_team_members_table.sql # ✅ CREATED
├── 20260502000000_create_roles_table.sql        # ✅ CREATED
├── 20260502000001_create_permissions_table.sql  # ✅ CREATED
├── 20260502000002_create_role_permissions_table.sql # ✅ CREATED
├── 20260502000003_create_user_roles_table.sql   # ✅ CREATED
├── 20260512000000_create_api_keys_table.sql     # ✅ CREATED
├── 20260512000001_create_audit_logs_table.sql   # ✅ CREATED
├── 20260513000000_create_cost_records_table.sql # ✅ CREATED
└── ... (remaining Phase 2 migrations - TODO)
```

***

## ✅ Definition of Done

Each task is considered complete when:

1. ✅ Code is written and follows clean architecture
2. ✅ Unit tests pass with > 80% coverage
3. ✅ Integration tests pass
4. ✅ Swagger documentation is updated
5. ✅ Code passes `golangci-lint`
6. ✅ Code is reviewed and merged

***

## 📝 Notes

- Start with M1 (Auth & RBAC) as it's foundational for other features
- OIDC integration may require external provider setup (Keycloak/Okta)
- OpenCost integration requires OpenCost to be deployed in cluster
- Budget alerts require email/Slack integration setup
- Consider feature flags for gradual rollout

***

## ✅ Completed Work

### M1 Week 1: User & Team Management (May 2026)

**Files Created:**

- `migrations/20260501000000_create_users_table.sql`
- `migrations/20260501000001_create_teams_table.sql`
- `migrations/20260501000002_create_team_members_table.sql`
- `internal/model/user/type.go`
- `internal/model/team/type.go`
- `internal/repository/user/init.go`, `user.go`
- `internal/repository/team/init.go`, `team.go`, `member.go`
- `internal/usecase/user/init.go`, `user.go`
- `internal/usecase/team/init.go`
- `internal/handler/http/user.go`
- `internal/handler/http/team.go`
- `internal/mocks/user_repository.go`
- `internal/mocks/team_repository.go`
- `internal/usecase/user/user_test.go`
- `internal/usecase/team/team_test.go`

**API Endpoints Added:**

| Method | Endpoint                        | Description                |
| ------ | ------------------------------- | -------------------------- |
| GET    | `/v1/users`                     | List users with pagination |
| POST   | `/v1/users`                     | Create a new user          |
| GET    | `/v1/users/:id`                 | Get user details           |
| PATCH  | `/v1/users/:id`                 | Update user                |
| DELETE | `/v1/users/:id`                 | Delete user                |
| PUT    | `/v1/users/:id/status`          | Update user status         |
| GET    | `/v1/teams`                     | List teams with pagination |
| POST   | `/v1/teams`                     | Create a new team          |
| GET    | `/v1/teams/:id`                 | Get team details           |
| PATCH  | `/v1/teams/:id`                 | Update team                |
| DELETE | `/v1/teams/:id`                 | Delete team                |
| GET    | `/v1/teams/:id/members`         | List team members          |
| POST   | `/v1/teams/:id/members`         | Add team member            |
| PATCH  | `/v1/teams/:id/members/:userId` | Update member role         |
| DELETE | `/v1/teams/:id/members/:userId` | Remove team member         |

### M1 Week 2: OIDC Integration & RBAC (May 2026)

**Files Created:**

- `migrations/20260502000000_create_roles_table.sql`
- `migrations/20260502000001_create_permissions_table.sql`
- `migrations/20260502000002_create_role_permissions_table.sql`
- `migrations/20260502000003_create_user_roles_table.sql`
- `internal/model/role/type.go`, `user_role.go`
- `internal/model/permission/type.go`
- `internal/repository/role/init.go`, `role.go`, `member.go`
- `internal/repository/permission/init.go`, `permission.go`
- `internal/usecase/role/init.go`, `role.go`, `role_test.go`
- `internal/usecase/auth/init.go`, `rbac.go`
- `internal/handler/http/role.go`
- `internal/handler/http/middleware/oidc.go`
- `internal/pkg/oidc/client.go`, `verifier.go`
- `internal/pkg/config/config.go` (updated with OIDC settings)
- `internal/mocks/role_repository.go`
- `internal/mocks/permission_repository.go`
- `internal/seed/permission.go`, `role.go`

**API Endpoints Added:**

| Method | Endpoint                                    | Description                |
| ------ | ------------------------------------------- | -------------------------- |
| GET    | `/v1/roles`                                 | List roles with pagination |
| POST   | `/v1/roles`                                 | Create a new role          |
| GET    | `/v1/roles/:id`                             | Get role details           |
| PATCH  | `/v1/roles/:id`                             | Update role                |
| DELETE | `/v1/roles/:id`                             | Delete role                |
| POST   | `/v1/roles/assign`                          | Assign role to user        |
| POST   | `/v1/roles/revoke`                          | Revoke role from user      |
| GET    | `/v1/users/:user_id/roles`                  | Get user's roles           |
| GET    | `/v1/teams/:team_id/members/:user_id/roles` | Get user's team roles      |

### M1 Week 3: API Keys & Audit Logging (May 2026)

**Files Created:**

- `migrations/20260512000000_create_api_keys_table.sql`
- `migrations/20260512000001_create_audit_logs_table.sql`
- `internal/model/auditlog/type.go` (AuditLog, Map, constants, request/response types, converters)
- `internal/repository/auditlog/init.go`, `auditlog.go`
- `internal/usecase/apikey/init.go`, `apikey.go`, `apikey_test.go`
- `internal/usecase/auditlog/init.go`, `auditlog_test.go`
- `internal/handler/http/apikey.go`
- `internal/handler/http/auditlog.go`
- `internal/handler/http/middleware/apikey.go` (API key auth middleware)
- `internal/handler/http/middleware/audit.go` (automatic audit logging middleware)
- `internal/mocks/apikey_repository.go`
- `internal/mocks/auditlog_repository.go`

**Existing Files (already from Phase 1):**

- `internal/model/apikey/type.go`
- `internal/repository/apikey/init.go`

**API Endpoints Added:**

| Method | Endpoint             | Description                |
| ------ | -------------------- | -------------------------- |
| GET    | `/v1/api-keys`       | List API keys              |
| POST   | `/v1/api-keys`       | Create a new API key       |
| GET    | `/v1/api-keys/:id`   | Get API key details        |
| PATCH  | `/v1/api-keys/:id`   | Update API key             |
| DELETE | `/v1/api-keys/:id`   | Delete (revoke) API key    |
| GET    | `/v1/audit-logs`     | List audit logs (filtered) |
| GET    | `/v1/audit-logs/:id` | Get audit log entry        |

**Key Features:**

- API key generation with `idp_` prefix and hex-encoded random suffix
- SHA-256 hashing for storage; plain key returned only on creation
- API key auth middleware supporting both `Authorization: Bearer` and `X-API-Key` headers
- Automatic audit logging middleware on all `/v1` routes (fire-and-forget)
- Audit log filtering by user, team, action, resource type, status, and date range

**Next Steps:**

- Add integration tests for role API, API key auth, and audit log retrieval
- Complete seed script for platform admin user
- Begin M2: FinOps (Week 4)

### M2 Week 4: OpenCost Integration & Cost Tracking (May 2026)

**Files Created:**

- `migrations/20260513000000_create_cost_records_table.sql`
- `internal/model/cost/type.go` (CostRecord, CostRecordResponse, CostListResponse, CostFilter, converters)
- `internal/repository/cost/init.go` (interface + implementation: Create, BatchCreate, List, GetByTeamAndPeriod)
- `internal/usecase/cost/init.go`, `cost.go`, `cost_test.go`
- `internal/handler/http/cost.go`
- `internal/pkg/opencost/client.go`, `types.go`
- `internal/pkg/prometheus/client.go` (stub for future queries)
- `internal/mocks/cost_repository.go`
- `internal/mocks/opencost_client.go`

**Existing Files Modified:**

- `internal/pkg/config/config.go` — added FinOpsConfig, OpenCostConfig, PrometheusConfig
- `configs/config.yaml` — added finops section
- `internal/handler/http/init.go` — added costUseCase
- `cmd/http/server.go` — added CostUseCase, cost routes
- `cmd/http/main.go` — wired opencost/prometheus clients, cost repo/usecase, sync goroutine

**API Endpoints Added:**

| Method | Endpoint         | Description                         |
| ------ | ---------------- | ----------------------------------- |
| GET    | `/v1/costs`      | List cost records with filtering    |
| GET    | `/v1/costs/team` | Get team cost records by time range |

**Key Features:**

- OpenCost Allocation API client with configurable base URL, API key, and timeout
- Cost data synced via `time.Ticker` goroutine (gated by `finops.enabled` config)
- Namespace-to-team mapping via OpenCost allocation labels
- Prometheus client stub for future rightsizing queries
- Cost records stored with NUMERIC(12,4) precision for all cost columns
- JSONB raw\_data column preserving original OpenCost response
- Indexes on `(team_id, period_start)` and `(namespace, period_start)` for query performance

**Next Steps:**

- Begin M2: FinOps (Week 5) — Budget Management & Alerts
- Integration tests: cost API

***

## 📎 References

- [PRD Phase 2](./prd/PRD_PHASE_2.md)
- [PRD Overview](./prd/PRD.md)
- [Development Guidelines](./DEV_GUIDELINE.md)
- [Test Documentation](./TEST.md)

