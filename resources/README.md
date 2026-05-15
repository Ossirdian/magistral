# Resources Service

HTTP-сервис для управления ресурсами школы: инструкторами, автомобилями, их занятостью и привязкой `car -> instructor`.

## Назначение

`resources` отвечает за:

- CRUD инструкторов
- CRUD автомобилей
- чтение и замену availability окон
- назначение автомобиля инструктору
- работу только в school-scoped контексте, полученном из trusted headers `envoy`

## Архитектура

```text
frontend -> /api/resources/*
  -> resources/internal/httpauth
    -> trusted x-user-role + x-school-id from envoy
  -> resources/internal/api/resources/v1
    -> service/car | service/instructor
      -> repository/car | repository/instructor
        -> Postgres
```

Точка сборки находится в [internal/app/app.go](/abs/path/c:/Users/vladd/Desktop/magistral/resources/internal/app/app.go:1):

- выполняются миграции при `AUTO_MIGRATE=true`
- создаётся `pgxpool`
- собирается HTTP API поверх car/instructor сервисов
- в middleware прокидывается school context из envoy headers

## Слои

- `cmd/http-server` - точка входа
- `internal/app` - wiring и lifecycle
- `internal/api/resources/v1` - HTTP API поверх OpenAPI-контракта
- `internal/httpauth` - проверка trusted headers и извлечение `school_id`
- `internal/service/car`, `internal/service/instructor` - бизнес-логика
- `internal/repository/car`, `internal/repository/instructor` - доступ к Postgres
- `internal/model` - доменные сущности и фильтры
- `migrations` - схема БД
- `tests/integration` - интеграционные тесты

## Модель доступа

Доступ к `/resources/*` получают только роли:

- `admin`
- `owner`
- `instructor`

Middleware требует:

- trusted headers от `envoy`
- роль `admin|owner|instructor`
- наличие `x-school-id`

Далее `school_id` читается из контекста и используется как tenant boundary для запросов.

## Основные сущности

- `Instructor`
  - привязан к `school_id` и `branch_id`
  - имеет `status`, `availability_status`, `current_car_id`
- `Car`
  - привязан к `school_id` и `branch_id`
  - может иметь `instructor_id`
  - имеет `status`, `availability_status`
- `AvailabilityWindow`
  - типы: `working_hours`, `vacation`, `blackout`

Статусы инструкторов:

- `active`
- `inactive`
- `on_leave`

Статусы автомобилей:

- `active`
- `inactive`
- `maintenance`

## API и контракты

Контракт описан в `shared/api/resources/v1` и генерируется в `shared/pkg/resources/v1`.

Ключевые маршруты:

- `GET /resources/instructors`
- `POST /resources/instructors`
- `GET /resources/instructors/{id}`
- `PATCH /resources/instructors/{id}`
- `DELETE /resources/instructors/{id}`
- `GET /resources/instructors/{id}/availability`
- `PUT /resources/instructors/{id}/availability`
- `GET /resources/cars`
- `POST /resources/cars`
- `GET /resources/cars/{id}`
- `PATCH /resources/cars/{id}`
- `DELETE /resources/cars/{id}`
- `GET /resources/cars/{id}/availability`
- `PUT /resources/cars/{id}/availability`
- `POST /resources/cars/{id}/assign`
- `GET /healthz`

## Внешние зависимости

- Postgres база `resources`
- вход только через `envoy`
- HTTP порт по умолчанию: `8080`

Основные env:

- `POSTGRES_DSN`
- `AUTO_MIGRATE`
- `RESOURCES_SCHOOL_ID`

Шаблон env: [deploy/env/resources.env.template](/abs/path/c:/Users/vladd/Desktop/magistral/deploy/env/resources.env.template:1)

## Запуск

Локально:

```powershell
cd resources
go run ./cmd/http-server
```

Через compose:

```powershell
task docker:up:resources
```

## Проверка и развитие

- healthcheck: `http://localhost:8080/healthz`
- интеграционные тесты: `resources/tests/integration`
- изменения OpenAPI требуют регенерации `shared/pkg/resources/v1`
