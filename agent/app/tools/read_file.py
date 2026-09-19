from pathlib import Path

def read_file(path:str, name:str):
    filepath = Path(path / name).resolve()

    if not filepath.is_relative_to(path):
        raise ValueError("Invalid file")
    
    if not filepath.is_file():
        raise FileNotFoundError("file not found")

    if filepath.stat().st_size > 100_000:
        raise ValueError("File is too large to read")

    try:
        content = filepath.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        raise ValueError("File is not a text file")

    return {
        "filepath": filepath,
        "content" : content,
    }