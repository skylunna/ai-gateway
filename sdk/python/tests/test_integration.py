"""Integration tests for the public luner API."""
import pytest
from unittest.mock import MagicMock

import luner
from luner.context import clear_context


def setup_function():
    clear_context()


def test_basic_flow():
    luner.init(
        gateway_url="http://localhost:8080",
        agent_name="test-agent",
        agent_version="v1.0",
    )

    ctx = luner.get_current_context()
    assert ctx is not None
    assert ctx.agent_name == "test-agent"
    assert ctx.agent_version == "v1.0"
    assert len(ctx.trace_id) == 32
    assert len(ctx.span_id) == 16


def test_set_user_after_init():
    luner.init(gateway_url="http://localhost:8080", agent_name="test")

    luner.set_user(user_id="alice", tenant_id="acme")

    ctx = luner.get_current_context()
    assert ctx.user_id == "alice"
    assert ctx.tenant_id == "acme"


def test_set_user_without_tenant():
    luner.init(gateway_url="http://localhost:8080", agent_name="test")
    luner.set_user("bob")

    ctx = luner.get_current_context()
    assert ctx.user_id == "bob"
    assert ctx.tenant_id is None


def test_set_session():
    luner.init(gateway_url="http://localhost:8080", agent_name="test")
    luner.set_session("sess-xyz")

    assert luner.get_current_context().session_id == "sess-xyz"


def test_tags():
    luner.init(gateway_url="http://localhost:8080", agent_name="test")

    luner.set_tag("version", "1.0")
    luner.set_tags({"env": "prod", "region": "us-west"})

    ctx = luner.get_current_context()
    assert ctx.tags["version"] == "1.0"
    assert ctx.tags["env"] == "prod"
    assert ctx.tags["region"] == "us-west"


def test_init_overwrites_previous_context():
    luner.init(gateway_url="http://localhost:8080", agent_name="agent-a")
    first_trace_id = luner.get_current_context().trace_id

    luner.init(gateway_url="http://localhost:8080", agent_name="agent-b")
    ctx = luner.get_current_context()

    assert ctx.agent_name == "agent-b"
    # new trace ID generated
    assert ctx.trace_id != first_trace_id


def test_set_user_no_context_is_noop():
    # Calling set_user before init should not raise
    luner.set_user("ghost")  # no context, should be safe


def test_set_tag_no_context_is_noop():
    luner.set_tag("k", "v")  # no context, should be safe


def test_wrap_requires_no_network():
    """wrap() should succeed without any network calls."""
    luner.init(gateway_url="http://localhost:8080", agent_name="test")

    # Simulate a minimal OpenAI-like client
    fake_client = MagicMock()
    fake_http = MagicMock()
    fake_http._luner_patched = False
    fake_http.send = MagicMock(return_value=MagicMock(status_code=200))

    import inspect
    # Make send a regular function (not coroutine)
    fake_http.send.__func__ = None

    fake_client._client = fake_http

    result = luner.wrap(fake_client)
    assert result is fake_client


def test_span_context_manager():
    luner.init(gateway_url="http://localhost:8080", agent_name="agent-x")
    parent_span_id = luner.get_current_context().span_id

    with luner.span("child-step") as child:
        assert child.parent_span_id == parent_span_id
        assert child.agent_name == "agent-x"
        # new span ID
        assert child.span_id != parent_span_id

    # Parent span restored after exiting
    assert luner.get_current_context().span_id == parent_span_id


def test_traced_decorator():
    luner.init(gateway_url="http://localhost:8080", agent_name="agent-d")
    captured = []

    @luner.traced("step-one")
    def my_step():
        ctx = luner.get_current_context()
        captured.append(ctx.span_id)

    outer_span_id = luner.get_current_context().span_id
    my_step()

    # Inside the decorated function, a new span_id was active
    assert len(captured) == 1
    assert captured[0] != outer_span_id
    # Restored after call
    assert luner.get_current_context().span_id == outer_span_id
