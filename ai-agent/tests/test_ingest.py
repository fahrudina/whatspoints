"""Tests for document reading + chunking (no DB, no network).

    cd ai-agent && python -m pytest tests/test_ingest.py
"""
import pytest

import ingest_document as ing


def test_read_txt_and_md(tmp_path):
    for ext in (".txt", ".md"):
        f = tmp_path / f"doc{ext}"
        f.write_text("hello world", encoding="utf-8")
        assert ing.read_document(str(f)) == "hello world"


def test_read_unsupported_raises(tmp_path):
    f = tmp_path / "doc.csv"
    f.write_text("a,b,c", encoding="utf-8")
    with pytest.raises(ValueError):
        ing.read_document(str(f))


def test_chunk_short_text_single_chunk():
    assert ing.chunk_text("short note") == ["short note"]
    assert ing.chunk_text("   ") == []


def test_chunk_long_text_splits_with_progress():
    text = ("kalimat tentang laundry. " * 200).strip()  # ~5000 chars
    chunks = ing.chunk_text(text, size=900, overlap=120)
    assert len(chunks) > 1
    assert all(c.strip() for c in chunks)          # no empty chunks
    assert all(len(c) <= 900 for c in chunks)      # respects size
    # Reassembled chunks cover all the words (overlap may repeat some).
    assert "laundry" in chunks[0] and "laundry" in chunks[-1]
