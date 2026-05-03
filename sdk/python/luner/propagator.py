"""
HTTP header propagation following W3C Trace Context + luner extensions.
"""
from __future__ import annotations

from typing import Optional

from .context import TraceContext


def inject_headers(ctx: TraceContext, headers: dict) -> dict:
    """
    Inject trace context into HTTP headers.

    Follows W3C Trace Context standard:
        traceparent: 00-{trace_id}-{span_id}-{flags}

    Plus luner custom headers:
        X-Luner-Agent: agent_name
        X-Luner-User: user_id
        etc.
    """
    headers = dict(headers)  # copy to avoid mutation

    # W3C traceparent
    flags = "01" if ctx.sampled else "00"
    headers["traceparent"] = f"00-{ctx.trace_id}-{ctx.span_id}-{flags}"

    # Luner custom headers (only if set)
    if ctx.agent_name:
        headers["X-Luner-Agent"] = ctx.agent_name
    if ctx.agent_version:
        headers["X-Luner-Agent-Version"] = ctx.agent_version
    if ctx.session_id:
        headers["X-Luner-Session"] = ctx.session_id
    if ctx.user_id:
        headers["X-Luner-User"] = ctx.user_id
    if ctx.tenant_id:
        headers["X-Luner-Tenant"] = ctx.tenant_id
    if ctx.environment:
        headers["X-Luner-Env"] = ctx.environment
    if ctx.parent_span_id:
        headers["X-Luner-Parent-Span"] = ctx.parent_span_id

    # Tags as CSV: k1=v1,k2=v2
    if ctx.tags:
        tags_str = ",".join(f"{k}={v}" for k, v in ctx.tags.items())
        headers["X-Luner-Tags"] = tags_str

    return headers


def extract_headers(headers: dict) -> Optional[TraceContext]:
    """
    Extract trace context from HTTP headers.

    Used for distributed tracing across service boundaries.
    """
    traceparent = headers.get("traceparent", "")
    if not traceparent:
        return None

    parts = traceparent.split("-")
    if len(parts) != 4:
        return None

    version, trace_id, span_id, flags = parts
    if version != "00":
        return None

    sampled = flags == "01"

    ctx = TraceContext(
        trace_id=trace_id,
        span_id=span_id,
        sampled=sampled,
        agent_name=headers.get("X-Luner-Agent"),
        agent_version=headers.get("X-Luner-Agent-Version"),
        session_id=headers.get("X-Luner-Session"),
        user_id=headers.get("X-Luner-User"),
        tenant_id=headers.get("X-Luner-Tenant"),
        environment=headers.get("X-Luner-Env"),
        parent_span_id=headers.get("X-Luner-Parent-Span"),
    )

    tags_str = headers.get("X-Luner-Tags", "")
    if tags_str:
        for pair in tags_str.split(","):
            if "=" in pair:
                k, v = pair.split("=", 1)
                ctx.tags[k.strip()] = v.strip()

    return ctx
