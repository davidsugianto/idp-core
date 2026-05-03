# Kubernetes Deployments

This directory contains Kubernetes manifests for deploying idp-core.

## Directory Structure

```
kubernetes/
├── base/                          # Base manifests
│   ├── namespace.yaml             # Namespace definition
│   ├── rbac.yaml                  # ServiceAccount, ClusterRole, ClusterRoleBinding
│   ├── configmap.yaml             # Application configuration
│   ├── secret.yaml                # Sensitive data (template)
│   ├── deployment.yaml            # Deployment with probes
│   ├── service.yaml               # ClusterIP services
│   └── kustomization.yaml         # Kustomize base
└── overlays/
    └── production/                # Production overrides
        └── kustomization.yaml
```

## Prerequisites

1. Kubernetes cluster (v1.28+)
2. ArgoCD installed in `argocd` namespace
3. PostgreSQL database (or use the provided StatefulSet)

## Quick Start

### Using kubectl

```bash
# Apply base manifests
kubectl apply -k base/

# Or apply production overlay
kubectl apply -k overlays/production/
```

### Using kustomize

```bash
# Build and preview
kustomize build base/

# Apply
kustomize build base/ | kubectl apply -f -
```

## Configuration

### Required Secrets

Before deploying, create the required secrets:

```bash
# Create namespace first
kubectl create namespace idp-core

# Create secrets
kubectl create secret generic idp-core-secrets \
  --from-literal=DB_USER=idp_user \
  --from-literal=DB_PASSWORD=your-secure-password \
  --from-literal=JWT_SECRET=your-jwt-secret \
  -n idp-core
```

### ConfigMap Overrides

Override configuration via ConfigMap or environment variables:

```bash
kubectl create configmap idp-core-config \
  --from-literal=DB_HOST=postgres.example.com \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_NAME=idp_core \
  -n idp-core
```

## Resource Requirements

| Environment | CPU Request | CPU Limit | Memory Request | Memory Limit | Replicas |
|-------------|-------------|-----------|----------------|--------------|----------|
| Development | 100m | 500m | 128Mi | 512Mi | 1 |
| Production | 200m | 1000m | 256Mi | 1Gi | 3 |

## Health Checks

| Probe | Endpoint | Initial Delay | Period |
|-------|----------|---------------|--------|
| Liveness | `/health` | 10s | 10s |
| Readiness | `/ready` | 5s | 5s |
| Startup | `/health` | 5s | 5s |

## RBAC Permissions

The ServiceAccount `idp-core` has the following permissions:

- **Namespaces**: create, get, list, watch, update, patch, delete
- **ResourceQuotas**: create, get, list, watch, update, patch, delete
- **NetworkPolicies**: create, get, list, watch, update, patch, delete
- **Pods**: get, list, watch
- **Deployments/StatefulSets/DaemonSets**: get, list, watch
- **ArgoCD Applications/AppProjects**: create, get, list, watch, update, patch, delete

See `base/rbac.yaml` for full details.

## Scaling

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: idp-core
  namespace: idp-core
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: idp-core
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

## Troubleshooting

```bash
# Check deployment status
kubectl get deployment idp-core -n idp-core

# Check pods
kubectl get pods -n idp-core -l app.kubernetes.io/name=idp-core

# View logs
kubectl logs -f -n idp-core -l app.kubernetes.io/name=idp-core

# Check events
kubectl get events -n idp-core --sort-by='.lastTimestamp'

# Describe pod
kubectl describe pod -n idp-core -l app.kubernetes.io/name=idp-core
```

## Security

- Runs as non-root user (UID 1000)
- Read-only root filesystem
- Drops all Linux capabilities
- No privilege escalation
