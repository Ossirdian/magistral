# Frontend

React/Vite SPA for manual end-to-end verification of the platform flows.

## Purpose

`frontend` now provides two working modes:

- `super_admin` -> SaaS control panel for schools, branches, and IAM school users
- `admin` / `owner` / `instructor` -> school workspace for resources, students, lessons, payments, documents, reporting, and planning requests

## Architecture

```text
browser
  -> envoy
    -> /                    -> app-frontend:80
    -> /api/iam/*           -> app-iam:8082
    -> /api/saas/*          -> app-saas:8081
    -> /api/resources/*     -> app-resources:8080
    -> /api/students/*      -> app-student:8083
    -> /api/payments/*      -> app-payment:8084
    -> /api/documents/*     -> app-document:8085
    -> /api/reporting/*     -> app-reporting:8086
```

Frontend nginx serves only static SPA assets.
Ingress routing lives in [deploy/compose/envoy/envoy.yaml](/C:/Users/vladd/Desktop/magistral/deploy/compose/envoy/envoy.yaml:1).

Protected API requests are authenticated at the gateway level:

- browser sends `Authorization` to `envoy`
- `envoy` calls IAM gRPC `Authorization.Check`
- IAM returns trusted `x-user-*` headers
- `envoy` forwards the HTTP request to the target service without the original auth header
- backend HTTP services derive `role` and `school_id` only from those forwarded headers

## App Layout

- [src/main.jsx](/C:/Users/vladd/Desktop/magistral/frontend/src/main.jsx:1) - React entrypoint
- [src/App.jsx](/C:/Users/vladd/Desktop/magistral/frontend/src/App.jsx:1) - session handling, routing by role, API calls, workspace screens
- [src/styles.css](/C:/Users/vladd/Desktop/magistral/frontend/src/styles.css:1) - visual system and layout rules
- [index.html](/C:/Users/vladd/Desktop/magistral/frontend/index.html:1) - SPA container and font imports

## Main Flows

### Login

- `POST /api/iam/auth/login`
- store session in `localStorage`
- `GET /api/iam/auth/me`
- route by role

### Super Admin Panel

- schools list and create
- bootstrap school with first admin
- branches for selected school
- IAM users for selected school

### School Workspace

- resources: cars, instructors, availability, car assignment
- students: students and groups with create/edit/delete flows
- lessons: create/edit/list/delete
- payments: list/create/edit/delete, student balance, contract total update
- documents: upload, generate from template, template registry with edit/delete, document metadata lookup, download, delete
- reporting: dashboard, student, lesson, instructor, car, and finance aggregates
- planning: submit `POST /api/lessons/planning/requests`

## Role Model in UI

- `super_admin` sees only SaaS panel
- `admin` sees full school workspace with write actions
- `owner` and `instructor` see the same school workspace in read-only mode for services that enforce GET-only access

## Run

Primary local flow:

```powershell
task docker:up:frontend
```

Vite dev mode:

```powershell
cd frontend
npm install
npm run dev
```

## Notes

- the UI uses token refresh through `/api/iam/auth/refresh`
- session is persisted in `localStorage`
- document upload is sent as JSON with `content_base64`
- planning requests are queued through `lesson` service and not persisted directly by `planning`
