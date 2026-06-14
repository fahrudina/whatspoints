"""API tests using FastAPI's TestClient with the graph mocked.

    cd ai-agent && pip install pytest httpx && python -m pytest
"""
from fastapi.testclient import TestClient

import app

client = TestClient(app.api)


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok", "service": "whatspoints-ai-agent"}


def test_ai_reply_shape(monkeypatch):
    # app imports `generate_reply` from graph lazily, so patch it on the graph module.
    import graph

    def fake_generate_reply(message, phone_number=""):
        return {
            "reply": "Masih kak",
            "intent": "ask_promo",
            "sources": [{"id": 1, "title": "Promo", "content": "isi", "category": "promo", "score": 0.12}],
        }

    monkeypatch.setattr(graph, "generate_reply", fake_generate_reply)

    resp = client.post("/ai/reply", json={"customer_message": "promo?", "phone_number": "628"})
    assert resp.status_code == 200
    body = resp.json()
    assert body["reply"] == "Masih kak"
    assert body["intent"] == "ask_promo"
    assert isinstance(body["sources"], list) and body["sources"][0]["id"] == 1
