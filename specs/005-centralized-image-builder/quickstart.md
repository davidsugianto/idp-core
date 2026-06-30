# Quickstart: Centralized Image Builder Validation

## Goal

Validate that the Centralized Image Builder feature lets a team register an application, trigger and observe builds, enforce required security verification, and update deployment source of truth without requiring a Dockerfile.

## Design assumptions

- Build execution and post-build callbacks preserve one canonical build record per accepted build intent
- Build status transitions are lifecycle-ordered (`queued` → `running` → terminal) and team-scoped
- Security verification outcomes are recorded on the same build before deployment-readiness is asserted
- Deployment automation uses the configured GitOps target and does not bypass policy-blocked builds

## Prerequisites

- Local development defaults disable OIDC bootstrap; if you need OIDC flows, re-enable `oidc.enabled` or set `OIDC_ENABLED=true`
- `idp-core` API is running locally on `http://localhost:8989`
- PostgreSQL is running and reachable by the API
- Required background workers or equivalent async processors are running for build orchestration
- A reachable source repository exists with a valid `application.yaml` descriptor and no required Dockerfile
- Approved builder profile, registry target, and GitOps target records exist for the test team
- Security verification integrations are configured so SBOM generation, image scanning, and image signing can complete
- The GitOps destination repository is reachable and writable by the platform
- A valid local kubeconfig is available either at the default location (`~/.kube/config`) or via `KUBECONFIG_PATH`
- The authenticated test user belongs to the target team

## Start the API (local validation)

```bash
make dev-app-up
```

If your kubeconfig is not at `~/.kube/config`, export `KUBECONFIG_PATH` to point at a valid local file before starting the API. If you want to exercise OIDC locally, also set `OIDC_ENABLED=true` and provide a reachable issuer/discovery endpoint.

## Fetch a bearer token

```bash
TOKEN=$(curl -s -X POST http://localhost:8989/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"test-user","team_id":"26bc2c5f-e250-4720-877a-a4b158bed1d4"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["token"])')
```

## Scenario 1: Register a buildable application

```bash
curl -s -X POST http://localhost:8989/v1/build-applications \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "go-http-server",
    "description": "Buildpacks-managed application",
    "repository_url": "https://github.com/davidsugianto/go-http-server.git",
    "application_descriptor_path": "application.yaml",
    "builder_profile_id": "<builder-profile-id>",
    "registry_target_id": "<registry-target-id>",
    "deployment_automation_enabled": true,
    "gitops_target_id": "<gitops-target-id>"
  }'
```

Expected results:
- response code is `201`
- response contains a new application `id`
- response includes the source repository and descriptor path
- response shows the selected builder and registry targets

## Scenario 2: View and update the application

```bash
APP_ID=<application-id>

curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/build-applications/$APP_ID

curl -s -X PATCH http://localhost:8989/v1/build-applications/$APP_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "deployment_automation_enabled": false
  }'
```

Expected results:
- detail response code is `200`
- patch response code is `200`
- updated response reflects the changed deployment automation setting
- list/detail responses remain team-scoped and do not expose another team’s application data

Delete semantics check:

```bash
curl -s -X DELETE http://localhost:8989/v1/build-applications/$APP_ID \
  -H "Authorization: Bearer $TOKEN"
```

Expected delete results:
- delete response code is `200`
- payload includes `message: "build application deleted"`
- subsequent list calls do not include the deleted application
- historical build/security/deployment records remain queryable via build endpoints

## Scenario 3: Trigger and observe a build

```bash
curl -s -X POST http://localhost:8989/v1/build-applications/$APP_ID/builds \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "source_revision": "main",
    "idempotency_key": "quickstart-build-001"
  }'
```

Save the returned build identifier:

```bash
BUILD_ID=<build-id>
```

Check status:

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/builds/$BUILD_ID
```

Expected results:
- trigger response code is `201`
- build detail response code is `200`
- build detail progresses through queued/running and reaches a terminal outcome
- repeated trigger using the same idempotency key does not create a conflicting build outcome

## Scenario 4: Stream build logs

```bash
curl -N -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/builds/$BUILD_ID/logs/stream
```

Expected results:
- response begins streaming while the build is active
- streamed content shows lifecycle progress
- stream closes cleanly or emits a terminal completion signal when the build finishes

## Scenario 5: Validate security verification and artifact publication

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/builds/$BUILD_ID
```

Expected results for a successful compliant build:
- `artifact.published_image_reference` is populated
- security verification shows SBOM, scan, and signing results
- policy gate result is visible
- build status becomes `deployment_ready` only after required verification passes

## Scenario 6: Validate build history and retry flow

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/build-applications/$APP_ID/builds

curl -s -X POST http://localhost:8989/v1/builds/$BUILD_ID/retry \
  -H "Authorization: Bearer $TOKEN"
```

Expected results:
- history response code is `200`
- history includes the original build
- retry response code is `200`
- retry result links back to the original build attempt

## Scenario 7: Validate cancel flow

1. Trigger a new build that remains active long enough to cancel.
2. Send the cancel request:

```bash
ACTIVE_BUILD_ID=<active-build-id>

curl -s -X POST http://localhost:8989/v1/builds/$ACTIVE_BUILD_ID/cancel \
  -H "Authorization: Bearer $TOKEN"
```

Expected results:
- response code is `200`
- build eventually reaches a terminal canceled state
- repeated cancel requests do not create contradictory outcomes

## Scenario 8: Validate security-gated deployment behavior

Run a build that fails one required verification step.

Expected results:
- image publication and security evidence remain visible on the build record when applicable
- deployment source update is not performed when required verification fails
- build record exposes a blocked or failed post-build outcome with a clear reason

## Scenario 9: Validate GitOps-first deployment update

Run a successful compliant build with deployment automation enabled.

Expected results:
- deployment update record is attached to the build
- GitOps repository change is visible and references the new image
- resulting deployment update status is `succeeded`
- build reaches `deployment_ready`

## Scenario 10: Validate failure behavior

Exercise at least one of the following:
- repository access failure
- unsupported runtime detection
- registry publication failure
- build cancellation after terminal completion
- unauthorized access to another team’s application or build

Expected results:
- client receives explicit sanitized failure information
- no secrets or credentials are exposed
- canonical build/application state remains internally consistent

## Scenario 11: Live Docker-based E2E smoke test

Use this flow when validating the feature against the local Docker-backed `idp-core` stack and the sample app repository at `../go-http-server`.

### Prepare the sample application repository

The repository must contain `application.yaml` at the repo root. The validated sample descriptor is:

```yaml
apiVersion: idp.dev/v1alpha1
kind: Application
metadata:
  name: go-http-server
spec:
  runtime: go
  port: 7979
  healthCheck:
    path: /v1/ping
  env:
    - name: ENV
      value: development
```

The sample repository used for validation is:
- local checkout: `../go-http-server`
- remote URL: `https://github.com/davidsugianto/go-http-server.git`

### Start the local stack

```bash
make dev-db-up
make dev-redis-up
make dev-app-up
make dev-cron-up
```

Health checks:

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8989/v1/ping
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8983/ping
```

Expected results:
- API health returns `200`
- cron health returns `200`

If the cron container was already running before centralized image builder scheduling changes were added, restart it so it reloads the current config:

```bash
docker-compose restart cron
```

### Authenticate and capture reusable variables

```bash
TOKEN=$(curl -s -X POST http://localhost:8989/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"test-user","team_id":"26bc2c5f-e250-4720-877a-a4b158bed1d4"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["token"])')

TEAM_ID=26bc2c5f-e250-4720-877a-a4b158bed1d4
APP_NAME=go-http-server-e2e-$(date +%s)
BUILDER_PROFILE_ID=<builder-profile-id>
REGISTRY_TARGET_ID=<registry-target-id>
GITOPS_TARGET_ID=<gitops-target-id>
```

### Register the sample application

```bash
APP_ID=$(curl -s -X POST http://localhost:8989/v1/build-applications \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\
    \"name\": \"$APP_NAME\",\
    \"description\": \"Live Docker E2E validation\",\
    \"repository_url\": \"https://github.com/davidsugianto/go-http-server.git\",\
    \"application_descriptor_path\": \"application.yaml\",\
    \"builder_profile_id\": \"$BUILDER_PROFILE_ID\",\
    \"registry_target_id\": \"$REGISTRY_TARGET_ID\",\
    \"deployment_automation_enabled\": true,\
    \"gitops_target_id\": \"$GITOPS_TARGET_ID\"\
  }" | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["id"])')
```

Expected results:
- create response returns `201`
- returned `application_descriptor_path` is `application.yaml`
- returned `repository_url` points at the sample repository

### Trigger a build from `master`

```bash
BUILD_ID=$(curl -s -X POST http://localhost:8989/v1/build-applications/$APP_ID/builds \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"source_revision":"master","idempotency_key":"live-e2e-build-001"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["id"])')
```

Inspect the accepted build:

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/builds/$BUILD_ID
```

Expected results:
- trigger response returns `201`
- initial build status is `queued`
- `source_revision_requested` is `master`

### Wait for async execution

```bash
for i in 1 2 3 4 5 6; do
  curl -s -H "Authorization: Bearer $TOKEN" \
    http://localhost:8989/v1/builds/$BUILD_ID
  echo
  sleep 10
done
```

Expected success path:
- build leaves `queued`
- build reaches `running`
- final build status becomes `deployment_ready`
- `source_revision_resolved` is populated
- artifact and security verification fields are populated

Inspect log streaming state:

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/builds/$BUILD_ID/logs/stream
```

Expected success path:
- `last_sequence` increases above `0`
- log lines include worker claim and completion messages

### Recovery path when the build remains queued

If the build is still `queued` after the wait window and `last_sequence` remains `0`, the worker path is present but scheduled dispatch has not fired yet in the running cron process. Manually dispatch the cron endpoint to validate the worker execution path directly:

```bash
curl -s -X POST http://localhost:8983/build-executor-dispatch
```

Then re-check build detail and logs:

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/builds/$BUILD_ID

curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8989/v1/builds/$BUILD_ID/logs/stream
```

Expected recovery results:
- build transitions to `deployment_ready`
- logs contain entries such as `build claimed by worker` and `build completed with status deployment_ready`
- artifact and security verification fields are present on the build detail response

### Clean up the test application

```bash
curl -s -X DELETE http://localhost:8989/v1/build-applications/$APP_ID \
  -H "Authorization: Bearer $TOKEN"
```

Expected results:
- delete response returns `200`
- payload includes `message: "build application deleted"`

## Validation run notes (2026-06-27)

- Swagger and API contract alignment validated after annotation refresh:
  - `swag init -g cmd/http/main.go -o docs/swagger --parseDependency --parseInternal`
  - `go test ./tests/contract`
- Build application/unit/repository regression checks are green via `go test ./...`.
- Manual cron dispatch via `POST http://localhost:8983/build-executor-dispatch` successfully drains queued builds and finalizes them to `deployment_ready`.
- Scheduled cron dispatch may require a cron container restart to reload the updated `build-executor-dispatch` schedule in Docker-backed local validation.

## Related artifacts

- `specs/005-centralized-image-builder/spec.md`
- `specs/005-centralized-image-builder/plan.md`
- `specs/005-centralized-image-builder/research.md`
- `specs/005-centralized-image-builder/data-model.md`
- `specs/005-centralized-image-builder/contracts/centralized-image-builder.md`
