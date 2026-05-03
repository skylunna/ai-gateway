"""Tests for OpenAI client wrapping and header injection."""
import pytest
from unittest.mock import MagicMock, patch
import httpx

import luner
from luner.context import clear_context, create_context, set_current_context, TraceContext
from luner.instrumentation import wrap, _patch_httpx_client


def setup_function():
    clear_context()


class FakeHTTPClient:
    """Minimal stand-in for httpx.Client."""

    def __init__(self):
        self._luner_patched = False
        self.last_request = None

    def send(self, request, *args, **kwargs):
        self.last_request = request
        return MagicMock(status_code=200)


class FakeOpenAIClient:
    """Minimal stand-in for openai.OpenAI."""

    def __init__(self):
        self._client = FakeHTTPClient()


def make_request(headers: dict | None = None) -> MagicMock:
    """Build a fake httpx.Request with mutable headers."""
    req = MagicMock(spec=httpx.Request)
    req.headers = dict(headers or {})
    return req


def test_wrap_returns_same_client():
    fake = FakeOpenAIClient()
    result = wrap(fake)
    assert result is fake


def test_wrap_is_idempotent():
    fake = FakeOpenAIClient()
    wrap(fake)
    original_send = fake._client.send
    wrap(fake)
    assert fake._client.send is original_send  # not re-patched


def test_headers_injected_when_context_set():
    ctx = create_context(agent_name="test-agent", user_id="alice")
    set_current_context(ctx)

    fake_http = FakeHTTPClient()
    _patch_httpx_client(fake_http)

    req = make_request({"Authorization": "Bearer key"})
    fake_http.send(req)

    captured = fake_http.last_request
    assert captured.headers.get("traceparent", "").startswith("00-")
    assert captured.headers.get("X-Luner-Agent") == "test-agent"
    assert captured.headers.get("X-Luner-User") == "alice"
    assert captured.headers.get("Authorization") == "Bearer key"  # preserved


def test_no_headers_without_context():
    fake_http = FakeHTTPClient()
    _patch_httpx_client(fake_http)

    req = make_request({"Authorization": "Bearer key"})
    fake_http.send(req)

    captured = fake_http.last_request
    assert "traceparent" not in captured.headers
    assert "X-Luner-Agent" not in captured.headers


def test_patch_sets_flag():
    fake_http = FakeHTTPClient()
    assert not fake_http._luner_patched
    _patch_httpx_client(fake_http)
    assert fake_http._luner_patched


def test_wrap_full_flow():
    luner.init(
        gateway_url="http://localhost:8080",
        agent_name="wrap-test",
        agent_version="v0.1",
    )

    fake = FakeOpenAIClient()
    wrapped = luner.wrap(fake)

    req = make_request()
    wrapped._client.send(req)

    assert req.headers.get("X-Luner-Agent") == "wrap-test"
    assert req.headers.get("X-Luner-Agent-Version") == "v0.1"
