# Reporting Service

HTTP service for read-only reports and school-scoped analytics.

## Responsibilities

- build dashboard metrics
- aggregate students, lessons, instructors, cars, and finance reports
- expose read-only analytics endpoints
- work strictly inside current `school_id`

## API

Contract lives in `shared/api/reporting/v1`, generated package is `shared/pkg/reporting/v1`.

Routes:

- `GET /reports/dashboard`
- `GET /reports/students`
- `GET /reports/lessons`
- `GET /reports/instructors/workload`
- `GET /reports/cars/utilization`
- `GET /reports/finance`

## Notes

- current MVP calculates reports on demand
- service is not a source of truth for business entities
- aggregates are built from `student`, `resources`, and `payment` databases
- optional filters: `branch_id`, `date_from`, `date_to`
