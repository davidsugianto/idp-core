# 📋 idp-core — Development TODO List (Phase 3)

> **Phase**: 3 - Platform
> **Timeline**: Q4 2026 (~10 weeks)
> **Status**: 📋 Planning
> **Last Updated**: May 21, 2026
> **Frontend**: [idp-ui](https://github.com/davidsugianto/idp-ui)
> **Frontend TODO**: [idp-ui DEV_TODO_LIST_PHASE_1](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md)

---

## 📊 Progress Overview

| Milestone | Status | Progress |
|-----------|--------|----------|
| M1: Frontend Foundation | ⬜ Not Started | 0% |
| M2: Environment UI | ⬜ Not Started | 0% |
| M3: Template System | ⬜ Not Started | 0% |
| M4: Multi-Cluster Support | ⬜ Not Started | 0% |
| M5: Real-time & Polish | ⬜ Not Started | 0% |

---

## 🗓️ M1: Frontend Foundation (Week 1-3)

> **Status**: ⬜ Not Started
> **Repo**: [idp-ui](https://github.com/davidsugianto/idp-ui)
> **TODO**: [idp-ui DEV_TODO_LIST_PHASE_1](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md)

### Scope

Frontend project setup, Ant Design theme integration, routing, layout components, API client, state management, and OIDC authentication flow. All tasks are tracked in the [idp-ui TODO list](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md#-m1-project-setup--auth-week-1-2).

#### Key Deliverables

- [ ] Vite + React + TypeScript project scaffold
- [ ] Ant Design theme and design tokens
- [ ] AppLayout with Header, Sidebar, and routing
- [ ] Axios API client with auth interceptors
- [ ] Zustand stores (auth, ui)
- [ ] OIDC login flow with Keycloak
- [ ] ProtectedRoute and AdminRoute guards

---

## 🗓️ M2: Environment UI (Week 4-5)

> **Repo**: [idp-ui](https://github.com/davidsugianto/idp-ui)
> **TODO**: [idp-ui DEV_TODO_LIST_PHASE_1](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md#-m2-dashboard--environments-week-3-4)

### Scope

Dashboard page with stat cards, environment table, and recommendations. Environment list with filtering, search, and pagination. Environment detail page with tabs (Overview, Workloads, Logs, Settings). Environment creation wizard with multi-step form. All tasks are tracked in the [idp-ui TODO list](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md#-m2-dashboard--environments-week-3-4).

#### Key Deliverables

- [ ] DashboardPage with StatCard, EnvironmentTable, RecommendationList
- [ ] EnvironmentListPage with FilterBar, search, pagination
- [ ] EnvironmentDetailPage with tabs (Overview, Workloads, Logs)
- [ ] EnvironmentCreatePage with multi-step wizard
- [ ] Sync and Delete actions with confirmation modals

---

## 🗓️ M3: Template System (Week 6-7)

> **Repos**: [idp-core](https://github.com/davidsugianto/idp-core) + [idp-ui](https://github.com/davidsugianto/idp-ui)

### Week 6: Backend — Template Management

#### Database & Models

- [ ] Create migration: `templates` table
- [ ] Create migration: `template_versions` table
- [ ] Create migration: `template_parameters` table
- [ ] Create migration: `template_resources` table
- [ ] Create migration: `template_instances` table
- [ ] Create model: `internal/model/template/type.go`
- [ ] Create model: `internal/model/template_version/type.go`
- [ ] Create model: `internal/model/template_parameter/type.go`
- [ ] Create model: `internal/model/template_resource/type.go`
- [ ] Create model: `internal/model/template_instance/type.go`

#### Repository Layer

- [ ] Create `internal/repository/template/init.go` (interface + struct)
- [ ] Create `internal/repository/template/template.go` (CRUD)
- [ ] Create `internal/repository/template/version.go` (version management)
- [ ] Create `internal/repository/template/parameter.go` (parameter CRUD)
- [ ] Create `internal/repository/template/resource.go` (resource CRUD)
- [ ] Create `internal/repository/template/instance.go` (instantiation)

#### Usecase Layer

- [ ] Create `internal/usecase/template/init.go`
- [ ] Create `internal/usecase/template/template.go`
- [ ] Implement CreateTemplate (validate YAML resources)
- [ ] Implement CreateVersion (set IsLatest, IsStable)
- [ ] Implement ValidateParameters (Zod-like validation)
- [ ] Implement Instantiate (render template + apply params → create environment)
- [ ] Implement built-in templates seeder

#### Handler Layer

- [ ] Create `internal/handler/http/template.go`
- [ ] Add template routes in `cmd/http/server.go`
- [ ] Wire dependencies in `cmd/http/main.go`
- [ ] Add Swagger annotations

### Week 7: Frontend — Template UI

> **Repo**: [idp-ui](https://github.com/davidsugianto/idp-ui)
> **TODO**: [idp-ui DEV_TODO_LIST_PHASE_2](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_2.md) (future)

#### Key Deliverables

- [ ] TemplateListPage with browse and filter
- [ ] TemplateDetailPage with params, resources, versions
- [ ] TemplateEditor with ParameterForm and ResourcePreview
- [ ] Template instantiation flow (template → environment)

---

## 🗓️ M4: Multi-Cluster Support (Week 8-9)

> **Repos**: [idp-core](https://github.com/davidsugianto/idp-core) + [idp-ui](https://github.com/davidsugianto/idp-ui)

### Week 8: Backend — Cluster Management

#### Database & Models

- [ ] Create migration: `clusters` table
- [ ] Create migration: `cluster_namespaces` table
- [ ] Create migration: `cluster_health` table
- [ ] Create model: `internal/model/cluster/type.go`
- [ ] Create model: `internal/model/cluster_namespace/type.go`
- [ ] Create model: `internal/model/cluster_health/type.go`

#### Repository Layer

- [ ] Create `internal/repository/cluster/init.go`
- [ ] Create `internal/repository/cluster/cluster.go` (CRUD)
- [ ] Create `internal/repository/cluster/namespace.go`
- [ ] Create `internal/repository/cluster/health.go`

#### Kubernetes Client Abstraction (multi-cluster)

- [ ] Create `internal/pkg/k8s/cluster_manager.go` (manage multiple clients)
- [ ] Implement cluster registration (validate endpoint + credentials)
- [ ] Implement per-cluster client-go client creation
- [ ] Implement cluster health check (API server connectivity)
- [ ] Implement cluster health metric collection (nodes, pods, resources)
- [ ] Update environment provisioner to accept target cluster

#### Usecase Layer

- [ ] Create `internal/usecase/cluster/init.go`
- [ ] Create `internal/usecase/cluster/cluster.go`
- [ ] Implement RegisterCluster (validate + store config)
- [ ] Implement HealthCheck (parallel checks across clusters)
- [ ] Implement MigrateEnvironment (move namespace between clusters)

#### Handler Layer

- [ ] Create `internal/handler/http/cluster.go`
- [ ] Add cluster routes in `cmd/http/server.go`
- [ ] Wire dependencies in `cmd/http/main.go`

#### Cron Jobs

- [ ] Create `internal/handler/cron/cluster.go`
- [ ] Register `cluster-health-check` cron job

#### Configuration

- [ ] Add `multi_cluster` config section
- [ ] Add cluster health check schedule

### Week 9: Frontend — Cluster UI

> **Repo**: [idp-ui](https://github.com/davidsugianto/idp-ui)
> **TODO**: [idp-ui DEV_TODO_LIST_PHASE_2](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_2.md) (future)

#### Key Deliverables

- [ ] ClusterManagementPage (admin: list, register, health)
- [ ] Cluster detail view (nodes, namespaces, metrics)
- [ ] Environment migration UI

---

## 🗓️ M5: Real-time & Polish (Week 10)

> **Repos**: [idp-core](https://github.com/davidsugianto/idp-core) + [idp-ui](https://github.com/davidsugianto/idp-ui)

### Backend — WebSocket Server

- [ ] Add WebSocket endpoint `/ws` in idp-core
- [ ] Implement WebSocket handshake with JWT auth
- [ ] Implement Redis pub/sub for message broadcasting
- [ ] Implement event types (status, sync, workload, cluster, notification)
- [ ] Implement log streaming from Kubernetes pods
- [ ] Add heartbeat/ping mechanism
- [ ] Handle connection close and cleanup

### Backend — Notifications

- [ ] Create migration: `notifications` table
- [ ] Create migration: `user_sessions` table
- [ ] Create model: `internal/model/notification/type.go`
- [ ] Create `internal/repository/notification/init.go`
- [ ] Create `internal/usecase/notification/init.go`
- [ ] Implement notification creation on environment events
- [ ] Create `internal/handler/http/notification.go`
- [ ] Add notification API endpoints

### Frontend — WebSocket & Notifications

> **Repo**: [idp-ui](https://github.com/davidsugianto/idp-ui)
> **TODO**: [idp-ui DEV_TODO_LIST_PHASE_2](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_2.md) (future)

#### Key Deliverables

- [ ] `useWebSocket` hook with connect, reconnect, subscribe
- [ ] Real-time environment status and sync progress
- [ ] NotificationCenter with bell icon, badge, mark-as-read

### Frontend — Polish

> **Repo**: [idp-ui](https://github.com/davidsugianto/idp-ui)
> **TODO**: [idp-ui DEV_TODO_LIST_PHASE_1](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md#-m4-settings--polish-week-6)

#### Key Deliverables

- [ ] Loading skeletons on all pages
- [ ] Error boundaries at page level
- [ ] Empty states for empty lists
- [ ] Responsive layout for mobile/tablet
- [ ] Breadcrumbs navigation
- [ ] 404 page

### Final Checks

- [ ] Run backend tests: `go test ./...`
- [ ] Build backend for production: `docker build`
- [ ] Verify Swagger docs
- [ ] Frontend final checks → [idp-ui TODO](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md#final-checks)

---

## 📊 Definition of Done

Each task is considered complete when:

1. ✅ Code follows project conventions (clean architecture for backend, component-based for frontend)
2. ✅ Unit tests pass with > 80% coverage
3. ✅ Integration tests pass
4. ✅ Swagger documentation is updated (backend)
5. ✅ Code passes `golangci-lint` (backend) / `eslint` (frontend)
6. ✅ TypeScript compiles without errors (frontend)
7. ✅ Lighthouse score > 90 (frontend pages)
8. ✅ PR reviewed and merged

---

## 📦 Dependencies to Add

### Backend (idp-core)

```go
require (
    // WebSocket
    github.com/gorilla/websocket v1.5.1

    // Redis pub/sub (for WebSocket broadcasting)
    github.com/go-redis/redis/v8 v8.11.5

    // YAML validation for templates
    gopkg.in/yaml.v3 v3.0.1
    sigs.k8s.io/yaml v1.4.0
)
```

### Frontend (idp-ui)

```json
{
  "dependencies": {
    "react": "^18.3.0",
    "react-dom": "^18.3.0",
    "react-router-dom": "^6.23.0",
    "antd": "^5.20.0",
    "@ant-design/icons": "^5.4.0",
    "axios": "^1.7.0",
    "@tanstack/react-query": "^5.45.0",
    "zustand": "^4.5.0",
    "react-hook-form": "^7.51.0",
    "zod": "^3.23.0",
    "@hookform/resolvers": "^3.6.0",
    "dayjs": "^1.11.0",
    "recharts": "^2.12.0",
    "react-syntax-highlighter": "^15.5.0"
  },
  "devDependencies": {
    "typescript": "^5.5.0",
    "vite": "^5.3.0",
    "@vitejs/plugin-react": "^4.3.0",
    "vitest": "^1.6.0",
    "@testing-library/react": "^16.0.0",
    "@testing-library/jest-dom": "^6.4.0",
    "eslint": "^8.57.0",
    "@typescript-eslint/parser": "^7.0.0",
    "prettier": "^3.3.0"
  }
}
```

---

## 🔧 Configuration Updates

Add to `configs/config.yaml`:

```yaml
# Phase 3 additions
frontend:
  enabled: true
  base_url: "https://idp.example.com"
  api_base_url: "https://api.idp.example.com"

websocket:
  enabled: true
  path: "/ws"
  heartbeat_interval: "30s"
  write_timeout: "10s"

templates:
  enabled: true
  git_repo_url: "${TEMPLATE_GIT_REPO}"
  git_branch: "main"
  sync_interval: "5m"

multi_cluster:
  enabled: true
  health_check_interval: "1m"
  default_cluster: "prod-us-east"

cron:
  schedules:  # new schedule
    cluster-health-check: "*/1 * * * *"
```

---

## 📁 File Structure (Phase 3)

### Backend (idp-core)

```
migrations/
├── 20260601000000_create_templates_table.up.sql
├── 20260601000001_create_template_versions_table.up.sql
├── 20260601000002_create_template_parameters_table.up.sql
├── 20260601000003_create_template_resources_table.up.sql
├── 20260601000004_create_template_instances_table.up.sql
├── 20260601000005_create_clusters_table.up.sql
├── 20260601000006_create_cluster_namespaces_table.up.sql
├── 20260601000007_create_cluster_health_table.up.sql
├── 20260601000008_create_notifications_table.up.sql
├── 20260601000009_create_user_sessions_table.up.sql

internal/
├── model/
│   ├── template/type.go            # → CREATED
│   ├── template_version/type.go    # → CREATED
│   ├── template_parameter/type.go  # → CREATED
│   ├── template_resource/type.go   # → CREATED
│   ├── template_instance/type.go   # → CREATED
│   ├── cluster/type.go             # → CREATED
│   ├── cluster_namespace/type.go   # → CREATED
│   ├── cluster_health/type.go      # → CREATED
│   ├── notification/type.go        # → CREATED
│   └── user_session/type.go        # → CREATED
│
├── repository/
│   ├── template/                   # → CREATED
│   ├── cluster/                    # → CREATED
│   └── notification/               # → CREATED
│
├── usecase/
│   ├── template/                   # → CREATED
│   ├── cluster/                    # → CREATED
│   └── notification/               # → CREATED
│
├── handler/
│   ├── http/
│   │   ├── template.go             # → CREATED
│   │   ├── cluster.go              # → CREATED
│   │   ├── notification.go         # → CREATED
│   │   └── ws.go                   # → CREATED (WebSocket)
│   └── cron/
│       └── cluster.go              # → CREATED
│
├── pkg/
│   ├── k8s/cluster_manager.go      # → CREATED
│   └── ws/                         # → CREATED (WebSocket hub)
│
└── mocks/                          # → UPDATED
```

### Frontend (idp-ui)

See [idp-ui DEV_TODO_LIST_PHASE_1](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md) for the complete frontend file structure.

---

## 📎 References

- [PRD Phase 3](./prd/PRD_PHASE_3.md)
- [PRD Overview](./prd/PRD.md)
- [idp-ui Repository](https://github.com/davidsugianto/idp-ui)
- [idp-ui PRD Phase 1](https://github.com/davidsugianto/idp-ui/blob/main/docs/prd/PRD_PHASE_1.md)
- [idp-ui TODO Phase 1](https://github.com/davidsugianto/idp-ui/blob/main/docs/DEV_TODO_LIST_PHASE_1.md)

---

## 📝 Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | May 21, 2026 | Platform Engineering | Initial Phase 3 TODO list |