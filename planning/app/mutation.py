from __future__ import annotations

from datetime import timedelta

from .models import CandidateLesson, PlanningContext


def mutate_population(lessons: list[CandidateLesson], context: PlanningContext) -> list[CandidateLesson]:
    if not lessons or not context.instructor_ids or not context.car_ids:
        return lessons

    mutated: list[CandidateLesson] = []
    for index, lesson in enumerate(lessons):
        if index % 2 == 0:
            mutated.append(
                CandidateLesson(
                    student_id=lesson.student_id,
                    instructor_id=context.instructor_ids[index % len(context.instructor_ids)],
                    car_id=context.car_ids[index % len(context.car_ids)],
                    start_at=lesson.start_at + timedelta(minutes=0),
                    end_at=lesson.end_at + timedelta(minutes=0),
                    status=lesson.status,
                )
            )
            continue
        mutated.append(lesson)
    return mutated
