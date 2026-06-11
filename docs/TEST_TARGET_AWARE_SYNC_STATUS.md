# Target-Aware Sync/Status E2E Test

This document captures the end-to-end validation flow for target-aware environment sync and GitOps status in `idp-core`.

It covers the validated path where:

- the environment resolves its assigned delivery target
- the API selects the correct target-specific ArgoCD control plane
- the generated ArgoCD `Application` uses the correct destination cluster server
- the production Kustomize overlay renders successfully
- ArgoCD sync/status can be read through `/v1/environments/:id/gitops/status`

## Goal

Validate that one running `idp-core` instance can:

1. create an environment on a delivery target
2. create an ArgoCD `Application` against the correct destination cluster
3. trigger sync through the resolved control plane
4. return GitOps status from that same resolved control plane

## Prerequisites

- `idp-core` API is running locally on `http://localhost:8989`
- PostgreSQL is running and reachable by the API
- Kubernetes cluster is running
- ArgoCD is installed in namespace `argocd`
- The delivery target used for the test already exists
- The Git repository revision used by ArgoCD contains the production overlay fix

## Required Target State

For same-cluster ArgoCD setups, the delivery target should use:

```text
cluster_server = https://kubernetes.default.svc
```

This value must match what ArgoCD recognizes as the destination cluster.

Do not use the ArgoCD API URL as `cluster_server`.

For example, this is wrong for `Application.spec.destination.server`:

```text
http://argocd-server.argocd.svc.cluster.local:80
```

That value is the ArgoCD server address, not the destination cluster identifier.

## Delivery Target Verification

Fetch a bearer token:

```bash
TOKEN=$(curl -s -X POST http://localhost:8989/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"test-user","team_id":"26bc2c5f-e250-4720-877a-a4b158bed1d4"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["token"])')
```

Inspect the target:

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/delivery-targets/9f65d098-8ae6-4ac3-9616-fd969a41e155
```

Expected fields:

```json
{
  "cluster_server": "https://kubernetes.default.svc",
  "argocd_namespace": "argocd",
  "argocd_server": "http://argocd-server.argocd.svc.cluster.local:80"
}
```

## Create a Fresh Validation Environment

```bash
NAME=e2e-target-aware-$(date +%H%M%S)

curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  http://localhost:8989/v1/environments \
  -d "{\"name\":\"$NAME\",\"git_repo_url\":\"https://github.com/davidsugianto/idp-core.git\",\"manifest_path\":\"deployments/kubernetes/overlays/production\",\"git_revision\":\"003-target-aware-sync-status\",\"delivery_target_id\":\"9f65d098-8ae6-4ac3-9616-fd969a41e155\"}"
```

Expected response:

- `201 Created`
- `cluster_server` is `https://kubernetes.default.svc`
- `argo_app_name` is populated

## Verify Generated ArgoCD Application

```bash
kubectl get applications.argoproj.io -n argocd <ARGO_APP_NAME> -o yaml
```

Expected spec:

```yaml
spec:
  destination:
    server: https://kubernetes.default.svc
```

This confirms the API is using the destination cluster server semantics, not the ArgoCD API URL.

## Trigger Sync

```bash
curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/<ENV_ID>/sync
```

Expected response:

```json
{
  "code": 200,
  "data": {
    "message": "sync triggered"
  }
}
```

## Read GitOps Status

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/<ENV_ID>/gitops/status
```

Expected successful shape:

```json
{
  "code": 200,
  "data": {
    "sync_status": "Synced",
    "health_status": "Progressing",
    "revision": "<git-sha>",
    "message": "successfully synced (all tasks run)"
  }
}
```

`health_status` may remain `Progressing` while workloads are still reconciling, but the target-aware routing is considered successful once:

- the sync reaches ArgoCD
- the revision is populated
- the previous destination-cluster registration error is gone

## ArgoCD Checks

Confirm the application status directly:

```bash
kubectl get applications.argoproj.io -n argocd <ARGO_APP_NAME> \
  -o jsonpath='{.status.sync.status}{"\n"}{.status.health.status}{"\n"}{.status.operationState.message}{"\n"}'
```

Expected output pattern:

```text
Synced
Progressing
successfully synced (all tasks run)
```

## Common Failure Modes

### 1. Cluster not configured

Example:

```text
cluster 'https://idp-test-control-plane:6443' has not been configured
```

Meaning:
- `Application.spec.destination.server` does not match a cluster ArgoCD recognizes

Fix:
- for same-cluster setups, use `https://kubernetes.default.svc`
- update the delivery target `cluster_server`
- recreate or refresh the application

### 2. ArgoCD API URL used as destination server

Example:

```text
cluster 'http://argocd-server.argocd.svc.cluster.local:80' has not been configured
```

Meaning:
- the system used `argocd_server` where it should have used `cluster_server`

Fix:
- ensure environment creation uses the destination cluster server

### 3. Kustomize overlay build failure

Example:

```text
id resid.ResId ... Name:"idp-core-config" ... does not exist; cannot merge or replace
```

Meaning:
- the production overlay references a broken `configMapGenerator` merge target

Fix:
- use the patched overlay in `deployments/kubernetes/overlays/production/kustomization.yaml`
- verify locally with:

```bash
kubectl kustomize deployments/kubernetes/overlays/production
```

### 4. Stale ArgoCD manifest cache

If the repo content is fixed and pushed but ArgoCD still reports the old manifest error, force a refresh and, if needed, restart ArgoCD components:

```bash
kubectl -n argocd patch application <ARGO_APP_NAME> --type merge \
  -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"hard"}}}'

kubectl -n argocd rollout restart deploy/argocd-repo-server
kubectl -n argocd rollout restart statefulset/argocd-application-controller
```

Then re-run sync/status.

## Success Criteria

The target-aware E2E path is considered healthy when all of these are true:

- delivery target exposes the expected non-secret control-plane metadata
- created environment stores the expected `delivery_target_id`
- generated ArgoCD `Application` uses `https://kubernetes.default.svc`
- sync endpoint returns `200`
- `/gitops/status` returns `200`
- the old `cluster has not been configured` error does not appear
- ArgoCD reports a real sync result instead of a target-resolution failure

## Related Files

- `internal/usecase/environment/environment.go`
- `internal/usecase/environment/environment_test.go`
- `deployments/kubernetes/overlays/production/kustomization.yaml`
- `specs/003-target-aware-sync-status/quickstart.md`
