import os
from dotenv import load_dotenv

load_dotenv()

OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY") or OPENAI_API_KEY

MODEL = os.getenv(
    "GEMINI_MODEL",
    os.getenv("OPENAI_MODEL", "gemini-2.5-flash"),
)

if MODEL == "gemini":
    MODEL = "gemini-3.6-flash"

MAX_TURNS = int(
    os.getenv(
        "TRIAGE_MAX_TURNS",
        "6",
    )
)

PORT = int(
    os.getenv(
        "PORT",
        "8001",
    )
)

CONTAINERIX_API_URL = os.getenv(
    "CONTAINERIX_API_URL", "https://localhost:8080",
)

if not GEMINI_API_KEY:
    raise RuntimeError("GEMINI_API_KEY is missing")