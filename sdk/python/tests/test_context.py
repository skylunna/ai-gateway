import pytest
from luner.context import (
    TraceContext,
    clear_context,
    create_context,
    generate_span_id,
    generate_trace_id,
    get_current_context,
    set_current_context,
)


def setup_function():
    clear_context()


def test_generate_ids():
    trace_id = generate_trace_id()
    assert len(trace_id) == 32
    assert all(c in "0123456789abcdef" for c in trace_id)

    span_id = generate_span_id()
    assert len(span_id) == 16
    assert all(c in "0123456789abcdef" for c in span_id)


def test_generate_ids_unique():
    ids = {generate_trace_id() for _ in range(100)}
    assert len(ids) == 100  # no collisions


def test_create_context():
    ctx = create_context(
        agent_name="test-agent",
        agent_version="v1.0",
        user_id="user-123",
    )

    assert ctx.agent_name == "test-agent"
    assert ctx.agent_version == "v1.0"
    assert ctx.user_id == "user-123"
    assert len(ctx.trace_id) == 32
    assert len(ctx.span_id) == 16
    assert ctx.tags == {}
    assert ctx.sampled is True


def test_create_context_with_tags():
    ctx = create_context(tags={"env": "prod", "region": "us-east"})
    assert ctx.tags["env"] == "prod"
    assert ctx.tags["region"] == "us-east"


def test_context_propagation():
    assert get_current_context() is None

    ctx = create_context(agent_name="test")
    set_current_context(ctx)

    retrieved = get_current_context()
    assert retrieved is not None
    assert retrieved.agent_name == "test"
    assert retrieved.trace_id == ctx.trace_id

    clear_context()
    assert get_current_context() is None


def test_context_mutation():
    ctx = create_context(agent_name="test")
    set_current_context(ctx)

    current = get_current_context()
    current.user_id = "alice"
    current.tags["env"] = "prod"

    retrieved = get_current_context()
    assert retrieved.user_id == "alice"
    assert retrieved.tags["env"] == "prod"


def test_context_defaults():
    ctx = create_context()
    assert ctx.parent_span_id is None
    assert ctx.session_id is None
    assert ctx.user_id is None
    assert ctx.tenant_id is None
    assert ctx.environment is None
