# Service Relations

## Entry Point

- `browser -> envoy -> frontend`
- `browser -> envoy -> iam|saas|resources|student|lesson|payment|document|reporting`

`envoy` is the only public HTTP ingress. It performs `ext_authz` against `iam` gRPC `Authorization.Check` and forwards trusted `x-user-*` headers to backend HTTP services.

## Source Of Truth

- `iam`: users, credentials, token validation
- `saas`: schools, branches
- `resources`: instructors, cars, availability
- `student`: students, groups
- `lesson`: lessons and schedule conflicts
- `payment`: payments, contracts, balances
- `document`: document metadata and templates
- `reporting`: read-only aggregates
- `planning`: schedule generation worker

## Runtime Relations

- `payment -> student` by gRPC for student ownership validation
- `document -> student` by gRPC for student ownership validation
- `reporting -> student/resources/lesson/payment` by direct storage reads
- `lesson -> Kafka planning.requests` for asynchronous planning request production
- `planning -> lesson` through the lesson batch apply contract
- `planning` reads domain data from other services and produces generated schedules

## Frontend Coverage

- `/api/iam/*`
- `/api/saas/*`
- `/api/resources/*`
- `/api/students/*`
- `/api/lessons/*`
- `/api/payments/*`
- `/api/documents/*`
- `/api/reporting/*`

The frontend sends planning requests through `lesson`: `/api/lessons/planning/requests`.
