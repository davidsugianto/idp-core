# 📋 idp-core — Product Requirements Document (PRD) Phase 3

> **Project**: `idp-core`
> **Phase**: 3 - Platform
> **Owner**: Platform Engineering Team
> **Last Updated**: June 10, 2026
> **Status**: 🚧 Backend implementation in progress
> **Timeline**: Q4 2026

---

## 🎯 Executive Summary

Phase 3 extends idp-core with backend platform capabilities for reusable templates, multi-cluster placement, environment movement workflows, and authenticated real-time operational updates. The developer portal remains part of Phase 3, but it is delivered from the separate `idp-ui` repository and consumes the backend contracts defined here.

### Phase 3 Goals

| Goal | Metric | Target |
|------|--------|--------|
| Developer self-service | API-backed workflows ready for portal consumption | 100% Phase 3 backend contracts |
| Template standardization | Template usage | 100% environments |
| Multi-cluster support | Cluster coverage | All production clusters |
| Developer productivity | Time to production | < 1 hour |
| Real-time operations | Event and log stream availability | All active environments |

---

## 🏗️ Architecture Overview

```mermaid
flowchart TB
    subgraph Frontend ["🖥️ Developer Portal (Separate idp-ui Repo)"]
        Dashboard["Dashboard"]
        EnvBrowser["Environment Browser"]
        TemplateEditor["Template Editor"]
        WorkloadViewer["Workload Viewer"]
    end

    subgraph API ["🔧 idp-core API (Go/Gin)"]
        HTTP["HTTP Handlers"]
        SSE["SSE Streams"]
    end

    subgraph Clusters ["☸️ Multi-Cluster Delivery Targets"]
        Cluster1["Cluster 1 (prod)"]
        Cluster2["Cluster 2 (staging)"]
        Cluster3["Cluster 3 (dev)"]
    end

    subgraph Storage ["💾 Data Layer"]
        DB[(PostgreSQL)]
        Redis[(Redis)]
    end

    Frontend --> API
    API --> Clusters
    API --> Storage
    SSE --> Frontend

    style Frontend fill:#e3f2fd,stroke:#2196f3
    style API fill:#e8f5e9,stroke:#4caf50
    style Clusters fill:#fff3e0,stroke:#ff9800
    style Storage fill:#fce4ec,stroke:#e91e63
```

---

## 🖥️ Feature 1: Developer Portal UI

### Overview

Build a modern, responsive web application that provides developers with a self-service interface for all platform capabilities. This UI is implemented in the separate `idp-ui` repository, while `idp-core` delivers the backend APIs and streaming contracts it consumes.

**UI Project**: [`idp-ui`](https://github.com/davidsugianto/idp-ui) (separate repository)

### User Stories

| ID | User Story | Priority |
|----|------------|----------|
| UI-001 | As a developer, I want a dashboard showing my environments so I can quickly access them | P0 |
| UI-002 | As a developer, I want to create environments through a wizard so I don't need to learn the API | P0 |
| UI-003 | As a developer, I want to see real-time workload status so I can monitor my applications | P0 |
| UI-004 | As a team lead, I want to manage team members and permissions through the UI | P1 |
| UI-005 | As a developer, I want to view cost reports in charts so I can understand spending visually | P1 |
| UI-006 | As a platform admin, I want an admin console to manage clusters and templates | P1 |

### UI Components

#### Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│  IDP Platform                                    [User ▼]   │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │ Envs: 12 │  │ Active: 8│  │ Cost:$2.4k│  │ Alerts: 2│     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │
│                                                             │
│  Recent Environments                    [+ New Environment] │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Name          Team     Status    Cost    Updated    │   │
│  │ feature-auth  backend  ✅ Ready  $120    2h ago     │   │
│  │ staging-api   backend  ⚠ Syncing $340    5m ago     │   │
│  │ dev-frontend  frontend ✅ Ready  $89     1d ago     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Recommendations                        [View All →]        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 💡 Scale down api-deployment (save ~$45/mo)          │   │
│  │ 💡 Increase memory for worker-pod (OOM risk)         │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

#### Environment Browser

```
┌─────────────────────────────────────────────────────────────┐
│  Environments                                    [+ Create]  │
├─────────────────────────────────────────────────────────────┤
│  Filter: [All Teams ▼] [All Clusters ▼] [Status ▼]  [Search]│
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐
│  │ 📦 feature-auth                                        │
│  │ Team: backend  |  Cluster: prod-us-east  |  ✅ Ready   │
│  │ ─────────────────────────────────────────────────────── │
│  │ Workloads: 4  |  Cost: $120/mo  |  Last sync: 2h ago   │
│  │                                                         │
│  │ [View Details] [Sync Now] [View Logs] [Delete]         │
│  └─────────────────────────────────────────────────────────┘
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐
│  │ 📦 staging-api                                         │
│  │ Team: backend  |  Cluster: staging  |  ⚠ Syncing       │
│  │ ─────────────────────────────────────────────────────── │
│  │ Workloads: 6  |  Cost: $340/mo  |  Last sync: 5m ago   │
│  │                                                         │
│  │ [View Details] [Cancel Sync] [View Logs] [Delete]      │
│  └─────────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────────┘
```

#### Template Editor

```
┌─────────────────────────────────────────────────────────────┐
│  Template: microservice-base                     [Save] [Preview]
├─────────────────────────────────────────────────────────────┤
│  Name: [microservice-base        ]  Version: [1.2.0 ▼]       │
│  Description: [Standard microservice template              ]│
│                                                             │
│  Parameters                                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Name        Type      Default    Required    Actions │   │
│  │ app_name    string    -          ✓           [Edit]  │   │
│  │ cpu_limit   string    500m       ✗           [Edit]  │   │
│  │ mem_limit   string    512Mi      ✗           [Edit]  │   │
│  │ replicas    number    2          ✗           [Edit]  │   │
│  │ [+ Add Parameter]                                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Resources                                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 📄 deployment.yaml     [Edit] [Preview]              │   │
│  │ 📄 service.yaml        [Edit] [Preview]              │   │
│  │ 📄 configmap.yaml      [Edit] [Preview]              │   │
│  │ [+ Add Resource]                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Technical Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Framework | React | 18+ |
| Build Tool | Vite | 5+ |
| UI Library | Ant Design | 5+ |
| State Management | Zustand / React Query | Latest |
| API Client | Axios / React Query | Latest |
| Real-time | WebSocket / SSE | - |
| Charts | Recharts / ECharts | Latest |
| Forms | React Hook Form + Zod | Latest |
| Router | React Router | 6+ |

### Pages & Routes

| Route | Page | Description |
|-------|------|-------------|
| `/` | Dashboard | Overview with metrics and quick actions |
| `/environments` | EnvironmentList | Browse and filter environments |
| `/environments/:id` | EnvironmentDetail | Environment details and actions |
| `/environments/new` | EnvironmentCreate | Create environment wizard |
| `/templates` | TemplateList | Browse available templates |
| `/templates/:id` | TemplateDetail | Template details and versions |
| `/templates/:id/edit` | TemplateEditor | Edit template |
| `/costs` | CostExplorer | Cost analysis and reports |
| `/costs/team/:id` | TeamCosts | Team-specific cost view |
| `/workloads` | WorkloadList | All workloads across environments |
| `/workloads/:id` | WorkloadDetail | Workload details and logs |
| `/settings` | Settings | User preferences |
| `/settings/api-keys` | APIKeys | Manage API keys |
| `/admin` | AdminConsole | Admin dashboard |
| `/admin/users` | UserManagement | Manage users |
| `/admin/teams` | TeamManagement | Manage teams |
| `/admin/roles` | RoleManagement | Manage roles |
| `/admin/clusters` | ClusterManagement | Manage clusters |

---

## 📋 Feature 2: Template Management Backend

### Overview

Implement a backend template management system that allows platform teams to define reusable environment templates with parameters, version lifecycle controls, validation, and historical instantiation records.

### User Stories

| ID | User Story | Priority |
|----|------------|----------|
| TM-001 | As a platform admin, I want to create templates so teams can standardize environments | P0 |
| TM-002 | As a developer, I want to browse templates so I can choose the right starting point | P0 |
| TM-003 | As a platform admin, I want to version templates so I can manage updates safely | P1 |
| TM-004 | As a developer, I want template validation so I know my parameters are correct | P1 |
| TM-005 | As a platform admin, I want a template marketplace so teams can share templates | P2 |

### Data Model

```go
// Template represents an environment template
type Template struct {
    ID          string    `gorm:"primaryKey"`
    Name        string    `gorm:"uniqueIndex;not null"`
    Slug        string    `gorm:"uniqueIndex;not null"`
    Description string    `gorm:"not null"`
    Category    string    `gorm:"index"` // microservice, job, batch, etc.
    Author      string    `gorm:"not null"`
    AuthorEmail string
    Visibility  string    `gorm:"not null"` // public, team, private
    TeamID      string    `gorm:"index"`    // for team templates
    Tags        string    `gorm:"type:text[]"`
    Status      string    `gorm:"not null"` // active, deprecated, archived
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TemplateVersion represents a version of a template
type TemplateVersion struct {
    ID          string    `gorm:"primaryKey"`
    TemplateID  string    `gorm:"index;not null"`
    Version     string    `gorm:"not null"` // semver
    Description string
    Changelog   string
    IsLatest    bool      `gorm:"default:false"`
    IsStable    bool      `gorm:"default:false"`
    Status      string    `gorm:"not null"` // stable, beta, deprecated
    CreatedAt   time.Time
}

// TemplateParameter represents a template parameter
type TemplateParameter struct {
    ID            string    `gorm:"primaryKey"`
    TemplateID    string    `gorm:"index;not null"`
    VersionID     string    `gorm:"index;not null"`
    Name          string    `gorm:"not null"`
    DisplayName   string    `gorm:"not null"`
    Description   string
    Type          string    `gorm:"not null"` // string, number, boolean, enum, array
    Default       string
    Required      bool      `gorm:"default:false"`
    Validation    string    `gorm:"type:jsonb"` // regex, min, max, options, etc.
    Order         int       // display order
    CreatedAt     time.Time
}

// TemplateResource represents a Kubernetes resource in a template
type TemplateResource struct {
    ID          string    `gorm:"primaryKey"`
    TemplateID  string    `gorm:"index;not null"`
    VersionID   string    `gorm:"index;not null"`
    Name        string    `gorm:"not null"`
    Type        string    `gorm:"not null"` // deployment, service, configmap, etc.
    Content     string    `gorm:"type:text;not null"` // YAML/JSON template
    Order       int       // application order
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// TemplateInstance represents a template used to create an environment
type TemplateInstance struct {
    ID            string    `gorm:"primaryKey"`
    TemplateID    string    `gorm:"index;not null"`
    VersionID     string    `gorm:"index;not null"`
    EnvironmentID string    `gorm:"index;not null"`
    Parameters    string    `gorm:"type:jsonb"` // parameter values used
    CreatedAt     time.Time
}
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/templates` | List templates (filterable) |
| POST | `/v1/templates` | Create template |
| GET | `/v1/templates/:id` | Get template details |
| PATCH | `/v1/templates/:id` | Update template |
| DELETE | `/v1/templates/:id` | Delete template |
| GET | `/v1/templates/:id/versions` | List template versions |
| POST | `/v1/templates/:id/versions` | Create new version |
| GET | `/v1/templates/:id/versions/:versionId` | Get specific version |
| PATCH | `/v1/templates/:id/versions/:versionId` | Update version lifecycle |
| PUT | `/v1/templates/:id/versions/:versionId/parameters` | Replace parameter definitions |
| PUT | `/v1/templates/:id/versions/:versionId/resources` | Replace resource definitions |
| POST | `/v1/templates/:id/versions/:versionId/validate` | Validate template inputs |
| POST | `/v1/environments` | Create environment from a template version |

### Built-in Templates

| Template | Description | Use Case |
|----------|-------------|----------|
| `microservice-base` | Standard microservice with Deployment, Service, ConfigMap | Backend services |
| `frontend-app` | Frontend application with CDN, Ingress | Web applications |
| `cron-job` | Scheduled job with CronJob resource | Batch processing |
| `worker-queue` | Worker with HPA and queue consumer | Async processing |
| `database-statefulset` | StatefulSet with PVC and backups | Databases |
| `ml-training-job` | ML training job with GPU support | ML workloads |

---

## 🌐 Feature 3: Multi-Cluster Placement Backend

### Overview

Enable idp-core to manage environment placement across multiple Kubernetes clusters through delivery targets, target-aware provisioning, and auditable movement workflows.

### User Stories

| ID | User Story | Priority |
|----|------------|----------|
| MC-001 | As a platform admin, I want to register clusters so I can manage multiple environments | P0 |
| MC-002 | As a developer, I want to deploy to specific clusters so I can use the right infrastructure | P0 |
| MC-003 | As a platform admin, I want cluster health monitoring so I can detect issues | P1 |
| MC-004 | As a developer, I want to migrate environments between clusters so I can promote through stages | P1 |
| MC-005 | As a platform admin, I want federated policies so I can enforce standards across clusters | P2 |

### Data Model

```go
// Cluster represents a registered Kubernetes cluster
type Cluster struct {
    ID              string    `gorm:"primaryKey"`
    Name            string    `gorm:"uniqueIndex;not null"`
    DisplayName     string    `gorm:"not null"`
    Description     string
    Endpoint        string    `gorm:"not null"` // API server endpoint
    Region          string    `gorm:"index"`    // us-east-1, eu-west-1, etc.
    Environment     string    `gorm:"index"`    // production, staging, development
    Provider        string    `gorm:"not null"` // aws, gcp, azure, on-prem
    Status          string    `gorm:"not null"` // active, unhealthy, maintenance
    HealthStatus    string    // healthy, degraded, unknown
    LastHealthCheck *time.Time
    Version         string    // K8s version
    NodeCount       int
    CapacityCPU     string
    CapacityMemory  string
    Config          string    `gorm:"type:jsonb"` // cluster-specific config
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// ClusterNamespace represents a namespace tracked in a cluster
type ClusterNamespace struct {
    ID          string    `gorm:"primaryKey"`
    ClusterID   string    `gorm:"index;not null"`
    Namespace   string    `gorm:"not null"`
    EnvironmentID string  `gorm:"index"` // linked environment
    Status      string    `gorm:"not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// ClusterHealth represents cluster health metrics
type ClusterHealth struct {
    ID                string    `gorm:"primaryKey"`
    ClusterID         string    `gorm:"index;not null"`
    Timestamp         time.Time `gorm:"index;not null"`
    NodeHealth        string    `gorm:"type:jsonb"` // node status map
    PodHealth         string    `gorm:"type:jsonb"` // pod counts by status
    ResourceUsage     string    `gorm:"type:jsonb"` // CPU, memory, storage usage
    ControlPlaneHealth string  // healthy, degraded, unknown
    CreatedAt         time.Time
}
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/delivery-targets` | List registered delivery targets |
| POST | `/v1/delivery-targets` | Register new delivery target |
| GET | `/v1/delivery-targets/:id` | Get delivery target details |
| PATCH | `/v1/delivery-targets/:id` | Update delivery target config |
| DELETE | `/v1/delivery-targets/:id` | Remove delivery target |
| POST | `/v1/environments/:id/movements` | Request environment movement to another target |
| GET | `/v1/environments/:id/movements` | List movement history for an environment |
| GET | `/v1/environments/:id/movements/:movementId` | Get movement details |

---

## 🔌 Feature 4: Real-time Update Backend

### Overview

Implement authenticated server-sent event streams and retained notification history for environment status changes, movement progress, notifications, and workload log streaming.

### User Stories

| ID | User Story | Priority |
|----|------------|----------|
| RT-001 | As a developer, I want real-time environment status so I don't need to refresh | P0 |
| RT-002 | As a developer, I want to stream logs in the UI so I can debug issues | P0 |
| RT-003 | As a developer, I want deployment progress updates so I know when it's done | P1 |
| RT-004 | As a developer, I want notifications for important events so I stay informed | P1 |

### Streamed Events

| Event | Transport | Description |
|-------|-----------|-------------|
| `status` | SSE | Environment status change |
| `progress` | SSE | Deployment or movement progress update |
| `notification` | SSE | New notification event |
| `log` | SSE | Workload log line |

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/notifications` | List retained notifications |
| GET | `/v1/environments/:id/events/stream` | Stream status, progress, and notification events |
| GET | `/v1/environments/:id/logs/stream` | Stream workload logs |

---

## 🔧 Technical Requirements

### Database Migrations

```
migrations/
├── 20260608000000_create_templates_table.sql
├── 20260608000001_create_template_versions_table.sql
├── 20260608000002_create_template_parameters_table.sql
├── 20260608000003_create_template_resources_table.sql
├── 20260608000004_create_template_instances_table.sql
├── 20260608000005_create_delivery_targets_table.sql
├── 20260608000006_create_environment_movements_table.sql
├── 20260608000007_create_notifications_table.sql
├── 20260608000008_alter_environments_add_phase3_refs.sql
```

### Configuration Updates

```yaml
# Phase 3 backend additions to config.yaml
frontend:
  enabled: true
  base_url: "https://idp.example.com"
  api_base_url: "https://api.idp.example.com"

streaming:
  enabled: true
  heartbeat_interval: "15s"
  subscription_ttl: "30m"

templates:
  enabled: true

multi_cluster:
  enabled: true
  default_delivery_target: "prod-us-east"
```

### External Dependencies

| Dependency | Purpose | Version |
|------------|---------|---------|
| ArgoCD | GitOps (existing) | 2.11+ |
| Prometheus | Metrics (existing) | 2.45+ |
| Redis | Existing background/coordination infrastructure | Current project version |
| idp-ui | Separate frontend consumer of these APIs | Separate repository |

---

## 📊 Success Metrics

| KPI | Target | Measurement |
|-----|--------|-------------|
| Template Usage | 100% environments | Environments from templates |
| Multi-cluster Coverage | All prod clusters | Registered delivery targets vs total |
| Real-time Latency | < 5s | SSE event delivery time |
| Validation Speed | < 3s | Template validation response time |
| Notification Availability | 100% authenticated callers | Notification list and stream access |

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Frontend complexity | Medium | Use established UI library; component-based architecture |
| WebSocket scalability | Medium | Redis pub/sub for horizontal scaling |
| Multi-cluster auth | High | Use service accounts per cluster; rotate credentials |
| Template security | High | Validate templates; sandbox rendering; review process |
| UI performance | Medium | Code splitting; lazy loading; caching |

---

## 🗓️ Timeline

| Milestone | Duration | Deliverables |
|-----------|----------|--------------|
| M1: Frontend Foundation | 3 weeks | React app, dashboard, auth integration |
| M2: Environment UI | 2 weeks | Environment browser, create wizard, details |
| M3: Template System | 2 weeks | Template CRUD, versioning, instantiation |
| M4: Multi-Cluster | 2 weeks | Cluster registration, cross-cluster deploy |
| M5: Real-time & Polish | 1 week | WebSocket, notifications, testing |

**Total Duration**: ~10 weeks

---

## 📎 References

- [PRD Overview](./PRD.md)
- [PRD Phase 1](./PRD_PHASE_1.md)
- [PRD Phase 2](./PRD_PHASE_2.md)
- [IDP UI Project](https://github.com/davidsugianto/idp-ui) - Developer Portal frontend
- [React Documentation](https://react.dev/)
- [Ant Design](https://ant.design/)
- [WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)

---

## 📝 Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | May 21, 2026 | Platform Engineering | Initial Phase 3 PRD |
