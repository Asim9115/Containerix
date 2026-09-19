def suggest_fix(root_cause: str, fix: str, confidence: str, steps_taken: list) -> str:
    valid = {"high", "medium", "low"}
    if confidence not in valid:
        confidence = "low"
    return "Diagnosis recorded. Analysis complete."