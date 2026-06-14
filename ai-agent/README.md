# whatspoints-ai-agent

AI RAG sidecar for whatspoints. Generates **suggested** WhatsApp replies (Bahasa
Indonesia, casual admin tone) from a pgvector knowledge base. It does **not**
auto-send anything — the Go service calls it over HTTP and returns the suggestion.

## Stack

FastAPI · LangGraph · OpenAI embeddings (`text-embedding-3-small`) · PostgreSQL + pgvector

## Setup

```bash
cd ai-agent
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env   # then fill DATABASE_URL and OPENAI_API_KEY
```

Environment variables:

| Var                  | Default                         | Purpose                                         |
|----------------------|---------------------------------|-------------------------------------------------|
| `DATABASE_URL`       | —                               | Postgres connection string                      |
| `OPENROUTER_API_KEY` | —                               | OpenRouter key for the chat LLM                 |
| `LLM_BASE_URL`       | `https://openrouter.ai/api/v1`  | Chat LLM base URL (OpenAI-compatible)           |
| `AI_MODEL`           | `openai/gpt-4o-mini`            | Chat model (OpenRouter-style name)              |
| `OPENAI_API_KEY`     | —                               | **OpenAI** key for embeddings (see note)        |
| `AI_HOST`            | `0.0.0.0`                       | Bind host                                       |
| `AI_PORT`            | `8090`                          | Bind port                                       |

> **Embeddings note:** OpenRouter has no embeddings endpoint, so
> `text-embedding-3-small` runs through real OpenAI (`OPENAI_API_KEY`). To use a
> different OpenAI-compatible embeddings provider, set `EMBEDDINGS_BASE_URL` and
> `EMBEDDINGS_API_KEY`. The chat LLM and embeddings keys are independent.

## Database

Apply the pgvector schema once (from repo root):

```bash
psql "$DATABASE_URL" -f database/vector_schema.sql
```

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
