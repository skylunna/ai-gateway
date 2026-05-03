import pytest
from luner.context import TraceContext, create_context
from luner.propagator import extract_headers, inject_headers


def test_inject_full_context():
    ctx = TraceContext(
        trace_id="4bf92f3577b34da6a3ce929d0e0e4736",
        span_id="00f067aa0ba902b7",
        sampled=True,
        agent_name="code-reviewer",
        agent_version="v2.1.3",
        user_id="alice",
        tenant_id="acme",
        environment="production",
        session_id="sess-123",
        parent_span_id="aabbccdd11223344",
        tags={"team": "infra", "region": "us-east"},
    )

    headers = inject_headers(ctx, {})

    assert headers["traceparent"] == "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
    assert headers["X-Luner-Agent"] == "code-reviewer"
    assert headers["X-Luner-Agent-Version"] == "v2.1.3"
    assert headers["X-Luner-User"] == "alice"
    assert headers["X-Luner-Tenant"] == "acme"
    assert headers["X-Luner-Env"] == "production"
    assert headers["X-Luner-Session"] == "sess-123"
    assert headers["X-Luner-Parent-Span"] == "aabbccdd11223344"
    assert "team=infra" in headers["X-Luner-Tags"]
    assert "region=us-east" in headers["X-Luner-Tags"]


def test_inject_minimal_context():
    ctx = TraceContext(
        trace_id="aaaa" * 8,
        span_id="bbbb" * 4,
        sampled=False,
    )
    headers = inject_headers(ctx, {})

    # Only traceparent should be set (no optional fields)
    assert "traceparent" in headers
    assert headers["traceparent"].endswith("-00")  # not sampled
    assert "X-Luner-Agent" not in headers
    assert "X-Luner-User" not in headers
    assert "X-Luner-Tags" not in headers


def test_inject_does_not_mutate_input():
    ctx = TraceContext(trace_id="a" * 32, span_id="b" * 16)
    original = {"Authorization": "Bearer token"}
    result = inject_headers(ctx, original)

    assert "Authorization" in result  # preserved
    assert "Authorization" in original  # not mutated
    assert "traceparent" not in original  # not added to original


def test_extract_full_headers():
    headers = {
        "traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
        "X-Luner-Agent": "my-agent",
        "X-Luner-Agent-Version": "v1.0",
        "X-Luner-User": "bob",
        "X-Luner-Tenant": "tenant-x",
        "X-Luner-Env": "staging",
        "X-Luner-Session": "s-abc",
        "X-Luner-Parent-Span": "deadbeef01234567",
        "X-Luner-Tags": "k1=v1,k2=v2",
    }

    ctx = extract_headers(headers)
    assert ctx is not None
    assert ctx.trace_id == "4bf92f3577b34da6a3ce929d0e0e4736"
    assert ctx.span_id == "00f067aa0ba902b7"
    assert ctx.sampled is True
    assert ctx.agent_name == "my-agent"
    assert ctx.user_id == "bob"
    assert ctx.tenant_id == "tenant-x"
    assert ctx.environment == "staging"
    assert ctx.session_id == "s-abc"
    assert ctx.parent_span_id == "deadbeef01234567"
    assert ctx.tags["k1"] == "v1"
    assert ctx.tags["k2"] == "v2"


def test_extract_missing_traceparent():
    assert extract_headers({}) is None
    assert extract_headers({"X-Luner-Agent": "x"}) is None


def test_extract_malformed_traceparent():
    assert extract_headers({"traceparent": "bad"}) is None
    assert extract_headers({"traceparent": "00-a-b"}) is None


def test_roundtrip():
    ctx = create_context(
        agent_name="roundtrip-agent",
        user_id="user-rt",
        tags={"x": "y"},
    )
    headers = inject_headers(ctx, {})
    restored = extract_headers(headers)

    assert restored is not None
    assert restored.trace_id == ctx.trace_id
    assert restored.span_id == ctx.span_id
    assert restored.agent_name == ctx.agent_name
    assert restored.user_id == ctx.user_id
    assert restored.tags["x"] == "y"
