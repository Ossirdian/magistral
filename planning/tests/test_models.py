from datetime import datetime, timezone
import unittest

from planning.app.models import CandidateLesson, Conflict, PlanningRequest, PlanningResult


class ModelsTestCase(unittest.TestCase):
    def test_planning_request_from_dict_parses_iso_dates(self) -> None:
        request = PlanningRequest.from_dict(
            {
                "school_id": "school-1",
                "period_start": "2026-04-21T09:00:00Z",
                "period_end": "2026-04-21T12:00:00Z",
                "student_ids": ["student-1"],
                "student_preferences": {
                    "student-1": {
                    "preferred_weekdays": ["monday"],
                    "preferred_time_windows": ["morning"],
                    "transmission": "manual",
                }
            },
                "group_ids": ["group-1"],
                "branch_ids": ["branch-1"],
                "replace_strategy": "merge",
                "requested_by_user_id": "user-1",
                "request_id": "request-1",
            }
        )

        self.assertEqual(datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc), request.period_start)
        self.assertEqual(datetime(2026, 4, 21, 12, 0, tzinfo=timezone.utc), request.period_end)
        self.assertEqual(["student-1"], request.student_ids)
        self.assertEqual(["monday"], request.student_preferences["student-1"].preferred_weekdays)
        self.assertEqual("manual", request.student_preferences["student-1"].transmission)
        self.assertEqual(["group-1"], request.group_ids)
        self.assertEqual("merge", request.replace_strategy)

    def test_planning_result_serializes_nested_objects(self) -> None:
        result = PlanningResult(
            request_id="request-1",
            school_id="school-1",
            lessons=[
                CandidateLesson(
                    student_id="student-1",
                    instructor_id="instructor-1",
                    car_id="car-1",
                    start_at=datetime(2026, 4, 21, 9, 0, tzinfo=timezone.utc),
                    end_at=datetime(2026, 4, 21, 10, 30, tzinfo=timezone.utc),
                )
            ],
            unassigned_students=["student-2"],
            conflicts=[Conflict(student_id="student-2", reason="resource_conflict")],
            score=95.0,
            resource_idle_seconds=3600.0,
        )

        payload = result.as_dict()

        self.assertEqual("request-1", payload["request_id"])
        self.assertEqual("student-1", payload["lessons"][0]["student_id"])
        self.assertEqual("student-2", payload["unassigned_students"][0])
        self.assertEqual("resource_conflict", payload["conflicts"][0]["reason"])
        self.assertEqual(3600.0, payload["resource_idle_seconds"])


if __name__ == "__main__":
    unittest.main()
