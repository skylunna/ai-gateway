"""
Mode 2 example - explicit span instrumentation for multi-step agents.

Demonstrates how to use luner.span() and @luner.traced() to create
a parent → child span hierarchy visible in the Gateway trace view.
"""
import luner
from openai import OpenAI


def search_knowledge_base(query: str) -> list[str]:
    """Simulated vector search."""
    return [f"doc about {query}", "related result"]


@luner.traced("summarise-docs")
def summarise(client: OpenAI, docs: list[str]) -> str:
    joined = "\n".join(docs)
    response = client.chat.completions.create(
        model="gpt-4o-mini",
        messages=[
            {"role": "system", "content": "Summarize the following documents."},
            {"role": "user", "content": joined},
        ],
    )
    return response.choices[0].message.content


def rag_pipeline(query: str) -> str:
    """A minimal RAG pipeline with explicit span tracking."""
    luner.init(
        gateway_url="http://localhost:8080",
        agent_name="rag-agent",
        agent_version="v1.0",
        environment="production",
    )

    client = luner.wrap(OpenAI(base_url="http://localhost:8080/v1", api_key="key"))

    # Step 1: retrieval span
    with luner.span("retrieval", tags={"query_len": str(len(query))}) as s:
        docs = search_knowledge_base(query)
        s.tags["docs_found"] = str(len(docs))

    # Step 2: summarisation span (via decorator)
    summary = summarise(client, docs)

    return summary


if __name__ == "__main__":
    result = rag_pipeline("AI observability best practices")
    print("Summary:", result)
