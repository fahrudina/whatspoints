"""Unit tests for the in-process conversation memory."""
import memory


def test_roundtrip_and_ttl(monkeypatch):
    memory._store.clear()
    memory.append("628", "halo", "hai kak")
    assert memory.get_history("628") == [("user", "halo"), ("assistant", "hai kak")]

    # Past the TTL the conversation is forgotten.
    monkeypatch.setattr(memory, "_TTL", -1)
    assert memory.get_history("628") == []


def test_empty_phone_is_noop():
    memory._store.clear()
    memory.append("", "x", "y")
    assert memory.get_history("") == []
    assert len(memory._store) == 0


def test_caps_to_recent_turns(monkeypatch):
    memory._store.clear()
    monkeypatch.setattr(memory, "_MAX_TURNS", 2)  # keep last 1 exchange
    memory.append("628", "q1", "a1")
    memory.append("628", "q2", "a2")
    assert memory.get_history("628") == [("user", "q2"), ("assistant", "a2")]


def test_evicts_oldest_over_cap(monkeypatch):
    memory._store.clear()
    monkeypatch.setattr(memory, "_MAX_CONVERSATIONS", 2)
    for p in ("a", "b", "c"):
        memory.append(p, "q", "r")
    assert "a" not in memory._store  # oldest evicted
    assert set(memory._store) == {"b", "c"}
