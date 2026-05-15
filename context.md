# Project Context

## Modules

- `platform`: shared infra utilities (`logger`, `closer`, `httpx`, `authn`, testcontainers helpers)
- `shared`: OpenAPI/proto contracts and generated code
- `iam`: identity and auth service, gRPC + HTTP
- `saas`: SaaS admin HTTP service for schools and branches
- `resources`: HTTP service for cars, instructors and availability
- `student`: HTTP + gRPC service for students and groups
- `payment`: HTTP service for student payments and contract balances
- `document`: HTTP service for student documents and document templates
- `reporting`: HTTP service for read-only analytics and aggregate reports
- `planning`: Python worker that consumes Kafka planning requests and produces planning results
- `student` stores students, groups, and lessons for school-scoped operations
- `frontend`: standalone SPA for manual end-to-end checks
- `envoy`: ingress gateway for browser traffic to SPA and HTTP APIs

## Contracts and Generation

- OpenAPI generation is wired through `ogen`
- `task ogen:gen:all` regenerates all HTTP contracts
- `task gen` regenerates proto and OpenAPI code

## Runtime Topology

- One shared `postgres-core` container is used by all services
- Kafka is provided by the shared `core` compose profile for async planning flow
- Separate databases inside the same Postgres instance:
  - `resources`
  - `saas`
  - `iam`
  - `student`
  - `payment`
  - `document`
  - `reporting`
- Shared Docker network: `magistral-microservices`
- Browser traffic enters the platform through `envoy`

## IAM

- Exposes gRPC and HTTP in the same process
- HTTP endpoints are intended for browser/frontend use:
  - `/auth/login`
  - `/auth/refresh`
  - `/auth/me`
  - `/admin/users`
  - `/admin/users/{id}`
  - `/admin/users/{id}/password`
- Bootstrap super-admin is configured through env:
  - `BOOTSTRAP_SUPER_ADMIN_EMAIL`
  - `BOOTSTRAP_SUPER_ADMIN_PASSWORD`

## Auth Model

- Shared auth validation model lives in `platform/pkg/authn`
- All external HTTP requests go through `envoy`
- `envoy` performs centralized authentication through IAM gRPC `Authorization.Check`
- On successful auth `envoy` injects trusted `x-user-*` headers into proxied HTTP requests
- HTTP services require trusted headers from `envoy`; direct IAM `ValidateToken` is no longer used in HTTP middleware
- `school_id` is taken from `x-school-id`
- `role` is taken from `x-user-role`
- `saas`:
  - full schools/branches/admin endpoints are `super_admin` only
  - `GET /branches?school_id=...` is additionally allowed for school-scoped roles
- `resources`:
  - available for `admin`, `owner`, `instructor`
  - school context is derived from IAM token
- `student`:
  - available for `admin`, `owner`, `instructor`
  - school context is derived from IAM token
- `lesson`:
  - available for `admin`, `owner`, `instructor`
  - school context is derived from IAM token
  - includes lesson scheduling with conflict checks for student, instructor, and car
  - exposes `POST /planning/requests` and publishes requests to Kafka
- `payment`:
  - available for `admin`, `owner`, `instructor`
  - school context is derived from IAM token
  - validates `student_id` ownership through `student` gRPC
  - stores payments and student contract totals, recalculating balance after each change
- `document`:
  - available for `admin`, `owner`, `instructor`
  - school context is derived from IAM token
  - validates `student_id` ownership through `student` gRPC
  - stores document metadata in Postgres and file content by `object_key` in storage
- `reporting`:
  - available for `admin`, `owner`, `instructor`
  - school context is derived from IAM token
  - exposes read-only dashboard and aggregate reporting endpoints
  - current MVP computes aggregates on demand from `student`, `resources`, and `payment` stores
- `planning`:
  - consumes planning requests from Kafka
  - expands groups into student selections
  - builds only individual lessons
  - exposes Prometheus metrics and OTLP traces for worker execution

## Frontend

- Located in `frontend/`
- Served by nginx as static SPA container
- Browser entrypoint is `envoy`
- `envoy` routes:
  - `/` -> `frontend`
  - `/api/iam/*` -> `iam`
  - `/api/saas/*` -> `saas`
  - `/api/resources/*` -> `resources`
  - `/api/students/*` -> `student`
  - `/api/payments/*` -> `payment`
  - `/api/documents/*` -> `document`
  - `/api/reporting/*` -> `reporting`
- `envoy` also terminates CORS and runs `ext_authz` against IAM before forwarding protected API calls
- Main flows:
  - `super_admin` -> SaaS control panel
  - `admin` / `owner` / `instructor` -> resources workspace

## Removed Temporary UI

- Embedded `test-ui` and `test-api` handlers were removed from `saas` and `resources`
- Manual verification now goes through the standalone `frontend`

## Useful Commands

```powershell
task env:generate
task gen
task format
task lint
task test
task test-integration
task docker:up:all
task docker:down:all
```

## Important Files

- `Taskfile.yml`
- `README.md`
- `frontend_integration_steps.md`
- `deploy/compose/*`
- `deploy/env/*`
