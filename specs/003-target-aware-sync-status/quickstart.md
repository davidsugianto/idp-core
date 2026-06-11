# Quickstart: Target-Aware Sync and Status Resolution Validation

## Prerequisites
- PostgreSQL is running and the API starts successfully.
- At least two delivery targets exist, each mapped to different control-plane metadata.
- You have a valid bearer token for a user with access to the target environments.
- The target control planes are reachable from the API runtime.

## 1. Run focused automated coverage
```bash
go test ./...
```

Expected result: the existing suite and new target-resolution coverage pass.

## 2. Verify delivery target metadata
Confirm the delivery targets expose distinct non-secret control-plane metadata.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/delivery-targets/$TARGET_A_ID

curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/delivery-targets/$TARGET_B_ID
```

Expected result: each target shows a different control-plane identifier or context, and no secrets appear in the response.

## 3. Create or identify two environments mapped to different targets
Each environment must already have an `argo_app_name` and a `delivery_target_id`.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_A_ID

curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_B_ID
```

Expected result: `ENV_A` is mapped to `TARGET_A` and `ENV_B` is mapped to `TARGET_B`.

## 4. Trigger sync for the first target
```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_A_ID/sync
```

Expected result: the request succeeds, and operator-visible logs or retained records show that `TARGET_A` and its control plane were selected without exposing tokens, kubeconfig contents, or other secrets.

## 5. Trigger sync for the second target
```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_B_ID/sync
```

Expected result: the request succeeds independently of `ENV_A`, and logs or retained records show that `TARGET_B` and its control plane were selected without cross-target leakage.

## 6. Read GitOps status from each target
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_A_ID/gitops/status

curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$ENV_B_ID/gitops/status
```

Expected result: each response comes from the correct target control plane, with no evidence of cross-target fallback, and any internal error text is sanitized before being returned.

## 7. Validate actionable failure for a broken target mapping
Use an environment whose target is missing required control-plane metadata, or temporarily configure a non-production test target to an unreachable control plane.

```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$BROKEN_ENV_ID/sync
```

Expected result: the request fails with a target-aware error that identifies the mapping or control-plane problem without exposing secrets.

## 8. Validate status failure for missing application in resolved control plane
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/environments/$BROKEN_ENV_ID/gitops/status
```

Expected result: the error explains that the application was not found in the resolved control plane, not that the system checked a different default target.

## 9. Validate unchanged single-target behavior
Run the same sync and status checks against an environment whose target resolves to the default shared control plane.

Expected result: the responses match current behavior while still passing through the target-resolution path.

## 10. Validate concurrent cross-target isolation
Trigger sync or status requests for `ENV_A` and `ENV_B` concurrently.

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8989/v1/environments/$ENV_A_ID/sync &
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8989/v1/environments/$ENV_B_ID/sync &
wait
```

Expected result: both requests resolve independently to their own targets and neither request reuses the other target's control plane.

## Success Checks
- Environments on different targets can be synced and queried for status independently in one running API instance.
- Missing or invalid target metadata produces actionable failures.
- Logs or retained records identify the selected delivery target/control plane for validated requests.
- No secrets, tokens, or kubeconfig contents appear in responses or logs.
