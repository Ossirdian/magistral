# Greedy + GA В Planning

Planning использует гибридный алгоритм:

```text
greedy seed -> ограниченный набор LessonOption -> genetic algorithm
```

Цель greedy-части - быстро построить валидное расписание без конфликтов. Цель GA - улучшить распределение по предпочтениям, равномерности и загрузке ресурсов.

GA не ищет расписание с нуля. Он работает только с заранее подготовленными валидными вариантами занятий.

Если greedy seed уже закрыл весь demand без конфликтов, GA переключается в режим `complete_plan_optimizer`:

- популяция строится из greedy-расписания;
- crossover отключается;
- мутация меняет отдельное занятие только если после decode остаются все занятия;
- особи с потерянными занятиями или конфликтами получают технический штраф ниже greedy.

В этом режиме GA оптимизирует предпочтения и равномерность, не борясь заново за закрытие demand.

## Входной Snapshot

Перед запуском алгоритма `Planner` загружает:

- студентов;
- целевой demand по каждому студенту;
- предпочтения студентов;
- трансмиссии студентов;
- активные автомобили;
- трансмиссии автомобилей;
- availability автомобилей;
- активных инструкторов;
- availability инструкторов;
- существующие занятия за период.

Привязка автомобиля к инструктору не используется. Любой доступный инструктор может быть выбран с любым подходящим доступным автомобилем.

## Подготовка Слотов

`Planner` строит дневные временные слоты. По умолчанию рабочее окно автомобиля:

```text
08:00-20:00
```

Если у автомобиля нет окон availability в `resources`, planning создает эти окна в памяти. Если окна есть, используются они.

Далее строятся автомобильные слоты:

```text
car_slot = time_slot + car
```

Инструкторы индексируются по доступным слотам на основе своих availability-окон.

## Построение Options

`option_builder` строит `valid_options_by_student`:

```text
student_id -> []LessonOption
```

`LessonOption` содержит:

- студента;
- инструктора;
- автомобиль;
- временной слот.

Option считается допустимым, если:

- автомобиль доступен;
- инструктор доступен;
- трансмиссия студента и автомобиля совпадает;
- существующие занятия не занимают тот же автомобиль или того же инструктора в этот слот.

## Greedy Seed

Greedy seed проходит по автомобильным слотам и выбирает:

1. доступный автомобиль;
2. доступного инструктора;
3. студента с оставшимся demand;
4. студента с подходящей трансмиссией;
5. вариант с лучшим совпадением предпочтений.

Для контроля конфликтов используются set-индексы:

```text
(start_at, student_id)
(start_at, instructor_id)
(start_at, car_id)
```

Слоты дефицитных автомобилей обрабатываются раньше. Это важно для датасетов, где есть один автомобиль на автомате и много автомобилей на механике.

Greedy seed добавляется в начало `valid_options_by_student`. Первая особь GA строится так, чтобы могла воспроизвести greedy-расписание.

## Ограничение Количества Options

Чтобы GA не работал с огромным количеством вариантов, options ограничиваются:

```text
max_options_per_student = min(600, max(48, 1800 * 600 / total_car_slots))
```

Параметры:

- минимум: `48` options на студента;
- максимум: `600` options на студента;
- масштаб: `1800`.

Чем больше автомобильных слотов, тем меньше лимит на одного студента. Это удерживает размер поиска в разумных пределах.

В диагностике результата выводятся:

- `valid_options` - общее количество options;
- `avg_options_per_student` - среднее количество options на студента;
- `students_with_valid_options` - сколько студентов получили хотя бы один option;
- `total_car_slots` - количество автомобильных слотов.

## Кодирование Особей

Особь GA состоит из пар генов на каждый demand:

```text
student_index, option_index
```

- `student_index` соответствует студенту demand-позиции;
- `option_index` выбирает один `LessonOption` из списка студента.

GA не выбирает отдельные `time_slot`, `car_id` или `instructor_id`. Эти значения уже зафиксированы внутри `LessonOption`.

## Decode

При decode алгоритм идет по `lesson_demands` и пытается поставить выбранный option.

Если выбранный option конфликтует с уже занятым студентом, инструктором или автомобилем, decode циклически ищет следующий option этого студента.

Если свободного option нет, demand помечается как:

```text
no_available_option
```

Сейчас такие случаи попадают в список conflicts результата. Это не обязательно означает физический конфликт ресурсов; часто это означает, что demand не удалось разместить в доступной емкости.

## Параметры GA

Параметры по умолчанию:

- `seed = 42`;
- `population_size = 100`;
- `generations = 50`;
- `cxpb = 0.8`;
- `mutpb = 0.3`;
- `indpb = 0.3`;
- `tournsize = 3`;
- `HallOfFame(1)`;
- elitism: лучшая особь переносится в следующее поколение.

Переменные окружения:

```text
PLANNING_GA_POPULATION
PLANNING_GA_GENERATIONS
PLANNING_GA_MAX_SECONDS
```

Для тестового task:

```yaml
PLANNING_GA_POPULATION: 20
PLANNING_GA_GENERATIONS: 8
PLANNING_GA_MAX_SECONDS: 60
```

`PLANNING_GA_MAX_SECONDS` ограничивает только GA-цикл. Если время вышло, возвращается лучший найденный результат из `HallOfFame`.

Эти параметры добавляются в `score_breakdown`:

- `ga_population`;
- `ga_generations`;
- `ga_max_seconds`.

## Verbose-Режим

Для подробного вывода:

```powershell
$env:PLANNING_VERBOSE='1'
$env:PLANNING_GA_VERBOSE='1'
```

Пример строк:

```text
[planning] loaded data: students=67 instructors=11 cars=11 existing_lessons=100 lesson_demands=1868
[planning] built time slots: slots_per_resource=240
[planning] built valid options: total_car_slots=2640 valid_options=26047 avg_options_per_student=388.8
[planning] greedy baseline: lessons=1868 unassigned_students=0 conflicts=0 score=201585.72 mode=complete_plan_optimizer
```

Обычный worker-лог также содержит:

```text
valid_options=...
avg_options_per_student=...
```

## Fitness

Формула:

```text
scheduled_bonus
- unmet_lesson_penalty
- conflict_penalty
- idle_penalty
- unevenness_penalty
- resource_unevenness_penalty
+ preference_bonus
```

Компоненты:

- `scheduled_bonus = scheduled_lessons * 100`;
- `unmet_lesson_penalty = missing_target_lessons * 300`;
- `conflict_penalty = conflicts * 250`;
- `idle_penalty = unused_effective_capacity * 20`;
- `unevenness_penalty` штрафует перекос закрытия программы между студентами;
- `resource_unevenness_penalty` штрафует перекос загрузки ресурсов;
- `preference_bonus` дает `+10` за совпадение предпочтения и `-5` за несовпадение.

Автомобили сравниваются по равномерности внутри своей трансмиссии, чтобы единственный автомобиль на автомате не штрафовался относительно механики.

## Текущий Месячный Датасет

Команда:

```powershell
task planning:dataset:test
```

Состав:

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

## Benchmark На 2 Месяца

Команда:

```powershell
task planning:dataset:benchmark:2months
```

Состав:

- `134` студента;
- `118` студентов на механике по `28` занятий;
- `16` студентов на автомате по `27` занятий;
- `11` автомобилей, из них `1` автомат;
- `11` инструкторов;
- `200` существующих занятий других групп;
- период `60` дней;
- цель основной группы: `3736` занятий;
- общий спрос с существующими занятиями: `3936` занятий;
- доступно `4000` слотов по `90` минут.

Параметры GA для benchmark task:

```yaml
PLANNING_GA_POPULATION: 30
PLANNING_GA_GENERATIONS: 12
PLANNING_GA_MAX_SECONDS: 60
```

Проверенный результат:

```text
scheduled_lessons: 3736
missing_target_lessons: 0
conflicts: 0
car_idle_percent: 1.60%
instructor_idle_percent: 1.60%
valid_options: 27275
avg_options_per_student: 203.54
greedy_fitness_score: 417600.888
best_fitness_score: 417630.288
```

HTML-визуализация benchmark:

```powershell
task planning:dataset:benchmark:2months:visualize
```

Создает:

```text
planning/generated/planning_benchmark_2months_visualization.html
```

## HTML-Отчет

Команда:

```powershell
task planning:dataset:visualize
```

Создает:

```text
planning/generated/planning_visualization.html
```

Отчет показывает:

- сводку по студентам, авто и инструкторам;
- календарь с количеством занятий по дням;
- загрузку и простой автомобилей;
- загрузку и простой инструкторов;
- fitness score и breakdown на русском;
- остаток занятий по студентам;
- итоговое расписание.
