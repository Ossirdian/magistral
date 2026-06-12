# Сервис Planning

`planning` строит расписание индивидуальных занятий по вождению. Сервис получает задачу из Kafka, загружает актуальный снимок данных из `student`, `resources` и `lesson`, строит варианты занятий и публикует результат обратно в Kafka.

Одно занятие длится `90` минут.

## Общий Поток

1. `lesson` принимает `POST /planning/requests`.
2. `lesson` публикует задачу в Kafka topic `planning.requests`.
3. `planning` загружает студентов, инструкторов, автомобили, доступность ресурсов и существующие занятия.
4. `planning` строит временные слоты по дням в окне `08:00-20:00`.
5. Для автомобилей без явно заданной доступности создаются дефолтные окна `08:00-20:00`.
6. Строятся допустимые варианты занятий: студент, слот, автомобиль, инструктор.
7. Greedy-часть строит базовое расписание без конфликтов.
8. GA получает greedy-расписание как первую особь и улучшает распределение.
9. Результат публикуется в Kafka result topic.
10. `lesson` применяет результат и переводит planning request в `completed`.

## Правила Планирования

- занятие всегда индивидуальное;
- студент не может получить два занятия в один временной слот;
- инструктор не может вести два занятия в один временной слот;
- автомобиль не может быть занят двумя занятиями в один временной слот;
- трансмиссия студента должна совпадать с трансмиссией автомобиля;
- существующие занятия блокируют соответствующего инструктора и автомобиль;
- доступность инструкторов и автомобилей является жестким ограничением;
- в режимах `merge` и `replace_generated_only` старые generated-занятия не блокируют новое расписание;
- привязка автомобиля к инструктору из `resources` не используется: planning выбирает любого доступного инструктора и любой подходящий автомобиль.

## Demand По Студентам

Целевое количество занятий считается в таком порядке:

1. `remaining_lessons_required_max`;
2. `remaining_lessons`;
3. `target_lessons`;
4. fallback из `remaining_hours / 1.5`.

Если используется fallback по часам, planning ограничивает целевой demand программой:

- `manual`: максимум `28` занятий;
- `automatic`: максимум `27` занятий.

Уже пройденные часы учитываются через `total_hours_completed`, либо через сумму `manual_hours_completed + lesson_hours_completed`.

## Greedy И GA

Planning использует гибридный подход:

```text
greedy seed -> ограниченный набор LessonOption -> GA
```

Greedy seed быстро строит валидную базу расписания. Он учитывает:

- доступность автомобиля;
- доступность инструктора;
- занятость студента, инструктора и автомобиля;
- трансмиссию;
- оставшийся demand студента;
- предпочтения по дням недели и времени.

GA не перебирает сырые комбинации `слот + автомобиль + инструктор`. Он работает только с заранее проверенными `LessonOption`, поэтому поиск ограничен валидными вариантами.

## Оценка Расписания

Итоговая оценка считается в `planning/app/fitness.py`:

```text
scheduled_bonus
- unmet_lesson_penalty
- conflict_penalty
- idle_penalty
- unevenness_penalty
- resource_unevenness_penalty
+ preference_bonus
```

Основные веса:

- `scheduled_lesson_bonus_weight = 100`;
- `missing_lesson_penalty_weight = 300`;
- `conflict_penalty_weight = 250`;
- `idle_capacity_penalty_weight = 20`;
- `student_unevenness_penalty_weight = 150`;
- `resource_unevenness_penalty_weight = 12`;
- `preference_match_bonus_weight = 10`;
- `preference_miss_penalty_weight = 5`.

`resource_unevenness_penalty` для автомобилей считается внутри трансмиссии. Единственный автомобиль на автомате не сравнивается с механикой.

В `score_breakdown` также попадают диагностические параметры:

- `valid_options`;
- `avg_options_per_student`;
- `students_with_valid_options`;
- `total_car_slots`;
- `ga_population`;
- `ga_generations`;
- `ga_max_seconds`;
- статистика по существующим занятиям: loaded, used, skipped generated, skipped deleted/cancelled.

## Логи

При включенном `PLANNING_VERBOSE=1` сервис печатает промежуточные этапы:

- загрузку данных;
- количество слотов;
- effective demand;
- количество valid options;
- среднее количество options на студента;
- старт и завершение GA.

Обычный worker-лог после завершения содержит `planning best solution score breakdown`, включая:

```text
scheduled_lessons
missing_target_lessons
conflicts
valid_options
avg_options_per_student
score
replace_strategy
```

## Параметры GA

Параметры по умолчанию:

- `PLANNING_GA_POPULATION = 100`;
- `PLANNING_GA_GENERATIONS = 50`;
- `PLANNING_GA_MAX_SECONDS = 60`.

Для тестового task используется быстрый режим:

```yaml
PLANNING_GA_POPULATION: 20
PLANNING_GA_GENERATIONS: 8
PLANNING_GA_MAX_SECONDS: 60
```

Переопределение в PowerShell:

```powershell
$env:PLANNING_GA_POPULATION='50'
$env:PLANNING_GA_GENERATIONS='25'
$env:PLANNING_GA_MAX_SECONDS='60'
task planning:dataset:test
```

`PLANNING_GA_MAX_SECONDS` ограничивает только GA-цикл. Если лимит достигнут раньше последнего поколения, сервис возвращает лучший найденный результат из `HallOfFame`.

## Тестовый Датасет

Команда:

```powershell
task planning:dataset:test
```

Текущий месячный датасет подобран так, чтобы суммарный простой автомобилей и инструкторов был меньше `5%`:

- `67` студентов;
- `59` студентов на механике по `28` занятий;
- `8` студентов на автомате по `27` занятий;
- `11` автомобилей, из них `1` автомат;
- `11` инструкторов;
- `100` существующих занятий других групп;
- период `30` дней;
- цель основной группы: `1868` занятий;
- общий спрос с существующими занятиями: `1968` занятий.

Ожидаемый результат:

```text
scheduled_lessons: 1868
missing_target_lessons: 0
conflicts: 0
car_idle_percent: 1.60%
instructor_idle_percent: 1.60%
```

Файлы отчета:

```text
planning/generated/planning_test_dataset.json
planning/generated/planning_run_report.json
```

## Benchmark На 2 Месяца

Команда:

```powershell
task planning:dataset:benchmark:2months
```

Датасет рассчитан на `60` дней:

- `134` студента;
- `118` студентов на механике по `28` занятий;
- `16` студентов на автомате по `27` занятий;
- `11` автомобилей, из них `1` автомат;
- `11` инструкторов;
- `200` существующих занятий других групп;
- цель основной группы: `3736` занятий;
- общий спрос с существующими занятиями: `3936` занятий;
- доступно `4000` слотов по `90` минут.

Проверенный результат:

```text
scheduled_lessons: 3736
missing_target_lessons: 0
conflicts: 0
car_idle_percent: 1.60%
instructor_idle_percent: 1.60%
valid_options: 27275
avg_options_per_student: 203.54
```

Файлы отчета:

```text
planning/generated/planning_benchmark_2months_dataset.json
planning/generated/planning_benchmark_2months_report.json
```

HTML-визуализация:

```powershell
task planning:dataset:benchmark:2months:visualize
```

```text
planning/generated/planning_benchmark_2months_visualization.html
```

## Визуализация

Команда:

```powershell
task planning:dataset:visualize
```

Создает:

```text
planning/generated/planning_visualization.html
```

HTML показывает:

- сводку по датасету;
- календарь с количеством занятий по дням;
- загрузку и простой автомобилей;
- загрузку и простой инструкторов;
- лучший fitness score и разбивку;
- остаток занятий по каждому студенту;
- итоговое расписание.

## Структура

- `planning/app/planner.py` - основной orchestration;
- `planning/app/data_loader.py` - загрузка данных из сервисов;
- `planning/app/option_builder.py` - greedy seed и построение options;
- `planning/app/genetic.py` - GA, мутации, decode особей;
- `planning/app/fitness.py` - fitness score;
- `planning/app/kafka_consumer.py` - Kafka worker;
- `planning/generate_test_dataset.py` - генератор тестового датасета;
- `planning/run_test_dataset.py` - прогон датасета через Planner;
- `planning/visualize_test_dataset.py` - HTML-визуализация.

## Проверка

```powershell
python -m pytest planning\tests\test_planning_flow.py -q
task planning:dataset:test
task planning:dataset:benchmark:2months
task planning:dataset:benchmark:2months:visualize
task planning:dataset:visualize
```
