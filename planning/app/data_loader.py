from __future__ import annotations

from dataclasses import dataclass

from .models import ExistingLesson, PlanningRequest, StudentPreference


@dataclass(frozen=True)
class LoadedPlanningData:
    student_ids: list[str]
    instructor_ids: list[str]
    car_ids: list[str]
    existing_lessons: list[ExistingLesson]
    student_preferences: dict[str, StudentPreference]
    student_transmissions: dict[str, str]
    car_transmissions: dict[str, str]


class DataLoader:
    def load(self, request: PlanningRequest) -> LoadedPlanningData:
        raise NotImplementedError


class StaticDataLoader(DataLoader):
    """
    Stub adapter until read contracts to Student/Resource/Lesson services are finalized.
    """

    def load(self, request: PlanningRequest) -> LoadedPlanningData:
        student_ids = _unique(request.student_ids + [member for group_id in request.group_ids for member in self._group_members(group_id)])
        return LoadedPlanningData(
            student_ids=student_ids,
            instructor_ids=["stub-instructor-1", "stub-instructor-2"],
            car_ids=["stub-car-1", "stub-car-2"],
            existing_lessons=[],
            student_preferences={
                student_id: request.student_preferences.get(student_id, StudentPreference())
                for student_id in student_ids
            },
            student_transmissions={
                student_id: request.student_preferences.get(student_id, StudentPreference()).transmission
                for student_id in student_ids
            },
            car_transmissions={
                "stub-car-1": "manual",
                "stub-car-2": "automatic",
            },
        )

    def _group_members(self, group_id: str) -> list[str]:
        return [f"{group_id}-student-1", f"{group_id}-student-2"]


def _unique(items: list[str]) -> list[str]:
    seen: set[str] = set()
    result: list[str] = []
    for item in items:
        item = item.strip()
        if not item or item in seen:
            continue
        seen.add(item)
        result.append(item)
    return result
