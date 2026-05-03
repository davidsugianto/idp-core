# On-Call Runbook

This runbook provides guidance for on-call engineers responding to alerts from idp-core.

## Table of Contents

1. [Alert Overview](#alert-overview)
2. [Common Alerts](#common-alerts)
3. [Troubleshooting](#troubleshooting)
4. [Escalation](#escalation)

## Alert Overview

idp-core exposes metrics at `/metrics` (Prometheus format) and health endpoints:
- `/health` - Liveness probe
- `/ready` - Readiness probe

### Key Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `http_request_duration_seconds` | HTTP request latency | p99 > 1s |
| `http_requests_total` | Total HTTP requests | Error rate > 5% |
| `environment_create_total` | Environments created | - |
| `environment_delete_total` | Environments deleted | - |
| `kubernetes_api_calls_total` | K8s API calls | Error rate > 10% |
| `argocd_api_calls_total` | ArgoCD API calls | Error rate > 10% |

## Common Alerts

### Alert: HighErrorRate

**Symptoms:** Error rate > 5% over 5 minutes

**Investigation:**
```bash
# Check application logs
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core --tail=100

# Check for recent errors
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core | grep -i error | tail -50

# Check pod status
kubectl get pods -n idp-core -l app.kubernetes.io/name=idp-core
```

**Common Causes:**
1. Database connectivity issues
2. Kubernetes API throttling
3. Invalid request payload

**Resolution:**
1. Check database connectivity
2. Check Kubernetes API server health
3. Review recent code deployments

---

### Alert: HighLatency

**Symptoms:** p99 latency > 1s over 5 minutes

**Investigation:**
```bash
# Check resource usage
kubectl top pods -n idp-core

# Check for slow queries
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core | grep -i "slow\|timeout"

# Check database connections
kubectl exec -n idp-core -it <pod> -- sh -c "pg_isready -h \$DB_HOST"
```

**Common Causes:**
1. Database query performance
2. Kubernetes API throttling
3. Memory pressure

**Resolution:**
1. Check for long-running queries in PostgreSQL
2. Increase resource limits if needed
3. Check for network issues

---

### Alert: PodCrashLooping

**Symptoms:** Pod restarting repeatedly

**Investigation:**
```bash
# Check pod status
kubectl describe pod -n idp-core <pod-name>

# Check previous logs
kubectl logs -n idp-core <pod-name> --previous

# Check events
kubectl get events -n idp-core --sort-by='.lastTimestamp'
```

**Common Causes:**
1. Failed database connection
2. Missing configuration/secrets
3. OOMKilled

**Resolution:**
1. Verify database is accessible
2. Check ConfigMap and Secret values
3. Increase memory limits if OOMKilled

---

### Alert: DatabaseConnectionFailed

**Symptoms:** Application cannot connect to database

**Investigation:**
```bash
# Check database pod
kubectl get pods -n idp-core -l app=postgres

# Check database service
kubectl get svc -n idp-core postgres

# Test connectivity from pod
kubectl exec -n idp-core -it <idp-core-pod> -- sh -c "nc -zv \$DB_HOST \$DB_PORT"

# Check credentials
kubectl get secret -n idp-core idp-core-secrets -o yaml
```

**Resolution:**
1. Verify database is running
2. Check network policies
3. Verify credentials in secret

---

### Alert: KubernetesAPIErrors

**Symptoms:** Errors calling Kubernetes API

**Investigation:**
```bash
# Check Kubernetes API server
kubectl get --raw='/healthz'

# Check RBAC permissions
kubectl auth can-i list namespaces --as=system:serviceaccount:idp-core:idp-core

# Check for throttling
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core | grep -i "throttl\|rate limit"
```

**Common Causes:**
1. RBAC permission issues
2. API server overload
3. Network issues

**Resolution:**
1. Verify ServiceAccount and RBAC
2. Check cluster health
3. Check for network policies blocking traffic

---

### Alert: ArgoCDIntegrationFailed

**Symptoms:** Errors creating/updating ArgoCD applications

**Investigation:**
```bash
# Check ArgoCD status
kubectl get pods -n argocd

# Check ArgoCD API server
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-server --tail=50

# Test ArgoCD Application creation
kubectl auth can-i create applications.argoproj.io --as=system:serviceaccount:idp-core:idp-core -n argocd

# Check existing ArgoCD applications
kubectl get applications -n argocd
```

**Common Causes:**
1. ArgoCD not installed
2. RBAC permission issues
3. Invalid Application spec

**Resolution:**
1. Verify ArgoCD is running
2. Check RBAC permissions for ArgoCD CRDs
3. Check ArgoCD logs for errors

## Troubleshooting

### Environment Creation Fails

```bash
# 1. Check idp-core logs
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core | grep -i "create\|environment"

# 2. Check if namespace was created
kubectl get ns | grep idp-

# 3. Check ArgoCD application
kubectl get application -n argocd | grep env-

# 4. Check events in target namespace
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
```

### Environment Deletion Stuck

```bash
# 1. Check environment status
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core | grep -i "delete"

# 2. Check if namespace exists
kubectl get ns <namespace>

# 3. Check for finalizers
kubectl get ns <namespace> -o jsonpath='{.spec.finalizers}'

# 4. Check ArgoCD application
kubectl get application -n argocd <app-name> -o yaml

# 5. Force delete if needed (caution!)
kubectl patch ns <namespace> -p '{"metadata":{"finalizers":[]}}' --type=merge
```

### Workload Status Not Updating

```bash
# 1. Check informer status
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core | grep -i "informer\|cache"

# 2. Restart pod to refresh informers
kubectl rollout restart deployment idp-core -n idp-core

# 3. Check for namespace access
kubectl auth can-i list pods -n <namespace> --as=system:serviceaccount:idp-core:idp-core
```

### Sync Not Triggering

```bash
# 1. Check ArgoCD application status
kubectl get application -n argocd <app-name> -o yaml

# 2. Check ArgoCD server logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-server --tail=100

# 3. Manual sync test
argocd app sync <app-name> --server argocd.example.com

# 4. Check Git repository access
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-repo-server --tail=100
```

## Useful Commands

### Quick Diagnostics

```bash
# Full status check
kubectl get all -n idp-core
kubectl get all -n argocd
kubectl get applications -n argocd

# Resource usage
kubectl top pods -n idp-core
kubectl top pods -n argocd

# Recent events
kubectl get events -n idp-core --sort-by='.lastTimestamp' | head -20

# Logs from all pods
kubectl logs -n idp-core -l app.kubernetes.io/name=idp-core --all-containers
```

### Database Operations

```bash
# Connect to PostgreSQL
kubectl exec -n idp-core -it <postgres-pod> -- psql -U idp_user -d idp_core

# Check environments table
kubectl exec -n idp-core -it <postgres-pod> -- psql -U idp_user -d idp_core -c "SELECT id, name, status FROM environments;"

# Check for locks
kubectl exec -n idp-core -it <postgres-pod> -- psql -U idp_user -d idp_core -c "SELECT * FROM pg_locks;"
```

### Kubernetes Operations

```bash
# List all idp-managed namespaces
kubectl get ns -l idp-core/managed-by=idp-core

# List all ArgoCD applications for environments
kubectl get applications -n argocd -l idp-core/managed-by=idp-core

# Check resource quotas
kubectl get resourcequota -A | grep idp-

# Check network policies
kubectl get networkpolicy -A | grep idp-
```

## Escalation

### Severity Levels

| Level | Response Time | Examples |
|-------|--------------|----------|
| P1 - Critical | 15 min | Service down, data loss |
| P2 - High | 1 hour | Degraded performance, partial outage |
| P3 - Medium | 4 hours | Single environment issue |
| P4 - Low | 24 hours | Minor issues, documentation |

### Escalation Path

1. **On-Call Engineer** - First responder
2. **Platform Team Lead** - P1/P2 escalation
3. **Engineering Manager** - Extended P1 incidents

### Communication Templates

**Incident Start:**
```
🚨 INCIDENT: [Title]
Severity: P[X]
Status: Investigating
Impact: [Description]
On-Call: [Name]
```

**Update:**
```
📊 UPDATE: [Title]
Status: [Investigating/Identified/Monitoring/Resolved]
Summary: [Brief update]
ETA: [Estimated resolution time]
```

**Resolution:**
```
✅ RESOLVED: [Title]
Duration: [Time]
Root Cause: [Brief description]
Action Items: [Follow-up tasks]
```

## Runbook Maintenance

- Review and update this runbook quarterly
- Add new alerts as they are discovered
- Document all incident postmortems
- Keep escalation contacts current
