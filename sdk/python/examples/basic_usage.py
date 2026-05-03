"""
Basic usage example - Mode 1 (minimal SDK integration).

Run:
    python examples/basic_usage.py

Requirements:
    - luner Gateway running at localhost:8080
    - pip install -e .
"""
import luner
from openai import OpenAI


def main():
    # 1. Initialize luner SDK (one-time setup)
    luner.init(
        gateway_url="http://localhost:8080",
        agent_name="code-reviewer",
        agent_version="v2.1.3",
        environment="production",
    )

    # 2. Wrap OpenAI client — all subsequent calls carry trace headers
    client = OpenAI(
        base_url="http://localhost:8080/v1",
        api_key="your-api-key",
    )
    client = luner.wrap(client)

    # 3. Set user context (optional; can be done per-request)
    luner.set_user(user_id="alice", tenant_id="acme-corp")

    # 4. Use OpenAI as normal
    response = client.chat.completions.create(
        model="gpt-4o-mini",
        messages=[
            {"role": "user", "content": "Review this Python code: def foo(): pass"}
        ],
    )

    print(f"Response: {response.choices[0].message.content}")
    print()

    ctx = luner.get_current_context()
    print("Headers sent by luner:")
    print(f"  traceparent: 00-{ctx.trace_id}-{ctx.span_id}-01")
    print(f"  X-Luner-Agent: {ctx.agent_name}")
    print(f"  X-Luner-User: {ctx.user_id}")
    print(f"  X-Luner-Tenant: {ctx.tenant_id}")

    # Gateway automatically recorded:
    # - agent_name: code-reviewer
    # - agent_version: v2.1.3
    # - user_id: alice
    # - tenant_id: acme-corp
    # - trace_id, cost, tokens, latency, etc.


if __name__ == "__main__":
    main()
