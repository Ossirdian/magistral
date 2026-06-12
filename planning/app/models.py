from __future__ import annotations

from dataclasses import asdict, dataclass, field
from datetime import datetime, timedelta
from typing import Any


LESSON_DURATION = timedelta(minutes=90)


@dataclass(frozen=True)
class StudentPreference:
    preferred_weekdays: list[str] = field(default_factory=list)
    preferred_time_windows: list[str] = field(default_factory=list)
    transmission: str = ""

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "StudentPreference":
        return cls(
            preferred_weekdays=[str(item) for item in (payload.get("preferred_weekdays") or [])],
            preferred_time_windows=[str(item) for item in (payload.get("preferred_time_windows") or [])],
            transmission=str(payload.get("transmission") or ""),
        )


@dataclass(frozen=True)
class PlanningRequest:
    school_id: str
    period_start: datetime
    period_end: datetime
    student_ids: list[str] = field(default_factory=list)
    student_preferences: dict[str, StudentPreference] = field(default_factory=dict)
    group_ids: list[str] = field(default_factory=list)
    branch_ids: list[str] = field(default_factory=list)
    replace_strategy: str = "append"
    requested_by_user_id: str = ""
    request_id: str = ""
    trace_context: dict[str, str] = field(default_factory=dict)

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "PlanningRequest":
        return cls(
            school_id=payload["school_id"],
            period_start=_parse_datetime(payload["period_start"]),
            period_end=_parse_datetime(payload["period_end"]),
            student_ids=list(payload.get("student_ids") or []),
            student_preferences={
                student_id: StudentPreference.from_dict(preference)
                for student_id, preference in dict(payload.get("student_preferences") or {}).items()
            },
            group_ids=list(payload.get("group_ids") or []),
            branch_ids=list(payload.get("branch_ids") or []),
            replace_strategy=payload.get("replace_strategy", "append"),
            requested_by_user_id=payload.get("requested_by_user_id", ""),
            request_id=payload.get("request_id", ""),
            trace_context={str(key): str(value) for key, value in dict(payload.get("trace_context") or {}).items()},
        )


@dataclass(frozen=True)
class ExistingLesson:
    student_id: str
    instructor_id: str
    car_id: str
    start_at: datetime
    end_at: datetime
    status: str = "scheduled"


@dataclass(frozen=True)
class AvailabilityWindow:
    resource_type: str
    resource_id: str
    type: str
    start_at: datetime
    end_at: datetime


@dataclass(frozen=True)
class TimeSlot:
    index: int
    start_at: datetime
    end_at: datetime


@dataclass(frozen=True)
class CarSlot:
    option_slot_id: int
    car_id: str
    time_slot: TimeSlot


@dataclass(frozen=True)
class InstructorAvailableTimeSlot:
    instructor_id: str
    time_slot: TimeSlot


@dataclass(frozen=True)
class LessonOption:
    id: int
    student_id: str
    instructor_id: str
    car_id: str
    time_slot: TimeSlot


@dataclass(frozen=True)
class CandidateLesson:
    student_id: str
    instructor_id: str
    car_id: str
    start_at: datetime
    end_at: datetime
    status: str = "scheduled"

    def as_dict(self) -> dict[str, Any]:
        return {
            "student_id": self.student_id,
            "instructor_id": self.instructor_id,
            "car_id": self.car_id,
            "start_at": _format_datetime(self.start_at),
            "end_at": _format_datetime(self.end_at),
            "status": self.status,
        }


@dataclass(frozen=True)
class Conflict:
    student_id: str
    reason: str

    def as_dict(self) -> dict[str, str]:
        return {"student_id": self.student_id, "reason": self.reason}


@dataclass(frozen=True)
class PlanningContext:
    students_to_plan: list[str]
    instructor_ids: list[str]
    car_ids: list[str]
    existing_lessons: list[ExistingLesson]
    student_preferences: dict[str, StudentPreference]
    student_transmissions: dict[str, str]
    car_transmissions: dict[str, str]
    period_start: datetime
    period_end: datetime
    instructor_availability: dict[str, list[AvailabilityWindow]] = field(default_factory=dict)
    car_availability: dict[str, list[AvailabilityWindow]] = field(default_factory=dict)
    lesson_demands: list[str] = field(default_factory=list)
    target_lessons_by_student: dict[str, int] = field(default_factory=dict)
    fill_all_slots: bool = False
    valid_options_by_student: dict[str, list[LessonOption]] = field(default_factory=dict)
    total_car_slots: int = 0


@dataclass(frozen=True)
class PlanningResult:
    request_id: str
    school_id: str
    replace_strategy: str
    lessons: list[CandidateLesson]
    unassigned_students: list[str]
    conflicts: list[Conflict]
    score: float
    resource_idle_seconds: float = 0.0
    score_breakdown: dict[str, float] = field(default_factory=dict)
    trace_context: dict[str, str] = field(default_factory=dict)

    def as_dict(self) -> dict[str, Any]:
        return {
            "request_id": self.request_id,
            "school_id": self.school_id,
            "replace_strategy": self.replace_strategy,
            "lessons": [lesson.as_dict() for lesson in self.lessons],
            "unassigned_students": self.unassigned_students,
            "conflicts": [conflict.as_dict() for conflict in self.conflicts],
            "score": self.score,
            "score_breakdown": self.score_breakdown,
            "resource_idle_seconds": self.resource_idle_seconds,
            "trace_context": self.trace_context,
        }


def _parse_datetime(value: str) -> datetime:
    if value.endswith("Z"):
        value = value[:-1] + "+00:00"
    return datetime.fromisoformat(value)


def _format_datetime(value: datetime) -> str:
    return value.isoformat()
