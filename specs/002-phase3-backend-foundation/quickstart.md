# Quickstart: Phase 3 Backend Foundation Validation

## Prerequisites
- PostgreSQL is running for the local API.
- The API starts successfully with `go run ./cmd/http`.
- Authentication is configured and you have a bearer token for a platform admin and a team-scoped developer.
- Kubernetes/provisioner access is available for environment status and log validation.

## 1. Run baseline tests
```bash
go test ./...
```

Expected result: existing suites pass before Phase 3 validation begins.

## 2. Create a reusable template
Use the admin token.

```bash
curl -X POST http://localhost:8989/v1/templates \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "golden-service",
    "description": "Reusable service environment",
    "category": "service",
    "author": "Platform Team",
    "author_email": "platform@example.com",
    "visibility": "team",
    "team_id": "TEAM_ID"
  }'
```

Expected result: `201` response with a template ID.

## 3. Publish a template version and validate inputs
Create a version, define parameters/resources, then validate a sample payload.

```bash
curl -X POST http://localhost:8989/v1/templates/$TEMPLATE_ID/versions \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"version":"v1.0.0","description":"Initial release","is_latest":true,"is_stable":true,"status":"stable"}'
```

```bash
curl -X POST http://localhost:8989/v1/templates/$TEMPLATE_ID/versions/$VERSION_ID/validate \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"inputs":{"service_name":"payments-api"}}'
```

Expected result: validation succeeds for allowed input and returns clear errors for invalid input.

## 4. Register a delivery target
Use the admin token.

```bash
curl -X POST http://localhost:8989/v1/delivery-targets \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "dev-cluster-a",
    "display_name": "Development Cluster A",
    "purpose": "dev",
    "cluster_name": "cluster-dev-a",
    "cluster_server": "https://10.0.0.10",
    "availability_state": "available",
    "health_state": "healthy"
  }'
```

Expected result: target appears in `GET /v1/delivery-targets` and is selectable for new environments.

## 5. Create an environment from a template on a target
Use the developer token.

```bash
curl -X POST http://localhost:8989/v1/environments \
  -H "Authorization: Bearer $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "payments-dev",
    "git_repo_url": "https://github.com/example/app-config.git",
    "manifest_path": "environments/payments-dev",
    "template_version_id": "'$VERSION_ID'",
    "template_inputs": {"service_name":"payments-api"},
    "delivery_target_id": "'$TARGET_ID'"
  }'
```

Expected result: environment creation succeeds, the response includes the resolved target metadata, and a template-instance history record is created.

## 6. Request an environment move
```bash
curl -X POST http://localhost:8989/v1/environments/$ENV_ID/movements \
  -H "Authorization: Bearer $DEV_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"destination_target_id":"'$SECOND_TARGET_ID'"}'
```

Expected result: a movement record is created with `pending` or `running` status and later transitions to `completed` or `failed`.

## 7. Validate live status and notification streaming
Open a stream in one terminal:

```bash
curl -N http://localhost:8989/v1/environments/$ENV_ID/events/stream?channels=status,progress,notification \
  -H "Authorization: Bearer $DEV_TOKEN" \
  -H "Accept: text/event-stream"
```

Trigger an environment sync or move in another terminal.

Expected result: ordered `status`, `progress`, and `notification` events arrive within a few seconds of the underlying change.

## 8. Validate live log streaming
```bash
curl -N "http://localhost:8989/v1/environments/$ENV_ID/logs/stream?workload=payments-api&tail_lines=20" \
  -H "Authorization: Bearer $DEV_TOKEN" \
  -H "Accept: text/event-stream"
```

Expected result: new log lines stream until the client disconnects, and disconnecting does not change environment state.

## 9. Validate historical review
- `GET /v1/environments/$ENV_ID/movements` returns the recorded move history.
- `GET /v1/notifications?environment_id=$ENV_ID` returns recent retained notifications.
- Reviewing the environment shows which template version and delivery target were used.

## Success Checks
- Template version validation catches bad input before provisioning starts.
- Delivery targets marked unavailable cannot be selected for placement.
- Environment movement progress is visible both through read APIs and the live stream.
- Unauthorized or expired stream access fails cleanly without leaking data.
