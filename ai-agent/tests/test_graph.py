"""Unit tests for the LangGraph nodes. No network calls — the LLM and the
pgvector search are faked.

    cd ai-agent && pip install pytest && python -m pytest
"""
import graph


class _FakeMsg:
    def __init__(self, content):
        self.content = content


class _FakeLLM:
    def __init__(self, reply):
        self._reply = reply

    def invoke(self, _):
        return _FakeMsg(self._reply)


def test_detect_intent_valid(monkeypatch):
    monkeypatch.setattr(graph, "_llm", _FakeLLM("ask_promo"))
    assert graph.detect_intent({"customer_message": "promo?"}) == {"intent": "ask_promo"}


def test_detect_intent_unknown_for_garbage(monkeypatch):
    monkeypatch.setattr(graph, "_llm", _FakeLLM("blahblah"))
    assert graph.detect_intent({"customer_message": "??"}) == {"intent": "unknown"}


def test_route_after_retrieval():
    assert graph.route_after_retrieval({"documents": [{"id": 1}]}) == "generate_answer"
    assert graph.route_after_retrieval({"documents": []}) == "fallback"


def test_fallback_returns_polite_message():
    out = graph.fallback({"customer_message": "x"})
    assert out["answer"] == graph.FALLBACK_REPLY
    assert out["sources"] == []


def test_generate_answer_uses_documents(monkeypatch):
    monkeypatch.setattr(graph, "_llm", _FakeLLM("Masih kak"))
    docs = [{"id": 1, "title": "Promo", "content": "isi", "category": "promo", "score": 0.1}]
    out = graph.generate_answer({"customer_message": "promo?", "intent": "ask_promo", "documents": docs})
    assert out["answer"] == "Masih kak"
    assert out["sources"] == docs
