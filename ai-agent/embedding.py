"""Embedding generation and pgvector similarity search.

Retrieval queries pgvector on every call, so newly indexed rows are picked up
without restarting the service.
"""
import os

import psycopg
from langchain_google_genai import GoogleGenerativeAIEmbeddings

# Embeddings run on Google AI Studio (Gemini). gemini-embedding-001 defaults to
# 3072 dims; we request 1536 so vectors match the VECTOR(1536) schema. Keep
# EMBED_DIM in sync with the schema if you change the model.
EMBED_MODEL = os.getenv("EMBEDDING_MODEL", "models/gemini-embedding-001")
EMBED_DIM = int(os.getenv("EMBEDDING_DIM", "1536"))
# Wider recall so multi-chunk answers (e.g. a pricelist split across rows) come
# back whole instead of just the top promo chunk. The relevance gate in graph.py
# still drops anything too far, so extra slots only help when context exists.
TOP_K = int(os.getenv("RETRIEVAL_TOP_K", "8"))

_embeddings = None


def _client() -> GoogleGenerativeAIEmbeddings:
    # Lazy init so importing this module (e.g. for /health) needs no API key.
    global _embeddings
    if _embeddings is None:
        _embeddings = GoogleGenerativeAIEmbeddings(
            model=EMBED_MODEL,
            google_api_key=os.getenv("GOOGLE_API_KEY"),
        )
    return _embeddings


def embed_text(text: str) -> list[float]:
    # Cosine distance (pgvector <=>) is scale-invariant, so the un-normalized
    # output Gemini returns below 3072 dims is fine here.
    return _client().embed_query(text, output_dimensionality=EMBED_DIM)


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
    # Request-path query: fail fast instead of hanging on DB/network stalls.
    with psycopg.connect(
        os.environ["DATABASE_URL"],
        connect_timeout=5,
        options="-c statement_timeout=5000",
    ) as conn:
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
