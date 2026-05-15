from __future__ import annotations

from datetime import timedelta

from .data_loader import DataLoader
from .genetic import evolve_plan
from .models import LESSON_DURATION, PlanningContext, PlanningRequest, PlanningResult


class Planner:
    def __init__(self, data_loader: DataLoader) -> None:
        self._data_loader = data_loader

    def plan(self, request: PlanningRequest) -> PlanningResult:
        loaded = self._data_loader.load(request)
        students_to_plan = loaded.student_ids
        context = PlanningContext(
            students_to_plan=students_to_plan,
            instructor_ids=loaded.instructor_ids,
            car_ids=loaded.car_ids,
            existing_lessons=loaded.existing_lessons,
            student_preferences=loaded.student_preferences,
            student_transmissions=loaded.student_transmissions,
            car_transmissions=loaded.car_transmissions,
            period_start=request.period_start,
            period_end=request.period_end,
        )

        slots = self._build_slots(request.period_start, request.period_end)
        lessons, unassigned_students, conflicts, score = evolve_plan(slots, context)
        return PlanningResult(
            request_id=request.request_id,
            school_id=request.school_id,
            lessons=lessons,
            unassigned_students=unassigned_students,
            conflicts=conflicts,
            score=score,
            resource_idle_seconds=self._resource_idle_seconds(context, lessons),
        )

    def _build_slots(self, period_start, period_end) -> list[tuple]:
        slots: list[tuple] = []
        cursor = period_start
        while cursor + LESSON_DURATION <= period_end:
            slots.append((cursor, cursor + LESSON_DURATION))
            cursor += LESSON_DURATION
        return slots

    def _resource_idle_seconds(self, context: PlanningContext, lessons) -> float:
        window = max(context.period_end - context.period_start, timedelta())
        resource_count = len(context.instructor_ids) + len(context.car_ids)
        if resource_count == 0:
            return 0.0

        total_capacity_seconds = window.total_seconds() * resource_count
        used_capacity_seconds = float(len(lessons) * 2) * LESSON_DURATION.total_seconds()
        return max(0.0, total_capacity_seconds - used_capacity_seconds)
