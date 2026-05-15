# IAM Service Implementation Steps

## Goal

Implement `IAM Service` in this repository following the same overall approach already used for `resources` and `saas`, with adaptation to the contract requirement that the internal service contract is `gRPC`.

## Working Assumptions

The contract leaves several strategic choices open. For implementation in this repo, the following assumptions are used:

- internal transport: `gRPC`
- access token strategy: signed JWT
- refresh token strategy: stateful refresh token stored in DB in hashed form
- password hashing: `bcrypt`
- password reset tokens: stored in DB in hashed form
- no MFA in current scope
- no dedicated login history / audit subsystem in current scope
- `ValidateToken` validates access JWT and returns identity context from token claims

## Step Log

### Step 1. Contract analysis

Read source contract:

- `C:\Users\vladd\Downloads\iam-service-contract.md`

Extracted main implementation areas:

- user management
- credentials storage
- password reset flow
- refresh session flow
- token validation
- identity context

### Step 2. Repo alignment

Compared current repo structure with existing services:

- `resources`
- `saas`
- `platform`
- `shared`

Target structure for new service:

- `shared/proto/...` for protobuf contract
- `shared/pkg/iam/v1` for generated protobuf Go code
- `iam/internal/...` for model/repository/service/api
- `iam/migrations/...`
- `iam/cmd/grpc-server/main.go`

### Step 3. Protobuf contract

Implemented internal gRPC contract in:

- `shared/proto/iam/v1/iam.proto`
- `shared/proto/buf.yaml`
- `shared/proto/buf.gen.yaml`

Generated code:

- `shared/pkg/iam/v1/iam.pb.go`
- `shared/pkg/iam/v1/iam_grpc.pb.go`

Covered methods:

- `CreateUser`
- `GetUser`
- `GetUserByEmail`
- `ListUsers`
- `UpdateUser`
- `ActivateUser`
- `DeactivateUser`
- `SetUserPassword`
- `Login`
- `RefreshSession`
- `RequestPasswordReset`
- `ConfirmPasswordReset`
- `ValidateToken`

### Step 4. IAM module scaffold

Created new module:

- `iam/go.mod`

Added module to workspace:

- `go.work`

Created base folders:

- `iam/internal/model`
- `iam/internal/config`
- `iam/internal/repository`
- `iam/internal/service`
- `iam/internal/api/grpc/v1`
- `iam/internal/converter`
- `iam/internal/migrator`
- `iam/cmd/grpc-server`
- `iam/tests/integration`
- `iam/migrations`

### Step 5. Data model and migrations

Implemented domain model in:

- `iam/internal/model/model.go`

Implemented SQL migration in:

- `iam/migrations/20260417195000_init_iam.sql`

Tables created:

- `users`
- `credentials`
- `password_reset_tokens`
- `refresh_tokens`

Main DB rules:

- unique email via `LOWER(email)` index
- role/status check constraints
- `super_admin` may exist without `school_id`
- all other roles require `school_id`
- credentials stored separately from users

### Step 6. Repository layer

Implemented Postgres repository in:

- `iam/internal/repository/repository.go`
- `iam/internal/repository/postgres.go`
- `iam/internal/repository/errors.go`
- `iam/internal/repository/builder.go`

Repository covers:

- create/get/list/update users
- password hash storage and update
- password reset token creation and consumption
- refresh token creation, lookup and rotation

Important implementation choice:

- create user + credentials is transactional
- password reset confirmation is transactional
- refresh token rotation is transactional

### Step 7. Service layer

Implemented business logic in:

- `iam/internal/service/service.go`
- `iam/internal/service/errors.go`
- `iam/internal/service/iam/service.go`

Implemented logic:

- email normalization
- password policy validation
- role/status validation
- bcrypt password hashing
- JWT access token issuing
- refresh token rotation
- password reset token hashing
- token validation with identity extraction

Important scope note:

- `RequestPasswordReset` stores reset token hash and logs creation event, but actual email delivery integration is still out of scope
- deactivate `reason` is accepted and logged, but not persisted in DB because the contract did not define storage for it

### Step 8. gRPC API layer

Implemented transport mapping in:

- `iam/internal/api/grpc/v1/api.go`
- `iam/internal/converter/convert.go`

Mapped errors to gRPC statuses:

- invalid input -> `InvalidArgument`
- auth failure -> `Unauthenticated`
- missing entity -> `NotFound`
- conflict -> `AlreadyExists`

### Step 9. Runtime bootstrap

Implemented runtime/config/bootstrap in:

- `iam/internal/config/config.go`
- `iam/internal/config/env/config.go`
- `iam/internal/migrator/migrator.go`
- `iam/cmd/grpc-server/main.go`

Runtime behavior:

- reads env config
- initializes logger
- optionally applies migrations
- opens postgres pool
- starts gRPC server
- registers IAM service
- registers gRPC health service

### Step 10. Deployment wiring

Added IAM deployment files:

- `deploy/docker/iam/Dockerfile`
- `deploy/compose/iam/docker-compose.yaml`
- `deploy/env/iam.env.template`

Updated shared deployment generation:

- `deploy/env/.env.template`
- `deploy/env/generate-env.ps1`

Updated root orchestration:

- `Taskfile.yml`

Changes in root tasking:

- added `iam` to `MODULES`
- added `iam` to `INTEGRATION_MODULES`
- `task test` now includes `iam`
- `task test-integration` now includes `iam`

### Step 11. Integration smoke

Implemented gRPC smoke test in:

- `iam/tests/integration/iam_e2e_test.go`

Test shape:

- Postgres via testcontainers
- migrations applied before test
- gRPC server started in-memory via `bufconn`
- smoke flow:
  - create user
  - login
  - validate access token
  - refresh session

Windows note:

- like existing repo integration tests, IAM e2e is skipped on Windows and intended to execute fully in Linux CI

## Verification Log

Verified successfully:

- protobuf generation via `buf generate`
- `go test ./...` in `shared`
- `go test ./...` in `iam`
- root `task test`
- root `task test-integration`
- `task env:generate`
- `docker compose -f deploy/compose/iam/docker-compose.yaml --env-file deploy/compose/iam/.env config`

## Remaining Limitations

- Docker image build for IAM was not verified in the current environment because Docker daemon access is not available here.
- Root `Taskfile.yml` still contains older POSIX-oriented tasks outside the IAM scope.
- Password reset flow is storage-complete, but not integrated with a real mail provider yet.
