"""
LangChain integration example.

luner works alongside LangChain because it patches at the httpx transport
layer — the same underlying HTTP client used by both openai and langchain-openai.

Run (requires langchain-openai):
    pip install langchain-openai
    python examples/langchain_example.py
"""
import luner

try:
    from langchain_openai import ChatOpenAI
except ImportError:
    raise SystemExit("Install langchain-openai: pip install langchain-openai")


def main():
    luner.init(
        gateway_url="http://localhost:8080",
        agent_name="langchain-agent",
        agent_version="v1.0",
    )
    luner.set_user("charlie", tenant_id="demo-tenant")

    # LangChain wraps the OpenAI SDK internally; luner patches the same transport
    llm = ChatOpenAI(
        model="gpt-4o-mini",
        base_url="http://localhost:8080/v1",
        api_key="your-api-key",
    )
    luner.wrap(llm.client)  # patch the inner openai.OpenAI client

    response = llm.invoke("What is the capital of France?")
    print("Answer:", response.content)


if __name__ == "__main__":
    main()
