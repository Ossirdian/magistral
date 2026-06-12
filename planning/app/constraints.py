from __future__ import annotations

from .models import AvailabilityWindow, CandidateLesson, ExistingLesson


def is_valid_time_range(start_at, end_at) -> bool:
    return start_at < end_at


def overlaps(left_start, left_end, right_start, right_end) -> bool:
    return left_start < right_end and right_start < left_end


def has_conflict(candidate: CandidateLesson, scheduled: list[CandidateLesson], existing: list[ExistingLesson]) -> bool:
    for lesson in scheduled:
        if _same_resource_overlap(candidate, lesson.student_id, lesson.instructor_id, lesson.car_id, lesson.start_at, lesson.end_at):
            return True
    for lesson in existing:
        if _same_resource_overlap(candidate, lesson.student_id, lesson.instructor_id, lesson.car_id, lesson.start_at, lesson.end_at):
            return True
    return False


def transmission_matches(student_transmission: str, car_transmission: str) -> bool:
    if not student_transmission or not car_transmission:
        return True
    return student_transmission == car_transmission


def available_for_lesson(candidate: CandidateLesson, windows: list[AvailabilityWindow]) -> bool:
    if not windows:
        return True

    working_windows = [window for window in windows if window.type == "working_hours"]
    blocking_windows = [window for window in windows if window.type in {"blackout", "vacation"}]

    if any(overlaps(candidate.start_at, candidate.end_at, window.start_at, window.end_at) for window in blocking_windows):
        return False

    if not working_windows:
        return True

    return any(
        window.start_at <= candidate.start_at and candidate.end_at <= window.end_at
        for window in working_windows
    )


def _same_resource_overlap(candidate: CandidateLesson, student_id: str, instructor_id: str, car_id: str, start_at, end_at) -> bool:
    if not overlaps(candidate.start_at, candidate.end_at, start_at, end_at):
        return False
    return (
        candidate.student_id == student_id
        or candidate.instructor_id == instructor_id
        or candidate.car_id == car_id
    )
