from __future__ import annotations

import json
import logging
import time
from types import SimpleNamespace

from .data_loader import StaticDataLoader
from .models import PlanningRequest
from .observability import Observability, ProcessingStats
from .planner import Planner
from .settings import Settings


class PlanningWorker:
    def __init__(self, settings: Settings) -> None:
        logging.basicConfig(level=getattr(logging, settings.log_level.upper(), logging.INFO))
        self._settings = settings
        self._logger = logging.getLogger("planning")
        self._planner = Planner(StaticDataLoader())
        self._observability = Observability(
            metrics_port=settings.metrics_port,
            otlp_endpoint=settings.tracing_collector_endpoint,
            service_version=settings.tracing_service_version,
        )

    def run_forever(self) -> None:
        self._observability.start_metrics_server()
        kafka = _import_kafka()
        consumer = kafka.KafkaConsumer(
            self._settings.kafka_request_topic,
            bootstrap_servers=self._settings.kafka_bootstrap_servers,
            group_id=self._settings.kafka_group_id,
            value_deserializer=lambda value: json.loads(value.decode("utf-8")),
            auto_offset_reset="earliest",
            enable_auto_commit=True,
        )
        producer = kafka.KafkaProducer(
            bootstrap_servers=self._settings.kafka_bootstrap_servers,
            value_serializer=lambda value: json.dumps(value).encode("utf-8"),
        )

        self._logger.info("planning worker started")
        while True:
            for message in consumer:
                timer = self._observability.measure()
                try:
                    request = PlanningRequest.from_dict(message.value)
                    with self._observability.request_span(request.request_id) as span:
                        span.set_attribute("planning.kafka_topic", self._settings.kafka_request_topic)
                        result = self._planner.plan(request)
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
                except Exception:
                    self._observability.observe_failure(timer.elapsed())
                    self._logger.exception("failed to process planning request")
            time.sleep(1)


def _import_kafka():
    from kafka import KafkaConsumer, KafkaProducer  # type: ignore

    return SimpleNamespace(KafkaConsumer=KafkaConsumer, KafkaProducer=KafkaProducer)
