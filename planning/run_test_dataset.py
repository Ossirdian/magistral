from __future__ import annotations

import argparse
import json
import os
from collections import Counter, defaultdict
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any


DEFAULT_DATASET = Path(__file__).resolve().parent / "generated" / "planning_test_dataset.json"


@dataclass(frozen=True)
class DatasetLoadedData:
    loaded: Any
    request: Any
    raw: dict[str, Any]


class DatasetDataLoader:
    def __init__(self, loaded: Any) -> None:
        self._loaded = loaded

    def load(self, request: Any) -> Any:
        del request
        return self._loaded


def main() -> None:
    args = _parse_args()
    print("[dataset] runner started", flush=True)
    if args.verbose:
        os.environ["PLANNING_VERBOSE"] = "1"
        os.environ["PLANNING_GA_VERBOSE"] = "1"
    print(f"[dataset] loading dataset: {args.dataset}", flush=True)
    dataset = _load_dataset(args.dataset)
    print(
        "[dataset] loaded: "
        f"students={len(dataset.raw['students'])} "
        f"cars={len(dataset.raw['cars'])} "
        f"instructors={len(dataset.raw['instructors'])} "
        f"existing_lessons={len(dataset.raw['existing_lessons'])}",
        flush=True,
    )
    print("[dataset] importing planning modules", flush=True)
    from planning.app.planner import Planner

    planner = Planner(DatasetDataLoader(dataset.loaded))
    print("[dataset] running Planner.plan()", flush=True)
    result = planner.plan(dataset.request)
    if args.report_json:
        args.report_json.parent.mkdir(parents=True, exist_ok=True)
        args.report_json.write_text(
            json.dumps(_result_payload(dataset, result), ensure_ascii=False, indent=2),
            encoding="utf-8",
        )
        print(f"[dataset] report json saved: {args.report_json}", flush=True)
    print("[dataset] Planner.plan() finished, printing diagnostics", flush=True)
    _print_report(dataset, result)


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run planning over the generated test dataset.")
    parser.add_argument("--dataset", type=Path, default=DEFAULT_DATASET)
    parser.add_argument("--verbose", action="store_true", help="Print planning progress and GA generation stats.")
    parser.add_argument("--report-json", type=Path, help="Write planning result diagnostics to JSON.")
    return parser.parse_args()


def _load_dataset(path: Path) -> DatasetLoadedData:
    from planning.app.data_loader import LoadedPlanningData
    from planning.app.models import AvailabilityWindow, ExistingLesson, PlanningRequest, StudentPreference

    payload = json.loads(path.read_text(encoding="utf-8"))
    metadata = payload["metadata"]
    students = list(payload["students"])
    cars = list(payload["cars"])
    instructors = list(payload["instructors"])
    existing_lessons = list(payload["existing_lessons"])

    target_lessons_by_student = {
        student["id"]: int(student["target_lessons"])
        for student in students
    }
    lesson_demands = [
        student_id
        for student_id, count in target_lessons_by_student.items()
        for _index in range(count)
    ]

    loaded = LoadedPlanningData(
        student_ids=[student["id"] for student in students],
        target_lessons_by_student=target_lessons_by_student,
        lesson_demands=lesson_demands,
        instructor_ids=[instructor["id"] for instructor in instructors],
        car_ids=[car["id"] for car in cars],
        existing_lessons=[
            ExistingLesson(
                student_id=lesson["student_id"],
                instructor_id=lesson["instructor_id"],
                car_id=lesson["car_id"],
                start_at=_parse_dt(lesson["start_at"]),
                end_at=_parse_dt(lesson["end_at"]),
                status=str(lesson.get("status") or "scheduled"),
            )
            for lesson in existing_lessons
        ],
        student_preferences={
            student["id"]: StudentPreference(
                preferred_weekdays=list(student["preferences"]["preferred_weekdays"]),
                preferred_time_windows=list(student["preferences"]["preferred_time_windows"]),
                transmission=str(student["transmission"]),
            )
            for student in students
        },
        student_transmissions={student["id"]: str(student["transmission"]) for student in students},
        car_transmissions={car["id"]: str(car["transmission"]) for car in cars},
        instructor_availability={
            instructor["id"]: [
                AvailabilityWindow(
                    resource_type=str(window["resource_type"]),
                    resource_id=str(window["resource_id"]),
                    type=str(window["type"]),
                    start_at=_parse_dt(window["start_at"]),
                    end_at=_parse_dt(window["end_at"]),
                )
                for window in instructor.get("availability", [])
            ]
            for instructor in instructors
        },
        car_availability={
            car["id"]: [
                AvailabilityWindow(
                    resource_type=str(window["resource_type"]),
                    resource_id=str(window["resource_id"]),
                    type=str(window["type"]),
                    start_at=_parse_dt(window["start_at"]),
                    end_at=_parse_dt(window["end_at"]),
                )
                for window in car.get("availability", [])
            ]
            for car in cars
        },
        diagnostics={
            "dataset_existing_lessons": len(existing_lessons),
            "dataset_students": len(students),
            "dataset_cars": len(cars),
            "dataset_instructors": len(instructors),
        },
    )
    request = PlanningRequest(
        school_id="school-dataset-test",
        period_start=_parse_dt(metadata["period_start"]),
        period_end=_parse_dt(metadata["period_end"]),
        student_ids=loaded.student_ids,
        group_ids=["group-main-01"],
        branch_ids=[],
        replace_strategy="append",
        requested_by_user_id="dataset-runner",
        request_id="planning-test-dataset-run",
    )
    return DatasetLoadedData(loaded=loaded, request=request, raw=payload)


def _print_report(dataset: DatasetLoadedData, result) -> None:
    raw = dataset.raw
    students = list(raw["students"])
    cars = list(raw["cars"])
    instructors = list(raw["instructors"])
    existing_lessons = list(raw["existing_lessons"])
    generated = result.lessons
    target_by_student = dataset.loaded.target_lessons_by_student
    generated_by_student = Counter(lesson.student_id for lesson in generated)
    existing_by_car = Counter(lesson["car_id"] for lesson in existing_lessons)
    existing_by_instructor = Counter(lesson["instructor_id"] for lesson in existing_lessons)
    generated_by_car = Counter(lesson.car_id for lesson in generated)
    generated_by_instructor = Counter(lesson.instructor_id for lesson in generated)
    lesson_hours = 1.5
    car_available_hours = sum(_availability_hours(car.get("availability", [])) for car in cars)
    instructor_available_hours = sum(_availability_hours(instructor.get("availability", [])) for instructor in instructors)
    car_busy_hours = sum((generated_by_car[car["id"]] + existing_by_car[car["id"]]) * lesson_hours for car in cars)
    instructor_busy_hours = sum(
        (generated_by_instructor[instructor["id"]] + existing_by_instructor[instructor["id"]]) * lesson_hours
        for instructor in instructors
    )
    car_idle_hours = max(0.0, car_available_hours - car_busy_hours)
    instructor_idle_hours = max(0.0, instructor_available_hours - instructor_busy_hours)

    print("Planning dataset run")
    print(f"  request_id: {result.request_id}")
    print(f"  period: {dataset.request.period_start.isoformat()} - {dataset.request.period_end.isoformat()}")
    print(f"  students: {len(students)}")
    print(f"  automatic_students: {sum(1 for item in students if item['transmission'] == 'automatic')}")
    print(f"  manual_students: {sum(1 for item in students if item['transmission'] == 'manual')}")
    print(f"  cars: {len(cars)}")
    print(f"  automatic_cars: {sum(1 for item in cars if item['transmission'] == 'automatic')}")
    print(f"  manual_cars: {sum(1 for item in cars if item['transmission'] == 'manual')}")
    print(f"  instructors: {len(instructors)}")
    print(f"  existing_lessons: {len(existing_lessons)}")
    print(f"  target_lessons: {sum(target_by_student.values())}")
    print(f"  generated_lessons: {len(generated)}")
    print(f"  unassigned_students: {len(result.unassigned_students)}")
    print(f"  conflicts: {len(result.conflicts)}")
    print(f"  best_fitness_score: {result.score:.2f}")
    print(f"  car_available_hours: {car_available_hours:.1f}")
    print(f"  car_busy_hours: {car_busy_hours:.1f}")
    print(f"  car_idle_hours: {car_idle_hours:.1f}")
    print(f"  car_idle_percent: {_percent(car_idle_hours, car_available_hours):.2f}%")
    print(f"  instructor_available_hours: {instructor_available_hours:.1f}")
    print(f"  instructor_busy_hours: {instructor_busy_hours:.1f}")
    print(f"  instructor_idle_hours: {instructor_idle_hours:.1f}")
    print(f"  instructor_idle_percent: {_percent(instructor_idle_hours, instructor_available_hours):.2f}%")

    print()
    print("Fitness score breakdown")
    for key, value in result.score_breakdown.items():
        if isinstance(value, float):
            print(f"  {key}: {value:.2f}")
        else:
            print(f"  {key}: {value}")

    first_lesson = min(generated, key=lambda item: item.start_at, default=None)
    last_lesson = max(generated, key=lambda item: item.start_at, default=None)
    print()
    print("Schedule bounds")
    print(f"  first_generated_lesson: {_lesson_line(first_lesson) if first_lesson else '-'}")
    print(f"  last_generated_lesson: {_lesson_line(last_lesson) if last_lesson else '-'}")
    print(f"  all_programs_completed: {all(generated_by_student[student['id']] >= target_by_student[student['id']] for student in students)}")

    print()
    print("Transmission usage")
    print("  type       students  target  generated  existing_car_bookings")
    for transmission in ["manual", "automatic"]:
        student_ids = [student["id"] for student in students if student["transmission"] == transmission]
        car_ids = [car["id"] for car in cars if car["transmission"] == transmission]
        print(
            f"  {transmission:<10} "
            f"{len(student_ids):>8} "
            f"{sum(target_by_student[student_id] for student_id in student_ids):>7} "
            f"{sum(generated_by_student[student_id] for student_id in student_ids):>10} "
            f"{sum(existing_by_car[car_id] for car_id in car_ids):>21}"
        )

    print()
    print("Car utilization")
    print("  car               type       avail_h  busy_h  idle_h  generated  existing")
    for car in cars:
        generated_count = generated_by_car[car["id"]]
        existing_count = existing_by_car[car["id"]]
        available_hours = _availability_hours(car.get("availability", []))
        busy_hours = (generated_count + existing_count) * lesson_hours
        print(
            f"  {car['id']:<17} "
            f"{car['transmission']:<10} "
            f"{available_hours:>7.1f} "
            f"{busy_hours:>6.1f} "
            f"{max(0.0, available_hours - busy_hours):>7.1f} "
            f"{generated_count:>9} "
            f"{existing_count:>8}"
        )

    print()
    print("Instructor utilization")
    print("  instructor        avail_h  busy_h  idle_h  generated  existing")
    for instructor in instructors:
        generated_count = generated_by_instructor[instructor["id"]]
        existing_count = existing_by_instructor[instructor["id"]]
        available_hours = _availability_hours(instructor.get("availability", []))
        busy_hours = (generated_count + existing_count) * lesson_hours
        print(
            f"  {instructor['id']:<16} "
            f"{available_hours:>7.1f} "
            f"{busy_hours:>6.1f} "
            f"{max(0.0, available_hours - busy_hours):>7.1f} "
            f"{generated_count:>9} "
            f"{existing_count:>8}"
        )

    print()
    print("Student completion")
    print("  student     type       target  generated  remaining  last_lesson")
    for student in students:
        generated_count = generated_by_student[student["id"]]
        remaining = max(0, target_by_student[student["id"]] - generated_count)
        last = max(
            (lesson.start_at.isoformat() for lesson in generated if lesson.student_id == student["id"]),
            default="-",
        )
        print(
            f"  {student['id']:<10} "
            f"{student['transmission']:<10} "
            f"{target_by_student[student['id']]:>6} "
            f"{generated_count:>9} "
            f"{remaining:>9} "
            f"{last}"
        )

    print()
    print("Generated lessons by date")
    for day, count in sorted(Counter(lesson.start_at.date().isoformat() for lesson in generated).items()):
        print(f"  {day}: {count}")

    print()
    print("Preference matches")
    _print_preference_matches(students, generated)

    if result.conflicts:
        print()
        print("Conflicts")
        for conflict in result.conflicts:
            print(f"  {conflict.student_id}: {conflict.reason}")


def _result_payload(dataset: DatasetLoadedData, result) -> dict[str, Any]:
    raw = dataset.raw
    students = list(raw["students"])
    cars = list(raw["cars"])
    instructors = list(raw["instructors"])
    existing_lessons = list(raw["existing_lessons"])
    generated = result.lessons
    target_by_student = dataset.loaded.target_lessons_by_student
    generated_by_student = Counter(lesson.student_id for lesson in generated)
    existing_by_car = Counter(lesson["car_id"] for lesson in existing_lessons)
    existing_by_instructor = Counter(lesson["instructor_id"] for lesson in existing_lessons)
    generated_by_car = Counter(lesson.car_id for lesson in generated)
    generated_by_instructor = Counter(lesson.instructor_id for lesson in generated)
    lesson_hours = 1.5
    car_available_hours = sum(_availability_hours(car.get("availability", [])) for car in cars)
    instructor_available_hours = sum(_availability_hours(instructor.get("availability", [])) for instructor in instructors)
    car_busy_hours = sum((generated_by_car[car["id"]] + existing_by_car[car["id"]]) * lesson_hours for car in cars)
    instructor_busy_hours = sum(
        (generated_by_instructor[instructor["id"]] + existing_by_instructor[instructor["id"]]) * lesson_hours
        for instructor in instructors
    )
    car_idle_hours = max(0.0, car_available_hours - car_busy_hours)
    instructor_idle_hours = max(0.0, instructor_available_hours - instructor_busy_hours)

    return {
        "request": {
            "request_id": result.request_id,
            "period_start": dataset.request.period_start.isoformat(),
            "period_end": dataset.request.period_end.isoformat(),
        },
        "summary": {
            "students": len(students),
            "automatic_students": sum(1 for item in students if item["transmission"] == "automatic"),
            "manual_students": sum(1 for item in students if item["transmission"] == "manual"),
            "cars": len(cars),
            "automatic_cars": sum(1 for item in cars if item["transmission"] == "automatic"),
            "manual_cars": sum(1 for item in cars if item["transmission"] == "manual"),
            "instructors": len(instructors),
            "existing_lessons": len(existing_lessons),
            "target_lessons": sum(target_by_student.values()),
            "generated_lessons": len(generated),
            "unassigned_students": len(result.unassigned_students),
            "conflicts": len(result.conflicts),
            "best_fitness_score": result.score,
            "score": result.score,
            "car_available_hours": car_available_hours,
            "car_busy_hours": car_busy_hours,
            "car_idle_hours": car_idle_hours,
            "car_idle_percent": _percent(car_idle_hours, car_available_hours),
            "instructor_available_hours": instructor_available_hours,
            "instructor_busy_hours": instructor_busy_hours,
            "instructor_idle_hours": instructor_idle_hours,
            "instructor_idle_percent": _percent(instructor_idle_hours, instructor_available_hours),
            "resource_idle_hours": car_idle_hours + instructor_idle_hours,
            "all_programs_completed": all(
                generated_by_student[student["id"]] >= target_by_student[student["id"]]
                for student in students
            ),
        },
        "fitness_score_breakdown": result.score_breakdown,
        "score_breakdown": result.score_breakdown,
        "generated_lessons": [lesson.as_dict() for lesson in generated],
        "cars": [
            {
                "id": car["id"],
                "transmission": car["transmission"],
                "generated_lessons": generated_by_car[car["id"]],
                "existing_lessons": existing_by_car[car["id"]],
                "available_hours": _availability_hours(car.get("availability", [])),
                "busy_hours": (generated_by_car[car["id"]] + existing_by_car[car["id"]]) * lesson_hours,
                "idle_hours": max(
                    0.0,
                    _availability_hours(car.get("availability", []))
                    - (generated_by_car[car["id"]] + existing_by_car[car["id"]]) * lesson_hours,
                ),
            }
            for car in cars
        ],
        "instructors": [
            {
                "id": instructor["id"],
                "generated_lessons": generated_by_instructor[instructor["id"]],
                "existing_lessons": existing_by_instructor[instructor["id"]],
                "available_hours": _availability_hours(instructor.get("availability", [])),
                "busy_hours": (
                    generated_by_instructor[instructor["id"]]
                    + existing_by_instructor[instructor["id"]]
                )
                * lesson_hours,
                "idle_hours": max(
                    0.0,
                    _availability_hours(instructor.get("availability", []))
                    - (
                        generated_by_instructor[instructor["id"]]
                        + existing_by_instructor[instructor["id"]]
                    )
                    * lesson_hours,
                ),
            }
            for instructor in instructors
        ],
        "students": [
            {
                "id": student["id"],
                "transmission": student["transmission"],
                "target_lessons": target_by_student[student["id"]],
                "generated_lessons": generated_by_student[student["id"]],
                "remaining_lessons": max(0, target_by_student[student["id"]] - generated_by_student[student["id"]]),
                "last_lesson_at": max(
                    (lesson.start_at.isoformat() for lesson in generated if lesson.student_id == student["id"]),
                    default=None,
                ),
                "preferred_weekdays": student["preferences"]["preferred_weekdays"],
                "preferred_time_windows": student["preferences"]["preferred_time_windows"],
            }
            for student in students
        ],
        "lessons_by_date": dict(
            sorted(Counter(lesson.start_at.date().isoformat() for lesson in generated).items())
        ),
        "conflicts": [conflict.as_dict() for conflict in result.conflicts],
    }


def _print_preference_matches(students: list[dict[str, Any]], lessons) -> None:
    students_by_id = {student["id"]: student for student in students}
    weekday_matches = 0
    time_matches = 0
    both_matches = 0
    total = len(lessons)
    by_student: dict[str, list[int]] = defaultdict(lambda: [0, 0, 0, 0])
    for lesson in lessons:
        student = students_by_id[lesson.student_id]
        preferences = student["preferences"]
        weekday_match = lesson.start_at.strftime("%A").lower() in preferences["preferred_weekdays"]
        time_match = _time_window(lesson.start_at.hour) in preferences["preferred_time_windows"]
        weekday_matches += int(weekday_match)
        time_matches += int(time_match)
        both_matches += int(weekday_match and time_match)
        row = by_student[lesson.student_id]
        row[0] += 1
        row[1] += int(weekday_match)
        row[2] += int(time_match)
        row[3] += int(weekday_match and time_match)

    print(f"  total_lessons: {total}")
    print(f"  weekday_matches: {weekday_matches}")
    print(f"  time_window_matches: {time_matches}")
    print(f"  both_matches: {both_matches}")
    print("  by_student: student total weekday time both")
    for student_id, values in sorted(by_student.items()):
        print(f"    {student_id:<10} {values[0]:>5} {values[1]:>7} {values[2]:>4} {values[3]:>4}")


def _lesson_line(lesson) -> str:
    return (
        f"{lesson.start_at.isoformat()} "
        f"student={lesson.student_id} "
        f"instructor={lesson.instructor_id} "
        f"car={lesson.car_id}"
    )


def _parse_dt(value: str) -> datetime:
    if value.endswith("Z"):
        value = value[:-1] + "+00:00"
    return datetime.fromisoformat(value)


def _availability_hours(availability: list[dict[str, Any]]) -> float:
    total = 0.0
    for window in availability:
        if window.get("type") != "working_hours":
            continue
        total += (_parse_dt(window["end_at"]) - _parse_dt(window["start_at"])).total_seconds() / 3600
    return total


def _percent(value: float, total: float) -> float:
    if total <= 0:
        return 0.0
    return value / total * 100


def _time_window(hour: int) -> str:
    if hour < 12:
        return "morning"
    if hour < 17:
        return "daytime"
    return "evening"


if __name__ == "__main__":
    main()
