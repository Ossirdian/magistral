from __future__ import annotations

import json
import logging
import time
from collections.abc import Callable
from typing import Any

from .settings import Settings


def connect_kafka(
    kafka: Any,
    settings: Settings,
    logger: logging.Logger,
    sleep: Callable[[float], None] = time.sleep,
) -> tuple[Any, Any]:
    attempt = 0
    while True:
        attempt += 1
        consumer = None
        try:
            consumer = kafka.KafkaConsumer(
                settings.kafka_request_topic,
                bootstrap_servers=settings.kafka_bootstrap_servers,
                group_id=settings.kafka_group_id,
                value_deserializer=lambda value: json.loads(value.decode("utf-8")),
                auto_offset_reset="earliest",
                enable_auto_commit=True,
            )
            producer = kafka.KafkaProducer(
                bootstrap_servers=settings.kafka_bootstrap_servers,
                value_serializer=lambda value: json.dumps(value).encode("utf-8"),
            )
            return consumer, producer
        except Exception:
            if consumer is not None:
                consumer.close()
            if settings.kafka_connect_max_attempts and attempt >= settings.kafka_connect_max_attempts:
                logger.exception(
                    "failed to connect to kafka",
                    extra={
                        "bootstrap_servers": settings.kafka_bootstrap_servers,
                        "attempt": attempt,
                    },
                )
                raise
            logger.warning(
                "kafka is unavailable; retrying connection",
                extra={
                    "bootstrap_servers": settings.kafka_bootstrap_servers,
                    "attempt": attempt,
                    "retry_interval_seconds": settings.kafka_connect_retry_interval_seconds,
                },
                exc_info=True,
            )
            sleep(settings.kafka_connect_retry_interval_seconds)
