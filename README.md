# Magistral

`Magistral` - монорепозиторий CRM-платформы для автошколы. В проекте есть единый HTTP ingress через `envoy`, SPA-интерфейс `Magistral CRM`, набор Go-сервисов, Python worker для планирования и общие OpenAPI/proto контракты.

## Состав Проекта

- `frontend` - React/Vite SPA `Magistral CRM`.
- `envoy` - единая внешняя HTTP-точка входа.
- `iam` - аутентификация, пользователи, роли, gRPC authz для Envoy.
- `saas` - школы и филиалы.
- `resources` - инструкторы, автомобили, окна доступности.
- `student` - ученики, группы, программы обучения.
- `lesson` - занятия, расписание, batch apply результата planning.
- `payment` - платежи, договоры, балансы.
- `document` - документы учеников и шаблоны.
- `reporting` - агрегированные read-only отчеты.
- `planning` - Python worker для генерации расписания через Kafka, greedy + GA.
- `audit` - журнал HTTP-действий пользователей.
- `shared` - OpenAPI, proto и сгенерированные клиенты/серверные интерфейсы.
- `platform` - общая инфраструктура: логирование, tracing, audit middleware, HTTP helpers.

## Архитектура

- Внешний HTTP-трафик идет через `envoy`.
- `envoy` вызывает `iam` по gRPC `envoy.service.auth.v3.Authorization/Check`.
- После успешной авторизации `envoy` прокидывает trusted headers `x-user-*` и `x-school-id`.
- Backend HTTP-сервисы доверяют только trusted headers от `envoy`.
- `payment` и `document` проверяют студентов через internal gRPC `student`.
- `reporting` в текущей версии строит агрегаты прямым чтением storage сервисов.
- `planning` работает асинхронно через Kafka.
- `lesson` принимает planning request, публикует его в Kafka, применяет результат planning и хранит in-memory статус запроса.

## Source Of Truth

- `iam` - пользователи, credentials, роли, token validation.
- `saas` - школы и филиалы.
- `resources` - инструкторы, автомобили, availability.
- `student` - ученики, группы, программы обучения.
- `lesson` - занятия и конфликты расписания.
- `payment` - платежи, договоры, балансы.
- `document` - документы и шаблоны.
- `audit` - журнал действий.
- `reporting` - read model, не источник истины.
- `planning` - генератор расписания, не источник истины.

## HTTP Маршруты Через Envoy

- `/` -> `frontend`
- `/api/iam/*` -> `iam`
- `/api/saas/*` -> `saas`
- `/api/audit/*` -> `audit`
- `/api/resources/*` -> `resources`
- `/api/students/*` -> `student`
- `/api/lessons/*` -> `lesson`
- `/api/payments/*` -> `payment`
- `/api/documents/*` -> `document`
- `/api/reporting/*` -> `reporting`

Отдельный compatibility route:

- `/api/students/planning/requests` -> `lesson /planning/requests`

## Основные Возможности

### Magistral CRM

Frontend работает в двух режимах:

- `super_admin` - школы, филиалы, пользователи IAM, audit.
- `admin|owner|instructor` - рабочее место школы.

В рабочем месте доступны:

- ресурсы: автомобили, инструкторы, availability-окна;
- ученики, группы и программы обучения;
- занятия с фильтрами и сортировкой;
- платежи и баланс ученика;
- документы и шаблоны;
- reporting;
- planning request с отображением статуса `queued -> completed`.

### Resources

`resources` управляет инструкторами, автомобилями и окнами доступности.

Важные правила:

- planning использует availability автомобилей из `resources`;
- если у авто нет окон availability, planning считает авто доступным каждый день `08:00-20:00`;
- привязка `car -> instructor` оставлена как legacy API, но основной UI и planning ее не используют;
- инструктор для занятия выбирается динамически по availability.

### Lesson

`lesson` отвечает за занятия и применение результата planning.

Поддержано:

- CRUD занятий;
- проверка конфликтов по студенту, инструктору и автомобилю;
- фильтрация по ученику, инструктору, автомобилю, статусу, периоду;
- сортировка `sort_by=start_at|student_id|instructor_id|car_id`;
- порядок `sort_order=asc|desc`;
- лимит списка занятий до `500`;
- календарная сводка занятий по дням во frontend;
- `POST /planning/requests` для постановки задачи planning;
- `GET /planning/requests/{request_id}` для статуса planning request.

Статус planning request хранится in-memory в `lesson`, поэтому после рестарта сервиса история статусов пропадает.

### Planning

`planning` строит индивидуальные занятия через гибрид:

```text
greedy seed -> ограниченный набор LessonOption -> GA
```

Основные правила:

- занятие длится `90` минут;
- студент, инструктор и автомобиль не могут пересекаться по времени;
- трансмиссия студента должна совпадать с трансмиссией авто;
- availability инструкторов и авто является hard constraint;
- старые generated-занятия игнорируются в режимах `merge` и `replace_generated_only`;
- привязка автомобиля к инструктору не используется.

Demand по студентам:

- сначала используются `remaining_lessons_required_max`, `remaining_lessons`, `target_lessons`;
- если их нет, fallback идет через `remaining_hours / 1.5`;
- fallback ограничивается программой: `manual = 28`, `automatic = 27`.

В `score_breakdown` попадают:

- fitness score и его компоненты;
- `valid_options`;
- `avg_options_per_student`;
- `students_with_valid_options`;
- `total_car_slots`;
- `ga_population`;
- `ga_generations`;
- `ga_max_seconds`;
- статистика существующих занятий.

Подробнее:

- [planning/README.md](./planning/README.md)
- [planning/GA.md](./planning/GA.md)

### Reporting

`reporting` отдает агрегаты:

- dashboard;
- отчет по ученикам;
- отчет по занятиям;
- финансы;
- загрузка инструкторов;
- utilization автомобилей.

## Контракты

OpenAPI:

- `shared/api/resources/v1/resources.openapi.yaml`
- `shared/api/saas-admin/v1/saas-admin.openapi.yaml`
- `shared/api/students/v1/students.openapi.yaml`
- `shared/api/lessons/v1/lessons.openapi.yaml`
- `shared/api/payments/v1/payments.openapi.yaml`
- `shared/api/documents/v1/documents.openapi.yaml`
- `shared/api/reporting/v1/reporting.openapi.yaml`

Сгенерированные пакеты:

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

После изменения контрактов:

```powershell
task gen
```

## Быстрый Старт

1. Сгенерировать env-файлы:

```powershell
task env:generate
```

2. При необходимости пересгенерировать контракты:

```powershell
task gen
```

3. Поднять окружение:

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

## Логин По Умолчанию

Super admin создается из env:

- `IAM_BOOTSTRAP_SUPER_ADMIN_EMAIL`
- `IAM_BOOTSTRAP_SUPER_ADMIN_PASSWORD`

Значения по умолчанию:

```text
superadmin@example.com
Passw0rd!
```

## Docker И Compose

Сборка образов идет через `docker buildx bake`.

Основные команды:

```powershell
task docker:build:all
task docker:up:all
task docker:down:all
```

Запуск отдельных сервисов:

```powershell
task docker:up:core
task docker:up:iam
task docker:up:saas
task docker:up:audit
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

После изменений во frontend или backend-коде пересоберите нужные контейнеры:

```powershell
docker compose up -d --build app-frontend app-lesson app-planning
```

Если имена сервисов отличаются:

```powershell
docker compose ps
```

## Очистка Локальных Артефактов

В `.gitignore` и `.dockerignore` исключены локальные cache/temp директории:

- `.gocache`;
- `.pytest_cache`;
- `__pycache__`;
- `planning/generated`;
- `planning/ga_tests`;
- `tmp-gomodcache`;
- `tmp-gocache`.

Если cache-файлы уже были добавлены в git, их нужно снять с индекса:

```powershell
git ls-files | Select-String -Pattern "(__pycache__|\.pyc$)" | ForEach-Object { git rm --cached -- $_.Line }
```

Очистка Docker:

```powershell
docker system df
docker builder prune
docker system prune -a
```

`docker volume prune` используйте осторожно: он может удалить локальные данные Postgres/Kafka.

## Observability

Подключены:

- Prometheus;
- Grafana;
- Jaeger;
- ELK/Kibana;
- audit middleware.

Метрики:

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

- Magistral CRM: `http://localhost:3000`
- Grafana: `http://localhost:3001`
- Prometheus: `http://localhost:9090`
- Kibana: `http://localhost:5601`
- Jaeger: `http://localhost:16686`

## Порты По Умолчанию

- `envoy`: `3000`
- `iam` gRPC: `50053`
- `student` gRPC: `50054`
- `planning` metrics: `9101`
- `postgres-core`: `5432`
- `jaeger`: `16686`

Остальные HTTP-сервисы доступны внутри Docker-сети и снаружи через `envoy`.

## Тесты И Проверки

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

Точечные проверки, которые часто нужны при работе с planning:

```powershell
python -m pytest planning\tests\test_planning_flow.py -q
task planning:dataset:test
task planning:dataset:benchmark:2months
task planning:dataset:benchmark:2months:visualize
task planning:dataset:visualize
```

Если Go test в локальной среде упирается в системный cache, можно задать cache внутри workspace:

```powershell
$env:GOCACHE='C:\Users\vladd\Desktop\magistral\.gocache'
go test ./lesson/...
```

## Важные Файлы

- [Taskfile.yml](./Taskfile.yml)
- [deploy/compose/envoy/envoy.yaml](./deploy/compose/envoy/envoy.yaml)
- [deploy/env/.env.template](./deploy/env/.env.template)
- [frontend/src/App.jsx](./frontend/src/App.jsx)
- [planning/README.md](./planning/README.md)
- [planning/GA.md](./planning/GA.md)

## Ограничения Текущей Версии

- Статусы planning request в `lesson` хранятся in-memory; для production нужна таблица в БД.
- `reporting` пока читает агрегаты напрямую из storage, а не через отдельные read API.
- Legacy endpoint `POST /resources/cars/{id}/assign` оставлен для совместимости, но planning и основной UI его не используют.
- Frontend рассчитан на основной запуск через контейнеры, `envoy` и nginx.
