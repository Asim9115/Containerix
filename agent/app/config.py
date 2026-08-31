import os
from dotenv import load_dotenv

load_dotenv()

ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY")

MODEL = os.getenv(
    "TRIAGE_MODEL",
    "claude-opus-4.5",
)

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