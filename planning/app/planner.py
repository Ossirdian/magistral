from __future__ import annotations

import os
from collections import Counter
from datetime import datetime, time, timedelta

from .data_loader import DataLoader
from .fitness import score_breakdown
from .genetic import evolve_plan
from .models import LESSON_DURATION, AvailabilityWindow, PlanningContext, PlanningRequest, PlanningResult
from .option_builder import build_valid_options
from .time_slots import build_daily_time_slot_cache


DEFAULT_CAR_AVAILABLE_FROM = time(8, 0)
DEFAULT_CAR_AVAILABLE_TO = time(20, 0)


class Planner:
    def __init__(self, data_loader: DataLoader) -> None:
        self._data_loader = data_loader

    def plan(self, request: PlanningRequest) -> PlanningResult:
        verbose = _env_flag("PLANNING_VERBOSE") or _env_flag("PLANNING_GA_VERBOSE")
        if verbose:
            _progress("loading planning data")
        loaded = self._data_loader.load(request)
        students_to_plan = loaded.student_ids
        if verbose:
            _progress(
                "loaded data: "
                f"students={len(students_to_plan)} "
                f"instructors={len(loaded.instructor_ids)} "
                f"cars={len(loaded.car_ids)} "
                f"existing_lessons={len(loaded.existing_lessons)} "
                f"lesson_demands={len(loaded.lesson_demands)}"
            )
        slot_cache = build_daily_time_slot_cache(
            request.period_start,
            request.period_end,
            DEFAULT_CAR_AVAILABLE_FROM,
            DEFAULT_CAR_AVAILABLE_TO,
        )
        if verbose:
            _progress(f"built time slots: slots_per_resource={len(slot_cache)}")
        lesson_demands = self._slot_demands(loaded.lesson_demands, len(slot_cache), len(loaded.car_ids))
        target_lessons_by_student = dict(Counter(lesson_demands))
        if verbose:
            _progress(
                "prepared demands: "
                f"effective_demands={len(lesson_demands)} "
                f"students_with_demands={len(target_lessons_by_student)}"
            )
        context = PlanningContext(
            students_to_plan=students_to_plan,
            lesson_demands=lesson_demands,
            target_lessons_by_student=target_lessons_by_student,
            instructor_ids=loaded.instructor_ids,
            car_ids=loaded.car_ids,
            existing_lessons=loaded.existing_lessons,
            student_preferences=loaded.student_preferences,
            student_transmissions=loaded.student_transmissions,
            car_transmissions=loaded.car_transmissions,
            period_start=request.period_start,
            period_end=request.period_end,
            instructor_availability=loaded.instructor_availability,
            car_availability=self._car_availability_with_defaults(loaded.car_ids, loaded.car_availability, request),
            fill_all_slots=True,
        )
        valid_options_by_student, total_car_slots = build_valid_options(slot_cache, context)
        valid_options = sum(len(options) for options in valid_options_by_student.values())
        avg_options_per_student = valid_options / max(1, len(valid_options_by_student))
        if verbose:
            _progress(
                "built valid options: "
                f"total_car_slots={total_car_slots} "
                f"valid_options={valid_options} "
                f"avg_options_per_student={avg_options_per_student:.1f}"
            )
        context = PlanningContext(
            students_to_plan=context.students_to_plan,
            lesson_demands=context.lesson_demands,
            target_lessons_by_student=context.target_lessons_by_student,
            instructor_ids=context.instructor_ids,
            car_ids=context.car_ids,
            existing_lessons=context.existing_lessons,
            student_preferences=context.student_preferences,
            student_transmissions=context.student_transmissions,
            car_transmissions=context.car_transmissions,
            period_start=context.period_start,
            period_end=context.period_end,
            instructor_availability=context.instructor_availability,
            car_availability=context.car_availability,
            fill_all_slots=False,
            valid_options_by_student=valid_options_by_student,
            total_car_slots=total_car_slots,
        )

        if verbose:
            _progress("starting genetic algorithm")
        ga_population = _env_int("PLANNING_GA_POPULATION", 100)
        ga_generations = _env_int("PLANNING_GA_GENERATIONS", 50)
        ga_max_seconds = _env_float("PLANNING_GA_MAX_SECONDS", 60.0)
        lessons, unassigned_students, conflicts, _score = evolve_plan(
            context,
            population_size=ga_population,
            generations=ga_generations,
            max_seconds=ga_max_seconds,
            verbose=_env_flag("PLANNING_GA_VERBOSE"),
        )
        if verbose:
            _progress(
                "finished genetic algorithm: "
                f"lessons={len(lessons)} "
                f"unassigned_students={len(unassigned_students)} "
                f"conflicts={len(conflicts)}"
            )
        breakdown = score_breakdown(lessons, unassigned_students, conflicts, context)
        breakdown["valid_options"] = float(valid_options)
        breakdown["avg_options_per_student"] = float(avg_options_per_student)
        breakdown["students_with_valid_options"] = float(len(valid_options_by_student))
        breakdown["total_car_slots"] = float(total_car_slots)
        breakdown["ga_population"] = float(ga_population)
        breakdown["ga_generations"] = float(ga_generations)
        breakdown["ga_max_seconds"] = float(ga_max_seconds)
        for key, value in loaded.diagnostics.items():
            breakdown[key] = float(value) if isinstance(value, int) else value
        return PlanningResult(
            request_id=request.request_id,
            school_id=request.school_id,
            replace_strategy=request.replace_strategy,
            lessons=lessons,
            unassigned_students=unassigned_students,
            conflicts=conflicts,
            score=breakdown["score"],
            resource_idle_seconds=self._resource_idle_seconds(context, lessons),
            score_breakdown=breakdown,
        )

    def _resource_idle_seconds(self, context: PlanningContext, lessons) -> float:
        resource_count = len(context.instructor_ids) + len(context.car_ids)
        if resource_count == 0:
            return 0.0

        total_capacity_seconds = sum(
            self._resource_available_seconds(context.instructor_availability.get(instructor_id, []), context)
            for instructor_id in context.instructor_ids
        ) + sum(
            self._resource_available_seconds(context.car_availability.get(car_id, []), context)
            for car_id in context.car_ids
        )
        used_capacity_seconds = float(len(lessons) * 2) * LESSON_DURATION.total_seconds()
        return max(0.0, total_capacity_seconds - used_capacity_seconds)

    def _resource_available_seconds(self, windows, context: PlanningContext) -> float:
        working_windows = [window for window in windows if window.type == "working_hours"]
        blocking_windows = [window for window in windows if window.type in {"blackout", "vacation"}]
        if working_windows:
            available_seconds = sum(
                self._overlap_seconds(window.start_at, window.end_at, context.period_start, context.period_end)
                for window in working_windows
            )
        else:
            available_seconds = max(timedelta(), context.period_end - context.period_start).total_seconds()
        blocked_seconds = sum(
            self._overlap_seconds(window.start_at, window.end_at, context.period_start, context.period_end)
            for window in blocking_windows
        )
        return max(0.0, available_seconds - blocked_seconds)

    def _overlap_seconds(self, start_at, end_at, period_start, period_end) -> float:
        start = max(start_at, period_start)
        end = min(end_at, period_end)
        return max(0.0, (end - start).total_seconds())

    def _slot_demands(self, lesson_demands: list[str], slot_count: int, car_count: int) -> list[str]:
        if not lesson_demands:
            return []
        if slot_count <= 0 or car_count <= 0:
            return lesson_demands

        capacity = slot_count * car_count
        remaining = Counter(lesson_demands)
        students = list(remaining.keys())
        result: list[str] = []
        while len(result) < capacity and any(count > 0 for count in remaining.values()):
            for student_id in students:
                if remaining[student_id] <= 0:
                    continue
                result.append(student_id)
                remaining[student_id] -= 1
                if len(result) >= capacity:
                    break
        return result

    def _default_car_availability(
        self,
        car_ids: list[str],
        request: PlanningRequest,
    ) -> dict[str, list[AvailabilityWindow]]:
        windows = [
            window
            for window in self._daily_default_car_windows(request.period_start, request.period_end)
            if window.end_at > request.period_start and window.start_at < request.period_end
        ]
        return {
            car_id: [
                AvailabilityWindow(
                    resource_type="car",
                    resource_id=car_id,
                    type="working_hours",
                    start_at=window.start_at,
                    end_at=window.end_at,
                )
                for window in windows
            ]
            for car_id in car_ids
        }

    def _car_availability_with_defaults(
        self,
        car_ids: list[str],
        availability: dict[str, list[AvailabilityWindow]],
        request: PlanningRequest,
    ) -> dict[str, list[AvailabilityWindow]]:
        defaults = self._default_car_availability(car_ids, request)
        return {
            car_id: availability.get(car_id) or defaults[car_id]
            for car_id in car_ids
        }

    def _daily_default_car_windows(self, period_start, period_end) -> list[AvailabilityWindow]:
        windows: list[AvailabilityWindow] = []
        current_date = period_start.date()
        end_date = period_end.date()
        while current_date <= end_date:
            windows.append(
                AvailabilityWindow(
                    resource_type="car",
                    resource_id="",
                    type="working_hours",
                    start_at=datetime.combine(current_date, DEFAULT_CAR_AVAILABLE_FROM, tzinfo=period_start.tzinfo),
                    end_at=datetime.combine(current_date, DEFAULT_CAR_AVAILABLE_TO, tzinfo=period_start.tzinfo),
                )
            )
            current_date += timedelta(days=1)
        return windows


def _env_flag(name: str) -> bool:
    return os.getenv(name, "").strip().lower() in {"1", "true", "yes", "on"}


def _env_int(name: str, default: int) -> int:
    try:
        value = int(os.getenv(name, ""))
    except ValueError:
        return default
    return value if value > 0 else default


def _env_float(name: str, default: float) -> float:
    try:
        value = float(os.getenv(name, ""))
    except ValueError:
        return default
    return value if value > 0 else default


def _progress(message: str) -> None:
    print(f"[planning] {message}", flush=True)
