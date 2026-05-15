# IAM Service

`iam` отвечает за аутентификацию, пользовательские сессии и identity-контекст, который дальше прокидывается в остальные HTTP-сервисы через `Envoy`.

Сервис поднимает два интерфейса:

- HTTP API для логина, refresh, `me` и админских операций с пользователями
- gRPC API для внутренних интеграций и `Envoy ext_authz`

## Роль в архитектуре

Текущий входной поток такой:

```text
HTTP client
  -> Envoy (:8080)
    -> ext_authz gRPC Check() в IAM (:50053)
      -> IAM ValidateToken(...)
      -> Envoy получает x-user-* заголовки
    -> downstream HTTP service
```

Что это означает на практике:

- весь внешний HTTP трафик идет через `Envoy`
- `Envoy` вызывает `IAM` по gRPC методом `Check`
- `IAM` валидирует `Bearer` token или `session-uuid`
- если токен валиден, `Envoy` добавляет identity headers:
  - `x-authenticated`
  - `x-user-id`
  - `x-user-role`
  - `x-school-id`
  - `x-user-status`
  - `x-user-active`
- downstream-сервисы больше не ходят в `IAM` за проверкой токена на каждый запрос

## За что отвечает сервис

- login и refresh пользовательских сессий
- выдача `access token`
- хранение пользователей, ролей и статусов
- bootstrap первого `super_admin`
- выдача identity-контекста для остальных сервисов через `Envoy ext_authz`
- gRPC контракт для внутренних сценариев IAM

## За что сервис больше не отвечает напрямую

- IAM не должен быть обязательным runtime-клиентом для каждого внутреннего HTTP-сервиса
- `school_id` и `role` приходят в downstream уже через заголовки от `Envoy`
- авторизация по HTTP строится вокруг прокинутых `x-user-*` headers, а не вокруг прямого gRPC запроса в IAM

## Транспорты

## HTTP API

Основные маршруты, которые `Envoy` проксирует в `iam`:

- `POST /api/iam/auth/login`
- `POST /api/iam/auth/refresh`
- `GET /api/iam/auth/me`
- `GET /api/iam/admin/users`
- `POST /api/iam/admin/users`
- `GET /api/iam/admin/users/{id}`
- `PATCH /api/iam/admin/users/{id}`
- `POST /api/iam/admin/users/{id}/password`
- `GET /healthz`

Особенности:

- `login` и `refresh` в `Envoy` исключены из `ext_authz`
- `/api/iam/*` после gateway переписывается в `/` внутри IAM HTTP сервера
- `/admin/*` дополнительно ограничивается логикой самого IAM

## gRPC API

На gRPC сервере зарегистрированы два интерфейса:

1. `IAMService`
   Контракт из [shared/proto/iam/v1/iam.proto](C:/Users/vladd/Desktop/magistral/shared/proto/iam/v1/iam.proto)

   Основные методы:
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

2. `AuthorizationServer`
   Envoy `ext_authz` интерфейс из `envoy.service.auth.v3`

   Основной метод:
   - `Check`

`Check` извлекает токен из:

- `Authorization: Bearer ...`
- `session-uuid`
- `x-session-uuid`

После успешной валидации метод возвращает `OK` и заголовки для downstream-сервисов. При ошибке возвращается `401 unauthorized`.

## Основные модули

- [cmd/grpc-server/main.go](C:/Users/vladd/Desktop/magistral/iam/cmd/grpc-server/main.go) — точка входа процесса
- [internal/app/app.go](C:/Users/vladd/Desktop/magistral/iam/internal/app/app.go) — wiring, запуск HTTP и gRPC серверов
- [internal/api/http](C:/Users/vladd/Desktop/magistral/iam/internal/api/http) — HTTP API для frontend/admin
- [internal/api/grpc/v1](C:/Users/vladd/Desktop/magistral/iam/internal/api/grpc/v1) — внутренний gRPC контракт IAM
- [internal/api/envoyauth/v1](C:/Users/vladd/Desktop/magistral/iam/internal/api/envoyauth/v1) — реализация `Envoy Check`
- [internal/service/iam](C:/Users/vladd/Desktop/magistral/iam/internal/service/iam) — бизнес-логика
- [internal/repository](C:/Users/vladd/Desktop/magistral/iam/internal/repository) — Postgres repository
- [migrations](C:/Users/vladd/Desktop/magistral/iam/migrations) — схема БД

## Доменные правила

Роли:

- `super_admin`
- `admin`
- `owner`
- `instructor`

Статусы пользователей:

- `active`
- `inactive`
- `suspended`

Ограничение tenant-модели:

- `super_admin` может существовать без `school_id`
- остальные роли обязаны иметь `school_id`

Это дополнительно зафиксировано в SQL-check constraints миграции.

## Зависимости

- Postgres база `iam`
- HTTP порт сервиса: `8082`
- gRPC порт сервиса: `50053`
- внешний HTTP доступ обычно идет не напрямую, а через `Envoy :8080`

Основные env:

- `POSTGRES_DSN`
- `JWT_SECRET`
- `JWT_ISSUER`
- `ACCESS_TOKEN_TTL`
- `REFRESH_TOKEN_TTL`
- `RESET_TOKEN_TTL`
- `BOOTSTRAP_SUPER_ADMIN_EMAIL`
- `BOOTSTRAP_SUPER_ADMIN_PASSWORD`
- `TRACING_COLLECTOR_ENDPOINT`

Шаблон env: [deploy/env/iam.env.template](C:/Users/vladd/Desktop/magistral/deploy/env/iam.env.template)

## Запуск

Локально:

```powershell
cd iam
go run ./cmd/grpc-server
```

Через контейнеры:

```powershell
task docker:up:iam
task docker:up:envoy
```

## Проверка

- IAM healthcheck: `http://localhost:8082/healthz`
- Gateway healthcheck: `http://localhost:8080/healthz`
- HTTP вход через gateway: `http://localhost:8080/api/iam/...`
- integration tests: `iam/tests/integration`
- unit tests на `ext_authz`: [internal/api/envoyauth/v1/api_test.go](C:/Users/vladd/Desktop/magistral/iam/internal/api/envoyauth/v1/api_test.go)

## Что важно помнить при изменениях

- если меняется логика identity headers, нужно синхронно проверять `platform/pkg/authn`
- если меняется auth flow, нужно обновлять `deploy/compose/envoy/envoy.yaml`
- если меняется gRPC контракт IAM, нужно регенерировать код в `shared/pkg/iam/v1`
- downstream HTTP-сервисы должны опираться на заголовки от `Envoy`, а не восстанавливать старые gRPC-клиенты в IAM
