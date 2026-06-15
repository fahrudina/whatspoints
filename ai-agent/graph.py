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

import memory
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
    "Classify the LATEST customer message into EXACTLY ONE label:\n"
    "- ask_promo: promotions, discounts, deals, prices, or rates.\n"
    "- ask_opening_hours: operating hours, open/close times, when they can come.\n"
    "- ask_service: what services exist (kiloan, satuan, antar-jemput, pengering, etc.).\n"
    "- ask_order_status: status or progress of an existing order/laundry.\n"
    "- complaint: a problem, dissatisfaction, or damaged/lost items.\n"
    "- unknown: greetings, small talk, or anything unrelated to the laundry business.\n"
    "Use the recent conversation ONLY to resolve short follow-ups (e.g. 'iya',\n"
    "'berapa lama?', 'yang itu'): inherit the topic the customer is continuing.\n"
    "Output ONLY the label in lowercase, with no punctuation or explanation.\n"
    "If unsure or unrelated to laundry, output: unknown.\n\n"
    "Recent conversation:\n{history}\n\n"
    "Latest message: {message}"
)

ANSWER_SYSTEM = (
    "You are a friendly laundry shop admin chatting with a customer on WhatsApp.\n"
    "Reply in Bahasa Indonesia with a natural, warm, conversational tone — like a "
    "real person texting, not a form letter.\n"
    "Use ALL the relevant details in the CONTEXT to answer fully. When the customer "
    "asks about prices or a price list, list the specific items and prices found in "
    "the context (not just the promo); if there are many, give the ones they asked "
    "about and mention more are available. Never invent prices, promos, branches, or "
    "order status that are not in the context.\n"
    "Do NOT end every message with a generic closing like 'jika ada yang ingin "
    "ditanyakan lebih lanjut, silakan'. Only ask a follow-up question when it "
    "genuinely helps; otherwise just give the answer and stop.\n"
    "If the context truly does not contain the answer, say so honestly and offer to "
    "have the admin check — do not guess.\n"
    "For complaints or order-status questions with insufficient data, ask for the "
    "specific detail you need (e.g. order id or name)."
)


class State(TypedDict, total=False):
    customer_message: str
    phone_number: str
    history: list  # [(role, text), ...] prior turns for this phone number
    intent: str
    documents: list
    answer: str
    sources: list
    should_reply: bool


def _format_history(history: list) -> str:
    if not history:
        return "(none)"
    return "\n".join(f"{role}: {text}" for role, text in history)


_llm = None


def _llm_client() -> ChatOpenAI:
    global _llm
    if _llm is None:
        _llm = ChatOpenAI(
            model=CHAT_MODEL,
            base_url=LLM_BASE_URL,
            api_key=os.getenv("LLM_API_KEY") or os.getenv("OPENROUTER_API_KEY"),
            # A bit warmer so replies vary instead of repeating a canned sign-off.
            temperature=float(os.getenv("AI_TEMPERATURE", "0.6")),
        )
    return _llm


def detect_intent(state: State) -> State:
    prompt = INTENT_PROMPT.format(
        history=_format_history(state.get("history")),
        message=state["customer_message"],
    )
    raw = _llm_client().invoke(prompt).content.strip().lower()
    # Lenient: the model may wrap the label in quotes/extra words. Pick the label
    # that appears earliest in the output (textual order, not list priority) so a
    # response mentioning several labels resolves to the one stated first.
    best = None
    for label in INTENT_LABELS:
        idx = raw.find(label)
        if idx != -1 and (best is None or idx < best[0]):
            best = (idx, label)
    return {"intent": best[1] if best else "unknown"}


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
        # Prior turns give the model continuity; roles map to human/ai messages.
        *(state.get("history") or []),
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
            {
                "customer_message": customer_message,
                "phone_number": phone_number,
                "history": memory.get_history(phone_number),
            }
        )
    except Exception:
        logger.exception("AI graph failed; returning no-reply")
        return {"reply": "", "intent": "unknown", "should_reply": False, "sources": []}

    reply = result.get("answer", FALLBACK_REPLY)
    should_reply = result.get("should_reply", True)  # False = unrelated, skip.
    # Only remember exchanges we actually replied to, so skipped/off-topic
    # chatter doesn't pollute the conversation context.
    if should_reply and reply.strip():
        memory.append(phone_number, customer_message, reply)
    return {
        "reply": reply,
        "intent": result.get("intent", "unknown"),
        "sources": result.get("sources", []),
        "should_reply": should_reply,
    }
