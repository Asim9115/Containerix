def suggest_fix(problem:str, evidence:str):
    return {        
        "task": "Explain the likely root cause and fix.",
        "problem": problem,
        "evidence": evidence,
        "restrictions": [
            "Do not modify files.",
            "Do not execute commands.",
            "Do not invent file contents.",
            "Mention uncertainty if evidence is insufficient.",
            "Explain how the user can verify the fix.",
        ],
    }