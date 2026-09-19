import os
from fastapi import FastAPI, HTTPException
from openai import APIError
from app.models import TriageRequest, TriageResponse
from app.agent import run_triage_loop

app = FastAPI(
    title="Containerix Triage Agent",
    version="1.0.0",
    description="Diagnoses docker build failures using an LLM tool loop."
)

@app.get("/health")
def health():
    return {"status": "ok", "service": "triage-agent"}

@app.post("/triage", response_model=TriageResponse)
def triage(req: TriageRequest):
    if not os.getenv("GEMINI_API_KEY"):
        raise HTTPException(status_code=503, detail="GEMINI_API_KEY not configured")

    try:
        return run_triage_loop(req)
    except APIError as e:
        raise HTTPException(status_code=502, detail=f"LLM API error: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    from app.config import PORT
    uvicorn.run("app.main:app", host="0.0.0.0", port=PORT, reload=False)