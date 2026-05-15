from __future__ import annotations

import logging
from dataclasses import dataclass
from time import perf_counter
from typing import Iterator

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from prometheus_client import Counter, Histogram, start_http_server


REQUESTS_PROCESSED = Counter(
    "planning_requests_processed_total",
    "Total number of successfully processed planning requests.",
)

REQUESTS_FAILED = Counter(
    "planning_requests_failed_total",
    "Total number of failed planning requests.",
)

ASSIGNED_LESSONS_TOTAL = Counter(
    "planning_assigned_lessons_total",
    "Total number of lessons assigned by the planning service.",
)

CONFLICTS_TOTAL = Counter(
    "planning_conflicts_total",
    "Total number of conflicts produced by the planning service.",
)

LESSONS_SCHEDULED = Histogram(
    "planning_scheduled_lessons",
    "Number of lessons generated for a planning request.",
    buckets=(0, 1, 5, 10, 20, 50, 100),
)

CONFLICTS_PER_REQUEST = Histogram(
    "planning_conflicts_per_request",
    "Number of conflicts generated for a planning request.",
    buckets=(0, 1, 5, 10, 20, 50, 100),
)

UNASSIGNED_STUDENTS = Histogram(
    "planning_unassigned_students",
    "Number of unassigned students in a planning result.",
    buckets=(0, 1, 5, 10, 20, 50, 100),
)

RESOURCE_IDLE_SECONDS = Histogram(
    "planning_resource_idle_seconds",
    "Idle time of instructors and cars within the planning window, in seconds.",
    buckets=(0, 300, 900, 1800, 3600, 7200, 14400, 28800, 57600, 86400),
)

REQUEST_DURATION = Histogram(
    "planning_request_duration_seconds",
    "Planning request processing duration in seconds.",
    buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60),
)


@dataclass(frozen=True)
class ProcessingStats:
    scheduled_lessons: int
    conflicts: int
    unassigned_students: int
    resource_idle_seconds: float
    duration_seconds: float


class Observability:
    def __init__(self, metrics_port: int, otlp_endpoint: str, service_version: str) -> None:
        self._metrics_port = metrics_port
        self._logger = logging.getLogger("planning.observability")
        self._tracer = trace.get_tracer("planning")
        self._metrics_started = False

        if otlp_endpoint:
            provider = TracerProvider(
                resource=Resource.create(
                    {
                        "service.name": "planning",
                        "service.version": service_version,
                    }
                )
            )
            provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=otlp_endpoint)))
            trace.set_tracer_provider(provider)
            self._tracer = trace.get_tracer("planning")

    def start_metrics_server(self) -> None:
        if self._metrics_started:
            return
        start_http_server(self._metrics_port)
        self._metrics_started = True
        self._logger.info("planning metrics server started", extra={"metrics_port": self._metrics_port})

    def request_span(self, request_id: str) -> Iterator[trace.Span]:
        return self._tracer.start_as_current_span(
            "planning.request",
            attributes={
                "planning.request_id": request_id,
            },
        )

    def measure(self) -> "_Timer":
        return _Timer()

    def observe_success(self, stats: ProcessingStats) -> None:
        REQUESTS_PROCESSED.inc()
        ASSIGNED_LESSONS_TOTAL.inc(stats.scheduled_lessons)
        CONFLICTS_TOTAL.inc(stats.conflicts)
        LESSONS_SCHEDULED.observe(stats.scheduled_lessons)
        CONFLICTS_PER_REQUEST.observe(stats.conflicts)
        UNASSIGNED_STUDENTS.observe(stats.unassigned_students)
        RESOURCE_IDLE_SECONDS.observe(stats.resource_idle_seconds)
        REQUEST_DURATION.observe(stats.duration_seconds)

    def observe_failure(self, duration_seconds: float) -> None:
        REQUESTS_FAILED.inc()
        REQUEST_DURATION.observe(duration_seconds)


class _Timer:
    def __init__(self) -> None:
        self._started = perf_counter()

    def elapsed(self) -> float:
        return perf_counter() - self._started
