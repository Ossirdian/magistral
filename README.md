# Magistral

Монорепозиторий платформы автошколы с HTTP ingress через `envoy`, отдельным frontend и набором backend-сервисов на Go и Python.

## Состав проекта

- `frontend` — SPA для ручной работы с системой
- `envoy` — единая внешняя HTTP точка входа
- `iam` — аутентификация, пользователи, роли, gRPC authz для Envoy
- `saas` — школы и филиалы
- `resources` — инструкторы, машины, доступность
- `student` — ученики и группы
- `lesson` — расписание, конфликты, batch apply для planning
- `payment` — платежи, договоры, балансы
- `document` — документы учеников и шаблоны
- `reporting` — агрегированные read-only отчёты
- `planning` — Python worker для генерации расписания через Kafka
- `shared` — OpenAPI, proto и сгенерированный код
- `platform` — общая инфраструктурная библиотека

## Актуальная архитектура

- Внешний HTTP трафик идёт только через `envoy`.
- `envoy` делает `ext_authz` в `iam` по gRPC `envoy.service.auth.v3.Authorization/Check`.
- После успешной проверки `envoy` прокидывает trusted headers `x-user-*` в backend HTTP сервисы.
- HTTP сервисы не ходят в `iam` за авторизацией напрямую и доверяют только заголовкам от `envoy`.
- Внутренние сервисы общаются между собой через gRPC там, где уже есть runtime-контракт.
- `payment` и `document` используют internal gRPC `student` для проверки принадлежности `student_id`.
- `planning` работает асинхронно через Kafka и применяет результат в `lesson` через batch endpoint.
- `reporting` в текущем MVP строит агрегаты прямым чтением storage `student`, `lesson`, `resources`, `payment`.

## Source Of Truth

- `iam` — users, credentials, token validation
- `saas` — schools, branches
- `resources` — instructors, cars, availability
- `student` — students, groups
- `lesson` — lessons, schedule conflicts
- `payment` — payments, contracts, balances
- `document` — documents, document templates
- `reporting` — только read model, не источник истины
- `planning` — только генерация, не источник истины

## HTTP маршруты через Envoy

- `/` -> `frontend`
- `/api/iam/*` -> `iam`
- `/api/saas/*` -> `saas`
- `/api/resources/*` -> `resources`
- `/api/students/*` -> `student`
- `/api/lessons/*` -> `lesson`
- `/api/payments/*` -> `payment`
- `/api/documents/*` -> `document`
- `/api/reporting/*` -> `reporting`

## Сервисы и ответственность

### `iam`

- login, refresh, `me`
- управление school-scoped users
- gRPC authz для `envoy`

### `saas`

- CRUD школ
- CRUD филиалов
- bootstrap школы с первым admin

### `resources`

- CRUD инструкторов
- CRUD автомобилей
- availability
- assignment `car -> instructor`

### `student`

- CRUD учеников
- CRUD групп
- internal gRPC `GetStudent`

Важно:

- оркестрация planning убрана из `student`
- `student` больше не владеет расписанием

### `lesson`

- CRUD занятий
- проверка конфликтов по `student_id`, `instructor_id`, `car_id`
- фильтрация расписания
- `POST /internal/v1/lessons/batch` для результата planning

### `payment`

- CRUD платежей
- договоры учеников
- расчёт остатка

### `document`

- документы учеников
- шаблоны документов
- генерация документов

### `reporting`

- dashboard
- students report
- lessons report
- finance report
- instructor workload
- car utilization

### `planning`

- чтение задач из Kafka
- генерация индивидуальных занятий
- отправка результата в `lesson`

## Контракты

OpenAPI:

- `shared/api/resources/v1/resources.openapi.yaml`
- `shared/api/saas-admin/v1/saas-admin.openapi.yaml`
- `shared/api/students/v1/students.openapi.yaml`
- `shared/api/lessons/v1/lessons.openapi.yaml`
- `shared/api/payments/v1/payments.openapi.yaml`
- `shared/api/documents/v1/documents.openapi.yaml`
- `shared/api/reporting/v1/reporting.openapi.yaml`

Generated packages:

- `shared/pkg/resources/v1`
- `shared/pkg/saasadmin/v1`
- `shared/pkg/students/v1`
- `shared/pkg/lessons/v1`
- `shared/pkg/payments/v1`
- `shared/pkg/documents/v1`
- `shared/pkg/reporting/v1`

Proto:

- `shared/proto/iam/v1`
- `shared/proto/student/v1`

## Быстрый старт

1. Сгенерировать env-файлы:

```powershell
task env:generate
```

2. При необходимости пересгенерировать код из контрактов:

```powershell
task gen
```

3. Поднять всё окружение:

```powershell
task docker:up:all
```

4. Открыть систему:

```text
http://localhost:3000
```

5. Остановить окружение:

```powershell
task docker:down:all
```

## Docker и compose

Сборка образов идёт через `docker buildx bake`.

Основные команды:

```powershell
task docker:build:lesson
task docker:build:student
task docker:build:payment
task docker:build:document
task docker:build:reporting
task docker:build:all
```

Запуск сервисов по отдельности:

```powershell
task docker:up:core
task docker:up:iam
task docker:up:saas
task docker:up:resources
task docker:up:student
task docker:up:lesson
task docker:up:payment
task docker:up:document
task docker:up:reporting
task docker:up:planning
task docker:up:frontend
task docker:up:envoy
```

## Observability

В проекте подключены:

- `Prometheus`
- `Grafana`
- `Jaeger`
- `ELK` для логов

Метрики снимаются с:

- `app-iam:8082/metrics`
- `app-saas:8081/metrics`
- `app-resources:8080/metrics`
- `app-student:8083/metrics`
- `app-lesson:8087/metrics`
- `app-payment:8084/metrics`
- `app-document:8085/metrics`
- `app-reporting:8086/metrics`
- `app-planning:9101/metrics`
- `app-envoy:9901/stats/prometheus`

UI:

- `Envoy / frontend entrypoint`: `http://localhost:3000`
- `Grafana`: `http://localhost:3001`
- `Prometheus`: `http://localhost:9090`
- `Kibana`: `http://localhost:5601`
- `Jaeger`: `http://localhost:16686`

## Порты по умолчанию

- `envoy`: `3000`
- `iam` gRPC: `50053`
- `student` gRPC: `50054`
- `planning` metrics: `9101`
- `postgres-core`: `5432`
- `jaeger`: `16686`

Остальные HTTP сервисы работают внутри Docker-сети и доступны снаружи через `envoy`.

## Логин по умолчанию

Super-admin создаётся автоматически из env:

- `IAM_BOOTSTRAP_SUPER_ADMIN_EMAIL`
- `IAM_BOOTSTRAP_SUPER_ADMIN_PASSWORD`

Значения по умолчанию:

```text
superadmin@example.com
Passw0rd!
```

## Полезные команды

Форматирование и линт:

```powershell
task format
task lint
```

Тесты:

```powershell
task test
task test-integration
```

Генерация:

```powershell
task ogen:gen:all
task proto:gen
task gen
```

Обновление зависимостей:

```powershell
task deps:update
```

## Важные файлы

- [Taskfile.yml](/C:/Users/vladd/Desktop/magistral/Taskfile.yml:1)
- [relations.md](/C:/Users/vladd/Desktop/magistral/relations.md:1)
- [deploy/compose/envoy/envoy.yaml](/C:/Users/vladd/Desktop/magistral/deploy/compose/envoy/envoy.yaml:1)
- [deploy/compose/core/prometheus/prometheus.yml](/C:/Users/vladd/Desktop/magistral/deploy/compose/core/prometheus/prometheus.yml:1)
- [deploy/env/.env.template](/C:/Users/vladd/Desktop/magistral/deploy/env/.env.template:1)

## Ограничения текущей версии

- `reporting` пока читает агрегаты напрямую из storage, а не через отдельные read APIs.
- `planning` остаётся worker-сервисом и ещё не покрыт полноценным runtime read-contract слоем ко всем источникам.
- frontend рассчитан на запуск в контейнерах через `envoy` и `nginx`.
