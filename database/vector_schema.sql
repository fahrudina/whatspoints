-- pgvector schema for the AI RAG assistant (whatspoints-ai-agent).
-- Apply with: psql "$DATABASE_URL" -f database/vector_schema.sql
-- On Supabase you can also paste this into the SQL editor.

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS knowledge_base (
    id BIGSERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    category TEXT,
    embedding VECTOR(1536), -- Google gemini-embedding-001 (output_dimensionality=1536)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Approximate nearest-neighbour index for cosine similarity search.
-- HNSW (not ivfflat): needs no training data, handles incremental inserts, and
-- can be created up front on an empty table. vector_cosine_ops matches the
-- cosine distance operator (<=>) used by the retrieval queries.
CREATE INDEX IF NOT EXISTS knowledge_base_embedding_idx
ON knowledge_base
USING hnsw (embedding vector_cosine_ops);

-- Optional seed data, inserted only when the table is empty so reruns don't
-- accumulate duplicates. Embeddings are NULL until you run index_knowledge.py.
-- (A unique constraint isn't used: document chunks legitimately share a title.)
INSERT INTO knowledge_base (title, content, category)
SELECT v.title, v.content, v.category
FROM (VALUES
    ('Jam Operasional', 'Ruang Laundry buka setiap hari jam 08:00 sampai 20:00 WIB.', 'business_info'),
    ('Promo Laundry', 'Promo Cuci 8KG Rp10.000 berlaku Senin sampai Rabu.', 'promo'),
    ('Layanan Laundry', 'Ruang Laundry menyediakan laundry kiloan, satuan, antar jemput, dan mesin pengering.', 'service')
) AS v(title, content, category)
WHERE NOT EXISTS (SELECT 1 FROM knowledge_base);
