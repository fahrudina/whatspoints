"""LangGraph RAG workflow for WhatsApp reply suggestions.

    START -> detect_intent -> retrieve_context -> route_after_retrieval
                                                    -> generate_answer (context found)
                                                    -> fallback        (no context)
                                                  -> END

Answers are Bahasa Indonesia, casual/polite WhatsApp-admin tone, and grounded
only in retrieved context (no hallucinated promo/price/branch/order data).
"""
import logging
import os
from typing import TypedDict

from langchain_openai import ChatOpenAI
from langgraph.graph import END, START, StateGraph

from embedding import search_knowledge

logger = logging.getLogger(__name__)

# Chat LLM runs through OpenRouter (OpenAI-compatible). Override base/key/model
# via env. Model names are OpenRouter-style, e.g. "openai/gpt-4o-mini".
CHAT_MODEL = os.getenv("AI_MODEL", "openai/gpt-4o-mini")
LLM_BASE_URL = os.getenv("LLM_BASE_URL", "https://openrouter.ai/api/v1")

# Match order matters: check specific labels before "unknown" when parsing the
# model's (possibly messy) output. A list, not a set, so iteration is stable.
INTENT_LABELS = [
    "ask_order_status",
    "ask_opening_hours",
    "ask_promo",
    "ask_service",
    "complaint",
    "unknown",
]
INTENTS = set(INTENT_LABELS)

# Only laundry-related intents get a reply. Anything else (e.g. "unknown") is
# skipped so the bot doesn't answer unrelated chatter.
LAUNDRY_INTENTS = INTENTS - {"unknown"}

# Cosine distance from search_knowledge (lower = more similar). Retrieved rows
# above this are treated as "not enough context" and routed to the fallback.
# Tune via env if the KB embeddings shift; 0.0 = identical, ~2.0 = opposite.
RELEVANCE_THRESHOLD = float(os.getenv("AI_RELEVANCE_THRESHOLD", "0.8"))

FALLBACK_REPLY = "Maaf kak, boleh dijelaskan lebih detail ya? Biar admin bantu cek 😊"

# Prompts are written in English for clarity/maintainability; the *replies* the
# model produces must stay in Bahasa Indonesia (enforced in ANSWER_SYSTEM).
INTENT_PROMPT = (
    "You are an intent classifier for a laundry business WhatsApp assistant.\n"
    "Classify the customer message into EXACTLY ONE label:\n"
    "- ask_promo: promotions, discounts, deals, prices, or rates.\n"
    "- ask_opening_hours: operating hours, open/close times, when they can come.\n"
    "- ask_service: what services exist (kiloan, satuan, antar-jemput, pengering, etc.).\n"
    "- ask_order_status: status or progress of an existing order/laundry.\n"
    "- complaint: a problem, dissatisfaction, or damaged/lost items.\n"
    "- unknown: greetings, small talk, or anything unrelated to the laundry business.\n"
    "Output ONLY the label in lowercase, with no punctuation or explanation.\n"
    "If unsure or unrelated to laundry, output: unknown.\n\n"
    "Message: {message}"
)

ANSWER_SYSTEM = (
    "You are a friendly, polite laundry shop admin on WhatsApp.\n"
    "ALWAYS reply to the customer in Bahasa Indonesia, in a casual, warm, concise tone.\n"
    "Answer ONLY using the CONTEXT provided below. Never invent promos, prices, "
    "branches, or order status that are not in the context.\n"
    "If the context does not contain enough information to answer, do NOT guess: "
    "reply in Bahasa Indonesia asking the customer for more detail, or say the admin "
    "will help check.\n"
    "For complaints and order-status questions, be careful: when data is insufficient, "
    "ask for more detail or say the admin will follow up."
)


class State(TypedDict, total=False):
    customer_message: str
    phone_number: str
    intent: str
    documents: list
    answer: str
    sources: list
    should_reply: bool


_llm = None


def _llm_client() -> ChatOpenAI:
    global _llm
    if _llm is None:
        _llm = ChatOpenAI(
            model=CHAT_MODEL,
            base_url=LLM_BASE_URL,
            api_key=os.getenv("LLM_API_KEY") or os.getenv("OPENROUTER_API_KEY"),
            temperature=0.3,
        )
    return _llm


def detect_intent(state: State) -> State:
    prompt = INTENT_PROMPT.format(message=state["customer_message"])
    raw = _llm_client().invoke(prompt).content.strip().lower()
    # Lenient: the model may wrap the label in quotes/extra words. Take the first
    # known label that appears (specific labels before "unknown").
    for label in INTENT_LABELS:
        if label in raw:
            return {"intent": label}
    return {"intent": "unknown"}


def route_after_intent(state: State) -> str:
    # Stop here for unrelated messages so the bot doesn't reply to everything.
    return "retrieve_context" if state.get("intent") in LAUNDRY_INTENTS else "skip"


def skip(state: State) -> State:
    return {"answer": "", "sources": [], "should_reply": False}


def retrieve_context(state: State) -> State:
    docs = search_knowledge(state["customer_message"])
    # Keep only rows close enough to count as real context; the rest are noise
    # and would otherwise push the model to answer without grounding.
    relevant = [d for d in docs if d["score"] <= RELEVANCE_THRESHOLD]
    return {"documents": relevant}


def route_after_retrieval(state: State) -> str:
    # No relevant context -> fallback so the bot still replies, but safely.
    return "generate_answer" if state.get("documents") else "fallback"


def generate_answer(state: State) -> State:
    docs = state["documents"]
    context = "\n".join(f"- {d['title']}: {d['content']}" for d in docs)
    messages = [
        ("system", ANSWER_SYSTEM),
        (
            "user",
            f"Intent: {state.get('intent')}\n\nKONTEKS:\n{context}\n\n"
            f"Pesan pelanggan: {state['customer_message']}",
        ),
    ]
    answer = _llm_client().invoke(messages).content.strip()
    return {"answer": answer, "sources": docs}


def fallback(state: State) -> State:
    return {"answer": FALLBACK_REPLY, "sources": []}


def build_graph():
    g = StateGraph(State)
    g.add_node("detect_intent", detect_intent)
    g.add_node("skip", skip)
    g.add_node("retrieve_context", retrieve_context)
    g.add_node("generate_answer", generate_answer)
    g.add_node("fallback", fallback)

    g.add_edge(START, "detect_intent")
    g.add_conditional_edges(
        "detect_intent",
        route_after_intent,
        {"retrieve_context": "retrieve_context", "skip": "skip"},
    )
    g.add_conditional_edges(
        "retrieve_context",
        route_after_retrieval,
        {"generate_answer": "generate_answer", "fallback": "fallback"},
    )
    g.add_edge("skip", END)
    g.add_edge("generate_answer", END)
    g.add_edge("fallback", END)
    return g.compile()


_graph = None


def generate_reply(customer_message: str, phone_number: str = "") -> dict:
    """Run the workflow and return {reply, intent, should_reply, sources}.

    Fail closed: on any LLM/DB error this returns a no-reply payload rather than
    raising, since the feature only produces optional suggestions.
    """
    global _graph
    if _graph is None:
        _graph = build_graph()
    try:
        result = _graph.invoke(
            {"customer_message": customer_message, "phone_number": phone_number}
        )
    except Exception:
        logger.exception("AI graph failed; returning no-reply")
        return {"reply": "", "intent": "unknown", "should_reply": False, "sources": []}
    return {
        "reply": result.get("answer", FALLBACK_REPLY),
        "intent": result.get("intent", "unknown"),
        "sources": result.get("sources", []),
        # False = unrelated message, caller should not reply.
        "should_reply": result.get("should_reply", True),
    }
