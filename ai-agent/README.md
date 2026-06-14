# whatspoints-ai-agent

AI RAG sidecar for whatspoints. Generates **suggested** WhatsApp replies (Bahasa
Indonesia, casual admin tone) from a pgvector knowledge base. It does **not**
auto-send anything — the Go service calls it over HTTP and returns the suggestion.

## Stack

FastAPI · LangGraph · Google Gemini embeddings (`gemini-embedding-001`) · PostgreSQL + pgvector

## Setup

```bash
cd ai-agent
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env   # then fill DATABASE_URL, OPENROUTER_API_KEY, GOOGLE_API_KEY
```

Environment variables:

| Var                  | Default                          | Purpose                                         |
|----------------------|----------------------------------|-------------------------------------------------|
| `DATABASE_URL`       | —                                | Postgres connection string                      |
| `OPENROUTER_API_KEY` | —                                | OpenRouter key for the chat LLM                 |
| `LLM_BASE_URL`       | `https://openrouter.ai/api/v1`   | Chat LLM base URL (OpenAI-compatible)           |
| `AI_MODEL`           | `openai/gpt-4o-mini`             | Chat model (OpenRouter-style name)              |
| `GOOGLE_API_KEY`     | —                                | Google AI Studio key for embeddings             |
| `EMBEDDING_MODEL`    | `models/gemini-embedding-001`    | Gemini embedding model                          |
| `EMBEDDING_DIM`      | `1536`                           | Output dims; must match `VECTOR(...)` in schema |
| `AI_HOST`            | `0.0.0.0`                        | Bind host                                       |
| `AI_PORT`            | `8090`                           | Bind port                                       |

> **Embeddings note:** the chat LLM runs through OpenRouter, but embeddings run on
> **Google AI Studio** (`gemini-embedding-001`). It defaults to 3072 dims; we
> request `EMBEDDING_DIM=1536` to match the `VECTOR(1536)` schema. If you change
> the model or dims, update `database/vector_schema.sql` and re-index. The chat and
> embedding providers/keys are independent.
>
> **Switching embedding providers invalidates existing vectors.** Reset and
> re-index: `UPDATE knowledge_base SET embedding = NULL;` then `python index_knowledge.py`.

## Database

Apply the pgvector schema once (from repo root):

```bash
psql "$DATABASE_URL" -f database/vector_schema.sql
```

> **Supabase on an IPv4 network:** use the **connection pooler** host, not the
> direct `db.<ref>.supabase.co` host. The direct host is IPv6-only, so on
> IPv4-only machines psycopg fails with `connection is bad: no error details
> available`. Use the pooler instead (session mode, port 5432):
>
> ```
> postgresql://postgres.<ref>:<password>@aws-0-<region>.pooler.supabase.com:5432/postgres?sslmode=require
> ```
>
> Copy it from Dashboard → Project Settings → Database → Connection pooler →
> Session. (Transaction mode on port 6543 also works, but with psycopg3 you must
> disable prepared statements via `prepare_threshold=None`; session mode avoids
> that.)

## Run

```bash
uvicorn app:api --host 0.0.0.0 --port 8090
```

Health check:

```bash
curl http://localhost:8090/health
# {"status":"ok","service":"whatspoints-ai-agent"}
```

## Indexing knowledge

New rows in `knowledge_base` start with `embedding = NULL`. Generate embeddings:

```bash
python index_knowledge.py
```

Safe to rerun — it only touches rows where `embedding IS NULL`. **No service
restart is needed** after inserting + indexing new data; retrieval queries
pgvector on every request.

## Endpoints

- `GET /health` — liveness.
- `POST /ai/reply` — generate a suggested reply.

### Intent gate

Only laundry-related intents (`ask_promo`, `ask_opening_hours`, `ask_service`,
`complaint`, `ask_order_status`) get a reply. Anything classified as `unknown`
is skipped so the bot doesn't answer unrelated chatter. The response always
includes `should_reply`; when it is `false`, `reply` is empty and the caller
should not respond:

```json
{ "reply": "", "intent": "unknown", "should_reply": false, "sources": [] }
```
