from __future__ import annotations

import argparse
import json
import random
from collections import Counter, defaultdict
from datetime import date, datetime, time, timedelta, timezone
from pathlib import Path
from typing import Any


LESSON_DURATION = timedelta(minutes=90)
DEFAULT_SEED = 20260527
DEFAULT_OUTPUT = Path(__file__).resolve().parent / "generated" / "planning_test_dataset.json"

PERIOD_DAYS = 30
STUDENTS_COUNT = 67
AUTOMATIC_STUDENTS_COUNT = 8
MANUAL_CARS_COUNT = 10
AUTOMATIC_CARS_COUNT = 1
INSTRUCTORS_COUNT = 11
EXISTING_LESSONS_COUNT = 100
MANUAL_REQUIRED_LESSONS = 28
AUTOMATIC_REQUIRED_LESSONS = 27

SLOT_START_TIMES = [
    time(8, 0),
    time(9, 30),
    time(11, 0),
    time(12, 30),
    time(14, 0),
    time(15, 30),
    time(17, 0),
    time(18, 30),
]

CAR_AVAILABLE_FROM = time(8, 0)
CAR_AVAILABLE_TO = time(20, 0)
INSTRUCTOR_WORK_FROM = time(8, 0)
INSTRUCTOR_WORK_TO = time(20, 0)
WEEKEND_DAYS = {"saturday", "sunday"}

WEEKDAY_NAMES = [
    "monday",
    "tuesday",
    "wednesday",
    "thursday",
    "friday",
    "saturday",
    "sunday",
]


def main() -> None:
    args = _parse_args()
    _apply_args(args)
    _progress("starting dataset generation")
    rng = random.Random(args.seed)
    period_start = datetime.combine(args.start_date, SLOT_START_TIMES[0], tzinfo=timezone.utc)
    period_end = period_start + timedelta(days=PERIOD_DAYS)

    _progress("building students, instructors, cars and slots")
    students = _build_students(rng)
    instructors = _build_instructors(rng, period_start, period_end)
    cars = _build_cars(period_start, period_end)
    slots = _build_slots(period_start, period_end)
    _progress(
        "base entities built: "
        f"students={len(students)} cars={len(cars)} instructors={len(instructors)} slots={len(slots)}"
    )
    _progress("building reference generated schedule before external load")
    reserved_generated_lessons = _build_schedule(students, slots, instructors, cars, [])
    _progress(f"reference reservation built: {len(reserved_generated_lessons)} lessons")
    _progress("placing existing lessons from other groups into remaining capacity")
    existing_lessons = _build_existing_lessons(rng, slots, instructors, cars, reserved_generated_lessons)
    _progress(f"existing lessons placed: {len(existing_lessons)}")
    _progress("building reference generated schedule for dataset report")
    generated_lessons = _build_schedule(students, slots, instructors, cars, existing_lessons)
    _progress(f"reference generated schedule built: {len(generated_lessons)} lessons")
    _progress("building dataset report")
    report = _build_report(period_start, period_end, students, instructors, cars, slots, existing_lessons, generated_lessons)

    payload = {
        "metadata": {
            "dataset_id": args.dataset_id,
            "seed": args.seed,
            "lesson_duration_minutes": int(LESSON_DURATION.total_seconds() // 60),
            "period_start": _format_dt(period_start),
            "period_end": _format_dt(period_end),
            "period_days": PERIOD_DAYS,
        },
        "students": students,
        "cars": cars,
        "instructors": instructors,
        "slots": [_slot_to_dict(slot) for slot in slots],
        "existing_lessons": existing_lessons,
        "generated_lessons": generated_lessons,
        "report": report,
    }

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    _progress(f"dataset saved: {args.output}")
    _print_console_report(payload, args.output)


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate a low-idle planning test dataset for driving-school students.")
    parser.add_argument("--seed", type=int, default=DEFAULT_SEED)
    parser.add_argument("--start-date", type=date.fromisoformat, default=date(2026, 6, 1))
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--dataset-id", default="planning-test-low-idle-month")
    parser.add_argument("--period-days", type=int, default=PERIOD_DAYS)
    parser.add_argument("--students", type=int, default=STUDENTS_COUNT)
    parser.add_argument("--automatic-students", type=int, default=AUTOMATIC_STUDENTS_COUNT)
    parser.add_argument("--manual-cars", type=int, default=MANUAL_CARS_COUNT)
    parser.add_argument("--automatic-cars", type=int, default=AUTOMATIC_CARS_COUNT)
    parser.add_argument("--instructors", type=int, default=INSTRUCTORS_COUNT)
    parser.add_argument("--existing-lessons", type=int, default=EXISTING_LESSONS_COUNT)
    return parser.parse_args()


def _apply_args(args: argparse.Namespace) -> None:
    global PERIOD_DAYS
    global STUDENTS_COUNT
    global AUTOMATIC_STUDENTS_COUNT
    global MANUAL_CARS_COUNT
    global AUTOMATIC_CARS_COUNT
    global INSTRUCTORS_COUNT
    global EXISTING_LESSONS_COUNT

    if args.period_days <= 0:
        raise ValueError("--period-days must be positive")
    if args.students <= 0:
        raise ValueError("--students must be positive")
    if args.automatic_students < 0 or args.automatic_students > args.students:
        raise ValueError("--automatic-students must be between 0 and --students")
    if args.manual_cars < 0 or args.automatic_cars < 0 or args.manual_cars + args.automatic_cars <= 0:
        raise ValueError("car counts must produce at least one car")
    if args.instructors <= 0:
        raise ValueError("--instructors must be positive")
    if args.existing_lessons < 0:
        raise ValueError("--existing-lessons must be non-negative")

    PERIOD_DAYS = args.period_days
    STUDENTS_COUNT = args.students
    AUTOMATIC_STUDENTS_COUNT = args.automatic_students
    MANUAL_CARS_COUNT = args.manual_cars
    AUTOMATIC_CARS_COUNT = args.automatic_cars
    INSTRUCTORS_COUNT = args.instructors
    EXISTING_LESSONS_COUNT = args.existing_lessons


def _build_students(rng: random.Random) -> list[dict[str, Any]]:
    weekday_sets = [
        ["monday", "wednesday", "friday"],
        ["tuesday", "thursday", "saturday"],
        ["monday", "tuesday", "thursday"],
        ["wednesday", "friday", "sunday"],
        ["saturday", "sunday"],
        ["monday", "thursday", "saturday"],
    ]
    time_window_sets = [
        ["morning"],
        ["daytime"],
        ["evening"],
        ["morning", "daytime"],
        ["daytime", "evening"],
    ]

    automatic_indexes = set(rng.sample(range(1, STUDENTS_COUNT + 1), AUTOMATIC_STUDENTS_COUNT))
    students: list[dict[str, Any]] = []
    for index in range(1, STUDENTS_COUNT + 1):
        transmission = "automatic" if index in automatic_indexes else "manual"
        lesson_count = AUTOMATIC_REQUIRED_LESSONS if transmission == "automatic" else MANUAL_REQUIRED_LESSONS
        remaining_hours = lesson_count * LESSON_DURATION.total_seconds() / 3600
        students.append(
            {
                "id": f"student-{index:02d}",
                "group_id": "group-main-01",
                "name": f"Student {index:02d}",
                "transmission": transmission,
                "remaining_lessons_required_max": lesson_count,
                "remaining_hours_estimated": remaining_hours,
                "target_lessons": lesson_count,
                "preferences": {
                    "preferred_weekdays": weekday_sets[(index - 1) % len(weekday_sets)],
                    "preferred_time_windows": time_window_sets[(index + rng.randint(0, 4)) % len(time_window_sets)],
                },
            }
        )
    return students


def _build_instructors(rng: random.Random, period_start: datetime, period_end: datetime) -> list[dict[str, Any]]:
    instructors: list[dict[str, Any]] = []
    for index in range(1, INSTRUCTORS_COUNT + 1):
        if index == 1:
            working_weekdays = list(WEEKDAY_NAMES)
            weekend_enabled = True
        else:
            working_weekdays = [weekday for weekday in WEEKDAY_NAMES if weekday not in WEEKEND_DAYS]
            weekend_enabled = False

        work_from = INSTRUCTOR_WORK_FROM
        work_to = INSTRUCTOR_WORK_TO
        instructors.append(
            {
                "id": f"instructor-{index:02d}",
                "name": f"Instructor {index:02d}",
                "status": "active",
                "working_hours": f"{work_from:%H:%M}-{work_to:%H:%M}",
                "working_weekdays": working_weekdays,
                "weekend_enabled": weekend_enabled,
                "availability": _build_availability_windows(
                    "instructor",
                    f"instructor-{index:02d}",
                    period_start,
                    period_end,
                    working_weekdays,
                    work_from,
                    work_to,
                ),
            }
        )
    return instructors


def _build_cars(period_start: datetime, period_end: datetime) -> list[dict[str, Any]]:
    manual = [
        {
            "id": f"manual-car-{index:02d}",
            "transmission": "manual",
            "status": "active",
            "working_hours": f"{CAR_AVAILABLE_FROM:%H:%M}-{CAR_AVAILABLE_TO:%H:%M}",
            "availability": _build_availability_windows(
                "car",
                f"manual-car-{index:02d}",
                period_start,
                period_end,
                [weekday for weekday in WEEKDAY_NAMES if weekday not in WEEKEND_DAYS],
                CAR_AVAILABLE_FROM,
                CAR_AVAILABLE_TO,
            ),
        }
        for index in range(1, MANUAL_CARS_COUNT + 1)
    ]
    automatic = [
        {
            "id": f"automatic-car-{index:02d}",
            "transmission": "automatic",
            "status": "active",
            "working_hours": f"{CAR_AVAILABLE_FROM:%H:%M}-{CAR_AVAILABLE_TO:%H:%M}",
            "availability": _build_availability_windows(
                "car",
                f"automatic-car-{index:02d}",
                period_start,
                period_end,
                WEEKDAY_NAMES,
                CAR_AVAILABLE_FROM,
                CAR_AVAILABLE_TO,
            ),
        }
        for index in range(1, AUTOMATIC_CARS_COUNT + 1)
    ]
    return manual + automatic


def _build_slots(period_start: datetime, period_end: datetime) -> list[dict[str, Any]]:
    slots: list[dict[str, Any]] = []
    current_date = period_start.date()
    while current_date < period_end.date():
        weekday = WEEKDAY_NAMES[current_date.weekday()]
        for slot_time in SLOT_START_TIMES:
            start_at = datetime.combine(current_date, slot_time, tzinfo=period_start.tzinfo)
            end_at = start_at + LESSON_DURATION
            if start_at < period_start or end_at > period_end:
                continue
            slots.append(
                {
                    "id": f"slot-{len(slots) + 1:04d}",
                    "start_at": start_at,
                    "end_at": end_at,
                    "weekday": weekday,
                    "time_window": _time_window(start_at),
                }
            )
        current_date += timedelta(days=1)
    return slots


def _build_availability_windows(
    resource_type: str,
    resource_id: str,
    period_start: datetime,
    period_end: datetime,
    working_weekdays: list[str],
    work_from: time,
    work_to: time,
) -> list[dict[str, str]]:
    windows: list[dict[str, str]] = []
    current_date = period_start.date()
    while current_date < period_end.date():
        weekday = WEEKDAY_NAMES[current_date.weekday()]
        if weekday in working_weekdays:
            start_at = datetime.combine(current_date, work_from, tzinfo=period_start.tzinfo)
            end_at = datetime.combine(current_date, work_to, tzinfo=period_start.tzinfo)
            windows.append(
                {
                    "resource_type": resource_type,
                    "resource_id": resource_id,
                    "type": "working_hours",
                    "start_at": _format_dt(start_at),
                    "end_at": _format_dt(end_at),
                }
            )
        current_date += timedelta(days=1)
    return windows


def _build_existing_lessons(
    rng: random.Random,
    slots: list[dict[str, Any]],
    instructors: list[dict[str, Any]],
    cars: list[dict[str, Any]],
    reserved_lessons: list[dict[str, Any]],
) -> list[dict[str, Any]]:
    lessons: list[dict[str, Any]] = []
    if EXISTING_LESSONS_COUNT <= 0:
        return lessons
    occupied: set[tuple[str, str, str]] = set()
    for lesson in reserved_lessons:
        slot_id = _slot_id_by_start(slots, lesson["start_at"])
        if slot_id is None:
            continue
        occupied.add((slot_id, lesson["instructor_id"], ""))
        occupied.add((slot_id, "", lesson["car_id"]))
        occupied.add((slot_id, lesson["instructor_id"], lesson["car_id"]))

    candidates: list[tuple[dict[str, Any], dict[str, Any], dict[str, Any]]] = []
    for slot in slots:
        for instructor in instructors:
            if not _slot_in_availability(slot, instructor["availability"]):
                continue
            for car in cars:
                if _slot_in_availability(slot, car["availability"]):
                    candidates.append((slot, instructor, car))
    rng.shuffle(candidates)

    for slot, instructor, car in candidates:
        key = (slot["id"], instructor["id"], car["id"])
        if any(item in occupied for item in [(slot["id"], instructor["id"], ""), (slot["id"], "", car["id"]), key]):
            continue
        occupied.add((slot["id"], instructor["id"], ""))
        occupied.add((slot["id"], "", car["id"]))
        occupied.add(key)
        lessons.append(
            {
                "id": f"existing-lesson-{len(lessons) + 1:03d}",
                "student_id": f"external-student-{len(lessons) + 1:03d}",
                "group_id": f"external-group-{rng.randint(1, 8):02d}",
                "instructor_id": instructor["id"],
                "car_id": car["id"],
                "start_at": _format_dt(slot["start_at"]),
                "end_at": _format_dt(slot["end_at"]),
                "status": "scheduled",
                "source": "pre_existing_other_group",
            }
        )
        if len(lessons) >= EXISTING_LESSONS_COUNT:
            break
    if len(lessons) < EXISTING_LESSONS_COUNT:
        raise RuntimeError(
            f"Could not place {EXISTING_LESSONS_COUNT} existing lessons; placed {len(lessons)} in remaining capacity."
        )
    return sorted(lessons, key=lambda item: (item["start_at"], item["student_id"]))


def _build_schedule(
    students: list[dict[str, Any]],
    slots: list[dict[str, Any]],
    instructors: list[dict[str, Any]],
    cars: list[dict[str, Any]],
    existing_lessons: list[dict[str, Any]],
) -> list[dict[str, Any]]:
    busy_instructors: set[tuple[str, str]] = set()
    busy_cars: set[tuple[str, str]] = set()
    busy_students: set[tuple[str, str]] = set()
    for lesson in existing_lessons:
        slot_key = lesson["start_at"]
        busy_instructors.add((slot_key, lesson["instructor_id"]))
        busy_cars.add((slot_key, lesson["car_id"]))

    generated: list[dict[str, Any]] = []
    remaining = {student["id"]: int(student["target_lessons"]) for student in students}
    students_by_id = {student["id"]: student for student in students}

    for slot in slots:
        if not any(count > 0 for count in remaining.values()):
            break

        available_instructors = [
            instructor["id"]
            for instructor in instructors
            if (slot["start_at"].isoformat(), instructor["id"]) not in busy_instructors
            and _slot_in_availability(slot, instructor["availability"])
        ]
        available_cars = [
            car
            for car in cars
            if (slot["start_at"].isoformat(), car["id"]) not in busy_cars
            and _slot_in_availability(slot, car["availability"])
        ]
        if not available_instructors or not available_cars:
            continue

        for student_id in _student_order_for_slot(students, remaining, slot):
            if remaining[student_id] <= 0 or not available_instructors:
                continue
            if (slot["start_at"].isoformat(), student_id) in busy_students:
                continue

            student = students_by_id[student_id]
            car_index = next(
                (
                    index
                    for index, car in enumerate(available_cars)
                    if car["transmission"] == student["transmission"]
                ),
                None,
            )
            if car_index is None:
                continue

            instructor_id = available_instructors.pop(0)
            car = available_cars.pop(car_index)
            busy_instructors.add((slot["start_at"].isoformat(), instructor_id))
            busy_cars.add((slot["start_at"].isoformat(), car["id"]))
            busy_students.add((slot["start_at"].isoformat(), student_id))
            remaining[student_id] -= 1
            generated.append(
                {
                    "id": f"generated-lesson-{len(generated) + 1:04d}",
                    "student_id": student_id,
                    "group_id": student["group_id"],
                    "instructor_id": instructor_id,
                    "car_id": car["id"],
                    "start_at": _format_dt(slot["start_at"]),
                    "end_at": _format_dt(slot["end_at"]),
                    "status": "scheduled",
                    "source": "generated_for_main_group",
                    "matches_preferred_weekday": slot["weekday"] in student["preferences"]["preferred_weekdays"],
                    "matches_preferred_time_window": slot["time_window"] in student["preferences"]["preferred_time_windows"],
                }
            )
            if not available_cars:
                break

    for slot in slots:
        if not any(count > 0 for count in remaining.values()):
            break

        available_instructors = [
            instructor["id"]
            for instructor in instructors
            if (slot["start_at"].isoformat(), instructor["id"]) not in busy_instructors
            and _slot_in_availability(slot, instructor["availability"])
        ]
        available_cars = [
            car
            for car in cars
            if (slot["start_at"].isoformat(), car["id"]) not in busy_cars
            and _slot_in_availability(slot, car["availability"])
        ]
        if not available_instructors or not available_cars:
            continue

        for student_id in sorted(remaining, key=lambda item: (-remaining[item], item)):
            if remaining[student_id] <= 0 or not available_instructors:
                continue
            if (slot["start_at"].isoformat(), student_id) in busy_students:
                continue

            student = students_by_id[student_id]
            car_index = next(
                (
                    index
                    for index, car in enumerate(available_cars)
                    if car["transmission"] == student["transmission"]
                ),
                None,
            )
            if car_index is None:
                continue

            instructor_id = available_instructors.pop(0)
            car = available_cars.pop(car_index)
            busy_instructors.add((slot["start_at"].isoformat(), instructor_id))
            busy_cars.add((slot["start_at"].isoformat(), car["id"]))
            busy_students.add((slot["start_at"].isoformat(), student_id))
            remaining[student_id] -= 1
            generated.append(
                {
                    "id": f"generated-lesson-{len(generated) + 1:04d}",
                    "student_id": student_id,
                    "group_id": student["group_id"],
                    "instructor_id": instructor_id,
                    "car_id": car["id"],
                    "start_at": _format_dt(slot["start_at"]),
                    "end_at": _format_dt(slot["end_at"]),
                    "status": "scheduled",
                    "source": "generated_for_main_group",
                    "matches_preferred_weekday": slot["weekday"] in student["preferences"]["preferred_weekdays"],
                    "matches_preferred_time_window": slot["time_window"] in student["preferences"]["preferred_time_windows"],
                }
            )
            if not available_cars:
                break

    return generated


def _student_order_for_slot(
    students: list[dict[str, Any]],
    remaining: dict[str, int],
    slot: dict[str, Any],
) -> list[str]:
    def score(student: dict[str, Any]) -> tuple[int, int, str]:
        preferences = student["preferences"]
        matched = int(slot["weekday"] in preferences["preferred_weekdays"]) + int(
            slot["time_window"] in preferences["preferred_time_windows"]
        )
        return (-matched, -remaining[student["id"]], student["id"])

    return [student["id"] for student in sorted(students, key=score)]


def _build_report(
    period_start: datetime,
    period_end: datetime,
    students: list[dict[str, Any]],
    instructors: list[dict[str, Any]],
    cars: list[dict[str, Any]],
    slots: list[dict[str, Any]],
    existing_lessons: list[dict[str, Any]],
    generated_lessons: list[dict[str, Any]],
) -> dict[str, Any]:
    all_lessons = existing_lessons + generated_lessons
    generated_counts = Counter(lesson["student_id"] for lesson in generated_lessons)
    last_generated_at = max((lesson["start_at"] for lesson in generated_lessons), default=None)
    remaining_by_student = []
    for student in students:
        scheduled = generated_counts[student["id"]]
        target = int(student["target_lessons"])
        remaining_lessons = max(0, target - scheduled)
        remaining_by_student.append(
            {
                "student_id": student["id"],
                "transmission": student["transmission"],
                "target_lessons": target,
                "scheduled_lessons": scheduled,
                "remaining_lessons": remaining_lessons,
                "remaining_hours": remaining_lessons * LESSON_DURATION.total_seconds() / 3600,
                "last_lesson_at": _last_lesson_at(student["id"], generated_lessons),
                "program_completed": remaining_lessons == 0,
            }
        )

    car_report = _resource_report(cars, all_lessons, slots, "car_id")
    instructor_report = _resource_report(instructors, all_lessons, slots, "instructor_id")
    car_available_hours = sum(item["available_hours"] for item in car_report)
    car_busy_hours = sum(item["busy_hours"] for item in car_report)
    instructor_available_hours = sum(item["available_hours"] for item in instructor_report)
    instructor_busy_hours = sum(item["busy_hours"] for item in instructor_report)
    car_idle_hours = max(0.0, car_available_hours - car_busy_hours)
    instructor_idle_hours = max(0.0, instructor_available_hours - instructor_busy_hours)
    return {
        "summary": {
            "students": len(students),
            "automatic_students": sum(1 for student in students if student["transmission"] == "automatic"),
            "manual_students": sum(1 for student in students if student["transmission"] == "manual"),
            "cars": len(cars),
            "automatic_cars": sum(1 for car in cars if car["transmission"] == "automatic"),
            "manual_cars": sum(1 for car in cars if car["transmission"] == "manual"),
            "instructors": len(instructors),
            "slots_per_resource": len(slots),
            "existing_lessons": len(existing_lessons),
            "generated_lessons": len(generated_lessons),
            "target_lessons": sum(int(student["target_lessons"]) for student in students),
            "period_start": _format_dt(period_start),
            "period_end": _format_dt(period_end),
            "main_group_completed": all(item["program_completed"] for item in remaining_by_student),
            "last_generated_lesson_at": last_generated_at,
            "car_available_hours": car_available_hours,
            "car_busy_hours": car_busy_hours,
            "car_idle_hours": car_idle_hours,
            "car_idle_percent": _percent(car_idle_hours, car_available_hours),
            "instructor_available_hours": instructor_available_hours,
            "instructor_busy_hours": instructor_busy_hours,
            "instructor_idle_hours": instructor_idle_hours,
            "instructor_idle_percent": _percent(instructor_idle_hours, instructor_available_hours),
        },
        "cars": car_report,
        "instructors": instructor_report,
        "students": remaining_by_student,
    }


def _resource_report(
    resources: list[dict[str, Any]],
    lessons: list[dict[str, Any]],
    slots: list[dict[str, Any]],
    key: str,
) -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = []
    for resource in resources:
        resource_lessons = [lesson for lesson in lessons if lesson[key] == resource["id"]]
        existing_count = sum(1 for lesson in resource_lessons if lesson["source"] == "pre_existing_other_group")
        generated_count = len(resource_lessons) - existing_count
        occupied_slots = len(resource_lessons)
        available_hours = _availability_hours(resource.get("availability") or [])
        busy_hours = occupied_slots * LESSON_DURATION.total_seconds() / 3600
        result.append(
            {
                "resource_id": resource["id"],
                "transmission": resource.get("transmission"),
                "available_hours": available_hours,
                "occupied_lessons": occupied_slots,
                "generated_lessons": generated_count,
                "existing_lessons": existing_count,
                "busy_hours": busy_hours,
                "idle_hours": max(0.0, available_hours - busy_hours),
            }
        )
    return result


def _slot_in_availability(slot: dict[str, Any], availability: list[dict[str, str]]) -> bool:
    for window in availability:
        if window["type"] != "working_hours":
            continue
        if _parse_dt(window["start_at"]) <= slot["start_at"] and slot["end_at"] <= _parse_dt(window["end_at"]):
            return True
    return False


def _availability_hours(availability: list[dict[str, str]]) -> float:
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


def _last_lesson_at(student_id: str, lessons: list[dict[str, Any]]) -> str | None:
    student_lessons = [lesson["start_at"] for lesson in lessons if lesson["student_id"] == student_id]
    return max(student_lessons) if student_lessons else None


def _time_window(start_at: datetime) -> str:
    if start_at.hour < 12:
        return "morning"
    if start_at.hour < 17:
        return "daytime"
    return "evening"


def _slot_to_dict(slot: dict[str, Any]) -> dict[str, Any]:
    return {
        **slot,
        "start_at": _format_dt(slot["start_at"]),
        "end_at": _format_dt(slot["end_at"]),
    }


def _slot_id_by_start(slots: list[dict[str, Any]], start_at: str) -> str | None:
    for slot in slots:
        if _format_dt(slot["start_at"]) == start_at:
            return slot["id"]
    return None


def _print_console_report(payload: dict[str, Any], output: Path) -> None:
    report = payload["report"]
    summary = report["summary"]
    print("Dataset generated")
    print(f"  file: {output}")
    print(f"  period: {summary['period_start']} - {summary['period_end']}")
    print(
        "  students: "
        f"{summary['students']} (manual={summary['manual_students']}, automatic={summary['automatic_students']})"
    )
    print(
        "  cars: "
        f"{summary['cars']} (manual={summary['manual_cars']}, automatic={summary['automatic_cars']})"
    )
    print(f"  instructors: {summary['instructors']}")
    print(f"  slots per resource: {summary['slots_per_resource']}")
    print(f"  existing lessons from other groups: {summary['existing_lessons']}")
    print(f"  generated lessons: {summary['generated_lessons']} / target {summary['target_lessons']}")
    print(f"  main group completed: {summary['main_group_completed']}")
    print(f"  last generated lesson: {summary['last_generated_lesson_at']}")
    print(
        "  car available/busy/idle hours: "
        f"{summary['car_available_hours']:.1f} / {summary['car_busy_hours']:.1f} / "
        f"{summary['car_idle_hours']:.1f} ({summary['car_idle_percent']:.2f}%)"
    )
    print(
        "  instructor available/busy/idle hours: "
        f"{summary['instructor_available_hours']:.1f} / {summary['instructor_busy_hours']:.1f} / "
        f"{summary['instructor_idle_hours']:.1f} ({summary['instructor_idle_percent']:.2f}%)"
    )

    print()
    print("Car utilization")
    print("  resource          avail_h  busy_h  idle_h  generated  existing")
    for item in report["cars"]:
        print(
            f"  {item['resource_id']:<16} "
            f"{item['available_hours']:>7.1f} "
            f"{item['busy_hours']:>6.1f} "
            f"{item['idle_hours']:>7.1f} "
            f"{item['generated_lessons']:>9} "
            f"{item['existing_lessons']:>8}"
        )

    print()
    print("Instructor utilization")
    print("  resource          avail_h  busy_h  idle_h  generated  existing")
    for item in report["instructors"]:
        print(
            f"  {item['resource_id']:<16} "
            f"{item['available_hours']:>7.1f} "
            f"{item['busy_hours']:>6.1f} "
            f"{item['idle_hours']:>7.1f} "
            f"{item['generated_lessons']:>9} "
            f"{item['existing_lessons']:>8}"
        )

    print()
    print("Students remaining lessons")
    print("  student     type       target  scheduled  remaining  remaining_h  last_lesson")
    for item in report["students"]:
        print(
            f"  {item['student_id']:<10} "
            f"{item['transmission']:<10} "
            f"{item['target_lessons']:>6} "
            f"{item['scheduled_lessons']:>10} "
            f"{item['remaining_lessons']:>9} "
            f"{item['remaining_hours']:>11.1f} "
            f"{item['last_lesson_at'] or '-'}"
        )


def _format_dt(value: datetime) -> str:
    return value.isoformat()


def _parse_dt(value: str) -> datetime:
    return datetime.fromisoformat(value)


def _progress(message: str) -> None:
    print(f"[dataset] {message}", flush=True)


if __name__ == "__main__":
    main()
