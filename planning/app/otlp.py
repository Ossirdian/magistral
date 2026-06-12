from __future__ import annotations

from urllib.parse import urlparse


def otlp_http_endpoint(endpoint: str) -> str:
    endpoint = endpoint.strip()
    if not endpoint:
        return endpoint
    parsed = urlparse(endpoint)
    if parsed.scheme in {"http", "https"}:
        return endpoint
    return f"http://{endpoint.rstrip('/')}/v1/traces"
