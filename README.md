# Containerix

> A self-hosted, lightweight Platform-as-a-Service (PaaS) engine — deploy any GitHub repository into an isolated, resource-controlled Docker container with a single HTTP request.

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat-square&logo=docker)](https://docs.docker.com/get-docker/)

---

## What is Containerix?

Containerix is a self-hosted PaaS engine built from scratch in Go. Give it a GitHub URL and it handles the rest — clones the repo, builds a Docker image (auto-generated Dockerfile for Node, Python, or Go; or uses an existing one), runs the container inside a **Linux cgroup v2** sandbox for kernel-level resource enforcement, and streams real-time build logs back over SSE.

Think Railway or Render, but something you own and run yourself.

**Status:** Core pipeline is functional end-to-end — user registration, API key auth, async deploy pipeline, cgroup v2 isolation, SSE log streaming, and startup reconciliation are all working. Custom domain routing and webhook-triggered redeploys are next.

---

## Key Features

| Feature | Detail |
|---|---|
| 🚀 **Async one-request deploys** | `POST /build` returns `202 Accepted` immediately with a `job_id`; the full pipeline runs in a background goroutine |
| 🔑 **API key authentication** | Every user gets a `ctx-` prefixed API key on registration; keys are SHA-256 hashed before storage, never stored raw |
| 🔑 **API key rotation** | `POST /users/api-key` invalidates the old key and issues a new one atomically |
| 🔒 **cgroup v2 resource isolation** | Each sandbox gets a named cgroup under `/sys/fs/cgroup/`; every container's PID is pinned to it after launch, enforcing CPU quota and memory hard limits at kernel level |
| 🛡️ **Docker security hardening** | Containers run with `--pids-limit`, `--memory-swap`, `--security-opt no-new-privileges`, and `CAP_DROP=ALL` per tier |
| 📡 **Three-phase SSE log streaming** | Phase 1: live build logs via buffered channel. Phase 2: job status polling until terminal state. Phase 3: live Docker stdout/stderr scoped to the client's request context |
| 🧱 **Shell injection prevention** | Build/start commands validated against a deny-list of shell metacharacters (`&&`, `||`, `;`, `|`, `$()`, etc.) |
| 🌐 **URL injection prevention** | Only `https://github.com/owner/repo` URLs accepted; scheme, host, credentials, query params, and fragments are all validated |
| 💾 **SQLite persistence (WAL mode)** | All users, deployments, jobs, and port allocations persisted across restarts; WAL mode + FK enforcement + 5 s busy timeout |
| 🔄 **Startup reconciliation** | On restart, cross-references live PIDs in `cgroup.procs` against the DB — stops orphaned host containers, marks ghost DB records stopped, and re-inserts missing port-allocation rows |
| 🚦 **Rate limiting** | Per-IP sliding-window rate limiter for all routes (default: 120 req/min); stricter limit on `POST /users` (default: 5 req/hour) |
| 📏 **Request size cap** | Configurable max request body (default: 1 MB) |
| ❤️ **Health & readiness endpoints** | `/health` (liveness) and `/ready` (DB ping) excluded from rate limiting for orchestrator probes |
| 🧩 **Repository pattern** | Interface-driven DB layer — swap SQLite for PostgreSQL without touching any business logic |
| 🗂️ **Graceful shutdown** | SIGINT/SIGTERM caught; server drains with a 30-second timeout before exit |

---

## Architecture Overview

```
HTTP Client
    │  POST /build  (with X-API-Key header)
    ▼
Gin Router  →  APIKeyAuth middleware  →  CreateDockerImage handler
    │
    │  202 Accepted  {job_id, logs_url}   ◄── returned immediately
    │
    └── goroutine spawned ──────────────────────────────────────────┐
                                                                    │
        Deploy Pipeline (internal/pipeline/deploy.go)              │
         1. Check sandbox resource budget                           │
         2. Validate GitHub URL (scheme, host, path, injection)     │
         3. Validate build request (shell metachar check)           │
         4. git clone --depth 1  →  tmp/<uuid>/                    │
         5. Generate Dockerfile (Node/Python/Go) or use existing    │
         6. docker build  →  image tag ctx-<uuid>                   │
         7. Allocate free host port (40000–50000 range)             │
         8. docker run (port map, env, CPU, memory, pids-limit,     │
                        no-new-privileges, CAP_DROP=ALL)            │
         9. docker inspect → get container PID                      │
        10. Write PID to /sys/fs/cgroup/<name>/cgroup.procs         │
        11. Update DB record  (status: running)                     │
        12. Emit SSE "deployed" event  →  live URL                  │
        13. Delete build image                                       │
                                                                    │
    ┌───────────────┬───────────────────────────────────────────────┘
    ▼               ▼
SQLite          Linux cgroup v2
data/           /sys/fs/cgroup/<name>/
containerix.db  ├── cpu.max        (quota / period)
WAL mode        ├── memory.max     (bytes)
FK enabled      └── cgroup.procs   (container PID)
```

### SSE Log Stream Flow

```
Pipeline goroutine         LogBus (chan SSEEvent, 256)    HTTP handler (/containers/:id/logs)
──────────────────         ──────────────────────────     ──────────────────────────────────
emit("cloning...")    →→   ch <- SSEEvent            →→   write + flush  (Phase 1)
emit("building...")   →→   ch <- SSEEvent            →→   write + flush
emit("deployed:...")  →→   ch <- SSEEvent            →→   write + flush
close(ch)                                                 phase 1 loop exits

                                                          poll job status every 500 ms (Phase 2)
                                                          ↓ job.Status == completed

                                                          docker logs -f --tail 100 ...  (Phase 3)
                                                          line-by-line → write + flush
                                                          (stops on client disconnect via ctx)
```

**Non-blocking emit:** if the channel is full, the oldest event is tail-dropped and replaced — a slow client never stalls the pipeline.

---

## Package Breakdown

| Package | Responsibility |
|---|---|
| `cmd/server/` | Entry point — init DB, repos, cgroup sandbox, pipeline, reconcile, HTTP server, graceful shutdown |
| `router/` | Gin route wiring, middleware composition |
| `internal/api/` | HTTP handlers (deploy, containers, jobs, deployments, users, SSE logs, health, readiness) |
| `internal/pipeline/` | Core deploy flow, container stop/delete/stop-all, startup sync (`SyncData`) |
| `internal/builder/` | `git clone`, Dockerfile generation (Node/Python/Go), `docker build`, URL + command validation |
| `internal/detector/` | TCP port scanner — probes 17 common framework ports with retry logic |
| `internal/docker/` | Docker CLI wrappers: run, stop, start, delete, inspect PID, get IP, stream logs, force-remove, container ID from PID via `/proc/<pid>/cgroup` |
| `internal/container/` | Container lifecycle (run, stop, start, delete, stop-all) |
| `internal/cgroup/` | cgroup v2 read/write: `Create`, `Update`, `AddProcess` (write PID), `GetProcesses` (read procs), `Destroy` |
| `internal/sandbox/` | In-memory CPU/memory budget manager with mutex; `CanAllocate`, `Allocate`, `Release`, `AddContainer`, `RemoveContainer` |
| `internal/ports/` | In-memory free-port allocator (mutex-protected map, ports 40000–50000); TCP-probes the OS before reserving |
| `internal/auth/` | API key generation (`crypto/rand`, 32 bytes, `ctx-` prefix) and SHA-256 hashing |
| `internal/middleware/` | `APIKeyAuth` (hash → DB lookup → set user_id), `GlobalRateLimit` (per-IP sliding window), `RegistrationRateLimit`, `MaxBody`, `RequestLogger` |
| `internal/config/` | Env-var driven config with typed defaults |
| `internal/state/` | Global singleton — sandbox manager + port manager |
| `internal/database/` | SQLite init, WAL mode, FK enforcement, 5 s busy timeout, embedded SQL migrations |
| `internal/repository/` | Interface definitions for all four repos (swap-ready for Postgres) |
| `internal/repository/sqllite/` | SQLite implementations of all repo interfaces |
| `internal/types/` | Shared domain types: `Config`, `Tier`, `LogBus`, `SSEEvent`, `BuildRequest`, `Container`, `Sandbox` |

---

## Resource Tiers

| Tier | CPU | Memory | PID Limit | Privileges |
|---|---|---|---|---|
| `tier1` | 0.2 cores | 500 MB | 100 | `CAP_DROP=ALL`, `no-new-privileges` |
| `tier2` | 0.5 cores | 750 MB | 150 | `CAP_DROP=ALL`, `no-new-privileges` |

Limits are enforced at two layers: Docker's `--cpus`, `--memory`, `--pids-limit` flags, **and** the Linux cgroup v2 `cpu.max`/`memory.max` files written directly after container start. The global sandbox tracks total allocated CPU and memory — new deploys are rejected if the budget would be exceeded.

---

## cgroup v2

Containerix uses **Linux cgroup v2** (the unified hierarchy), writing directly to the kernel interface under `/sys/fs/cgroup/`:

| File | What Containerix writes |
|---|---|
| `cpu.max` | `<quota> 100000` — e.g. `20000 100000` = 0.2 cores |
| `memory.max` | Hard memory limit in bytes |
| `cgroup.procs` | Container PID written here after `docker run`; kernel auto-removes it on container exit |

> **Note:** cgroup v2 is the default on Ubuntu 22.04+, Fedora 31+, Debian 11+. Verify: `stat -fc %T /sys/fs/cgroup` — should return `cgroup2fs`.

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| [Go](https://go.dev/dl/) | 1.21+ | Tested on 1.26 |
| [Docker](https://docs.docker.com/get-docker/) | 20.10+ | Daemon must be running |
| [Git](https://git-scm.com/) | Any | Used to clone target repos |
| GCC / CGO | Any | `sudo apt install gcc` — required to compile `go-sqlite3` |
| Linux (cgroup v2) | Kernel 5.2+ | Required for cgroup v2; run with `sudo` for `/sys/fs/cgroup` access |

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

All config is via environment variables (see [Configuration](#configuration) below). The defaults work out of the box for local development.

### 3. Install Go Dependencies

```bash
go mod download
```

> `go-sqlite3` uses CGO. Make sure `gcc` is installed: `sudo apt install gcc`

### 4. Build

```bash
make build
# or
go build -o server ./cmd/server
```

### 5. Run (root required for cgroup v2)

```bash
sudo ./server
# or
make run   # builds then runs with sudo
```

You should see:

```
Initializing Sandbox
Name: containerix
CPU: 2
RAM: 3221225472
[sync] Running sync data
server listening on :8080
```

---

## Quick Start: Deploy a Repo

```bash
# 1. Create a user account
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'
# → returns {"api_key": "ctx-...", "warning": "save this API key now.."}

# 2. Deploy a repository (use your api_key from step 1)
curl -X POST http://localhost:8080/build \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ctx-..." \
  -d '{
    "url": "https://github.com/owner/repo",
    "language": "node",
    "start_command": "node index.js",
    "port": 3000,
    "tier": "tier1"
  }'
# → returns {"job_id": "a1b2c3d4", "status": "queued", "logs": "/containers/a1b2c3d4/logs"}

# 3. Stream build + runtime logs
curl -N http://localhost:8080/containers/a1b2c3d4/logs \
  -H "X-API-Key: ctx-..."
# → SSE stream: cloning... building... deployed: http://localhost:49152
```

---

## API Reference

All routes under `/` (except `POST /users`, `/health`, `/ready`) require authentication via `X-API-Key: <key>` or `Authorization: Bearer <key>` header.

### User Management

| Method | Route | Auth | Description |
|---|---|---|---|
| `POST` | `/users` | ❌ | Register a new user. Returns a one-time raw API key. |
| `GET` | `/users/me` | ✅ | Get your own profile (id, name, email, created_at). |
| `POST` | `/users/api-key` | ✅ | Rotate your API key. Old key is immediately invalidated. |

**Register:**
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'
```
```json
{
  "user_id": "550e8400-...",
  "email": "alice@example.com",
  "api_key": "ctx-3f2a...",
  "warning": "save this API key now.."
}
```

---

### Deploy

#### `POST /build` — Deploy a Repository

Triggers the async pipeline. Returns immediately with a `job_id`.

```bash
curl -X POST http://localhost:8080/build \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ctx-..." \
  -d '{
    "url": "https://github.com/owner/repo",
    "language": "node",
    "start_command": "node index.js",
    "port": 3000,
    "tier": "tier1",
    "env": { "NODE_ENV": "production" }
  }'
```

**Request body:**

| Field | Required | Type | Description |
|---|---|---|---|
| `url` | ✅ | string | HTTPS GitHub URL (`https://github.com/owner/repo`) |
| `language` | ✅ | string | `"node"`, `"python"`, `"go"`, or `"docker"` (uses existing Dockerfile) |
| `start_command` | ✅ | string | Command to start the app (e.g. `"node index.js"`, `"python app.py"`) |
| `port` | ✅ | int | Port your app listens on inside the container |
| `tier` | ❌ | string | `"tier1"` (default) or `"tier2"` |
| `build_command` | ❌ | string | Optional build step (e.g. `"npm install"`) — required for `node` and `go` |
| `root_directory` | ❌ | string | Subdirectory within the repo to build from |
| `env` | ❌ | object | Key-value environment variables injected into the container |

> **Note:** `PORT` and `HOST=0.0.0.0` are always injected by the platform, overriding any user-provided `PORT`.

**Response `202 Accepted`:**
```json
{
  "job_id": "a1b2c3d4",
  "status": "queued",
  "logs": "/containers/a1b2c3d4/logs"
}
```

---

### Jobs

| Method | Route | Description |
|---|---|---|
| `GET` | `/jobs` | List all your jobs |
| `GET` | `/jobs/:id` | Get job status, step, container ID, host port, error, timestamps |

**Job statuses:** `queued` → `building` → `completed` / `failed`

**Job response:**
```json
{
  "job_id": "a1b2c3d4",
  "deployment_id": "a1b2c3d4",
  "status": "completed",
  "step": "starting container",
  "error": "",
  "created_at": "2026-09-13T10:00:00Z",
  "completed_at": "2026-09-13T10:02:30Z",
  "logs": "/containers/a1b2c3d4/logs"
}
```

---

### Log Streaming

#### `GET /containers/:id/logs` — Stream Logs (SSE)

```bash
curl -N http://localhost:8080/containers/a1b2c3d4/logs \
  -H "X-API-Key: ctx-..."
```

Three sequential phases on a single HTTP connection:

| Phase | What streams |
|---|---|
| 1 — Build logs | Live pipeline steps: cloning, building, starting |
| 2 — Job resolution | Polls until job reaches `completed` or `failed` |
| 3 — Container logs | Live `docker logs -f` output (last 100 lines + follow) |

| SSE Event | Data |
|---|---|
| `log` | Build step or container log line |
| `deployed` | `http://localhost:<port>` — your live app URL |
| `error` | Error message (build failure or container error) |
| `done` | Stream ended (container stopped or client disconnect) |

---

### Containers

| Method | Route | Description |
|---|---|---|
| `GET` | `/containers` | List all your containers (running and stopped) |
| `GET` | `/containers/:id` | Get container details |
| `POST` | `/containers/:id/stop` | Stop a running container |
| `POST` | `/containers/stop-all` | Stop all your running containers |
| `DELETE` | `/containers/:id` | Stop, remove container, free port, release resources, delete from DB |

---

### Deployments

| Method | Route | Description |
|---|---|---|
| `GET` | `/deployments` | List all your deployments |
| `GET` | `/deployments/:id` | Get deployment details |
| `DELETE` | `/deployments/:id` | Stop and delete a deployment (same as `DELETE /containers/:id`) |

---

### System

| Method | Route | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | ❌ | Liveness probe — always `{"status": "ok"}` |
| `GET` | `/ready` | ❌ | Readiness probe — checks DB connectivity |
| `GET` | `/cgroup` | ❌ | View current sandbox cgroup state (CPU, memory, containers) |
| `DELETE` | `/cgroup` | ❌ | Destroy the sandbox cgroup |
| `GET` | `/dbports` | ❌ | View all port allocations in DB |

---

## Configuration

All configuration is via environment variables. Every variable has a sensible default.

| Variable | Default | Description |
|---|---|---|
| `CONTAINERIX_LISTEN` | `:8080` | Address and port to listen on |
| `CONTAINERIX_DB_PATH` | `data/containerix.db` | SQLite database file path |
| `CONTAINERIX_SANDBOX_NAME` | `containerix` | cgroup v2 group name under `/sys/fs/cgroup/` |
| `CONTAINERIX_SANDBOX_CPU` | `2` | Total CPU cores available across all containers |
| `CONTAINERIX_SANDBOX_MEMORY` | `3221225472` | Total memory budget in bytes (default: 3 GB) |
| `CONTAINERIX_ALLOW_REGISTRATION` | `true` | Set to `false` to disable `POST /users` |
| `CONTAINERIX_MAX_REQUEST_BODY` | `1048576` | Max request body size in bytes (default: 1 MB) |
| `CONTAINERIX_GLOBAL_RATE_LIMIT` | `120` | Max requests per IP per window (0 = disabled) |
| `CONTAINERIX_GLOBAL_RATE_WINDOW` | `60` | Global rate limit window in seconds |
| `CONTAINERIX_REGISTRATION_RATE_LIMIT` | `5` | Max registration attempts per IP per window |
| `CONTAINERIX_REGISTRATION_RATE_WINDOW` | `3600` | Registration rate limit window in seconds |

---

## Database Schema

An embedded SQL migration (`internal/database/migrations/001_init_.sql`) runs automatically at startup:

- **`users`** — name, email, hashed API key, created_at
- **`deployments`** — full lifecycle: status, container ID, image tag, host port, container port, tier CPU/memory, env JSON, error message
- **`jobs`** — background task tracking: status, step, error, created_at, completed_at
- **`port_allocations`** — host port ↔ container ID mapping, allocated_at

SQLite is opened in WAL mode with foreign keys enabled and a 5-second busy timeout for safe concurrent access.

---

## Startup Reconciliation

On every startup, `SyncData()` runs before accepting any traffic:

1. Reads live PIDs from `cgroup.procs`
2. Resolves each PID to a container name via `/proc/<pid>/cgroup` (64-char Docker ID extraction) with a Docker API fallback
3. **Orphan cleanup:** containers alive on the host but absent from DB → stopped
4. **Ghost cleanup:** deployments marked `running` in DB but not alive on host → marked `stopped`, port freed
5. **Port table repair:** re-inserts missing port-allocation rows for live containers; removes stale rows for dead containers
6. Returns a snapshot so `main.go` can restore the in-memory sandbox budget and port manager without a second DB query

---

## How Go Concurrency Powers the Pipeline

### Non-blocking deploys via goroutines

```go
// handler returns 202 immediately
go func() {
    containerID, err := h.Pipeline.Deploy(userID, jobId, logBus, &body)
    // updates job state in DB when done
}()
c.JSON(http.StatusAccepted, gin.H{"job_id": jobId, ...})
```

### Buffered channel as log bus

```go
type LogBus struct {
    Ch chan SSEEvent   // buffered, capacity 256
}
```

The pipeline goroutine emits to the channel; the SSE handler ranges over it. When the build finishes, `close(ch)` signals the handler to move to Phase 3.

### Non-blocking emit (tail-drop)

```go
select {
case logBus.Ch <- event:      // fast path
default:
    <-logBus.Ch               // drop oldest event
    logBus.Ch <- event        // write new event
}
```

A disconnected or slow client never stalls the build.

### Mutex-free log registry

`BuildLogRegistry` uses `sync.RWMutex` to let multiple SSE clients concurrently look up the same log bus without a bottleneck.

---

## Project Structure

```
containerix/
├── cmd/server/main.go              # Entry point
├── router/router.go                # Gin route wiring
├── internal/
│   ├── api/                        # HTTP handlers, BuildLogRegistry, SSE streaming
│   │   ├── handler.go              # POST /build, GET /jobs
│   │   ├── logs.go                 # GET /containers/:id/logs (3-phase SSE)
│   │   ├── containers.go           # Container CRUD + stop
│   │   ├── deployments.go          # Deployment CRUD
│   │   ├── user.go                 # User registration, GetMe, RotateAPIKey
│   │   ├── state.go                # GlobalState, /health, /ready
│   │   ├── ports.go                # GET /dbports
│   │   └── store.go                # BuildLogRegistry (in-memory log bus map)
│   ├── pipeline/
│   │   ├── deploy.go               # Core async deploy flow (steps 1–13)
│   │   ├── container.go            # DeleteContainer, StopContainer, StopAllContainers
│   │   ├── sync.go                 # SyncData — startup reconciliation
│   │   └── state.go                # Pipeline State struct
│   ├── builder/
│   │   ├── builder.go              # git clone (--depth 1)
│   │   ├── dockerimage.go          # Dockerfile generation + docker build
│   │   ├── validate.go             # Shell metachar + build request validation
│   │   └── templates/              # Node, Python, Go Dockerfile templates
│   ├── detector/
│   │   └── detector.go             # TCP port scanner (17 common ports, 5 retries)
│   ├── docker/
│   │   └── docker.go               # Docker CLI wrappers + Moby API client
│   ├── container/
│   │   └── containermanager.go     # Run, Stop, Start, Delete, StopAll
│   ├── cgroup/
│   │   ├── create.go               # mkdir + write cpu.max + memory.max
│   │   ├── addprocess.go           # Write PID to cgroup.procs
│   │   ├── getprocess.go           # Read PIDs from cgroup.procs
│   │   ├── update.go               # Update cpu.max + memory.max
│   │   └── destroy.go              # Remove cgroup directory
│   ├── sandbox/
│   │   ├── init.go                 # Create cgroup + init SandboxManager
│   │   ├── resource.go             # CanAllocate, Allocate, Release
│   │   ├── config.go               # AddContainer, RemoveContainer
│   │   ├── update.go               # Update sandbox limits
│   │   └── stats.go                # GetState, Stats, Remaining
│   ├── ports/
│   │   └── manager.go              # Free-port allocator (40000–50000, mutex-protected)
│   ├── auth/
│   │   └── auth.go                 # GenerateAPIKey (crypto/rand), HashApiKey (SHA-256)
│   ├── middleware/
│   │   ├── auth.go                 # APIKeyAuth (hash lookup, sets user_id in context)
│   │   ├── ratelimit.go            # GlobalRateLimit, RegistrationRateLimit (sliding window)
│   │   ├── logger.go               # RequestLogger
│   │   └── request_size.go         # MaxBody
│   ├── config/
│   │   └── config.go               # Env-var driven config with typed defaults
│   ├── state/
│   │   └── state.go                # Global SB singleton (sandbox + port manager)
│   ├── database/
│   │   ├── database.go             # SQLite init, WAL, FK, busy timeout
│   │   └── migrations/001_init_.sql
│   ├── repository/
│   │   ├── interfaces.go           # DeploymentRepo, JobRepo, PortsRepo, UserRepo interfaces
│   │   └── sqllite/                # SQLite implementations of all interfaces
│   └── types/
│       ├── types.go                # Config, Tier, LogBus, SSEEvent, BuildRequest, Container, Sandbox
│       ├── status.go               # Deploy/job status constants
│       └── helper.go               # MemoryToBytes helper
├── frontend/                       # React + Vite dashboard (MVP)
├── data/                           # SQLite DB (created at runtime)
├── tmp/                            # Cloned repos (created + auto-cleaned after each build)
├── go.mod / go.sum
├── Makefile
├── .env.example
└── README.md
```

---

## Makefile

```bash
make build    # compile → ./server
make run      # build + sudo ./server
make deps     # go mod download
make tidy     # go mod tidy
make clean    # remove binary + tmp/
```

---

## Common Issues

| Issue | Fix |
|---|---|
| `permission denied` on `/sys/fs/cgroup` | Run with `sudo ./server` |
| `cgo: C compiler not found` | `sudo apt install gcc` |
| `docker: command not found` | Install Docker, ensure daemon is running |
| `git clone failed` | Ensure the repo is public and the URL is a valid `https://github.com/owner/repo` |
| cgroup not found | Verify cgroup v2: `stat -fc %T /sys/fs/cgroup` → should be `cgroup2fs` |
| Port scan timeout | App may take >12 s to start; the pipeline falls back to `DefaultAppPort` (10000) |

---

## Roadmap

- [ ] Git webhook auto-redeploy on `git push` to `main`
- [ ] Custom domain routing via reverse proxy (Caddy)
- [ ] PostgreSQL backend (repository interfaces already designed for it)
- [ ] Readiness probe (currently a fixed 5 s sleep — `internal/readiness` is written, pending re-wire)
- [ ] Admin API (view all users, all containers, sandbox stats)
- [ ] Deploy from private repos (SSH key or GitHub token support)

---

## License

[MIT](./LICENSE)
