"""
Instrumentation for OpenAI SDK - automatic header injection.
"""
from __future__ import annotations

import inspect
from typing import Any

from .context import get_current_context
from .propagator import inject_headers


def wrap(client: Any) -> Any:
    """
    Wrap an OpenAI (or AsyncOpenAI) client to automatically inject luner headers.

    Intercepts the underlying httpx transport so every request carries the
    current TraceContext as HTTP headers, with zero changes to call-site code.

    Example:
        client = OpenAI(base_url="http://localhost:8080/v1")
        client = luner.wrap(client)
        client.chat.completions.create(...)  # headers injected automatically
    """
    _patch_httpx_client(client._client)
    return client


def wrap_async(client: Any) -> Any:
    """
    Wrap an AsyncOpenAI client.  Same approach as the sync version.
    """
    _patch_httpx_client(client._client)
    return client


# ── internals ────────────────────────────────────────────────────────────────


def _patch_httpx_client(http_client: Any) -> None:
    """
    Monkey-patch the httpx client's send method to inject luner headers.

    Works for both httpx.Client (sync) and httpx.AsyncClient (async) because
    we detect the coroutine at call time.
    """
    if getattr(http_client, "_luner_patched", False):
        return  # idempotent

    original_send = http_client.send

    if inspect.iscoroutinefunction(original_send):
        async def _send(request: Any, *args: Any, **kwargs: Any) -> Any:
            _inject(request)
            return await original_send(request, *args, **kwargs)
    else:
        def _send(request: Any, *args: Any, **kwargs: Any) -> Any:  # type: ignore[misc]
            _inject(request)
            return original_send(request, *args, **kwargs)

    http_client.send = _send
    http_client._luner_patched = True


def _inject(request: Any) -> None:
    """Inject the current context into an httpx.Request object."""
    ctx = get_current_context()
    if ctx is None:
        return

    new_headers = inject_headers(ctx, dict(request.headers))
    # httpx.Headers is immutable; replace via the MutableHeaders proxy
    for key, value in new_headers.items():
        request.headers[key] = value
