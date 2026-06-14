# AI Reply Suggestion — High-Level Design

A high-level overview of the optional AI RAG assistant added to WhatsPoints. For
hands-on setup and commands, see [`ai-agent/README.md`](ai-agent/README.md) and
the "AI Reply Suggestion" section of the main [`README.md`](README.md).

## 1. Purpose

Generate **suggested** WhatsApp replies (Bahasa Indonesia, casual admin tone)
grounded in a business knowledge base, so a human admin can answer customers
faster.

**In scope (this phase):** produce reply suggestions on demand via an API.
**Explicitly out of scope:** auto-sending replies. The feature only *suggests* —
`ENABLE_AI_AUTO_SEND` is a reserved flag with no behavior yet.

## 2. Design goals

- **Optional & non-invasive.** Off by default; the existing WhatsApp API
  (`/api/send-message`, `/api/status`, `/api/senders`, `/health`) is unaffected
  whether AI is on or off.
- **Isolated.** All AI logic lives in a separate Python sidecar (`ai-agent/`).
  The Go service stays a thin, clean-architecture gateway.
- **Grounded, not chatty.** Answers come only from retrieved knowledge; the bot
  stays silent on unrelated messages (intent gate) and never invents promos,
  prices, branches, or order status.
- **No restart to learn.** New knowledge is searchable as soon as it's embedded
  and stored — retrieval queries pgvector on every request.

## 3. Architecture

```text
                         Basic Auth
   Client ──HTTP──▶  Go WhatsApp API (Gin)  ──HTTP──▶  Python AI sidecar (FastAPI)
                     POST /api/ai/reply               POST /ai/reply
                          │                                 │
                          │                                 ├─▶ OpenRouter  (chat LLM)
                          │                                 ├─▶ Google AI Studio (Gemini embeddings)
                          │                                 └─▶ PostgreSQL + pgvector (knowledge_base)
                          │                                        ▲
                          └────────── shares the same DB ──────────┘
```

- The **Go service** owns auth, routing, and the feature toggle. It calls the
  sidecar over HTTP and returns the suggestion; it never blocks on AI for its
  core messaging endpoints.
- The **Python sidecar** owns the RAG workflow (intent detection, retrieval,
  answer generation) and all LLM/DB access for AI.
- Both connect to the **same Postgres** (the sidecar uses pgvector).

## 4. Components

### Go (clean architecture, module `github.com/wa-serv`)

| Layer | File | Responsibility |
|-------|------|----------------|
| Config | `config/env.go` (`AIConfig`, `LoadAIConfig`) | Parse `ENABLE_AI_RESPONSE` / `ENABLE_AI_AUTO_SEND` / `AI_SERVICE_URL` |
| Domain | `internal/domain/{entities,interfaces}.go` | `AIReplyRequest/Response/Source`, `AIService` + `AIClient` interfaces, errors |
| Infrastructure | `internal/infrastructure/ai_client.go` | HTTP client to the sidecar (15s timeout, non-2xx → error, no sensitive logging) |
| Application | `internal/application/ai_service.go` | Validate input, enforce the enabled/disabled rule, call the client |
| Presentation | `internal/presentation/ai_handler.go` + `router.go` | `POST /api/ai/reply` under Basic Auth; 503 when disabled, 400 on empty message |

### Python sidecar (`ai-agent/`)

| File | Responsibility |
|------|----------------|
| `app.py` | FastAPI app: `GET /health`, `POST /ai/reply` |
| `graph.py` | LangGraph RAG workflow + intent gate; chat via OpenRouter |
| `embedding.py` | Gemini embeddings + pgvector cosine search |
| `index_knowledge.py` | Backfill embeddings for rows where `embedding IS NULL` (rerun-safe) |
| `ingest_document.py` | Chunk a `.txt/.md/.pdf` into `knowledge_base` rows, then embed |

## 5. Request flow — `POST /api/ai/reply`

```text
Client → Go AIHandler   (inbound body: {message, phone_number})
  ├─ AI disabled?              → 503 {"success":false,"message":"AI response feature is disabled"}
  ├─ empty "message"?          → 400 {"success":false,"message":"message is required"}
  └─ enabled →
       AIService → AIClient → POST {AI_SERVICE_URL}/ai/reply  {customer_message, phone_number}
                              (the client maps the inbound "message" → "customer_message")
                                    └─ sidecar runs the RAG graph
       ← 200 {reply, intent, should_reply, sources[]}
```

The route is **always registered**, even when disabled, so clients get a clear
503 instead of a 404.

## 6. RAG workflow (LangGraph)

```text
START → detect_intent → route_after_intent
          ├─ skip   (intent = unknown / non-laundry)  ─────────────▶ END   should_reply=false, reply=""
          └─ retrieve_context → route_after_retrieval
                                  ├─ generate_answer (context found) ─▶ END   should_reply=true
                                  └─ fallback        (no context)    ─▶ END   polite "ask for detail"
```

- **Intent gate.** `detect_intent` (LLM) classifies the message into
  `ask_promo`, `ask_opening_hours`, `ask_service`, `complaint`,
  `ask_order_status`, or `unknown`. Only the laundry-related intents proceed;
  `unknown` is **skipped** so the bot doesn't reply to unrelated chatter. The
  response always carries `should_reply` (false ⇒ caller should not respond).
- **Retrieval.** `retrieve_context` embeds the message and runs a pgvector
  cosine search (top 3), every request — newly added knowledge is picked up
  without a restart.
- **Answering.** `generate_answer` replies in Bahasa Indonesia using only the
  retrieved context. If nothing relevant is found, `fallback` politely asks for
  more detail. Complaints and order-status questions stay conservative.

## 7. Data model

`knowledge_base` (see [`database/vector_schema.sql`](database/vector_schema.sql)):

| Column | Notes |
|--------|-------|
| `id` | `BIGSERIAL` PK |
| `title` | source label (manual title, or the document file name) |
| `content` | the text chunk (one row = one chunk) |
| `category` | free-text tag (`promo`, `service`, `business_info`, …) |
| `embedding` | `VECTOR(1536)` — Gemini `gemini-embedding-001` @ 1536 dims |
| `created_at` / `updated_at` | timestamps |

Index: **HNSW** with `vector_cosine_ops` (matches the `<=>` cosine queries).
HNSW is used over IVFFlat because it needs no training data, handles incremental
inserts, and can be built up front.

## 8. Knowledge lifecycle

```text
add knowledge ──▶ embedding = NULL ──▶ embed ──▶ searchable (no restart)
```

Three ways to add knowledge, all converging on the same embed step:
1. **Manual SQL** `INSERT` into `knowledge_base`, then `python index_knowledge.py`.
2. **Document ingest** `python ingest_document.py file.pdf --category service`
   (chunks → rows → embeds in one go).
3. **Seed data** shipped in `vector_schema.sql`.

Switching the embedding model/provider invalidates existing vectors — reset with
`UPDATE knowledge_base SET embedding = NULL;` and re-index.

## 9. Configuration

| Variable | Service | Default | Meaning |
|----------|---------|---------|---------|
| `ENABLE_AI_RESPONSE` | Go | `false` | Master toggle (`true/1/yes/on`) |
| `ENABLE_AI_AUTO_SEND` | Go | `false` | Reserved; auto-send not implemented |
| `AI_SERVICE_URL` | Go | `http://localhost:8090` (when enabled) | Sidecar URL (`http://ai-agent:8090` in Docker) |
| `DATABASE_URL` | sidecar | — | Postgres (use the Supabase **pooler** on IPv4 networks) |
| `OPENROUTER_API_KEY` | sidecar | — | Chat LLM key |
| `AI_MODEL` / `LLM_BASE_URL` | sidecar | `openai/gpt-4o-mini` / OpenRouter | Chat model/endpoint |
| `GOOGLE_API_KEY` | sidecar | — | Gemini embeddings key |
| `EMBEDDING_MODEL` / `EMBEDDING_DIM` | sidecar | `models/gemini-embedding-001` / `1536` | Must match the `VECTOR(...)` size |

**Provider split:** chat runs through **OpenRouter**; embeddings run through
**Google AI Studio** (OpenRouter has no embeddings endpoint). The two keys/providers
are independent.

## 10. Behavior when disabled (default)

- `POST /api/ai/reply` → **503** with `{"success": false, "message": "AI response feature is disabled"}`.
- The AI client/service are **not constructed**; a missing `AI_SERVICE_URL` does not block startup.
- All existing WhatsApp endpoints work exactly as before. Manual sending is never affected by the AI toggle.

## 11. Security & operational notes

- `/api/ai/reply` sits behind the same **Basic Auth** group as `/api/send-message`.
- The Go AI client uses a **context-aware 15s timeout** and does **not log
  message contents or upstream error bodies**.
- Secrets (`*_API_KEY`, `DATABASE_URL`) come from the environment only; `.env`
  is gitignored. No keys are committed.
- The sidecar fails closed: on retrieval/LLM errors the Go handler returns a
  generic 500 without leaking upstream details.

## 12. Deployment

- **Local:** run the Go service, then the sidecar
  (`uvicorn app:api --host 0.0.0.0 --port 8090`); set `ENABLE_AI_RESPONSE=true`
  and `AI_SERVICE_URL=http://localhost:8090`.
- **Docker Compose:** `docker-compose.yml` defines both `whatspoints` and
  `ai-agent`. Inside the compose network use `AI_SERVICE_URL=http://ai-agent:8090`.
  Compose reads the **root `.env`**.

## 13. Future work

- Implement `ENABLE_AI_AUTO_SEND` (auto-reply) as a separate, deliberate phase —
  kept independent from suggestion generation by design.
- Optional knowledge-source tracking (e.g. a `source` column) for cleaner
  document re-ingestion and provenance.
