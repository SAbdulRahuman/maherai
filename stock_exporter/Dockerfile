# ─── Node.js Build Stage (Next.js UI) ─────────────────────
FROM node:20-alpine AS ui-builder

WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# ─── Go Build Stage ──────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy the Next.js build output into the Go embed path
COPY --from=ui-builder /app/web/out/ ./internal/ui/static/

ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X main.version=${VERSION} \
      -X main.gitCommit=${GIT_COMMIT} \
      -X main.buildDate=${BUILD_DATE}" \
    -o stock_exporter ./cmd/

# ─── Runtime Stage ────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/stock_exporter /usr/local/bin/stock_exporter
COPY --from=builder /app/config.yaml /etc/stock_exporter/config.yaml
COPY --from=builder /app/config.tadawul.yaml /etc/stock_exporter/config.tadawul.yaml

EXPOSE 9101 9102

USER nonroot:nonroot

ENTRYPOINT ["stock_exporter"]
CMD ["serve", "--config", "/etc/stock_exporter/config.yaml"]