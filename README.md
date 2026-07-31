# Containerix

> A self-hosted, lightweight Platform-as-a-Service (PaaS) engine — deploy any GitHub repository into an isolated, resource-controlled Docker container with a single HTTP request.

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat-square&logo=docker)](https://docs.docker.com/get-docker/)
[![Nixpacks](https://img.shields.io/badge/Nixpacks-Required-purple?style=flat-square)](https://nixpacks.com/)

---

## What is Containerix?

Containerix is a **work-in-progress** self-hosted PaaS engine built from scratch in Go. The core deployment engine is actively being developed and hardened. Give it a GitHub URL and it handles the rest — clones the repo, auto-detects the language and runtime, builds a Docker image (via Nixpacks or an existing Dockerfile), runs the container inside a **Linux cgroup v2** sandbox for resource enforcement, and returns a live URL.

Think Railway or Render, but something you own and run yourself.

> **Status:** The core pipeline is functional end-to-end. User management, a polished API surface, and a web dashboard are not built yet — those come after the core system is robust.

---

## Key Features

| Feature | Detail |
|---|---|
| 🚀 **One-request deploys** | `POST /build` with a GitHub URL triggers the full async pipeline |
| 🔍 **Auto build detection** | Nixpacks if no Dockerfile found; falls back to `docker build` if one exists |
| 🔒 **cgroup v2 resource isolation** | Each container is added to a Linux cgroup v2 group with CPU and memory limits |
| 📡 **Real-time SSE log streaming** | `/containers/:id/logs` streams build + runtime logs in two phases |
| 💾 **SQLite persistence** | All deployments, jobs, and port allocations are persisted across restarts |
| 🔄 **Startup reconciliation** | On restart, syncs host containers against the DB — marks orphaned ones stopped |
| 🛡️ **URL validation** | Only HTTPS GitHub URLs accepted; option-injection guarded |
| 🧩 **Repository pattern** | Interface-driven DB layer — designed to swap SQLite for PostgreSQL without touching business logic |

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                        HTTP Client                           │
└─────────────────────────┬────────────────────────────────────┘
                          │  POST /build
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                    Gin HTTP Router                           │
│             router/router.go → internal/api/                 │
└─────────────────────────┬────────────────────────────────────┘
                          │ 202 Accepted  (job_id returned immediately)
                          │ goroutine spawned ─────────────────────────┐
                          ▼                                             │
┌──────────────────────────────────────────────────────────────┐        │
│              Async Pipeline (internal/pipeline/)             │        │
│  1.  Create DB record          (status: building)            │        │
│  2.  Check sandbox resource budget                           │        │
│  3.  Validate GitHub URL                                     │        │
│  4.  git clone → tmp/<uuid>/                                 │        │
│  5.  nixpacks build / docker build → image tag               │        │
│  6.  Probe container → TCP scan → detect container port      │        │
│  7.  Allocate free host port + persist to SQLite             │        │
│  8.  docker run  (port mapping + env vars)                   │        │
│  9.  Get container PID → write to cgroup v2                  │        │
│  10. Update DB record          (status: running)             │        │
│  11. Emit SSE "deployed" event with live URL                 │        │
└─────────────────────────┬────────────────────────────────────┘        │
                          │                                             │
           ┌──────────────┴──────────────┐                             │
           ▼                             ▼                             │
┌─────────────────────┐    ┌──────────────────────────┐               │
│ SQLite              │    │ Linux cgroup v2           │               │
│ data/containerix.db │    │ /sys/fs/cgroup/<name>/    │               │
│ WAL mode, FK on     │    │ cpu.max + memory.max      │               │
└─────────────────────┘    └──────────────────────────┘               │
                                                                       │
                           SSE log stream ◄──────────────────────────── ┘
```

### Package Breakdown

| Package | Responsibility |
|---|---|
| `cmd/server` | Entry point — initialises DB, repos, sandbox, pipeline, HTTP server |
| `router/` | Gin route wiring |
| `internal/api/` | HTTP handlers, in-memory job store, SSE streaming |
| `internal/pipeline/` | Core deploy flow + startup sync (`SyncData`) |
| `internal/builder/` | `git clone`, Dockerfile detection, `nixpacks build` / `docker build` |
| `internal/detector/` | TCP port scanning on probe containers |
| `internal/docker/` | Thin wrappers around Docker CLI (inspect, logs, PID, run) |
| `internal/container/` | Container lifecycle — run, stop, delete |
| `internal/cgroup/` | Linux cgroup v2 read/write (`cpu.max`, `memory.max`, `cgroup.procs`) |
| `internal/sandbox/` | CPU/memory resource budget tracker |
| `internal/state/` | Global singleton — sandbox + ports manager |
| `internal/ports/` | In-memory free-port allocator |
| `internal/database/` | SQLite init, WAL mode, embedded SQL migrations |
| `internal/repository/` | Repository interfaces (swap-ready for Postgres) |
| `internal/repository/sqllite/` | SQLite implementations of all repos |
| `internal/types/` | Shared types: `Config`, `Tier`, `LogBus`, `SSEEvent` |

---

## How Goroutines and Channels Power the Pipeline

Go's concurrency primitives are central to how Containerix works — not just a nice-to-have.

### Non-blocking deploys via goroutines

The moment `POST /build` is handled, a goroutine is spawned and the HTTP handler returns `202 Accepted` with a `job_id` immediately. The entire pipeline — cloning, building, port detection, running the container — runs concurrently in the background without blocking any other request.

```go
go func() {
    containerID, err := h.Pipeline.Deploy(jobId, logBus, body.Url, tier, body.Env)
    // ... update job state
}()

// returned already — client has job_id
c.JSON(http.StatusAccepted, gin.H{"job_id": jobId, ...})
```

### Real-time log streaming via channels

Each deployment gets a `LogBus` — a buffered channel of `SSEEvent` structs. The pipeline writes to it as each step completes. The SSE HTTP handler ranges over that channel and flushes events to the client as they arrive:

```
Pipeline goroutine           LogBus (channel)          HTTP SSE handler
──────────────────────       ─────────────────          ─────────────────────
emit("cloning repo...")  →→  ch <- SSEEvent        →→  write + flush
emit("building image")   →→  ch <- SSEEvent        →→  write + flush
emit("deployed: :49152") →→  ch <- SSEEvent        →→  write + flush
close(ch)                                               range exits → Phase B
                             containerBus.Ch        →→  live docker logs...
```

When the build channel closes (`close(ch)`), the SSE handler uses that as a natural signal to transition to Phase B — ranging over a second channel that streams live Docker container logs. No polling, no timers.

### Non-blocking emit pattern

A slow or disconnected client never stalls the pipeline. If the log channel is full, the oldest event is dropped and replaced:

```go
select {
case logBus.Ch <- event:      // fast path — client keeping up
default:
    <-logBus.Ch               // drop oldest
    logBus.Ch <- event        // write new
}
```

The pipeline keeps moving regardless of what any client is doing.

### Channels as synchronisation

Log delivery between the pipeline goroutine and the HTTP handler requires no explicit mutexes. The channel is the synchronisation primitive — a safe, idiomatic handoff between concurrent goroutines.

---

## cgroup v2

Containerix uses **Linux cgroup v2** (the unified hierarchy), writing directly to the kernel interface under `/sys/fs/cgroup/`:

| File | Purpose |
|---|---|
| `cpu.max` | CPU quota in the format `<quota> <period>` (e.g. `50000 100000` = 0.5 cores) |
| `memory.max` | Memory limit in bytes |
| `cgroup.procs` | PID of the container's root process, written after container start |

cgroup v2 uses a **single unified hierarchy** — no separate subsystem mounts like the old v1 `/sys/fs/cgroup/memory/`, `/sys/fs/cgroup/cpu/` split. One directory per sandbox group, all controllers in one place.

> **Note:** cgroup v2 is the default on most modern Linux distributions (Ubuntu 22.04+, Fedora 31+, Debian 11+). You can verify: `stat -fc %T /sys/fs/cgroup` — should return `cgroup2fs`.

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.21+ (tested on 1.26) | — |
| [Docker](https://docs.docker.com/get-docker/) | 20.10+ | Daemon must be running |
| [Git](https://git-scm.com/) | Any | Used to clone target repos |
| [Nixpacks](https://nixpacks.com/docs/install) | Latest | Auto-build engine — see below |
| GCC / CGO | Any | `sudo apt install gcc` — required to compile `go-sqlite3` |
| Linux (cgroup v2) | Kernel 5.2+ | Required for cgroup v2 support |

### Installing Nixpacks

Nixpacks detects a project's language and runtime and builds a Docker image automatically — no `Dockerfile` needed.

```bash
# Linux
curl -sSL https://nixpacks.com/install.sh | bash

# Verify
nixpacks --version
```

Or via npm: `npm install -g @nixpacks/nixpacks`

> When a repo has no `Dockerfile`, Containerix calls `nixpacks build` automatically. It handles Node.js, Python, Go, Ruby, and more out of the box.

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/asim9115/containerix.git
cd containerix
```

### 2. Configure Environment

```bash
cp .env.example .env
```

```env
PORT="8080"
SANDBOX_NAME="containerix"
SANDBOX_CPU="2"
SANDBOX_MEMORY="3221225472"   # 3 GB in bytes
```

### 3. Install Go Dependencies

```bash
go mod download
```

> `go-sqlite3` uses CGO. Make sure `gcc` is installed: `sudo apt install gcc`

### 4. Build the Binary

```bash
go build -o server ./cmd/server
```

Or with make:

```bash
make build
```

### 5. Run as Root (Required for cgroup v2 access)

Containerix writes directly to `/sys/fs/cgroup/`, which requires root:

```bash
sudo ./server
```

Or:

```bash
make run   # builds then runs with sudo
```

You should see:

```
2026/07/31 00:00:00 Database initialized: data/containerix.db
2026/07/31 00:00:00 [GIN-debug] Listening and serving HTTP on :8080
```

---

## Makefile Targets

```bash
make build    # compile → ./server
make run      # build + sudo ./server
make deps     # go mod download
make tidy     # go mod tidy
make clean    # remove binary + tmp/
```

---

## API Reference

### `POST /build` — Deploy a Repository

Triggers the async pipeline. Returns immediately with a `job_id`.

```bash
curl -X POST http://localhost:8080/build \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/owner/repo", "tier": "tier1"}'
```

**Body:**
```json
{
  "url":  "https://github.com/owner/repo",
  "tier": "tier1",
  "env":  { "NODE_ENV": "production" }
}
```

| Field | Required | Values |
|---|---|---|
| `url` | ✅ | HTTPS GitHub URL only |
| `tier` | ❌ | `"tier1"` (default) or `"tier2"` |
| `env` | ❌ | Key-value pairs injected into the container |

**Response `202`:**
```json
{
  "job_id": "a1b2c3d4",
  "status": "queued",
  "logs":   "/containers/a1b2c3d4/logs"
}
```

---

### `GET /containers/:id/logs` — Stream Logs (SSE)

```bash
curl -N http://localhost:8080/containers/a1b2c3d4/logs
```

Streams in two phases:
- **Phase A** — Build logs (clone → build → port detect → start)
- **Phase B** — Live container stdout/stderr once running

| SSE Event | Data |
|---|---|
| `log` | Build step message |
| `deployed` | `http://localhost:<port>` — your live app URL |
| `error` | Error message |
| `done` | Container stopped |

---

### Other Endpoints

| Method | Route | Description |
|---|---|---|
| `GET` | `/jobs/:id` | Check job status, container ID, port, error |
| `GET` | `/jobs` | List all jobs |
| `GET` | `/containers` | List active containers |
| `GET` | `/containers/:id` | Get container details |
| `DELETE` | `/containers/:id` | Stop and remove a container |
| `GET` | `/cgroup` | View sandbox resource usage |
| `DELETE` | `/cgroup` | Destroy the sandbox cgroup |

---

## Resource Tiers

| Tier | CPU | Memory |
|---|---|---|
| `tier1` | 0.5 cores | 512 MB |
| `tier2` | 1.0 cores | 1 GB |

The global sandbox enforces a total resource budget set at startup. New deployments are rejected if the budget would be exceeded.

---

## Technology Stack

| Layer | Technology |
|---|---|
| **Language** | Go 1.26 |
| **HTTP Framework** | [Gin](https://github.com/gin-gonic/gin) |
| **Database** | SQLite via [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) (CGO) |
| **Container Engine** | Docker (CLI subprocess) |
| **Auto Build** | [Nixpacks](https://nixpacks.com/) |
| **Resource Isolation** | Linux cgroup v2 (`cpu.max`, `memory.max`) |
| **Unique IDs** | [google/uuid](https://github.com/google/uuid) |
| **Log Streaming** | Server-Sent Events (SSE) over HTTP |
| **Port Detection** | TCP scan via probe container |
| **Persistence Pattern** | Repository interfaces (swap-ready for PostgreSQL) |

---

## Database Schema

An embedded SQL migration (`internal/database/migrations/001_init_.sql`) runs automatically at startup:

- **`users`** — tenant table (seeded with a default `test` user for now)
- **`deployments`** — full lifecycle record: status, container ID, ports, image tag, env, error
- **`jobs`** — background task tracking: step, status, timestamps
- **`port_allocations`** — tracks which host ports are reserved to which containers

SQLite is opened in **WAL mode** with foreign keys enabled and a 5-second busy timeout for safe concurrent access.

---

## Project Structure

```
containerix/
├── cmd/server/main.go            # Entry point
├── router/router.go              # Gin route wiring
├── internal/
│   ├── api/                      # HTTP handlers, job store, SSE
│   ├── builder/                  # git clone, Nixpacks/Docker build
│   ├── cgroup/                   # cgroup v2 read/write
│   ├── container/                # Container lifecycle
│   ├── database/                 # SQLite init + embedded migrations
│   │   └── migrations/
│   ├── detector/                 # TCP port scanning
│   ├── docker/                   # Docker CLI wrappers
│   ├── pipeline/                 # Core deploy + startup sync
│   ├── ports/                    # In-memory port allocator
│   ├── repository/               # Interfaces + SQLite implementations
│   │   └── sqllite/
│   ├── sandbox/                  # CPU/memory budget manager
│   ├── state/                    # Global singleton
│   └── types/                    # Shared domain types
├── data/                         # SQLite DB (created at runtime)
├── tmp/                          # Cloned repos (created + cleaned at runtime)
├── go.mod / go.sum
├── Makefile
├── .env.example
└── README.md
```

---

## Common Issues

| Issue | Fix |
|---|---|
| `permission denied` on `/sys/fs/cgroup` | Run with `sudo ./server` |
| `cgo: C compiler not found` | `sudo apt install gcc` |
| `nixpacks: command not found` | `curl -sSL https://nixpacks.com/install.sh \| bash` |
| `docker: command not found` | Install Docker, ensure daemon is running |
| `git clone failed` | Ensure the repo is public and URL is correct |
| cgroup not found | Verify cgroup v2: `stat -fc %T /sys/fs/cgroup` → should be `cgroup2fs` |

---

## Roadmap

- [ ] User management and API key authentication
- [ ] Polished developer-facing API surface
- [ ] Web dashboard
- [ ] Custom domain routing via reverse proxy
- [ ] PostgreSQL backend (interfaces already designed for it)
- [ ] Webhook-triggered auto-redeploys on `git push`

---

## License

[MIT](./LICENSE)
