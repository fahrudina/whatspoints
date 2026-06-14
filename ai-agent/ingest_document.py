"""Ingest a document into knowledge_base, then embed it.

Splits a .txt/.md/.pdf file into chunks, inserts each chunk as a row (title =
source name, content = chunk, category = your tag), and runs the normal
embedding step. No schema change — chunks are ordinary knowledge_base rows.

    cd ai-agent
    python ingest_document.py path/to/price-list.pdf --category service
    python ingest_document.py faq.md --title "FAQ" --category business_info

Re-ingesting the same title is blocked unless you pass --replace.
"""
import argparse
import os
from pathlib import Path

import psycopg
from dotenv import load_dotenv

load_dotenv()

CHUNK_SIZE = 900
CHUNK_OVERLAP = 120


def read_document(path: str) -> str:
    suffix = Path(path).suffix.lower()
    if suffix in (".txt", ".md"):
        return Path(path).read_text(encoding="utf-8")
    if suffix == ".pdf":
        from pypdf import PdfReader  # lazy: only PDFs need this dependency

        reader = PdfReader(path)
        return "\n".join((page.extract_text() or "") for page in reader.pages)
    raise ValueError(f"unsupported file type {suffix!r} (supported: .txt, .md, .pdf)")


def chunk_text(text: str, size: int = CHUNK_SIZE, overlap: int = CHUNK_OVERLAP) -> list[str]:
    """Split text into ~`size`-char chunks with `overlap`, breaking on whitespace
    boundaries so words aren't cut mid-token. Empty chunks are dropped."""
    text = text.strip()
    if len(text) <= size:
        return [text] if text else []

    chunks: list[str] = []
    start = 0
    while start < len(text):
        end = start + size
        if end < len(text):
            window = text[start:end]
            # Prefer to break at a paragraph/sentence/word boundary near the end.
            brk = max(window.rfind("\n"), window.rfind(". "), window.rfind(" "))
            if brk > size // 2:
                end = start + brk + 1
        chunk = text[start:end].strip()
        if chunk:
            chunks.append(chunk)
        start = max(end - overlap, start + 1)  # +1 guarantees forward progress
    return chunks


def main() -> None:
    ap = argparse.ArgumentParser(description="Chunk a document into knowledge_base rows and embed it.")
    ap.add_argument("path", help="path to a .txt, .md, or .pdf file")
    ap.add_argument("--category", default=None, help="category tag for the rows")
    ap.add_argument("--title", default=None, help="source title (defaults to the file name)")
    ap.add_argument("--replace", action="store_true", help="replace existing chunks with the same title")
    ap.add_argument("--no-embed", action="store_true", help="insert rows only; embed later via index_knowledge.py")
    args = ap.parse_args()

    title = args.title or Path(args.path).name
    chunks = chunk_text(read_document(args.path))
    if not chunks:
        print("No text extracted; nothing to ingest.")
        return

    with psycopg.connect(os.environ["DATABASE_URL"]) as conn:
        with conn.cursor() as cur:
            existing = cur.execute(
                "SELECT count(*) FROM knowledge_base WHERE title = %s", (title,)
            ).fetchone()[0]
            if existing and not args.replace:
                print(f"'{title}' already has {existing} chunks. Use --replace to re-ingest.")
                return
            if existing:  # --replace
                cur.execute("DELETE FROM knowledge_base WHERE title = %s", (title,))
                print(f"Removed {existing} existing chunks for '{title}'.")
            for chunk in chunks:
                cur.execute(
                    "INSERT INTO knowledge_base (title, content, category) VALUES (%s, %s, %s)",
                    (title, chunk, args.category),
                )
        conn.commit()
    print(f"Inserted {len(chunks)} chunks from '{title}'.")

    if args.no_embed:
        print("Skipped embedding (--no-embed). Run: python index_knowledge.py")
        return
    import index_knowledge  # lazy: reuse the exact embedding path

    index_knowledge.main()


if __name__ == "__main__":
    main()
