"""Embedding generation and pgvector similarity search.

Retrieval queries pgvector on every call, so newly indexed rows are picked up
without restarting the service.
"""
import os

import psycopg
from langchain_openai import OpenAIEmbeddings

EMBED_MODEL = "text-embedding-3-small"  # 1536 dims, matches VECTOR(1536)
TOP_K = 3

_embeddings = None


def _client() -> OpenAIEmbeddings:
    # Lazy init so importing this module doesn't require an API key.
    # Embeddings use real OpenAI (OpenRouter has no embeddings endpoint); keep
    # this key separate from the OpenRouter chat key. EMBEDDINGS_BASE_URL lets you
    # point at another OpenAI-compatible embeddings provider if needed.
    global _embeddings
    if _embeddings is None:
        _embeddings = OpenAIEmbeddings(
            model=EMBED_MODEL,
            base_url=os.getenv("EMBEDDINGS_BASE_URL") or None,
            api_key=os.getenv("EMBEDDINGS_API_KEY") or os.getenv("OPENAI_API_KEY"),
        )
    return _embeddings


def embed_text(text: str) -> list[float]:
    return _client().embed_query(text)


def vector_to_pgvector(vector: list[float]) -> str:
    """Render a Python float list as a pgvector literal: '[0.1,0.2,0.3]'."""
    return "[" + ",".join(str(x) for x in vector) + "]"


def search_knowledge(query: str, top_k: int = TOP_K) -> list[dict]:
    """Return the top_k most similar knowledge_base rows for `query`.

    `score` is cosine distance (lower = more similar). Rows without an embedding
    are ignored.
    """
    embedding = vector_to_pgvector(embed_text(query))
    sql = """
        SELECT id, title, content, category,
               embedding <=> %s::vector AS score
        FROM knowledge_base
        WHERE embedding IS NOT NULL
        ORDER BY embedding <=> %s::vector
        LIMIT %s
    """
    with psycopg.connect(os.environ["DATABASE_URL"]) as conn:
        rows = conn.execute(sql, (embedding, embedding, top_k)).fetchall()

    return [
        {
            "id": r[0],
            "title": r[1],
            "content": r[2],
            "category": r[3],
            "score": float(r[4]),
        }
        for r in rows
    ]
