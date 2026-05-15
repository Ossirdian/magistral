# Planning Service

`Planning Service` is a Python worker that consumes planning requests from Kafka,
builds a candidate lesson plan with a genetic algorithm, and publishes the
result back to Kafka.

## Responsibilities

- consume planning requests from Kafka
- expand `group_ids` into a unique student selection
- build individual lessons with fixed `90` minute duration
- enforce hard constraints for students, instructors, cars, and existing lessons
- optimize the plan score with soft student preferences
- publish planning results for downstream orchestration

## Non-Responsibilities

- no CRUD for students, groups, resources, or lessons
- no direct lesson persistence
- no authentication

## Runtime Flow

1. `lesson` service accepts `POST /planning/requests`
2. `lesson` publishes a message to `planning.requests`
3. `planning` consumes the request
4. `planning` loads student/resource/existing lesson data through adapters
5. `planning` generates time slots for the requested window
6. `planning` runs GA evolution and decodes the best individual
7. `planning` publishes the result to `planning.results`

## Request Contract

Minimal payload:

```json
{
  "school_id": "uuid",
  "period_start": "2026-04-27T09:00:00Z",
  "period_end": "2026-04-27T18:00:00Z",
  "student_ids": ["uuid"],
  "group_ids": ["uuid"],
  "branch_ids": ["uuid"],
  "replace_strategy": "append",
  "request_id": "string"
}
```

Payload with soft preferences:

```json
{
  "school_id": "uuid",
  "period_start": "2026-04-27T09:00:00Z",
  "period_end": "2026-04-27T18:00:00Z",
  "student_ids": ["student-1"],
  "student_preferences": {
    "student-1": {
      "preferred_weekdays": ["monday", "wednesday"],
      "preferred_time_windows": ["daytime"]
    }
  },
  "group_ids": [],
  "branch_ids": [],
  "replace_strategy": "append",
  "request_id": "req-1"
}
```

## Scheduling Model

- `Group` is only a student selection mechanism
- output lessons are always individual
- lesson duration is fixed at `90` minutes
- `student_ids` and expanded `group_ids` are merged into `students_to_plan`
- each student is encoded by 3 genes:
  - slot index
  - instructor index
  - car index

## Hard Constraints

The planner rejects a candidate lesson if it overlaps an already scheduled or
existing lesson for the same:

- student
- instructor
- car

Conflict reasons currently emitted by the decoder:

- `no_available_slot`
- `resource_conflict`

## Soft Constraints

Student preferences are modeled as soft constraints and affect score, but do
not block plan generation.

Supported preferences:

- `preferred_weekdays`: `monday` ... `sunday`
- `preferred_time_windows`: `morning`, `daytime`, `evening`

Current scoring effect per scheduled lesson:

- `+20` when weekday matches a preferred weekday
- `-15` when weekday preferences exist and the lesson misses them
- `+20` when time window matches a preferred time window
- `-15` when time preferences exist and the lesson misses them

Time windows:

- `morning`: hour `< 12`
- `daytime`: hour `< 17`
- `evening`: hour `>= 17`

This follows the same product meaning as student-preference timetabling
literature: preferences are optimization targets, not hard feasibility rules.

## Genetic Algorithm

Implementation uses `DEAP` and a simple evolutionary loop in
`planning/app/genetic.py`.

Current defaults:

- deterministic seed: `42`
- population size: `80`
- generations: `40`
- selection: tournament, `tournsize=3`
- crossover: two-point crossover
- mutation probability per individual: `0.25`
- mutation probability per gene in a triplet: `0.2`

If there are no students, no instructors, no cars, or no time slots, the
planner returns an empty result with all students marked unassigned.

## Fitness Function

Final score is computed as:

```text
scheduled_bonus
- backlog_penalty
- conflict_penalty
- idle_penalty
+ preference_bonus
```

Current components:

- `scheduled_bonus = lessons * 100`
- `backlog_penalty = unassigned_students * 50`
- `conflict_penalty = conflicts * 100`
- `idle_penalty` depends on unused effective resource utilization
- `preference_bonus` rewards matching weekday/time-window preferences

## Result Contract

Result payload includes:

- `request_id`
- `school_id`
- `lessons`
- `unassigned_students`
- `conflicts`
- `score`
- `resource_idle_seconds`

`resource_idle_seconds` is computed over the requested planning window as:

- total instructor and car capacity in seconds
- minus used capacity for generated lessons

Each generated lesson consumes one instructor and one car.

## Observability

Prometheus metrics are exposed on `:${PLANNING_METRICS_PORT}/metrics`.

Counters:

- `planning_requests_processed_total`
- `planning_requests_failed_total`
- `planning_assigned_lessons_total`
- `planning_conflicts_total`

Histograms:

- `planning_scheduled_lessons`
- `planning_conflicts_per_request`
- `planning_unassigned_students`
- `planning_resource_idle_seconds`
- `planning_request_duration_seconds`

Tracing:

- OTLP traces are sent to `Jaeger`
- service name: `planning`

Logging:

- structured logs are emitted for successful and failed planning requests

## Environment

- `PLANNING_KAFKA_BOOTSTRAP_SERVERS`
- `PLANNING_KAFKA_REQUEST_TOPIC`
- `PLANNING_KAFKA_RESULT_TOPIC`
- `PLANNING_KAFKA_GROUP_ID`
- `PLANNING_LOG_LEVEL`
- `PLANNING_METRICS_PORT`
- `TRACING_COLLECTOR_ENDPOINT`
- `TRACING_SERVICE_VERSION`

## Layout

- `planning/main.py` - service entrypoint
- `planning/app/settings.py` - environment configuration
- `planning/app/kafka_consumer.py` - Kafka consumer/producer loop
- `planning/app/models.py` - request/result models
- `planning/app/data_loader.py` - external data loading interfaces
- `planning/app/constraints.py` - hard-constraint checks
- `planning/app/fitness.py` - scoring with soft preferences
- `planning/app/genetic.py` - population evolution and decoding
- `planning/app/mutation.py` - mutation step
- `planning/app/crossover.py` - crossover step
- `planning/app/repair.py` - repair step
- `planning/app/planner.py` - orchestration and idle-capacity calculation

## Status

This implementation is a runnable GA-based planning worker with hard
resource-conflict rules and soft student-preference optimization. External data
loading remains adapter-based and is currently suitable for integration wiring
and local verification.
