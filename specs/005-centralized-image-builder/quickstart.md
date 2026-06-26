# Quickstart: Centralized Image Builder Validation

## Goal

Validate that the Centralized Image Builder feature lets a team register an application, trigger and observe builds, enforce required security verification, and update deployment source of truth without requiring a Dockerfile.

## Prerequisites

- `idp-core` API is running locally on `http://localhost:8989`
- PostgreSQL is running and reachable by the API
- Required background workers or equivalent async processors are running for build orchestration
- A reachable source repository exists with a valid `application.yaml` descriptor and no required Dockerfile
- Approved builder profile, registry target, and GitOps target records exist for the test team
- Security verification integrations are configured so SBOM generation, image scanning, and image signing can complete
- The GitOps destination repository is reachable and writable by the platform
- The authenticated test user belongs to the target team

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
    "repository_url": "https://github.com/example/go-http-server.git",
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
- trigger response code is `202`
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
- retry response code is `202`
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
- response code is `202`
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

## Related artifacts

- `specs/005-centralized-image-builder/spec.md`
- `specs/005-centralized-image-builder/plan.md`
- `specs/005-centralized-image-builder/research.md`
- `specs/005-centralized-image-builder/data-model.md`
- `specs/005-centralized-image-builder/contracts/centralized-image-builder.md`
