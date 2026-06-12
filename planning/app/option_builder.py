from __future__ import annotations

import os
from collections import Counter, defaultdict
from collections.abc import Sequence
from time import perf_counter

from .constraints import available_for_lesson, transmission_matches
from .models import (
    CandidateLesson,
    CarSlot,
    InstructorAvailableTimeSlot,
    LessonOption,
    PlanningContext,
    TimeSlot,
)


MIN_OPTIONS_PER_STUDENT = 48
MAX_OPTIONS_PER_STUDENT = 600
OPTION_SCALE_FACTOR = 1800


def build_car_slots(time_slots: Sequence[TimeSlot], car_ids: Sequence[str]) -> list[CarSlot]:
    car_slots: list[CarSlot] = []
    for time_slot in time_slots:
        for car_id in car_ids:
            car_slots.append(
                CarSlot(
                    option_slot_id=len(car_slots),
                    car_id=car_id,
                    time_slot=time_slot,
                )
            )
    return car_slots


def build_instructor_available_time_slots(
    time_slots: Sequence[TimeSlot],
    context: PlanningContext,
) -> dict[str, list[InstructorAvailableTimeSlot]]:
    result: dict[str, list[InstructorAvailableTimeSlot]] = {}
    for instructor_id in context.instructor_ids:
        slots: list[InstructorAvailableTimeSlot] = []
        for time_slot in time_slots:
            probe = CandidateLesson(
                student_id="",
                instructor_id=instructor_id,
                car_id="",
                start_at=time_slot.start_at,
                end_at=time_slot.end_at,
            )
            if available_for_lesson(probe, context.instructor_availability.get(instructor_id, [])):
                slots.append(InstructorAvailableTimeSlot(instructor_id=instructor_id, time_slot=time_slot))
        result[instructor_id] = slots
    return result


def build_valid_options(
    time_slots: Sequence[TimeSlot],
    context: PlanningContext,
) -> tuple[dict[str, list[LessonOption]], int]:
    verbose = _env_flag("PLANNING_VERBOSE")
    started_at = perf_counter()
    if verbose:
        _progress(
            "build_valid_options started: "
            f"students={len(context.students_to_plan)} "
            f"time_slots={len(time_slots)} "
            f"cars={len(context.car_ids)} "
            f"instructors={len(context.instructor_ids)}"
        )
    car_slots = build_car_slots(time_slots, context.car_ids)
    instructor_slots = build_instructor_available_time_slots(time_slots, context)
    instructors_by_slot = _instructors_by_slot(instructor_slots)
    max_options = max_options_per_student(len(car_slots))
    if verbose:
        _progress(
            "option input prepared: "
            f"car_slots={len(car_slots)} "
            f"instructor_slot_links={sum(len(slots) for slots in instructor_slots.values())} "
            f"max_options_per_student={max_options}"
        )

    options_by_student: dict[str, list[LessonOption]] = {}
    next_option_id = 0
    busy_instructors, busy_cars = _existing_resource_busy_sets(context)
    seed_options_by_student, next_option_id = _build_greedy_seed_options(
        car_slots,
        instructors_by_slot,
        context,
        busy_instructors,
        busy_cars,
        next_option_id,
    )
    for student_number, student_id in enumerate(context.students_to_plan, start=1):
        seed_options = seed_options_by_student.get(student_id, [])
        seen_option_keys = {
            (option.time_slot.index, option.car_id, option.instructor_id)
            for option in seed_options
        }
        compatible_car_slots = [
            car_slot
            for car_slot in car_slots
            if _student_can_use_car(student_id, car_slot.car_id, context)
            and (car_slot.time_slot.start_at, car_slot.car_id) not in busy_cars
            and _car_available(car_slot, context)
        ]
        selected_car_slots = _sample_rotated(
            compatible_car_slots,
            max(0, max_options - len(seed_options)),
            rotation=(student_number - 1) * max(1, max_options),
        )
        options: list[LessonOption] = list(seed_options)
        for car_slot_index, car_slot in enumerate(selected_car_slots):
            instructors = instructors_by_slot.get(car_slot.time_slot.index, [])
            instructor_id = _select_available_instructor(
                instructors,
                busy_instructors,
                car_slot.time_slot.start_at,
                car_slot_index,
            )
            if not instructor_id:
                continue
            option_key = (car_slot.time_slot.index, car_slot.car_id, instructor_id)
            if option_key in seen_option_keys:
                continue
            options.append(
                LessonOption(
                    id=next_option_id,
                    student_id=student_id,
                    instructor_id=instructor_id,
                    car_id=car_slot.car_id,
                    time_slot=car_slot.time_slot,
                )
            )
            seen_option_keys.add(option_key)
            next_option_id += 1
        options_by_student[student_id] = options
        if verbose:
            _progress(
                "student options built: "
                f"{student_number}/{len(context.students_to_plan)} "
                f"student={student_id} "
                f"seed_options={len(seed_options)} "
                f"compatible_car_slots={len(compatible_car_slots)} "
                f"limited_options={len(options_by_student[student_id])} "
                f"elapsed={perf_counter() - started_at:.1f}s"
            )

    if verbose:
        _progress(
            "build_valid_options finished: "
            f"students={len(options_by_student)} "
            f"valid_options={sum(len(options) for options in options_by_student.values())} "
            f"elapsed={perf_counter() - started_at:.1f}s"
        )
    return options_by_student, len(car_slots)


def max_options_per_student(total_car_slots: int) -> int:
    if total_car_slots <= 0:
        return 0
    scaled = max(MIN_OPTIONS_PER_STUDENT, OPTION_SCALE_FACTOR * MAX_OPTIONS_PER_STUDENT // total_car_slots)
    return min(MAX_OPTIONS_PER_STUDENT, scaled)


def _limit_options_evenly(options: list[LessonOption], limit: int) -> list[LessonOption]:
    if limit <= 0 or len(options) <= limit:
        return options

    coverage_options = _coverage_options(options)
    if len(coverage_options) >= limit:
        return _sample_evenly(coverage_options, limit)

    selected_ids = {option.id for option in coverage_options}
    remaining = [option for option in options if option.id not in selected_ids]
    return coverage_options + _sample_evenly(remaining, limit - len(coverage_options))


def _coverage_options(options: list[LessonOption]) -> list[LessonOption]:
    grouped: dict[tuple[int, str], list[LessonOption]] = defaultdict(list)
    for option in options:
        grouped[(option.time_slot.index, option.car_id)].append(option)

    coverage: list[LessonOption] = []
    for group_index, group in enumerate(grouped.values()):
        coverage.append(group[group_index % len(group)])
    return coverage


def _sample_evenly(options: list[LessonOption], limit: int) -> list[LessonOption]:
    if limit <= 0:
        return []
    if len(options) <= limit:
        return options
    if limit == 1:
        return [options[0]]

    last = len(options) - 1
    return [options[round(index * last / (limit - 1))] for index in range(limit)]


def _sample_rotated(options: list[CarSlot], limit: int, rotation: int) -> list[CarSlot]:
    if limit <= 0:
        return []
    if len(options) <= limit:
        return options
    start = rotation % len(options)
    return [options[(start + index) % len(options)] for index in range(limit)]


def _instructors_by_slot(
    instructor_slots: dict[str, list[InstructorAvailableTimeSlot]],
) -> dict[int, list[str]]:
    result: dict[int, list[str]] = defaultdict(list)
    for instructor_id, slots in instructor_slots.items():
        for slot in slots:
            result[slot.time_slot.index].append(instructor_id)
    return dict(result)


def _existing_resource_busy_sets(context: PlanningContext) -> tuple[set[tuple[object, str]], set[tuple[object, str]]]:
    busy_instructors = {
        (lesson.start_at, lesson.instructor_id)
        for lesson in context.existing_lessons
    }
    busy_cars = {
        (lesson.start_at, lesson.car_id)
        for lesson in context.existing_lessons
    }
    return busy_instructors, busy_cars


def _build_greedy_seed_options(
    car_slots: list[CarSlot],
    instructors_by_slot: dict[int, list[str]],
    context: PlanningContext,
    existing_busy_instructors: set[tuple[object, str]],
    existing_busy_cars: set[tuple[object, str]],
    next_option_id: int,
) -> tuple[dict[str, list[LessonOption]], int]:
    remaining = Counter(context.lesson_demands)
    if not remaining:
        return {}, next_option_id

    options_by_student: dict[str, list[LessonOption]] = defaultdict(list)
    busy_students: set[tuple[object, str]] = set()
    busy_instructors = set(existing_busy_instructors)
    busy_cars = set(existing_busy_cars)
    student_ids = list(context.students_to_plan)

    for car_slot in _seed_car_slot_order(car_slots, context):
        if not any(count > 0 for count in remaining.values()):
            break
        start_at = car_slot.time_slot.start_at
        if (start_at, car_slot.car_id) in busy_cars:
            continue
        if not _car_available(car_slot, context):
            continue

        instructor_id = _select_available_instructor(
            instructors_by_slot.get(car_slot.time_slot.index, []),
            busy_instructors,
            start_at,
            car_slot.option_slot_id,
        )
        if not instructor_id:
            continue

        student_id = _select_greedy_student(student_ids, remaining, car_slot, context, busy_students)
        if not student_id:
            continue

        options_by_student[student_id].append(
            LessonOption(
                id=next_option_id,
                student_id=student_id,
                instructor_id=instructor_id,
                car_id=car_slot.car_id,
                time_slot=car_slot.time_slot,
            )
        )
        next_option_id += 1
        remaining[student_id] -= 1
        busy_students.add((start_at, student_id))
        busy_instructors.add((start_at, instructor_id))
        busy_cars.add((start_at, car_slot.car_id))

    return dict(options_by_student), next_option_id


def _seed_car_slot_order(car_slots: list[CarSlot], context: PlanningContext) -> list[CarSlot]:
    compatible_counts: dict[str, int] = {}
    for car_id in context.car_ids:
        compatible_counts[car_id] = sum(
            1
            for student_id in context.students_to_plan
            if _student_can_use_car(student_id, car_id, context)
        )
    return sorted(
        car_slots,
        key=lambda car_slot: (
            car_slot.time_slot.index,
            compatible_counts.get(car_slot.car_id, len(context.students_to_plan)),
            car_slot.car_id,
        ),
    )


def _select_greedy_student(
    student_ids: list[str],
    remaining: Counter,
    car_slot: CarSlot,
    context: PlanningContext,
    busy_students: set[tuple[object, str]],
) -> str:
    start_at = car_slot.time_slot.start_at
    candidates = [
        student_id
        for student_id in student_ids
        if remaining[student_id] > 0
        and (start_at, student_id) not in busy_students
        and _student_can_use_car(student_id, car_slot.car_id, context)
    ]
    if not candidates:
        return ""

    candidates.sort(
        key=lambda student_id: (
            -remaining[student_id],
            -_preference_score(student_id, car_slot, context),
            student_id,
        )
    )
    return candidates[0]


def _preference_score(student_id: str, car_slot: CarSlot, context: PlanningContext) -> int:
    preference = context.student_preferences.get(student_id)
    if preference is None:
        return 0
    score = 0
    weekday = car_slot.time_slot.start_at.strftime("%A").lower()
    hour = car_slot.time_slot.start_at.hour
    time_window = "morning" if hour < 12 else "daytime" if hour < 17 else "evening"
    if weekday in preference.preferred_weekdays:
        score += 1
    if time_window in preference.preferred_time_windows:
        score += 1
    return score


def _select_available_instructor(
    instructors: list[str],
    busy_instructors: set[tuple[object, str]],
    start_at,
    rotation_index: int,
) -> str:
    if not instructors:
        return ""
    for offset in range(len(instructors)):
        instructor_id = instructors[(rotation_index + offset) % len(instructors)]
        if (start_at, instructor_id) not in busy_instructors:
            return instructor_id
    return ""


def _student_can_use_car(student_id: str, car_id: str, context: PlanningContext) -> bool:
    return transmission_matches(
        context.student_transmissions.get(student_id, ""),
        context.car_transmissions.get(car_id, ""),
    )


def _car_available(car_slot: CarSlot, context: PlanningContext) -> bool:
    probe = CandidateLesson(
        student_id="",
        instructor_id="",
        car_id=car_slot.car_id,
        start_at=car_slot.time_slot.start_at,
        end_at=car_slot.time_slot.end_at,
    )
    return available_for_lesson(probe, context.car_availability.get(car_slot.car_id, []))


def _env_flag(name: str) -> bool:
    return os.getenv(name, "").strip().lower() in {"1", "true", "yes", "on"}


def _progress(message: str) -> None:
    print(f"[planning] {message}", flush=True)
