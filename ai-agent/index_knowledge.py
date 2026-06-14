"""Backfill embeddings for knowledge_base rows that don't have one yet.

Safe to rerun: only touches rows where embedding IS NULL.

    cd ai-agent
    python index_knowledge.py
"""
import os

import psycopg
from dotenv import load_dotenv

from embedding import embed_text, vector_to_pgvector

load_dotenv()


def main() -> None:
    with psycopg.connect(os.environ["DATABASE_URL"]) as conn:
        rows = conn.execute(
            "SELECT id, title, content FROM knowledge_base "
            "WHERE embedding IS NULL ORDER BY id"
        ).fetchall()

        print(f"Found {len(rows)} documents without embeddings")
        for row_id, title, content in rows:
            print(f"Indexing ID {row_id}: {title}")
            vector = vector_to_pgvector(embed_text(content))
            conn.execute(
                "UPDATE knowledge_base "
                "SET embedding = %s::vector, updated_at = CURRENT_TIMESTAMP "
                "WHERE id = %s",
                (vector, row_id),
            )
            # Commit per row so a late failure keeps earlier progress.
            conn.commit()

    print("Embedding indexing completed")


if __name__ == "__main__":
    main()
