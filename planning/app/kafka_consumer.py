from __future__ import annotations

import json
import logging
import time
from types import SimpleNamespace

from .data_loader import HTTPDataLoader
from .kafka_connection import connect_kafka
from .models import PlanningRequest
from .observability import Observability, ProcessingStats
from .planner import Planner
from .settings import Settings


class PlanningWorker:
    def __init__(self, settings: Settings) -> None:
        logging.basicConfig(level=getattr(logging, settings.log_level.upper(), logging.INFO))
        self._settings = settings
        self._logger = logging.getLogger("planning")
        self._planner = Planner(HTTPDataLoader(settings))
        self._observability = Observability(
            metrics_port=settings.metrics_port,
            otlp_endpoint=settings.tracing_collector_endpoint,
            service_version=settings.tracing_service_version,
        )

    def run_forever(self) -> None:
        self._observability.start_metrics_server()
        kafka = _import_kafka()
        consumer, producer = connect_kafka(kafka, self._settings, self._logger)

        self._logger.info("planning worker started")
        while True:
            for message in consumer:
                timer = self._observability.measure()
                try:
                    request = PlanningRequest.from_dict(message.value)
                    with self._observability.request_span(request.request_id, request.trace_context) as span:
                        span.set_attribute("planning.kafka_topic", self._settings.kafka_request_topic)
                        result = self._planner.plan(request)
                        result.trace_context.update(self._observability.inject_trace_context())
                        producer.send(self._settings.kafka_result_topic, result.as_dict()).get(timeout=10)
                        stats = ProcessingStats(
                            scheduled_lessons=len(result.lessons),
                            conflicts=len(result.conflicts),
                            unassigned_students=len(result.unassigned_students),
                            resource_idle_seconds=result.resource_idle_seconds,
                            duration_seconds=timer.elapsed(),
                        )
                        self._observability.observe_success(stats)
                        span.set_attribute("planning.scheduled_lessons", stats.scheduled_lessons)
                        span.set_attribute("planning.conflicts", stats.conflicts)
                        span.set_attribute("planning.unassigned_students", stats.unassigned_students)
                        span.set_attribute("planning.resource_idle_seconds", stats.resource_idle_seconds)
                        span.set_attribute("planning.result_topic", self._settings.kafka_result_topic)
                        for key, value in result.score_breakdown.items():
                            span.set_attribute(f"planning.score.{key}", value)
                    self._logger.info(
                        "planning request processed",
                        extra={
                            "request_id": request.request_id,
                            "scheduled_lessons": len(result.lessons),
                            "conflicts": len(result.conflicts),
                            "resource_idle_seconds": result.resource_idle_seconds,
                            "unassigned_students": len(result.unassigned_students),
                        },
                    )
                    self._logger.info(
                        "planning best solution score breakdown: "
                        "scheduled_bonus=%s scheduled_lessons=%s "
                        "unmet_lesson_penalty=%s missing_target_lessons=%s "
                        "conflict_penalty=%s conflicts=%s "
                        "idle_penalty=%s unevenness_penalty=%s "
                        "valid_options=%s avg_options_per_student=%s "
                        "preference_bonus=%s score=%s "
                        "existing_loaded=%s existing_used=%s "
                        "existing_skipped_deleted_or_cancelled=%s existing_skipped_generated=%s "
                        "replace_strategy=%s",
                        result.score_breakdown.get("scheduled_bonus", 0.0),
                        result.score_breakdown.get("scheduled_lessons", 0.0),
                        result.score_breakdown.get("unmet_lesson_penalty", 0.0),
                        result.score_breakdown.get("missing_target_lessons", 0.0),
                        result.score_breakdown.get("conflict_penalty", 0.0),
                        result.score_breakdown.get("conflicts", 0.0),
                        result.score_breakdown.get("idle_penalty", 0.0),
                        result.score_breakdown.get("unevenness_penalty", 0.0),
                        result.score_breakdown.get("valid_options", 0.0),
                        result.score_breakdown.get("avg_options_per_student", 0.0),
                        result.score_breakdown.get("preference_bonus", 0.0),
                        result.score_breakdown.get("score", result.score),
                        result.score_breakdown.get("existing_lessons_loaded", 0.0),
                        result.score_breakdown.get("existing_lessons_used", 0.0),
                        result.score_breakdown.get("existing_lessons_skipped_deleted_or_cancelled", 0.0),
                        result.score_breakdown.get("existing_lessons_skipped_generated", 0.0),
                        result.score_breakdown.get("replace_strategy", ""),
                        extra={
                            "request_id": request.request_id,
                            **{f"score_{key}": value for key, value in result.score_breakdown.items()},
                        },
                    )
                except Exception:
                    self._observability.observe_failure(timer.elapsed())
                    self._logger.exception("failed to process planning request")
            time.sleep(1)


def _import_kafka():
    from kafka import KafkaConsumer, KafkaProducer  # type: ignore

    return SimpleNamespace(KafkaConsumer=KafkaConsumer, KafkaProducer=KafkaProducer)
