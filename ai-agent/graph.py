"""LangGraph RAG workflow for WhatsApp reply suggestions.

    START -> detect_intent -> retrieve_context -> route_after_retrieval
                                                    -> generate_answer (context found)
                                                    -> fallback        (no context)
                                                  -> END

Answers are Bahasa Indonesia, casual/polite WhatsApp-admin tone, and grounded
only in retrieved context (no hallucinated promo/price/branch/order data).
"""
import os
from typing import TypedDict

from langchain_openai import ChatOpenAI
from langgraph.graph import END, START, StateGraph

from embedding import search_knowledge

# Chat LLM runs through OpenRouter (OpenAI-compatible). Override base/key/model
# via env. Model names are OpenRouter-style, e.g. "openai/gpt-4o-mini".
CHAT_MODEL = os.getenv("AI_MODEL", "openai/gpt-4o-mini")
LLM_BASE_URL = os.getenv("LLM_BASE_URL", "https://openrouter.ai/api/v1")

INTENTS = {
    "ask_promo",
    "ask_opening_hours",
    "ask_service",
    "complaint",
    "ask_order_status",
    "unknown",
}

# Only laundry-related intents get a reply. Anything else (e.g. "unknown") is
# skipped so the bot doesn't answer unrelated chatter.
LAUNDRY_INTENTS = INTENTS - {"unknown"}

FALLBACK_REPLY = "Maaf kak, boleh dijelaskan lebih detail ya? Biar admin bantu cek 😊"

INTENT_PROMPT = (
    "Klasifikasikan pesan pelanggan ke SATU label berikut: "
    "ask_promo, ask_opening_hours, ask_service, complaint, ask_order_status, unknown. "
    "Jawab HANYA dengan label, tanpa penjelasan.\n\nPesan: {message}"
)

ANSWER_SYSTEM = (
    "Kamu adalah admin laundry yang ramah dan sopan di WhatsApp. "
    "Balas dalam Bahasa Indonesia dengan gaya santai, singkat, dan ramah. "
    "Jawab HANYA berdasarkan KONTEKS di bawah. Jangan mengarang promo, harga, "
    "cabang, atau status pesanan yang tidak ada di konteks. "
    "Jika konteks tidak cukup untuk menjawab, minta pelanggan menjelaskan lebih detail "
    "atau katakan admin akan bantu cek. "
    "Untuk keluhan (complaint) dan status pesanan (ask_order_status), bersikap hati-hati: "
    "bila data tidak cukup, minta detail tambahan atau sampaikan admin akan bantu cek."
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
    return {"intent": raw if raw in INTENTS else "unknown"}


def route_after_intent(state: State) -> str:
    # Stop here for unrelated messages so the bot doesn't reply to everything.
    return "retrieve_context" if state.get("intent") in LAUNDRY_INTENTS else "skip"


def skip(state: State) -> State:
    return {"answer": "", "sources": [], "should_reply": False}


def retrieve_context(state: State) -> State:
    docs = search_knowledge(state["customer_message"])
    return {"documents": docs}


def route_after_retrieval(state: State) -> str:
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
    """Run the workflow and return {reply, intent, sources}."""
    global _graph
    if _graph is None:
        _graph = build_graph()
    result = _graph.invoke(
        {"customer_message": customer_message, "phone_number": phone_number}
    )
    return {
        "reply": result.get("answer", FALLBACK_REPLY),
        "intent": result.get("intent", "unknown"),
        "sources": result.get("sources", []),
        # False = unrelated message, caller should not reply.
        "should_reply": result.get("should_reply", True),
    }
