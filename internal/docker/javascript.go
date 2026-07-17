package docker

import (
	"fmt"

	"github.com/asim9115/containerix/internal/detector"
)

func generateNode(d detector.DetectResult) (string, error) {
	version := d.Version
	if version == "" {
		version = "20"
	}
	pm := d.PackageManager
	if pm == "" {
		pm = "npm"
	}

	switch d.Framework {
	case "nextjs":
		return generateNextJS(version, pm), nil
	case "react", "vite":
		return generateReact(version, pm), nil
	case "express":
		return generateExpress(version, pm, d), nil
	case "fastify":
		return generateFastify(version, pm, d), nil
	case "koa":
		return generateKoa(version, pm, d), nil
	default:
		return generatePlainNode(version, pm, d), nil
	}
}

// ---------------------------------------------------------------------------
// Next.js  (multi-stage, standalone output)
// ---------------------------------------------------------------------------

func generateNextJS(version, pm string) string {
	installCmd := pmInstall(pm)
	return fmt.Sprintf(`FROM node:%s-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN %s

FROM node:%s-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN %s run build

FROM node:%s-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
`, version, installCmd, version, pm, version)
}

// ---------------------------------------------------------------------------
// React / Vite  (static build served by nginx)
// ---------------------------------------------------------------------------

func generateReact(version, pm string) string {
	installCmd := pmInstall(pm)
	return fmt.Sprintf(`FROM node:%s-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN %s
COPY . .
RUN %s run build

FROM nginx:alpine AS runner
COPY --from=builder /app/dist /usr/share/nginx/html
COPY --from=builder /app/build /usr/share/nginx/html 2>/dev/null || true

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
`, version, installCmd, pm)
}

// ---------------------------------------------------------------------------
// Express
// ---------------------------------------------------------------------------

func generateExpress(version, pm string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 3000
	}
	installCmd := pmInstall(pm)
	return fmt.Sprintf(`FROM node:%s-alpine

WORKDIR /app

COPY package*.json ./
RUN %s --omit=dev

COPY . .

ENV NODE_ENV=production

EXPOSE %d

CMD ["node", "index.js"]
`, version, installCmd, port)
}

// ---------------------------------------------------------------------------
// Fastify
// ---------------------------------------------------------------------------

func generateFastify(version, pm string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 3000
	}
	installCmd := pmInstall(pm)
	return fmt.Sprintf(`FROM node:%s-alpine

WORKDIR /app

COPY package*.json ./
RUN %s --omit=dev

COPY . .

ENV NODE_ENV=production
ENV FASTIFY_PORT=%d

EXPOSE %d

CMD ["node", "index.js"]
`, version, installCmd, port, port)
}

// ---------------------------------------------------------------------------
// Koa
// ---------------------------------------------------------------------------

func generateKoa(version, pm string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 3000
	}
	installCmd := pmInstall(pm)
	return fmt.Sprintf(`FROM node:%s-alpine

WORKDIR /app

COPY package*.json ./
RUN %s --omit=dev

COPY . .

ENV NODE_ENV=production

EXPOSE %d

CMD ["node", "index.js"]
`, version, installCmd, port)
}

// ---------------------------------------------------------------------------
// Plain Node
// ---------------------------------------------------------------------------

func generatePlainNode(version, pm string, d detector.DetectResult) string {
	port := d.Port
	if port == 0 {
		port = 3000
	}
	installCmd := pmInstall(pm)
	entry := "index.js"
	if len(d.RunCommand) > 1 {
		entry = d.RunCommand[len(d.RunCommand)-1]
	}
	return fmt.Sprintf(`FROM node:%s-alpine

WORKDIR /app

COPY package*.json ./
RUN %s --omit=dev

COPY . .

ENV NODE_ENV=production

EXPOSE %d

CMD ["node", "%s"]
`, version, installCmd, port, entry)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// pmInstall returns the install command for the given package manager.
func pmInstall(pm string) string {
	switch pm {
	case "yarn":
		return "yarn install --frozen-lockfile"
	case "pnpm":
		return "pnpm install --frozen-lockfile"
	default:
		return "npm ci"
	}
}
