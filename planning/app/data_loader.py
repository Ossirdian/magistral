from __future__ import annotations

import json
import logging
import math
from dataclasses import dataclass
from typing import Any
from urllib.parse import urlencode
from urllib.request import Request, urlopen

from .models import AvailabilityWindow, ExistingLesson, PlanningRequest, StudentPreference
from .models import LESSON_DURATION
from .settings import Settings

DEFAULT_REQUIRED_LESSONS_BY_TRANSMISSION = {
    "manual": 28,
    "automatic": 27,
}

try:
    from opentelemetry import propagate, trace
except ModuleNotFoundError:
    propagate = None
    trace = None


@dataclass(frozen=True)
class LoadedPlanningData:
    student_ids: list[str]
    target_lessons_by_student: dict[str, int]
    lesson_demands: list[str]
    instructor_ids: list[str]
    car_ids: list[str]
    existing_lessons: list[ExistingLesson]
    student_preferences: dict[str, StudentPreference]
    student_transmissions: dict[str, str]
    car_transmissions: dict[str, str]
    instructor_availability: dict[str, list[AvailabilityWindow]]
    car_availability: dict[str, list[AvailabilityWindow]]
    diagnostics: dict[str, int | str]


class DataLoader:
    def load(self, request: PlanningRequest) -> LoadedPlanningData:
        raise NotImplementedError


class _NoopSpan:
    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return False

    def set_attribute(self, key: str, value: Any) -> None:
        pass


class _NoopTracer:
    def start_as_current_span(self, name: str, attributes: dict[str, Any] | None = None) -> _NoopSpan:
        return _NoopSpan()


@dataclass
class PlanningRequestDataCache:
    existing_lessons: list[ExistingLesson]
    instructor_availability: dict[str, list[AvailabilityWindow]]
    diagnostics: dict[str, int | str]


class HTTPDataLoader(DataLoader):
    def __init__(self, settings: Settings, client: "ServiceClient | None" = None) -> None:
        self._client = client or ServiceClient(settings)
        self._tracer = _tracer()

    def load(self, request: PlanningRequest) -> LoadedPlanningData:
        with self._tracer.start_as_current_span(
            "planning.adapters.load",
            attributes={
                "planning.request_id": request.request_id,
                "planning.school_id": request.school_id,
            },
        ) as span:
            students = self._load_students(request)
            programs = self._client.list_student_service("/study-programs", request.school_id, {})
            program_transmissions = {
                item["id"]: str(item.get("transmission") or "")
                for item in programs
            }

            instructors = self._load_instructors(request)
            cars = self._load_cars(request)
            data_cache = self._build_request_cache(request, instructors)
            span.set_attribute("planning.adapters.students", len(students))
            span.set_attribute("planning.adapters.instructors", len(instructors))
            span.set_attribute("planning.adapters.cars", len(cars))
            span.set_attribute("planning.adapters.existing_lessons", len(data_cache.existing_lessons))

        target_lessons_by_student = _target_lessons_by_student(students, program_transmissions)
        lesson_demands = [
            student_id
            for student_id, target_lessons in target_lessons_by_student.items()
            for _index in range(target_lessons)
        ]

        return LoadedPlanningData(
            student_ids=[student["id"] for student in students],
            target_lessons_by_student=target_lessons_by_student,
            lesson_demands=lesson_demands,
            instructor_ids=[instructor["id"] for instructor in instructors],
            car_ids=[car["id"] for car in cars],
            existing_lessons=data_cache.existing_lessons,
            student_preferences={
                student["id"]: StudentPreference(
                    preferred_weekdays=list(student.get("preferred_weekdays") or []),
                    preferred_time_windows=list(student.get("preferred_time_windows") or []),
                    transmission=program_transmissions.get(student.get("program_id", ""), ""),
                )
                for student in students
            },
            student_transmissions={
                student["id"]: program_transmissions.get(student.get("program_id", ""), "")
                for student in students
            },
            car_transmissions={
                car["id"]: str(car.get("transmission") or "")
                for car in cars
            },
            instructor_availability=data_cache.instructor_availability,
            car_availability={
                car["id"]: self._load_resource_availability("car", car["id"], request)
                for car in cars
            },
            diagnostics=data_cache.diagnostics,
        )

    def _build_request_cache(
        self,
        request: PlanningRequest,
        instructors: list[dict[str, Any]],
    ) -> PlanningRequestDataCache:
        existing_lessons, diagnostics = self._load_existing_lessons(request)
        return PlanningRequestDataCache(
            existing_lessons=existing_lessons,
            instructor_availability={
                instructor["id"]: self._load_resource_availability("instructor", instructor["id"], request)
                for instructor in instructors
            },
            diagnostics=diagnostics,
        )

    def _load_students(self, request: PlanningRequest) -> list[dict[str, Any]]:
        selected: dict[str, dict[str, Any]] = {}
        branch_ids = set(request.branch_ids)

        for student_id in request.student_ids:
            student = self._client.get_student_service(f"/students/{student_id}", request.school_id)
            if not branch_ids or student.get("branch_id") in branch_ids:
                selected[student["id"]] = student

        for group_id in request.group_ids:
            for branch_id in request.branch_ids or [None]:
                params = {"group_id": group_id}
                if branch_id:
                    params["branch_id"] = branch_id
                for student in self._client.list_student_service("/students", request.school_id, params):
                    selected[student["id"]] = student

        return list(selected.values())

    def _load_instructors(self, request: PlanningRequest) -> list[dict[str, Any]]:
        instructors: dict[str, dict[str, Any]] = {}
        for branch_id in request.branch_ids or [None]:
            params: dict[str, Any] = {"status": "active"}
            if branch_id:
                params["branch_id"] = branch_id
            for instructor in self._client.list_resources_service("/resources/instructors", request.school_id, params):
                instructors[instructor["id"]] = instructor
        return list(instructors.values())

    def _load_cars(self, request: PlanningRequest) -> list[dict[str, Any]]:
        cars: dict[str, dict[str, Any]] = {}
        for branch_id in request.branch_ids or [None]:
            params: dict[str, Any] = {"status": "active"}
            if branch_id:
                params["branch_id"] = branch_id
            for car in self._client.list_resources_service("/resources/cars", request.school_id, params):
                cars[car["id"]] = car
        return list(cars.values())

    def _load_existing_lessons(self, request: PlanningRequest) -> tuple[list[ExistingLesson], dict[str, int | str]]:
        lessons: dict[str, dict[str, Any]] = {}
        branch_ids = set(request.branch_ids)
        skip_replaceable_generated = request.replace_strategy in {"merge", "replace_generated_only"}
        loaded_count = 0
        skipped_deleted_or_cancelled = 0
        skipped_generated = 0
        params: dict[str, Any] = {
            "date_from": request.period_start.isoformat(),
            "date_to": request.period_end.isoformat(),
        }
        for item in self._client.list_lesson_service("/lessons", request.school_id, params):
            loaded_count += 1
            if branch_ids and item.get("branch_id") not in branch_ids:
                continue
            if _is_deleted_or_cancelled_lesson(item):
                skipped_deleted_or_cancelled += 1
                continue
            if skip_replaceable_generated and _is_replaceable_generated_lesson(item):
                skipped_generated += 1
                continue
            lessons[item["id"]] = item
        diagnostics: dict[str, int | str] = {
            "replace_strategy": request.replace_strategy,
            "existing_lessons_loaded": loaded_count,
            "existing_lessons_used": len(lessons),
            "existing_lessons_skipped_deleted_or_cancelled": skipped_deleted_or_cancelled,
            "existing_lessons_skipped_generated": skipped_generated,
        }
        diagnostic_message = (
            "planning existing lessons loaded: "
            f"request_id={request.request_id} "
            f"replace_strategy={request.replace_strategy} "
            f"loaded={loaded_count} "
            f"used={len(lessons)} "
            f"skipped_deleted_or_cancelled={skipped_deleted_or_cancelled} "
            f"skipped_generated={skipped_generated}"
        )
        logging.getLogger("planning").warning("%s", diagnostic_message)
        print(diagnostic_message, flush=True)
        return [
            ExistingLesson(
                student_id=item["student_id"],
                instructor_id=item["instructor_id"],
                car_id=item["car_id"],
                start_at=_parse_datetime(item["start_at"]),
                end_at=_parse_datetime(item["end_at"]),
                status=str(item.get("status") or "scheduled"),
            )
            for item in lessons.values()
        ], diagnostics

    def _load_resource_availability(
        self,
        resource_type: str,
        resource_id: str,
        request: PlanningRequest,
    ) -> list[AvailabilityWindow]:
        path = (
            f"/resources/cars/{resource_id}/availability"
            if resource_type == "car"
            else f"/resources/instructors/{resource_id}/availability"
        )
        return [
            AvailabilityWindow(
                resource_type=str(item.get("resource_type") or resource_type),
                resource_id=str(item.get("resource_id") or resource_id),
                type=str(item.get("type") or ""),
                start_at=_parse_datetime(item["start_at"]),
                end_at=_parse_datetime(item["end_at"]),
            )
            for item in self._client.list_resources_service(path, request.school_id, {})
            if item.get("start_at") and item.get("end_at")
        ]


class ServiceClient:
    def __init__(self, settings: Settings) -> None:
        self._student_base_url = settings.student_service_url.rstrip("/")
        self._resources_base_url = settings.resources_service_url.rstrip("/")
        self._lesson_base_url = settings.lesson_service_url.rstrip("/")
        self._service_user_id = settings.service_user_id
        self._tracer = _tracer()

    def list_student_service(self, path: str, school_id: str, params: dict[str, Any]) -> list[dict[str, Any]]:
        return self._list_all(self._student_base_url, path, school_id, params)

    def get_student_service(self, path: str, school_id: str) -> dict[str, Any]:
        return dict(self._get(self._student_base_url, path, school_id, {}).get("data") or {})

    def list_resources_service(self, path: str, school_id: str, params: dict[str, Any]) -> list[dict[str, Any]]:
        return self._list_all(self._resources_base_url, path, school_id, params)

    def list_lesson_service(self, path: str, school_id: str, params: dict[str, Any]) -> list[dict[str, Any]]:
        return self._list_all(self._lesson_base_url, path, school_id, params)

    def _list_all(self, base_url: str, path: str, school_id: str, params: dict[str, Any]) -> list[dict[str, Any]]:
        result: list[dict[str, Any]] = []
        offset = 0
        limit = 100
        while True:
            page_params = {**params, "limit": limit, "offset": offset}
            payload = self._get(base_url, path, school_id, page_params)
            items = _ensure_list(payload.get("data"))
            result.extend(items)
            meta = dict(payload.get("meta") or {})
            count = int(meta.get("count") or len(items))
            if len(items) < limit or offset + len(items) >= count:
                return result
            offset += len(items)

    def _get(self, base_url: str, path: str, school_id: str, params: dict[str, Any]) -> dict[str, Any]:
        query = urlencode({key: value for key, value in params.items() if value not in ("", None)})
        url = f"{base_url}{path}"
        if query:
            url = f"{url}?{query}"

        with self._tracer.start_as_current_span(
            "planning.adapter.http",
            attributes={
                "http.method": "GET",
                "url.path": path,
                "server.address": base_url,
            },
        ) as span:
            headers = {
                "accept": "application/json",
                "x-user-id": self._service_user_id,
                "x-user-role": "admin",
                "x-school-id": school_id,
                "x-user-status": "active",
                "x-user-active": "true",
                "x-authenticated": "true",
            }
            if propagate is not None:
                propagate.inject(headers)
            request = Request(
                url,
                headers=headers,
                method="GET",
            )
            with urlopen(request, timeout=10) as response:
                span.set_attribute("http.response.status_code", response.status)
                return json.loads(response.read().decode("utf-8"))


def _ensure_list(value: Any) -> list[dict[str, Any]]:
    if not isinstance(value, list):
        return []
    return [dict(item) for item in value if isinstance(item, dict)]


def _is_replaceable_generated_lesson(item: dict[str, Any]) -> bool:
    if bool(item.get("is_generated")):
        return True
    return str(item.get("comment") or "").startswith("generated by planning")


def _is_deleted_or_cancelled_lesson(item: dict[str, Any]) -> bool:
    if item.get("deleted_at"):
        return True
    return str(item.get("status") or "").strip().lower() == "cancelled"


def _target_lessons_by_student(
    students: list[dict[str, Any]],
    program_transmissions: dict[str, str] | None = None,
) -> dict[str, int]:
    lesson_hours = LESSON_DURATION.total_seconds() / 3600
    targets: dict[str, int] = {}
    for student in students:
        student_id = str(student.get("id") or "")
        if not student_id:
            continue
        remaining_lessons = _optional_int(
            _first_present(
                student.get("remaining_lessons_required_max"),
                student.get("remaining_lessons"),
                student.get("target_lessons"),
            )
        )
        if remaining_lessons is not None:
            targets[student_id] = max(0, remaining_lessons)
            continue
        remaining_hours = _float_or_zero(student.get("remaining_hours"))
        target = max(0, math.ceil(remaining_hours / lesson_hours))
        transmission = _student_transmission(student, program_transmissions or {})
        default_required_lessons = DEFAULT_REQUIRED_LESSONS_BY_TRANSMISSION.get(transmission)
        if default_required_lessons is not None:
            completed_hours = _float_or_zero(student.get("total_hours_completed"))
            if completed_hours == 0:
                completed_hours = (
                    _float_or_zero(student.get("manual_hours_completed"))
                    + _float_or_zero(student.get("lesson_hours_completed"))
                )
            completed_lessons = max(0, math.floor(completed_hours / lesson_hours))
            target = min(target, max(0, default_required_lessons - completed_lessons))
        targets[student_id] = target
    return targets


def _student_transmission(student: dict[str, Any], program_transmissions: dict[str, str]) -> str:
    transmission = str(student.get("transmission") or "").strip().lower()
    if transmission:
        return transmission
    return str(program_transmissions.get(str(student.get("program_id") or ""), "")).strip().lower()


def _optional_int(value: Any) -> int | None:
    if value in ("", None):
        return None
    try:
        return int(value)
    except (TypeError, ValueError):
        return None


def _first_present(*values: Any) -> Any:
    for value in values:
        if value not in ("", None):
            return value
    return None


def _float_or_zero(value: Any) -> float:
    try:
        return float(value or 0)
    except (TypeError, ValueError):
        return 0.0


def _parse_datetime(value: str):
    from .models import _parse_datetime as parse_datetime

    return parse_datetime(value)


def _tracer():
    if trace is None:
        return _NoopTracer()
    return trace.get_tracer("planning")
