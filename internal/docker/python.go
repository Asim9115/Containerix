package docker

import (
	"fmt"

	"github.com/asim9115/containerix/internal/detector"
)

func generatePython(d detector.DetectResult) (string, error) {
	version := d.Version
	if version == "" {
		version = "3.12"
	}

	switch d.Framework {
	case "flask":
		return generateFlask(version, d), nil
	case "fastapi":
		return generateFastAPI(version, d), nil
	case "django":
		return generateDjango(version, d), nil
	case "tornado":
		return generateTornado(version, d), nil
	default:
		return generatePlainPython(version, d), nil
	}
}

// ---------------------------------------------------------------------------
// Flask
// ---------------------------------------------------------------------------

func generateFlask(version string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 5000
	}
	depFile := d.DependencyFile
	if depFile == "" {
		depFile = "requirements.txt"
	}
	return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

# Install dependencies first for better layer caching
COPY %s .
RUN pip install --no-cache-dir -r %s

COPY . .

ENV FLASK_APP=main.py
ENV FLASK_RUN_HOST=0.0.0.0
ENV FLASK_RUN_PORT=%d
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

EXPOSE %d

CMD ["flask", "run"]
`, version, depFile, depFile, port, port)
}

// ---------------------------------------------------------------------------
// FastAPI
// ---------------------------------------------------------------------------

func generateFastAPI(version string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 8000
	}
	depFile := d.DependencyFile
	if depFile == "" {
		depFile = "requirements.txt"
	}
	return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

COPY %s .
RUN pip install --no-cache-dir -r %s

COPY . .

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

EXPOSE %d

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "%d"]
`, version, depFile, depFile, port, port)
}

// ---------------------------------------------------------------------------
// Django
// ---------------------------------------------------------------------------

func generateDjango(version string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 8000
	}
	depFile := d.DependencyFile
	if depFile == "" {
		depFile = "requirements.txt"
	}
	return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

COPY %s .
RUN pip install --no-cache-dir -r %s

COPY . .

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV DJANGO_SETTINGS_MODULE=core.settings

EXPOSE %d

RUN python manage.py collectstatic --noinput || true

CMD ["python", "manage.py", "runserver", "0.0.0.0:%d"]
`, version, depFile, depFile, port, port)
}

// ---------------------------------------------------------------------------
// Tornado
// ---------------------------------------------------------------------------

func generateTornado(version string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 8888
	}
	depFile := d.DependencyFile
	if depFile == "" {
		depFile = "requirements.txt"
	}
	return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

COPY %s .
RUN pip install --no-cache-dir -r %s

COPY . .

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

EXPOSE %d

CMD ["python", "main.py"]
`, version, depFile, depFile, port)
}

// ---------------------------------------------------------------------------
// Plain Python (no framework detected)
// ---------------------------------------------------------------------------

func generatePlainPython(version string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 8000
	}
	depFile := d.DependencyFile
	if depFile == "" {
		depFile = "requirements.txt"
	}
	return fmt.Sprintf(`FROM python:%s-slim

WORKDIR /app

COPY %s .
RUN pip install --no-cache-dir -r %s

COPY . .

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

EXPOSE %d

CMD ["python", "main.py"]
`, version, depFile, depFile, port)
}