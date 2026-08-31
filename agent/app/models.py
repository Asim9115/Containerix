from typing import Optional
from pydantic import BaseModel

class TriageRequest(BaseModel):
    job_id:str
    error:str
    repo_path:str
    repo_url:str
    build_log:Optional[str] = None

class TriageResponse(BaseModel):
    job_id:str
    root_cause:str
    fix:str
    confidence:str
    steps_taken:list[str]
    turns_used:int
    