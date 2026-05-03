"""
Context management using contextvars for async-safe context propagation.
"""
from __future__ import annotations

import contextvars
import secrets
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class TraceContext:
    """Agent trace context that gets propagated via HTTP headers."""

    trace_id: str
    span_id: str
    parent_span_id: Optional[str] = None
    session_id: Optional[str] = None

    # Business context
    agent_name: Optional[str] = None
    agent_version: Optional[str] = None
    user_id: Optional[str] = None
    tenant_id: Optional[str] = None
    environment: Optional[str] = None

    tags: dict = field(default_factory=dict)
    sampled: bool = True


# Context variable for storing current trace context
_current_context: contextvars.ContextVar[Optional[TraceContext]] = contextvars.ContextVar(
    "luner_context", default=None
)


def get_current_context() -> Optional[TraceContext]:
    """Get the current trace context from contextvars."""
    return _current_context.get()


def set_current_context(ctx: TraceContext) -> contextvars.Token:
    """Set the current trace context."""
    return _current_context.set(ctx)


def clear_context() -> None:
    """Clear the current trace context."""
    _current_context.set(None)


def generate_trace_id() -> str:
    """Generate a random trace ID (32 hex chars)."""
    return secrets.token_hex(16)


def generate_span_id() -> str:
    """Generate a random span ID (16 hex chars)."""
    return secrets.token_hex(8)


def create_context(
    agent_name: Optional[str] = None,
    agent_version: Optional[str] = None,
    user_id: Optional[str] = None,
    tenant_id: Optional[str] = None,
    environment: Optional[str] = None,
    session_id: Optional[str] = None,
    tags: Optional[dict] = None,
) -> TraceContext:
    """Create a new trace context with generated IDs."""
    return TraceContext(
        trace_id=generate_trace_id(),
        span_id=generate_span_id(),
        agent_name=agent_name,
        agent_version=agent_version,
        user_id=user_id,
        tenant_id=tenant_id,
        environment=environment,
        session_id=session_id,
        tags=tags or {},
    )
