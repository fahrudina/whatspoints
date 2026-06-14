-- pgvector schema for the AI RAG assistant (whatspoints-ai-agent).
-- Apply with: psql "$DATABASE_URL" -f database/vector_schema.sql
-- On Supabase you can also paste this into the SQL editor.

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS knowledge_base (
    id BIGSERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    category TEXT,
    embedding VECTOR(1536), -- OpenAI text-embedding-3-small
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Approximate nearest-neighbour index for cosine similarity search.
CREATE INDEX IF NOT EXISTS knowledge_base_embedding_idx
ON knowledge_base
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Optional seed data. Embeddings are NULL until you run index_knowledge.py.
INSERT INTO knowledge_base (title, content, category)
VALUES
('Jam Operasional', 'Ruang Laundry buka setiap hari jam 08:00 sampai 20:00 WIB.', 'business_info'),
('Promo Laundry', 'Promo Cuci 8KG Rp10.000 berlaku Senin sampai Rabu.', 'promo'),
('Layanan Laundry', 'Ruang Laundry menyediakan laundry kiloan, satuan, antar jemput, dan mesin pengering.', 'service')
ON CONFLICT DO NOTHING;
