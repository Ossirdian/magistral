from __future__ import annotations

from collections import Counter
from datetime import datetime, time, timedelta, timezone

from planning.app.genetic import (
    DEFAULT_CROSSOVER_PROBABILITY,
    DEFAULT_GENE_MUTATION_PROBABILITY,
    DEFAULT_GENERATIONS,
    DEFAULT_MUTATION_PROBABILITY,
    DEFAULT_POPULATION_SIZE,
    DEFAULT_SEED,
    DEFAULT_TOURNAMENT_SIZE,
    evolve_plan,
)
from planning.app.models import LESSON_DURATION, ExistingLesson, PlanningContext, StudentPreference, TimeSlot
from planning.app.option_builder import build_valid_options


SCHOOL_ID = "school-demo"
REQUEST_ID = "planning-demo-30-days"
PERIOD_DAYS = 30
STUDENTS_COUNT = 60
MANUAL_CARS_COUNT = 10
AUTOMATIC_CARS_COUNT = 1
INSTRUCTORS_COUNT = 10
EXISTING_LESSONS_PER_WEEK = 10


def main() -> None:
    period_start = datetime(2026, 5, 7, 9, 0, tzinfo=timezone.utc)
    period_end = period_start + timedelta(days=PERIOD_DAYS)
    students = [f"student-{index:02d}" for index in range(1, STUDENTS_COUNT + 1)]
    instructors = [f"instructor-{index:02d}" for index in range(1, INSTRUCTORS_COUNT + 1)]
    manual_cars = [f"manual-car-{index:02d}" for index in range(1, MANUAL_CARS_COUNT + 1)]
    automatic_cars = [f"automatic-car-{index:02d}" for index in range(1, AUTOMATIC_CARS_COUNT + 1)]
    cars = manual_cars + automatic_cars

    preferences, student_transmissions = _build_student_preferences(students)
    existing_lessons = _build_existing_lessons(period_start, instructors, manual_cars)
    target_lessons_by_student = {student_id: 4 for student_id in students}
    context = PlanningContext(
        students_to_plan=students,
        instructor_ids=instructors,
        car_ids=cars,
        existing_lessons=existing_lessons,
        student_preferences=preferences,
        student_transmissions=student_transmissions,
        car_transmissions={**{car_id: "manual" for car_id in manual_cars}, **{car_id: "automatic" for car_id in automatic_cars}},
        period_start=period_start,
        period_end=period_end,
        lesson_demands=[
            student_id
            for student_id, target_lessons in target_lessons_by_student.items()
            for _index in range(target_lessons)
        ],
        target_lessons_by_student=target_lessons_by_student,
    )
    slots = _build_working_slots(period_start, PERIOD_DAYS)
    options_by_student, total_car_slots = build_valid_options(slots, context)
    context = _with_options(context, options_by_student, total_car_slots)

    _print_input_summary(context, slots)
    lessons, unassigned_students, conflicts, score = evolve_plan(context)
    lessons = sorted(lessons, key=lambda lesson: (lesson.start_at, lesson.student_id))

    print()
    print("GA parameters")
    print(f"  seed: {DEFAULT_SEED}")
    print(f"  population_size: {DEFAULT_POPULATION_SIZE}")
    print(f"  generations: {DEFAULT_GENERATIONS}")
    print(f"  crossover: two-point, probability={DEFAULT_CROSSOVER_PROBABILITY}")
    print(f"  mutation_probability: {DEFAULT_MUTATION_PROBABILITY}")
    print(f"  gene_mutation_probability: {DEFAULT_GENE_MUTATION_PROBABILITY}")
    print(f"  selection: tournament, size={DEFAULT_TOURNAMENT_SIZE}")

    print()
    print("Planning result")
    print(f"  request_id: {REQUEST_ID}")
    print(f"  school_id: {SCHOOL_ID}")
    print(f"  period: {_format_dt(period_start)} - {_format_dt(period_end)} ({PERIOD_DAYS} days)")
    print(f"  score: {score:.2f}")
    print(f"  lessons: {len(lessons)}")
    print(f"  unassigned_students: {len(unassigned_students)}")
    print(f"  conflicts: {len(conflicts)}")

    _print_existing_schedule(context.existing_lessons)
    _print_schedule(lessons, context)
    _print_conflicts(unassigned_students, conflicts)


def _build_student_preferences(students: list[str]) -> tuple[dict[str, StudentPreference], dict[str, str]]:
    weekday_sets = [
        ["monday", "wednesday", "friday"],
        ["tuesday", "thursday"],
        ["saturday", "sunday"],
        ["monday", "tuesday", "thursday"],
        ["wednesday", "friday", "saturday"],
    ]
    time_windows = [["morning"], ["daytime"], ["evening"], ["morning", "daytime"], ["daytime", "evening"]]

    preferences: dict[str, StudentPreference] = {}
    transmissions: dict[str, str] = {}
    for index, student_id in enumerate(students):
        transmission = "automatic" if index in {4, 13, 22, 31} else "manual"
        preferences[student_id] = StudentPreference(
            preferred_weekdays=weekday_sets[index % len(weekday_sets)],
            preferred_time_windows=time_windows[(index * 2) % len(time_windows)],
            transmission=transmission,
        )
        transmissions[student_id] = transmission
    return preferences, transmissions


def _build_existing_lessons(
    period_start: datetime,
    instructors: list[str],
    manual_cars: list[str],
) -> list[ExistingLesson]:
    weekly_offsets = [
        (0, time(9, 0)),
        (0, time(13, 30)),
        (1, time(10, 30)),
        (1, time(15, 0)),
        (2, time(9, 0)),
        (2, time(16, 30)),
        (3, time(12, 0)),
        (4, time(10, 30)),
        (4, time(15, 0)),
        (5, time(12, 0)),
    ]
    lessons: list[ExistingLesson] = []
    start_day = period_start.date()
    weeks = (PERIOD_DAYS + 6) // 7
    for week_index in range(weeks):
        week_start = start_day + timedelta(days=week_index * 7)
        available_days = min(7, PERIOD_DAYS - week_index * 7)
        available_offsets = [
            (day_offset, slot_time)
            for day_offset, slot_time in weekly_offsets
            if day_offset < available_days
        ]
        lesson_offsets = available_offsets[:EXISTING_LESSONS_PER_WEEK]
        if len(lesson_offsets) < EXISTING_LESSONS_PER_WEEK:
            all_week_offsets = [
                (day_offset, slot_time)
                for day_offset in range(available_days)
                for slot_time in [time(9, 0), time(10, 30), time(12, 0), time(13, 30), time(15, 0), time(16, 30)]
            ]
            lesson_offsets = all_week_offsets[:EXISTING_LESSONS_PER_WEEK]

        for lesson_index, (day_offset, slot_time) in enumerate(lesson_offsets):
            lesson_day = week_start + timedelta(days=day_offset)
            start_at = datetime.combine(lesson_day, slot_time, tzinfo=period_start.tzinfo)
            lessons.append(
                ExistingLesson(
                    student_id=f"existing-student-{week_index + 1:02d}-{lesson_index + 1:02d}",
                    instructor_id=instructors[(week_index + lesson_index) % len(instructors)],
                    car_id=manual_cars[(week_index * 3 + lesson_index) % len(manual_cars)],
                    start_at=start_at,
                    end_at=start_at + LESSON_DURATION,
                )
            )
    return lessons


def _build_working_slots(period_start: datetime, days: int) -> list[TimeSlot]:
    daily_starts = [time(9, 0), time(10, 30), time(12, 0), time(13, 30), time(15, 0), time(16, 30)]
    slots: list[TimeSlot] = []
    start_day = period_start.date()
    for day_offset in range(days):
        day = start_day + timedelta(days=day_offset)
        for slot_time in daily_starts:
            start_at = datetime.combine(day, slot_time, tzinfo=period_start.tzinfo)
            slots.append(TimeSlot(index=len(slots), start_at=start_at, end_at=start_at + LESSON_DURATION))
    return slots


def _print_input_summary(context: PlanningContext, slots: list[TimeSlot]) -> None:
    student_transmissions = Counter(context.student_transmissions.values())
    car_transmissions = Counter(context.car_transmissions.values())

    print("Planning demo input")
    print(f"  students: {len(context.students_to_plan)} (manual={student_transmissions['manual']}, automatic={student_transmissions['automatic']})")
    print(f"  cars: {len(context.car_ids)} (manual={car_transmissions['manual']}, automatic={car_transmissions['automatic']})")
    print(f"  instructors: {len(context.instructor_ids)}")
    print(f"  existing_lessons: {len(context.existing_lessons)} ({EXISTING_LESSONS_PER_WEEK} per week)")
    print(f"  working_slots: {len(slots)}")
    print(f"  car_slots: {context.total_car_slots}")
    print(f"  valid_options: {sum(len(options) for options in context.valid_options_by_student.values())}")
    print(f"  lesson_duration_minutes: {int(LESSON_DURATION.total_seconds() // 60)}")


def _with_options(
    context: PlanningContext,
    options_by_student,
    total_car_slots: int,
) -> PlanningContext:
    return PlanningContext(
        students_to_plan=context.students_to_plan,
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
        lesson_demands=context.lesson_demands,
        target_lessons_by_student=context.target_lessons_by_student,
        valid_options_by_student=options_by_student,
        total_car_slots=total_car_slots,
    )


def _print_existing_schedule(existing_lessons: list[ExistingLesson]) -> None:
    print()
    print("Existing schedule")
    print("  date        start  end    student                 instructor     car")
    print("  ----------  -----  -----  ----------------------  -------------  -------------")
    for lesson in sorted(existing_lessons, key=lambda item: (item.start_at, item.student_id)):
        print(
            "  "
            f"{lesson.start_at:%Y-%m-%d}  "
            f"{lesson.start_at:%H:%M}  "
            f"{lesson.end_at:%H:%M}  "
            f"{lesson.student_id:<22}  "
            f"{lesson.instructor_id:<13}  "
            f"{lesson.car_id}"
        )


def _print_schedule(lessons, context: PlanningContext) -> None:
    print()
    print("Generated schedule")
    print("  date        start  end    student     pref_days                 pref_time        transmission  instructor     car")
    print("  ----------  -----  -----  ----------  ------------------------  ---------------  ------------  -------------  ----------------")
    for lesson in lessons:
        preference = context.student_preferences[lesson.student_id]
        print(
            "  "
            f"{lesson.start_at:%Y-%m-%d}  "
            f"{lesson.start_at:%H:%M}  "
            f"{lesson.end_at:%H:%M}  "
            f"{lesson.student_id:<10}  "
            f"{','.join(preference.preferred_weekdays):<24}  "
            f"{','.join(preference.preferred_time_windows):<15}  "
            f"{context.student_transmissions[lesson.student_id]:<12}  "
            f"{lesson.instructor_id:<13}  "
            f"{lesson.car_id}"
        )


def _print_conflicts(unassigned_students: list[str], conflicts) -> None:
    if not unassigned_students and not conflicts:
        return

    print()
    print("Unassigned and conflicts")
    for conflict in conflicts:
        print(f"  {conflict.student_id}: {conflict.reason}")


def _format_dt(value: datetime) -> str:
    return value.isoformat()


if __name__ == "__main__":
    main()
