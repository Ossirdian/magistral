from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    kafka_bootstrap_servers: str
    kafka_request_topic: str
    kafka_result_topic: str
    kafka_group_id: str
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
            log_level=os.getenv("PLANNING_LOG_LEVEL", "INFO"),
            metrics_port=int(os.getenv("PLANNING_METRICS_PORT", "9101")),
            tracing_collector_endpoint=os.getenv("TRACING_COLLECTOR_ENDPOINT", ""),
            tracing_service_version=os.getenv("TRACING_SERVICE_VERSION", "dev"),
        )
