"""
luner - Python SDK for AI Control Plane

Usage:
    import luner
    from openai import OpenAI

    luner.init(gateway_url="http://localhost:8080", agent_name="my-agent")
    client = luner.wrap(OpenAI(base_url="http://localhost:8080/v1"))

    response = client.chat.completions.create(
        model="gpt-4o-mini",
        messages=[{"role": "user", "content": "Hello"}]
    )
"""
from __future__ import annotations

from typing import Optional

from .context import (
    TraceContext,
    clear_context,
    create_context,
    get_current_context,
    set_current_context,
)
from .instrumentation import wrap, wrap_async
from .spans import span, traced
from .version import __version__

# ── global configuration ──────────────────────────────────────────────────────

_config: dict = {
    "gateway_url": None,
    "agent_name": None,
    "agent_version": None,
    "environment": None,
}


def init(
    gateway_url: str,
    agent_name: str,
    agent_version: str = "default",
    environment: str = "production",
    user_id: Optional[str] = None,
    tenant_id: Optional[str] = None,
    session_id: Optional[str] = None,
    tags: Optional[dict] = None,
) -> None:
    """
    Initialize luner SDK with global configuration.

    Creates a trace context that is automatically injected into every
    subsequent LLM call made through a wrapped client.

    Args:
        gateway_url: luner Gateway base URL (e.g. "http://localhost:8080")
        agent_name: Name of your agent (e.g. "code-reviewer")
        agent_version: Version/variant identifier (e.g. "v2.1.3")
        environment: Deployment environment — production / staging / dev
        user_id: Optional end-user identifier
        tenant_id: Optional tenant identifier
        session_id: Optional session identifier
        tags: Optional free-form key-value tags
    """
    _config["gateway_url"] = gateway_url
    _config["agent_name"] = agent_name
    _config["agent_version"] = agent_version
    _config["environment"] = environment

    ctx = create_context(
        agent_name=agent_name,
        agent_version=agent_version,
        user_id=user_id,
        tenant_id=tenant_id,
        environment=environment,
        session_id=session_id,
        tags=tags,
    )
    set_current_context(ctx)


def set_user(user_id: str, tenant_id: Optional[str] = None) -> None:
    """
    Set user context for the current trace.

    Useful when user identity is resolved after init() — e.g. after an
    authentication step.

    Args:
        user_id: End-user identifier
        tenant_id: Optional tenant identifier
    """
    ctx = get_current_context()
    if ctx:
        ctx.user_id = user_id
        if tenant_id:
            ctx.tenant_id = tenant_id


def set_session(session_id: str) -> None:
    """Set session ID for the current trace."""
    ctx = get_current_context()
    if ctx:
        ctx.session_id = session_id


def set_tag(key: str, value: str) -> None:
    """Add a single tag to the current trace context."""
    ctx = get_current_context()
    if ctx:
        ctx.tags[key] = value


def set_tags(tags: dict) -> None:
    """Add multiple tags to the current trace context."""
    ctx = get_current_context()
    if ctx:
        ctx.tags.update(tags)


__all__ = [
    # Core API
    "init",
    "wrap",
    "wrap_async",
    # Span instrumentation (Mode 2)
    "span",
    "traced",
    # Context management
    "set_user",
    "set_session",
    "set_tag",
    "set_tags",
    "get_current_context",
    "clear_context",
    # Types
    "TraceContext",
    # Version
    "__version__",
]
