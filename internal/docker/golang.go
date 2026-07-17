package docker

import (
	"fmt"

	"github.com/asim9115/containerix/internal/detector"
)

func generateGo(d detector.DetectResult) (string, error) {
	version := d.Version
	if version == "" {
		version = "1.22"
	}

	port := d.Port
	if port == 0 {
		port = 8080
	}

	switch d.Framework {
	case "fiber":
		return generateGoFiber(version, port), nil
	case "gin":
		return generateGoGin(version, port), nil
	case "echo":
		return generateGoEcho(version, port), nil
	case "chi", "gorilla":
		return generateGoStdlib(version, port, d.Framework), nil
	default:
		return generateGoGeneric(version, port), nil
	}
}

// ---------------------------------------------------------------------------
// Generic multi-stage (distroless runtime) — used for plain Go / chi / gorilla
// ---------------------------------------------------------------------------

func generateGoGeneric(version string, port int) string {
	return fmt.Sprintf(`FROM golang:%s-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o app ./...

# Minimal runtime image
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/app /app

EXPOSE %d

ENTRYPOINT ["/app"]
`, version, port)
}

// ---------------------------------------------------------------------------
// Gin
// ---------------------------------------------------------------------------

func generateGoGin(version string, port int) string {
	return fmt.Sprintf(`FROM golang:%s-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o app ./...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/app /app

ENV GIN_MODE=release

EXPOSE %d

ENTRYPOINT ["/app"]
`, version, port)
}

// ---------------------------------------------------------------------------
// Echo
// ---------------------------------------------------------------------------

func generateGoEcho(version string, port int) string {
	return fmt.Sprintf(`FROM golang:%s-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o app ./...

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/app /app

EXPOSE %d

ENTRYPOINT ["/app"]
`, version, port)
}

// ---------------------------------------------------------------------------
// Fiber
// ---------------------------------------------------------------------------

func generateGoFiber(version string, port int) string {
	// Fiber uses fasthttp which requires a slightly heavier base image
	return fmt.Sprintf(`FROM golang:%s-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o app ./...

FROM alpine:3.20

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE %d

ENTRYPOINT ["./app"]
`, version, port)
}

// ---------------------------------------------------------------------------
// stdlib-based routers (chi, gorilla/mux)
// ---------------------------------------------------------------------------

func generateGoStdlib(version string, port int, framework string) string {
	_ = framework // reserved for future framework-specific ENV vars
	return generateGoGeneric(version, port)
}
