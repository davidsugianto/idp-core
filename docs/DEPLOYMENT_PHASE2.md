# Phase 2 Deployment Guide

This guide covers deployment of Phase 2 features: OIDC/RBAC, FinOps, Rightsizing, and Service Catalog.

## Prerequisites

### Infrastructure Requirements

| Component | Version | Purpose |
|-----------|---------|---------|
| Kubernetes | 1.25+ | Container orchestration |
| PostgreSQL | 15+ | Primary database |
| Redis Sentinel | 7.0+ | Distributed locking for cron jobs |
| ArgoCD | 2.6+ | GitOps deployment |
| OpenCost | 1.100+ | Cost monitoring |
| Prometheus | 2.40+ | Metrics & rightsizing data |

### External Services (Optional)

| Service | Purpose | Required For |
|---------|---------|--------------|
| OIDC Provider (Keycloak/Okta) | SSO authentication | OIDC flow |
| Slack | Budget alert notifications | Budget alerts |

---

## 1. OIDC & RBAC Deployment

### 1.1 OIDC Provider Configuration

Configure your OIDC provider (Keycloak example):

```yaml
# Keycloak Client Configuration
client_id: idp-core
client_secret: ${OIDC_CLIENT_SECRET}
redirect_uris:
  - https://idp.example.com/auth/callback
post_logout_redirect_uris:
  - https://idp.example.com/
web_origins:
  - https://idp.example.com

# Group Mapping
groups_claim: groups
admin_group: platform-admins
```

### 1.2 IDP-Core OIDC Configuration

```yaml
# configs/config.yaml
auth:
  jwt:
    secret: ${JWT_SECRET}
    expiry: 24h
  oidc:
    enabled: true
    issuer_url: https://keycloak.example.com/realms/idp
    client_id: idp-core
    client_secret: ${OIDC_CLIENT_SECRET}
    redirect_url: https://idp.example.com/auth/callback
    scopes:
      - openid
      - profile
      - email
      - groups
    groups_claim: groups
    admin_group: platform-admins
```

### 1.3 RBAC Seed Data

Default roles and permissions are seeded automatically on startup:

| Role | Permissions |
|------|-------------|
| `platform_admin` | Full access to all resources |
| `team_admin` | Manage team resources, members |
| `developer` | View/create environments, view costs |
| `viewer` | Read-only access |

Seed the platform admin user:

```bash
# Create platform admin user
curl -X POST https://idp.example.com/v1/users \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "name": "Platform Admin",
    "team_id": "platform"
  }'

# Assign platform_admin role
curl -X POST https://idp.example.com/v1/roles/assign \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid",
    "role_id": "platform_admin"
  }'
```

### 1.4 Kubernetes RBAC Sync

The OIDC groups are automatically synced to Kubernetes RBAC:

```yaml
# deployments/kubernetes/base/rbac-sync.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: idp-platform-admins
subjects:
  - kind: Group
    name: platform-admins
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
```

---

## 2. FinOps Deployment (Cost Tracking & Budgets)

### 2.1 Deploy OpenCost

```bash
# Install OpenCost via Helm
helm repo add opencost https://opencost.github.io/opencost-helm-chart
helm install opencost opencost/opencost \
  --namespace opencost \
  --create-namespace \
  --set opencost.prometheus.external.enabled=true \
  --set opencost.prometheus.external.name=prometheus-server.monitoring.svc.cluster.local
```

Verify OpenCost is running:

```bash
kubectl port-forward -n opencost svc/opencost 9003:9003
curl http://localhost:9003/allocation/compute?window=1d
```

### 2.2 Deploy Prometheus (if not exists)

```bash
# Install kube-prometheus-stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

### 2.3 Configure FinOps in IDP-Core

```yaml
# configs/config.yaml
finops:
  enabled: true
  opencost:
    base_url: http://opencost.opencost.svc.cluster.local:9003
    poll_interval: 1h
  prometheus:
    url: http://prometheus-server.monitoring.svc.cluster.local:80
```

### 2.4 Configure Slack for Budget Alerts

```yaml
# configs/config.yaml
slack:
  enabled: true
  webhook_url: ${SLACK_WEBHOOK_URL}
  channel: "#platform-alerts"
```

Create a Slack webhook:

1. Go to https://api.slack.com/apps
2. Create new app → Incoming Webhooks
3. Add webhook to workspace
4. Copy webhook URL to `SLACK_WEBHOOK_URL` secret

### 2.5 Create Budget

```bash
curl -X POST https://idp.example.com/v1/budgets \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Budget",
    "team_id": "team-1",
    "limit": 5000.00,
    "period": "monthly",
    "thresholds": [80, 90, 100],
    "channels": ["slack"]
  }'
```

---

## 3. Rightsizing Deployment

### 3.1 Prometheus Metrics Requirements

Ensure Prometheus is scraping kube-state-metrics and cAdvisor:

```yaml
# Prometheus scrape config
- job_name: 'kubernetes-pods'
  kubernetes_sd_configs:
    - role: pod
  relabel_configs:
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
      action: keep
      regex: true
```

### 3.2 Configure Rightsizing

```yaml
# configs/config.yaml
rightsizing:
  enabled: true
  analysis_interval: 24h
  lookback_days: 7
  cpu_utilization_threshold: 0.5  # Scale down if < 50% utilized
  memory_utilization_threshold: 0.5
  cpu_safety_buffer: 1.2  # Add 20% buffer to recommendations
  memory_safety_buffer: 1.3  # Add 30% buffer to recommendations
```

### 3.3 Cron Job Schedule

The rightsizing recommendation generator runs as a cron job:

```yaml
# configs/config.yaml
cron:
  schedules:
    rightsizing-generate: "0 6 * * *"  # Daily at 6 AM
    cost-sync: "0 * * * *"             # Hourly
    budget-alert-check: "*/15 * * * *" # Every 15 minutes
```

### 3.4 Apply Recommendations

View recommendations:

```bash
curl https://idp.example.com/v1/rightsizing/recommendations?status=pending \
  -H "Authorization: Bearer ${TOKEN}"
```

Apply a recommendation:

```bash
curl -X POST https://idp.example.com/v1/rightsizing/recommendations/${REC_ID}/apply \
  -H "Authorization: Bearer ${TOKEN}"
```

Rollback if needed:

```bash
curl -X POST https://idp.example.com/v1/rightsizing/recommendations/${REC_ID}/rollback \
  -H "Authorization: Bearer ${TOKEN}"
```

---

## 4. Resource Quotas Deployment

### 4.1 Deploy Admission Webhook

The admission webhook is deployed as part of idp-core:

```yaml
# deployments/kubernetes/base/webhook.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: idp-admission-webhook
webhooks:
  - name: validate.idp.example.com
    rules:
      - apiGroups: [""]
        apiVersions: ["v1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["pods"]
    clientConfig:
      service:
        name: idp-webhook
        namespace: idp-core
        path: /validate
    admissionReviewVersions: ["v1"]
    sideEffects: None
```

### 4.2 Create Resource Quota

```bash
curl -X POST https://idp.example.com/v1/quotas \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "team-1-production",
    "team_id": "team-1",
    "cpu_request_limit": "4",
    "memory_request_limit": "8Gi",
    "pod_count_limit": 50,
    "enforce": true
  }'
```

### 4.3 Verify Quota Enforcement

When a pod exceeds quota, the admission webhook rejects it:

```json
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "response": {
    "allowed": false,
    "status": {
      "message": "quota exceeded: cpu_request limit 4, current 3.8, requested 0.5"
    }
  }
}
```

---

## 5. Service Catalog Deployment

### 5.1 No additional infrastructure required

Service Catalog uses the existing PostgreSQL database.

### 5.2 Register a Service

```bash
curl -X POST https://idp.example.com/v1/services \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "user-service",
    "description": "User management microservice",
    "team_id": "team-1",
    "visibility": "team"
  }'
```

### 5.3 Add Service Version

```bash
curl -X POST https://idp.example.com/v1/services/${SERVICE_ID}/versions \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.2.0",
    "git_ref": "refs/tags/v1.2.0",
    "changelog": "Added user profile endpoints"
  }'
```

### 5.4 Add Endpoint

```bash
curl -X POST https://idp.example.com/v1/services/${SERVICE_ID}/versions/${VERSION_ID}/endpoints \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://user-service.example.com",
    "type": "http",
    "description": "Production endpoint"
  }'
```

### 5.5 Add Dependency

```bash
curl -X POST https://idp.example.com/v1/services/${SERVICE_ID}/dependencies \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "depends_on_service_id": "database-service",
    "dependency_type": "runtime"
  }'
```

### 5.6 Deploy Version to Environment

```bash
curl -X POST https://idp.example.com/v1/services/${SERVICE_ID}/versions/${VERSION_ID}/deploy \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "environment_id": "env-production"
  }'
```

---

## 6. Redis Sentinel Deployment

Redis Sentinel is required for distributed locking in cron jobs.

### 6.1 Deploy Redis Sentinel

```bash
# Using Helm
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install redis bitnami/redis \
  --namespace redis \
  --create-namespace \
  --set sentinel.enabled=true \
  --set auth.password=${REDIS_PASSWORD}
```

### 6.2 Configure IDP-Core

```yaml
# configs/config.yaml
redis:
  master_name: mymaster
  address: redis-sentinel.redis.svc.cluster.local:26379
  password: ${REDIS_PASSWORD}
```

---

## 7. Cron Server Deployment

The cron server handles scheduled jobs for cost sync, budget alerts, and rightsizing.

### 7.1 Deploy Cron Server

```bash
kubectl apply -k deployments/kubernetes/overlays/production
```

The cron server runs as a separate deployment:

```yaml
# deployments/kubernetes/base/cron-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: idp-cron
spec:
  replicas: 1  # Single replica with Redis locking
  template:
    spec:
      containers:
        - name: cron
          image: idp-core-cron:v2
          ports:
            - containerPort: 8983
```

### 7.2 Manual Job Trigger

Jobs can be triggered manually via HTTP:

```bash
# Trigger cost sync
curl -X POST http://cron-service:8983/cost-sync

# Trigger rightsizing generation
curl -X POST http://cron-service:8983/rightsizing-generate

# Trigger budget alert check
curl -X POST http://cron-service:8983/budget-alert-check
```

---

## 8. Complete Deployment Checklist

### Pre-Deployment

- [ ] PostgreSQL database created and migrated
- [ ] Redis Sentinel deployed
- [ ] Kubernetes cluster ready
- [ ] ArgoCD installed and configured
- [ ] OpenCost deployed and connected to Prometheus
- [ ] Prometheus scraping all namespaces
- [ ] OIDC provider configured (if using OIDC)
- [ ] Slack webhook created (if using budget alerts)

### Deployment

- [ ] Deploy idp-core API server
- [ ] Deploy idp-core cron server
- [ ] Deploy admission webhook
- [ ] Verify health endpoints
- [ ] Run database migrations
- [ ] Seed default roles and permissions

### Post-Deployment

- [ ] Create platform admin user
- [ ] Assign platform_admin role
- [ ] Create initial teams
- [ ] Create initial environments
- [ ] Set up resource quotas for namespaces
- [ ] Configure budgets for teams
- [ ] Verify cost sync is working
- [ ] Verify budget alerts are firing
- [ ] Verify rightsizing recommendations are generated

---

## 9. Monitoring & Troubleshooting

### Health Checks

```bash
# API server health
curl https://idp.example.com/ping

# Cron server health
curl http://cron-service:8983/ping

# OpenCost health
curl http://opencost:9003/healthz
```

### Log Inspection

```bash
# API server logs
kubectl logs -n idp-core deployment/idp-api

# Cron server logs
kubectl logs -n idp-core deployment/idp-cron

# Admission webhook logs
kubectl logs -n idp-core deployment/idp-webhook
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Cost sync not working | OpenCost unreachable | Check OpenCost service and Prometheus connection |
| Budget alerts not firing | Slack webhook invalid | Verify SLACK_WEBHOOK_URL secret |
| Rightsizing empty | Prometheus not scraping | Check Prometheus targets and metrics |
| Quota not enforced | Webhook not registered | Verify ValidatingWebhookConfiguration |
| OIDC login fails | Invalid client config | Check issuer URL and client secret |
| Cron jobs not running | Redis lock held | Check Redis Sentinel, may need lock cleanup |

---

## 10. Security Considerations

### Secrets Management

Store all secrets in Kubernetes Secrets or external secret manager:

```yaml
# Required secrets
JWT_SECRET
OIDC_CLIENT_SECRET
REDIS_PASSWORD
SLACK_WEBHOOK_URL
DATABASE_URL
```

### Network Policies

Restrict traffic between components:

```yaml
# Only allow idp-core to access OpenCost
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: opencost-access
spec:
  podSelector:
    matchLabels:
      app: opencost
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: idp-api
```

### RBAC in Kubernetes

Limit idp-core service account permissions:

```yaml
# Only allow get/list/watch on pods/deployments
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: idp-core
rules:
  - apiGroups: [""]
    resources: ["pods", "namespaces", "resourcequotas"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets"]
    verbs: ["get", "list", "watch", "update"]
```
