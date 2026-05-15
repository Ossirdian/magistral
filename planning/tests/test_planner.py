from datetime import datetime, timedelta, timezone
import unittest

from planning.app.constraints import has_conflict, overlaps
from planning.app.data_loader import DataLoader, LoadedPlanningData, StaticDataLoader
from planning.app.fitness import score_plan
from planning.app.genetic import decode_individual
from planning.app.models import CandidateLesson, ExistingLesson, PlanningContext, PlanningRequest, StudentPreference
from planning.app.planner import Planner


class StubDataLoader(DataLoader):
    def __init__(
        self,
        student_ids: list[str],
        instructor_ids: list[str],
        car_ids: list[str],
        existing_lessons: list[ExistingLesson] | None = None,
    ) -> None:
        self._data = LoadedPlanningData(
            student_ids=student_ids,
            instructor_ids=instructor_ids,
            car_ids=car_ids,
            existing_lessons=existing_lessons or [],
            student_preferences={student_id: StudentPreference() for student_id in student_ids},
            student_transmissions={},
            car_transmissions={},
        )

    def load(self, request: PlanningRequest) -> LoadedPlanningData:
        return self._data


class PlannerTestCase(unittest.TestCase):
    def test_planner_creates_individual_lessons(self) -> None:
        planner = Planner(StaticDataLoader())
        request = PlanningRequest(
            school_id="school-1",
            period_start=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
            period_end=datetime(2026, 4, 21, 15, 0, tzinfo=timezone.utc),
            student_ids=["student-1", "student-2"],
            group_ids=[],
            branch_ids=[],
            replace_strategy="append",
            requested_by_user_id="user-1",
            request_id="request-1",
        )

        result = planner.plan(request)

        self.assertEqual("request-1", result.request_id)
        self.assertEqual(2, len(result.lessons))
        self.assertEqual([], result.unassigned_students)
        self.assertTrue(all(lesson.student_id in {"student-1", "student-2"} for lesson in result.lessons))
        self.assertTrue(all((lesson.end_at - lesson.start_at) == timedelta(minutes=90) for lesson in result.lessons))

    def test_planner_marks_students_unassigned_when_resources_missing(self) -> None:
        planner = Planner(StubDataLoader(student_ids=["student-1"], instructor_ids=[], car_ids=[]))
        request = _request(student_ids=["student-1"])

        result = planner.plan(request)

        self.assertEqual([], result.lessons)
        self.assertEqual(["student-1"], result.unassigned_students)
        self.assertEqual("no_available_slot", result.conflicts[0].reason)

    def test_planner_avoids_existing_lesson_conflicts(self) -> None:
        existing = ExistingLesson(
            student_id="student-1",
            instructor_id="instructor-1",
            car_id="car-1",
            start_at=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
            end_at=datetime(2026, 4, 21, 10, 30, tzinfo=timezone.utc),
        )
        planner = Planner(
            StubDataLoader(
                student_ids=["student-1"],
                instructor_ids=["instructor-1"],
                car_ids=["car-1"],
                existing_lessons=[existing],
            )
        )
        request = _request(student_ids=["student-1"], period_end=datetime(2026, 4, 21, 12, 0, tzinfo=timezone.utc))

        result = planner.plan(request)

        self.assertEqual(1, len(result.lessons))
        self.assertEqual(datetime(2026, 4, 21, 10, 30, tzinfo=timezone.utc), result.lessons[0].start_at)

    def test_static_loader_expands_groups_and_deduplicates_students(self) -> None:
        loader = StaticDataLoader()
        request = _request(student_ids=["group-1-student-1"], group_ids=["group-1"])

        data = loader.load(request)

        self.assertEqual(
            ["group-1-student-1", "group-1-student-2"],
            data.student_ids,
        )

    def test_decode_individual_marks_invalid_slot_as_unassigned(self) -> None:
        context = PlanningContext(
            students_to_plan=["student-1"],
            instructor_ids=["instructor-1"],
            car_ids=["car-1"],
            existing_lessons=[],
            student_preferences={},
            student_transmissions={},
            car_transmissions={},
            period_start=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
            period_end=datetime(2026, 4, 21, 12, 0, tzinfo=timezone.utc),
        )
        slots = [(context.period_start, context.period_start + timedelta(minutes=90))]

        lessons, unassigned, conflicts = decode_individual([-1, 0, 0], context, slots)

        self.assertEqual([], lessons)
        self.assertEqual(["student-1"], unassigned)
        self.assertEqual("no_available_slot", conflicts[0].reason)

    def test_constraint_helpers_detect_overlap_and_conflict(self) -> None:
        start_at = datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc)
        end_at = start_at + timedelta(minutes=90)
        other_start = datetime(2026, 4, 21, 9, 30, tzinfo=timezone.utc)
        other_end = other_start + timedelta(minutes=90)

        self.assertTrue(overlaps(start_at, end_at, other_start, other_end))

        scheduled = [
            CandidateLesson(
                student_id="student-1",
                instructor_id="instructor-1",
                car_id="car-1",
                start_at=start_at,
                end_at=end_at,
            )
        ]
        candidate = CandidateLesson(
            student_id="student-2",
            instructor_id="instructor-1",
            car_id="car-2",
            start_at=other_start,
            end_at=other_end,
        )
        self.assertTrue(has_conflict(candidate, scheduled, []))

    def test_preference_scoring_rewards_matching_time_window(self) -> None:
        context = PlanningContext(
            students_to_plan=["student-1"],
            instructor_ids=["instructor-1"],
            car_ids=["car-1"],
            existing_lessons=[],
            student_preferences={
                "student-1": StudentPreference(
                    preferred_weekdays=["monday"],
                    preferred_time_windows=["daytime"],
                )
            },
            student_transmissions={},
            car_transmissions={},
            period_start=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
            period_end=datetime(2026, 4, 21, 18, 0, tzinfo=timezone.utc),
        )
        preferred_lesson = CandidateLesson(
            student_id="student-1",
            instructor_id="instructor-1",
            car_id="car-1",
            start_at=datetime(2026, 4, 21, 12, 0, tzinfo=timezone.utc),
            end_at=datetime(2026, 4, 21, 13, 30, tzinfo=timezone.utc),
        )
        non_preferred_lesson = CandidateLesson(
            student_id="student-1",
            instructor_id="instructor-1",
            car_id="car-1",
            start_at=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
            end_at=datetime(2026, 4, 21, 10, 30, tzinfo=timezone.utc),
        )

        preferred_score = score_plan([preferred_lesson], [], [], context)
        non_preferred_score = score_plan([non_preferred_lesson], [], [], context)

        self.assertGreater(preferred_score, non_preferred_score)

    def test_decode_individual_rejects_transmission_mismatch(self) -> None:
        context = PlanningContext(
            students_to_plan=["student-1"],
            instructor_ids=["instructor-1"],
            car_ids=["car-1"],
            existing_lessons=[],
            student_preferences={},
            student_transmissions={"student-1": "manual"},
            car_transmissions={"car-1": "automatic"},
            period_start=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
            period_end=datetime(2026, 4, 21, 12, 0, tzinfo=timezone.utc),
        )
        slots = [(context.period_start, context.period_start + timedelta(minutes=90))]

        lessons, unassigned, conflicts = decode_individual([0, 0, 0], context, slots)

        self.assertEqual([], lessons)
        self.assertEqual(["student-1"], unassigned)
        self.assertEqual("transmission_mismatch", conflicts[0].reason)


def _request(
    student_ids: list[str],
    group_ids: list[str] | None = None,
    period_end: datetime | None = None,
) -> PlanningRequest:
    return PlanningRequest(
        school_id="school-1",
        period_start=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
        period_end=period_end or datetime(2026, 4, 21, 15, 0, tzinfo=timezone.utc),
        student_ids=student_ids,
        student_preferences={},
        group_ids=group_ids or [],
        branch_ids=[],
        replace_strategy="append",
        requested_by_user_id="user-1",
        request_id="request-1",
    )


if __name__ == "__main__":
    unittest.main()
