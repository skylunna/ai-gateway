"""
Span decorator for Mode 2 - explicit instrumentation of agent steps.

Usage:
    @luner.span("my-tool-call")
    def call_tool(query: str) -> str:
        ...
"""
from __future__ import annotations

import functools
from contextlib import contextmanager
from typing import Any, Callable, Generator, Optional

from .context import (
    TraceContext,
    create_context,
    get_current_context,
    set_current_context,
)


@contextmanager
def span(
    name: str,
    tags: Optional[dict] = None,
) -> Generator[TraceContext, None, None]:
    """
    Context manager that creates a child span for a block of work.

    The new span inherits agent/user context from the parent and sets
    parent_span_id so Gateway can reconstruct the call tree.

    Example:
        with luner.span("retrieval") as s:
            s.tags["query"] = query
            results = vector_search(query)
    """
    parent = get_current_context()

    child = create_context(
        agent_name=parent.agent_name if parent else None,
        agent_version=parent.agent_version if parent else None,
        user_id=parent.user_id if parent else None,
        tenant_id=parent.tenant_id if parent else None,
        environment=parent.environment if parent else None,
        session_id=parent.session_id if parent else None,
        tags=dict(parent.tags) if parent else {},
    )
    if parent:
        child.parent_span_id = parent.span_id
    if tags:
        child.tags.update(tags)

    token = set_current_context(child)
    try:
        yield child
    finally:
        # Restore parent context so sibling spans see the original span_id
        set_current_context(parent)  # type: ignore[arg-type]


def traced(name: Optional[str] = None) -> Callable:
    """
    Decorator that wraps a function in a luner span.

    Example:
        @luner.traced("search-step")
        def search(query: str) -> list:
            ...
    """
    def decorator(fn: Callable) -> Callable:
        span_name = name or fn.__qualname__

        @functools.wraps(fn)
        def wrapper(*args: Any, **kwargs: Any) -> Any:
            with span(span_name):
                return fn(*args, **kwargs)

        @functools.wraps(fn)
        async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
            with span(span_name):
                return await fn(*args, **kwargs)

        import inspect
        return async_wrapper if inspect.iscoroutinefunction(fn) else wrapper

    return decorator
