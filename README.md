# Containerix
## 🚀 Getting Started

### Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| [Go](https://go.dev/dl/) | 1.21+ | Tested on Go 1.26 |
| [Docker](https://docs.docker.com/get-docker/) | 20.10+ | Must be running |
| [Git](https://git-scm.com/) | Any | Used to clone target repos |

### 1. Clone the Repository

```bash
git clone https://github.com/asim9115/containerix.git
cd containerix
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Open `.env` and set your desired port (defaults to `8080` if omitted):

```env
PORT="8080"
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run the Server

```bash
go run ./cmd/server
```

You should see:

```
2026/06/25 00:00:00 starting containerix on port 8080
```

### 5. Build a Binary (Optional)

```bash
go build -o containerix ./cmd/server
./containerix
```
