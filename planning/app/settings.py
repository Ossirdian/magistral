from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    kafka_bootstrap_servers: str
    kafka_request_topic: str
    kafka_result_topic: str
    kafka_group_id: str
    student_service_url: str = "http://app-student:8083"
    resources_service_url: str = "http://app-resources:8080"
    lesson_service_url: str = "http://app-lesson:8087"
    service_user_id: str = "planning-service"
    kafka_connect_retry_interval_seconds: float = 5.0
    kafka_connect_max_attempts: int = 0
    log_level: str = "INFO"
    metrics_port: int = 9101
    tracing_collector_endpoint: str = ""
    tracing_service_version: str = "dev"

    @classmethod
    def from_env(cls) -> "Settings":
        return cls(
            kafka_bootstrap_servers=os.getenv("PLANNING_KAFKA_BOOTSTRAP_SERVERS", "kafka:9092"),
            kafka_request_topic=os.getenv("PLANNING_KAFKA_REQUEST_TOPIC", "planning.requests"),
            kafka_result_topic=os.getenv("PLANNING_KAFKA_RESULT_TOPIC", "planning.results"),
            kafka_group_id=os.getenv("PLANNING_KAFKA_GROUP_ID", "planning-service"),
            student_service_url=os.getenv("PLANNING_STUDENT_SERVICE_URL", "http://app-student:8083"),
            resources_service_url=os.getenv("PLANNING_RESOURCES_SERVICE_URL", "http://app-resources:8080"),
            lesson_service_url=os.getenv("PLANNING_LESSON_SERVICE_URL", "http://app-lesson:8087"),
            service_user_id=os.getenv("PLANNING_SERVICE_USER_ID", "planning-service"),
            kafka_connect_retry_interval_seconds=float(
                os.getenv("PLANNING_KAFKA_CONNECT_RETRY_INTERVAL_SECONDS", "5")
            ),
            kafka_connect_max_attempts=int(os.getenv("PLANNING_KAFKA_CONNECT_MAX_ATTEMPTS", "0")),
            log_level=os.getenv("PLANNING_LOG_LEVEL", "INFO"),
            metrics_port=int(os.getenv("PLANNING_METRICS_PORT", "9101")),
            tracing_collector_endpoint=os.getenv("TRACING_COLLECTOR_ENDPOINT", ""),
            tracing_service_version=os.getenv("TRACING_SERVICE_VERSION", "dev"),
        )
