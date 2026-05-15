from __future__ import annotations

from .constraints import has_conflict
from .models import CandidateLesson, ExistingLesson


def repair_population(lessons: list[CandidateLesson], existing_lessons: list[ExistingLesson]) -> list[CandidateLesson]:
    repaired: list[CandidateLesson] = []
    for lesson in lessons:
        if has_conflict(lesson, repaired, existing_lessons):
            continue
        repaired.append(lesson)
    return repaired
