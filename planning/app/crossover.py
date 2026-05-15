from __future__ import annotations

from .models import CandidateLesson


def crossover_population(primary: list[CandidateLesson], secondary: list[CandidateLesson]) -> list[CandidateLesson]:
    if not primary:
        return list(secondary)
    if not secondary:
        return list(primary)

    pivot = len(primary) // 2
    return list(primary[:pivot]) + list(secondary[pivot:])
