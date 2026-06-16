# Quickstart: Environment Status Accuracy Validation

## Goal

Validate that environment operational endpoints return accurate namespace-scoped workload data for the requested environment and delivery target.

## Prerequisites

- `idp-core` API is running locally on `http://localhost:8989`
- PostgreSQL is running and reachable by the API
- Kubernetes cluster is running and reachable through the delivery target assigned to the test environment
- A test environment already exists with:
  - a valid `delivery_target_id`
  - a namespace that contains at least one workload for the primary validation flow
- A second environment or namespace is available for validating the empty-namespace path

## Fetch a bearer token

```bash
TOKEN=$(curl -s -X POST http://localhost:8989/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"test-user","team_id":"26bc2c5f-e250-4720-877a-a4b158bed1d4"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["token"])')
```

## Scenario 1: Validate environment status for a populated namespace

```bash
ENV_ID=<environment-id>

curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_ID/status
```

Expected results:
- response code is `200`
- `pod_summary.total` is greater than `0`
- `deployment_summary.desired` is greater than `0`
- `argo_status` remains populated when the environment has an ArgoCD application
- `namespace` and `delivery_target_id` match the requested environment

Cross-check with Kubernetes:

```bash
kubectl get deploy -n <namespace>
kubectl get pods -n <namespace>
```

The deployment and pod counts should align with the API summary.

## Scenario 2: Validate workload list for the same environment

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_ID/workloads
```

Expected results:
- response code is `200`
- `environment_id` equals `$ENV_ID`
- `namespace` is populated
- `summary.total_workloads` is greater than `0`
- `workloads` is a non-empty list

## Scenario 3: Validate workload detail from the workload list

1. Pick one workload name returned by Scenario 2.
2. Request its details:

```bash
WORKLOAD_NAME=<workload-name>

curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_ID/workloads/$WORKLOAD_NAME
```

Expected results:
- response code is `200`
- `name` matches `$WORKLOAD_NAME`
- `desired_replicas`, `ready_replicas`, and `image` align with the workload list response

## Scenario 4: Validate empty namespace behavior

Run the same `status` and `workloads` requests against an environment whose namespace exists but has no workloads.

Expected results:
- `/status` returns `200` with zero-valued workload summaries that accurately represent the empty namespace
- `/workloads` returns `200` with the correct `environment_id` and `namespace`, zero summary totals, and an empty `workloads` list
- neither endpoint returns blank identifiers or null workload data

## Scenario 5: Validate failure behavior

Exercise one of the following:
- break the delivery target mapping for a test environment
- stop the API's access to the resolved target cluster
- request a workload name that does not exist

Expected results:
- target-resolution or namespace-read failures return explicit errors instead of misleading zero/blank success payloads
- nonexistent workload names return a not-found style failure

## Related artifacts

- `specs/004-fix-environment-status/spec.md`
- `specs/004-fix-environment-status/plan.md`
- `specs/004-fix-environment-status/contracts/environment-operational-status.md`
- `specs/004-fix-environment-status/data-model.md`
