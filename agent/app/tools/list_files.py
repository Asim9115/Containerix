from pathlib import Path

def list_repo_files(repopath:str):
    if not repopath.is_dir():
        raise NotADirectoryError("Directory not found")

    ignored = {
        ".git",
        "node_modules",
        "venv",
        ".venv",
        "__pycache__",
    }

    files = []

    for item in repopath.rglob("*"):
        if any(part in ignored for part in item.parts):
            continue

        if item.is_file():
            files.append(str(item.relative_to(repopath)))

        if len(files) > 200:
            break

    return {
            "files": files,
            "truncated": len(files) > 200
        }