"""In-process short-term conversation memory, keyed by phone number.

ponytail: a global OrderedDict guarded by one lock. The ai-agent runs as a
single process, so this is enough; swap for Redis/Postgres if you ever run
multiple replicas. History is intentionally lost on restart — it only needs to
keep a chat feeling continuous for the length of one conversation.
"""
import os
import threading
import time
from collections import OrderedDict, deque


def _env_num(name: str, default, cast, min_value):
    # Fall back to default on junk, clamp to min so we never crash deque(maxlen).
    try:
        return max(cast(os.getenv(name, default)), min_value)
    except (TypeError, ValueError):
        return default


# Keep the last N exchanges (each = one user + one assistant message).
_MAX_TURNS = _env_num("HISTORY_TURNS", 5, int, 1) * 2
_TTL = _env_num("HISTORY_TTL_SECONDS", 1800.0, float, 0.0)  # 30 min idle -> forget
_MAX_CONVERSATIONS = _env_num("HISTORY_MAX_CONVERSATIONS", 1000, int, 1)

_lock = threading.Lock()
# phone -> {"turns": deque[(role, text)], "ts": last_activity_epoch}
_store: "OrderedDict[str, dict]" = OrderedDict()


def _expired(entry: dict, now: float) -> bool:
    return now - entry["ts"] > _TTL


def get_history(phone: str) -> list[tuple[str, str]]:
    """Return [(role, text), ...] for recent turns, or [] if none/expired."""
    if not phone:
        return []
    now = time.time()
    with _lock:
        entry = _store.get(phone)
        if entry is None or _expired(entry, now):
            _store.pop(phone, None)
            return []
        _store.move_to_end(phone)
        return list(entry["turns"])


def append(phone: str, user_msg: str, bot_reply: str) -> None:
    """Record one user+assistant exchange for `phone`."""
    if not phone:
        return
    now = time.time()
    with _lock:
        entry = _store.get(phone)
        if entry is None or _expired(entry, now):
            entry = {"turns": deque(maxlen=_MAX_TURNS), "ts": now}
            _store[phone] = entry
        entry["turns"].append(("user", user_msg))
        entry["turns"].append(("assistant", bot_reply))
        entry["ts"] = now
        _store.move_to_end(phone)
        _prune(now)


def _prune(now: float) -> None:
    # Called under _lock: drop expired convos, then evict oldest over the cap.
    for phone in [p for p, e in _store.items() if _expired(e, now)]:
        del _store[phone]
    while len(_store) > _MAX_CONVERSATIONS:
        _store.popitem(last=False)
