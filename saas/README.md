# SaaS Service

HTTP-сервис административной панели SaaS-уровня. Управляет школами и филиалами, а также координирует создание первого school-admin через IAM.

## Назначение

`saas` отвечает за:

- CRUD школ
- CRUD филиалов
- bootstrap школы вместе с первым `admin`
- авторизацию запросов через trusted headers от `envoy`
- разграничение доступа между `super_admin` и school-scoped ролями

## Архитектура

```text
frontend -> /api/saas/*
  -> saas/internal/httpauth
    -> trusted x-user-* headers from envoy
  -> saas/internal/api/saasadmin/v1
    -> service/school | service/branch
      -> repository/school | repository/branch
        -> Postgres

frontend -> /api/saas/admin/bootstrap-school
  -> saas/internal/api/admin
    -> service/school
      -> Postgres + IAM HTTP POST /admin/users
```

Сборка приложения описана в [internal/app/app.go](/abs/path/c:/Users/vladd/Desktop/magistral/saas/internal/app/app.go:1):

- при старте накатываются миграции
- создаётся `pgxpool`
- для bootstrap-school поднимается HTTP client в IAM
- собираются `school` и `branch` сервисы
- HTTP handler оборачивается в auth middleware и CORS

## Слои

- `cmd/http-server` - точка входа
- `internal/app` - wiring и lifecycle
- `internal/api/saasadmin/v1` - OpenAPI-совместимый HTTP API
- `internal/api/admin` - специальный endpoint bootstrap школы
- `internal/httpauth` - проверка trusted headers от envoy
- `internal/service/school`, `internal/service/branch` - бизнес-логика
- `internal/client/iam` - HTTP-клиент для bootstrap-school
- `internal/repository/school`, `internal/repository/branch` - доступ к Postgres
- `internal/model` - доменные модели
- `migrations` - схема БД
- `tests/integration` - интеграционные тесты

## Модель доступа

`saas/internal/httpauth` реализует две ветки доступа:

- `super_admin` имеет полный доступ к `/schools`, `/branches`, `/admin/*`
- `admin`, `owner`, `instructor` могут только читать `GET /branches?school_id=<their school>`

Это нужно, чтобы `frontend` мог получать справочник филиалов для school-scoped пользователей без доступа к SaaS панели целиком.

## Основные сущности

- `School`
  - `id`, `name`, `status`, `created_at`, `updated_at`, `deleted_at`
- `Branch`
  - `id`, `school_id`, `name`, `address`, `status`, `created_at`, `updated_at`, `deleted_at`

Статусы школ:

- `active`
- `inactive`
- `suspended`

Статусы филиалов:

- `active`
- `inactive`

## API и контракты

Контракт сервиса описан в `shared/api/saas-admin/v1` и генерируется в `shared/pkg/saasadmin/v1`.

Ключевые маршруты:

- `GET /schools`
- `POST /schools`
- `GET /schools/{id}`
- `PATCH /schools/{id}`
- `DELETE /schools/{id}`
- `GET /branches`
- `POST /branches`
- `GET /branches/{id}`
- `PATCH /branches/{id}`
- `DELETE /branches/{id}`
- `POST /admin/bootstrap-school`
- `GET /healthz`

## Внешние зависимости

- Postgres база `saas`
- IAM HTTP для создания первого school-admin
- HTTP порт по умолчанию: `8081`

Основные env:

- `POSTGRES_DSN`
- `IAM_HTTP_BASE_URL`
- `AUTO_MIGRATE`

Шаблон env: [deploy/env/saas.env.template](/abs/path/c:/Users/vladd/Desktop/magistral/deploy/env/saas.env.template:1)

## Запуск

Локально:

```powershell
cd saas
go run ./cmd/http-server
```

Через compose:

```powershell
task docker:up:saas
```

## Проверка и развитие

- healthcheck: `http://localhost:8081/healthz`
- интеграционные тесты: `saas/tests/integration`
- изменения OpenAPI требуют регенерации `shared/pkg/saasadmin/v1`
