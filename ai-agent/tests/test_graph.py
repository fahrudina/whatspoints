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


def test_detect_intent_picks_earliest_label(monkeypatch):
    # When several labels appear, the one stated first textually wins.
    monkeypatch.setattr(graph, "_llm", _FakeLLM("ask_service then maybe ask_promo"))
    assert graph.detect_intent({"customer_message": "x"}) == {"intent": "ask_service"}


def test_route_after_intent_skips_unrelated():
    assert graph.route_after_intent({"intent": "ask_promo"}) == "retrieve_context"
    assert graph.route_after_intent({"intent": "complaint"}) == "retrieve_context"
    assert graph.route_after_intent({"intent": "unknown"}) == "skip"
    assert graph.route_after_intent({}) == "skip"


def test_skip_does_not_reply():
    out = graph.skip({"customer_message": "hai bro apa kabar"})
    assert out["should_reply"] is False
    assert out["answer"] == ""
    assert out["sources"] == []


def test_route_after_retrieval():
    assert graph.route_after_retrieval({"documents": [{"id": 1}]}) == "generate_answer"
    assert graph.route_after_retrieval({"documents": []}) == "fallback"


def test_retrieve_context_drops_irrelevant(monkeypatch):
    # score is cosine distance: 0.2 is close, 0.95 is far (above threshold).
    docs = [
        {"id": 1, "title": "a", "content": "x", "category": "c", "score": 0.2},
        {"id": 2, "title": "b", "content": "y", "category": "c", "score": 0.95},
    ]
    monkeypatch.setattr(graph, "search_knowledge", lambda q, **k: docs)
    monkeypatch.setattr(graph, "RELEVANCE_THRESHOLD", 0.8)
    out = graph.retrieve_context({"customer_message": "promo?"})
    assert out["documents"] == [docs[0]]


def test_retrieve_context_all_irrelevant_triggers_fallback(monkeypatch):
    docs = [{"id": 1, "title": "a", "content": "x", "category": "c", "score": 1.4}]
    monkeypatch.setattr(graph, "search_knowledge", lambda q, **k: docs)
    monkeypatch.setattr(graph, "RELEVANCE_THRESHOLD", 0.8)
    out = graph.retrieve_context({"customer_message": "cuaca hari ini?"})
    assert out["documents"] == []
    assert graph.route_after_retrieval(out) == "fallback"


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
