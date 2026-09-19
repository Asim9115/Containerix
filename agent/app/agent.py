import json
from openai import OpenAI

from app.config import GEMINI_API_KEY, MODEL, MAX_TURNS
from app.prompts import build_system_prompt
from app.models import TriageRequest, TriageResponse
from app.tools.build_logs import read_build_logs
from app.tools.read_file import read_file
from app.tools.list_files import list_repo_files
from app.tools.suggest_fix import suggest_fix

client = OpenAI(
    api_key=GEMINI_API_KEY,
    base_url="https://generativelanguage.googleapis.com/v1beta/openai/"
)


TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "read_build_logs",
            "description": "Read the Docker build output. ALWAYS call this first.",
            "parameters": {
                "type": "object",
                "properties": {
                    "last_n_lines": {
                        "type": "integer",
                        "description": "How many tail lines to return. Default 80."
                    }
                }
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "list_repo_files",
            "description": "List all files in the cloned repo. Use before read_file to discover filenames.",
            "parameters": {"type": "object", "properties": {}}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "read_file",
            "description": "Read a specific file from the repo by relative path.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Relative path from repo root. e.g. 'Dockerfile', 'package.json'"
                    }
                },
                "required": ["path"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "suggest_fix",
            "description": "Call ONLY when you have identified the root cause. This ends your analysis.",
            "parameters": {
                "type": "object",
                "properties": {
                    "root_cause": {"type": "string"},
                    "fix": {"type": "string"},
                    "confidence": {
                        "type": "string",
                        "enum": ["high", "medium", "low"]
                    },
                    "steps_taken": {
                        "type": "array",
                        "items": {"type": "string"}
                    }
                },
                "required": ["root_cause", "fix", "confidence", "steps_taken"]
            }
        }
    }
]

def execute_tool(name: str, args: dict, req: TriageRequest) -> str:
    """
    Executes the tool the model requested.
    Returns a plain string — this is what the model sees as the tool result.
    """
    try:
        if name == "read_build_logs":
            n = args.get("last_n_lines", 80)
            return read_build_logs(req.build_log or "", n)
        elif name == "list_repo_files":
            result = list_repo_files(req.repo_path)
            return "\n".join(result["files"]) + (
                "\n\n[truncated at 200 files]" if result.get("truncated") else ""
            )
        elif name == "read_file":
            path = args.get("path", "")
            result = read_file(req.repo_path, path)
            content = result["content"]
            # Cap at 4000 chars to control token spend
            if len(content) > 4000:
                content = content[:4000] + "\n\n[... file truncated at 4000 chars]"
            return content
        elif name == "suggest_fix":
            return suggest_fix(**args)
        else:
            return f"Unknown tool: {name}"
    except Exception as e:
        # Return the error AS A STRING to the model — don't crash the loop.
        # The model can adapt: e.g., "file not found" → try a different filename.
        return f"Tool error: {str(e)}"

def run_triage_loop(req: TriageRequest) -> TriageResponse:
    """
    The multi-turn tool loop. The model decides:
    - Which tools to call
    - In what order
    - When it has enough information to conclude
    
    Your code only executes what the model requests.
    """
    messages = [
        {
            "role": "system",
            "content": build_system_prompt(req.repo_url, req.error)
        },
        {
            "role": "user",
            "content": (
                f"Docker build failed for: {req.repo_url}\n"
                f"Error: {req.error}\n\n"
                "Diagnose the root cause and suggest a fix."
            )
        }
    ]
    final_result = None
    turns = 0
    while turns < MAX_TURNS:
        turns += 1
        response = client.chat.completions.create(
            model=MODEL,
            messages=messages,
            tools=TOOLS,
            tool_choice="auto",  
        )
        choice = response.choices[0]
        msg = choice.message
        messages.append(msg)  
        finish_reason = choice.finish_reason

        if finish_reason == "stop" and not msg.tool_calls:
            break

        if msg.tool_calls:
            tool_results = []
            for call in msg.tool_calls:
                name = call.function.name
                args = json.loads(call.function.arguments)
                # Detect terminal tool BEFORE executing
                if name == "suggest_fix":
                    final_result = args   # capture the structured output
                    # Still feed a result back so conversation stays valid
                    tool_results.append({
                        "role": "tool",
                        "tool_call_id": call.id,
                        "content": suggest_fix(**args)
                    })
                    break  # stop processing more tools in this turn
                else:
                    result_text = execute_tool(name, args, req)
                    tool_results.append({
                        "role": "tool",
                        "tool_call_id": call.id,
                        "content": result_text
                    })
            messages.extend(tool_results)
            # Exit the outer loop if suggest_fix was called
            if final_result is not None:
                break
    # Build the response
    if final_result:
        return TriageResponse(
            job_id=req.job_id,
            root_cause=final_result.get("root_cause", "Unknown"),
            fix=final_result.get("fix", "No fix suggested."),
            confidence=final_result.get("confidence", "low"),
            steps_taken=final_result.get("steps_taken", []),
            turns_used=turns,
        )
    else:
        # Max turns exceeded without a conclusion
        return TriageResponse(
            job_id=req.job_id,
            root_cause="Analysis did not converge within the turn limit.",
            fix=f"Manually inspect the build output. Raw error: {req.error}",
            confidence="low",
            steps_taken=[f"Exhausted {MAX_TURNS} turns without reaching a conclusion."],
            turns_used=turns,
        )