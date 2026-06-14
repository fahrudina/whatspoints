"""FastAPI sidecar for whatspoints AI reply suggestions.

Run with:
    uvicorn app:api --host 0.0.0.0 --port 8090
"""
import os

from dotenv import load_dotenv
from fastapi import FastAPI
from pydantic import BaseModel

load_dotenv()

api = FastAPI(title="whatspoints-ai-agent")


class ReplyRequest(BaseModel):
    customer_message: str
    phone_number: str = ""


@api.get("/health")
def health():
    return {"status": "ok", "service": "whatspoints-ai-agent"}


@api.post("/ai/reply")
def ai_reply(req: ReplyRequest):
    # Imported lazily so /health works without OpenAI/DB configured.
    from graph import generate_reply

    return generate_reply(req.customer_message, req.phone_number)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        api,
        host=os.getenv("AI_HOST", "0.0.0.0"),
        port=int(os.getenv("AI_PORT", "8090")),
    )
