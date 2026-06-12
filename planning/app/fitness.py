from __future__ import annotations

from .models import CandidateLesson, Conflict, PlanningContext


SCHEDULED_LESSON_BONUS = 100.0
MISSING_LESSON_PENALTY = 300.0
CONFLICT_PENALTY = 250.0
IDLE_CAPACITY_PENALTY = 20.0
STUDENT_UNEVENNESS_PENALTY = 150.0
RESOURCE_UNEVENNESS_PENALTY = 12.0
PREFERENCE_MATCH_BONUS = 10.0
PREFERENCE_MISS_PENALTY = 5.0


def score_plan(lessons: list[CandidateLesson], unassigned_students: list[str], conflicts: list[Conflict], context: PlanningContext) -> float:
    return score_breakdown(lessons, unassigned_students, conflicts, context)["score"]


def score_breakdown(lessons: list[CandidateLesson], unassigned_students: list[str], conflicts: list[Conflict], context: PlanningContext) -> dict[str, float]:
    scheduled_bonus = float(len(lessons) * SCHEDULED_LESSON_BONUS)
    missing_target_lessons = float(_missing_target_lessons(lessons, context))
    unmet_penalty = missing_target_lessons * MISSING_LESSON_PENALTY
    conflict_penalty = float(len(conflicts) * CONFLICT_PENALTY)
    idle_penalty = _idle_resource_penalty(lessons, context)
    unevenness_penalty = _student_distribution_penalty(lessons, context)
    resource_unevenness_penalty = _resource_distribution_penalty(lessons, context)
    preference_bonus = _student_preference_bonus(lessons, context)
    score = (
        scheduled_bonus
        - unmet_penalty
        - conflict_penalty
        - idle_penalty
        - unevenness_penalty
        - resource_unevenness_penalty
        + preference_bonus
    )
    return {
        "score": score,
        "scheduled_lessons": float(len(lessons)),
        "scheduled_bonus": scheduled_bonus,
        "missing_target_lessons": missing_target_lessons,
        "unmet_lesson_penalty": unmet_penalty,
        "conflicts": float(len(conflicts)),
        "conflict_penalty": conflict_penalty,
        "idle_penalty": idle_penalty,
        "unevenness_penalty": unevenness_penalty,
        "resource_unevenness_penalty": resource_unevenness_penalty,
        "preference_bonus": preference_bonus,
        "scheduled_lesson_bonus_weight": SCHEDULED_LESSON_BONUS,
        "missing_lesson_penalty_weight": MISSING_LESSON_PENALTY,
        "conflict_penalty_weight": CONFLICT_PENALTY,
        "idle_capacity_penalty_weight": IDLE_CAPACITY_PENALTY,
        "student_unevenness_penalty_weight": STUDENT_UNEVENNESS_PENALTY,
        "resource_unevenness_penalty_weight": RESOURCE_UNEVENNESS_PENALTY,
        "preference_match_bonus_weight": PREFERENCE_MATCH_BONUS,
        "preference_miss_penalty_weight": PREFERENCE_MISS_PENALTY,
    }


def _missing_target_lessons(lessons: list[CandidateLesson], context: PlanningContext) -> int:
    scheduled = _scheduled_counts(lessons)
    missing = 0
    for student_id, target in context.target_lessons_by_student.items():
        missing += max(0, target - scheduled.get(student_id, 0))
    return missing


def _idle_resource_penalty(lessons: list[CandidateLesson], context: PlanningContext) -> float:
    if not context.instructor_ids or not context.car_ids:
        return 25.0
    resource_lesson_capacity = min(
        _total_resource_capacity(context.instructor_ids, context.instructor_availability, context),
        _total_resource_capacity(context.car_ids, context.car_availability, context),
    )
    required_lessons = sum(context.target_lessons_by_student.values())
    effective_capacity = min(resource_lesson_capacity, required_lessons)
    return float(max(0, effective_capacity - len(lessons)) * IDLE_CAPACITY_PENALTY)


def _student_distribution_penalty(lessons: list[CandidateLesson], context: PlanningContext) -> float:
    ratios: list[float] = []
    scheduled = _scheduled_counts(lessons)
    for student_id, target in context.target_lessons_by_student.items():
        if target <= 0:
            continue
        ratios.append(scheduled.get(student_id, 0) / target)
    if len(ratios) < 2:
        return 0.0
    mean = sum(ratios) / len(ratios)
    return sum(abs(ratio - mean) for ratio in ratios) * STUDENT_UNEVENNESS_PENALTY


def _resource_distribution_penalty(lessons: list[CandidateLesson], context: PlanningContext) -> float:
    instructor_counts = _resource_counts(lessons, "instructor")
    return _capacity_distribution_penalty(
        instructor_counts,
        _resource_capacities(context.instructor_ids, context.instructor_availability, context),
    ) + _car_distribution_penalty(lessons, context)


def _car_distribution_penalty(lessons: list[CandidateLesson], context: PlanningContext) -> float:
    car_counts = _resource_counts(lessons, "car")
    car_capacities = _resource_capacities(context.car_ids, context.car_availability, context)
    penalty = 0.0
    for transmission in sorted({value for value in context.car_transmissions.values() if value}):
        transmission_car_ids = [
            car_id
            for car_id in context.car_ids
            if context.car_transmissions.get(car_id) == transmission
        ]
        if not transmission_car_ids:
            continue
        transmission_counts = {
            car_id: car_counts.get(car_id, 0)
            for car_id in transmission_car_ids
        }
        transmission_capacities = {
            car_id: car_capacities.get(car_id, 0)
            for car_id in transmission_car_ids
        }
        penalty += _capacity_distribution_penalty(transmission_counts, transmission_capacities)
    unknown_car_ids = [
        car_id
        for car_id in context.car_ids
        if not context.car_transmissions.get(car_id)
    ]
    if unknown_car_ids:
        penalty += _capacity_distribution_penalty(
            {car_id: car_counts.get(car_id, 0) for car_id in unknown_car_ids},
            {car_id: car_capacities.get(car_id, 0) for car_id in unknown_car_ids},
        )
    return penalty


def _resource_counts(lessons: list[CandidateLesson], resource_type: str) -> dict[str, int]:
    counts: dict[str, int] = {}
    for lesson in lessons:
        resource_id = lesson.instructor_id if resource_type == "instructor" else lesson.car_id
        counts[resource_id] = counts.get(resource_id, 0) + 1
    return counts


def _capacity_distribution_penalty(counts: dict[str, int], capacities: dict[str, int]) -> float:
    active_capacities = {resource_id: capacity for resource_id, capacity in capacities.items() if capacity > 0}
    if len(active_capacities) < 2:
        return 0.0
    total_lessons = sum(counts.get(resource_id, 0) for resource_id in active_capacities)
    total_capacity = sum(active_capacities.values())
    if total_lessons <= 0 or total_capacity <= 0:
        return 0.0
    return sum(
        abs(counts.get(resource_id, 0) - total_lessons * capacity / total_capacity)
        for resource_id, capacity in active_capacities.items()
    ) * RESOURCE_UNEVENNESS_PENALTY


def _total_resource_capacity(resource_ids: list[str], availability, context: PlanningContext) -> int:
    return sum(_resource_capacities(resource_ids, availability, context).values())


def _resource_capacities(resource_ids: list[str], availability, context: PlanningContext) -> dict[str, int]:
    return {
        resource_id: _resource_capacity(resource_id, availability.get(resource_id, []), context)
        for resource_id in resource_ids
    }


def _resource_capacity(resource_id: str, windows, context: PlanningContext) -> int:
    del resource_id
    lesson_seconds = 90 * 60
    working_windows = [window for window in windows if window.type == "working_hours"]
    blocking_windows = [window for window in windows if window.type in {"blackout", "vacation"}]

    if working_windows:
        available_seconds = sum(
            _overlap_seconds(window.start_at, window.end_at, context.period_start, context.period_end)
            for window in working_windows
        )
    else:
        available_seconds = max(0.0, (context.period_end - context.period_start).total_seconds())

    blocked_seconds = sum(
        _overlap_seconds(window.start_at, window.end_at, context.period_start, context.period_end)
        for window in blocking_windows
    )
    return max(0, int((available_seconds - blocked_seconds) // lesson_seconds))


def _overlap_seconds(start_at, end_at, period_start, period_end) -> float:
    start = max(start_at, period_start)
    end = min(end_at, period_end)
    return max(0.0, (end - start).total_seconds())


def _student_preference_bonus(lessons: list[CandidateLesson], context: PlanningContext) -> float:
    total = 0.0
    for lesson in lessons:
        preference = context.student_preferences.get(lesson.student_id)
        if preference is None:
            continue
        weekday = lesson.start_at.strftime("%A").lower()
        time_window = _time_window(lesson.start_at.hour)

        if preference.preferred_weekdays:
            total += PREFERENCE_MATCH_BONUS if weekday in preference.preferred_weekdays else -PREFERENCE_MISS_PENALTY
        if preference.preferred_time_windows:
            total += PREFERENCE_MATCH_BONUS if time_window in preference.preferred_time_windows else -PREFERENCE_MISS_PENALTY
    return total


def _scheduled_counts(lessons: list[CandidateLesson]) -> dict[str, int]:
    counts: dict[str, int] = {}
    for lesson in lessons:
        counts[lesson.student_id] = counts.get(lesson.student_id, 0) + 1
    return counts


def _time_window(hour: int) -> str:
    if hour < 12:
        return "morning"
    if hour < 17:
        return "daytime"
    return "evening"
