from __future__ import annotations

from .models import CandidateLesson, Conflict, PlanningContext


def score_plan(lessons: list[CandidateLesson], unassigned_students: list[str], conflicts: list[Conflict], context: PlanningContext) -> float:
    scheduled_bonus = float(len(lessons) * 100)
    backlog_penalty = float(len(unassigned_students) * 50)
    conflict_penalty = float(len(conflicts) * 100)
    idle_penalty = _idle_resource_penalty(lessons, context)
    preference_bonus = _student_preference_bonus(lessons, context)
    return scheduled_bonus - backlog_penalty - conflict_penalty - idle_penalty + preference_bonus


def _idle_resource_penalty(lessons: list[CandidateLesson], context: PlanningContext) -> float:
    if not context.instructor_ids or not context.car_ids:
        return 25.0
    utilization = min(len(lessons), len(context.instructor_ids) + len(context.car_ids))
    return float(max(0, len(context.students_to_plan) - utilization) * 5)


def _student_preference_bonus(lessons: list[CandidateLesson], context: PlanningContext) -> float:
    total = 0.0
    for lesson in lessons:
        preference = context.student_preferences.get(lesson.student_id)
        if preference is None:
            continue
        weekday = lesson.start_at.strftime("%A").lower()
        time_window = _time_window(lesson.start_at.hour)

        if preference.preferred_weekdays:
            total += 20.0 if weekday in preference.preferred_weekdays else -15.0
        if preference.preferred_time_windows:
            total += 20.0 if time_window in preference.preferred_time_windows else -15.0
    return total


def _time_window(hour: int) -> str:
    if hour < 12:
        return "morning"
    if hour < 17:
        return "daytime"
    return "evening"
