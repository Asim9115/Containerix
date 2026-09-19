def read_build_logs(build_log:str, last_n_lines: int = 80):
    if not build_log or not build_log.strip():
        return (
            "Build log is not available. "
            "Reason from the raw error string in the initial context."
        )
    lines = build_log.strip().splitlines()
    if len(lines) <= last_n_lines:
        return "\n".join(lines)

    tail = lines[-last_n_lines:]
    header = f"[Log truncated — showing last {last_n_lines} of {len(lines)} lines]\n\n"
    return header + "\n".join(tail)
