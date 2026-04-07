# ─── Build Stage ──────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
      -X main.version=${VERSION} \
      -X main.gitCommit=${GIT_COMMIT} \
      -X main.buildDate=${BUILD_DATE}" \
    -o stock_exporter ./cmd/main.go

# ─── Runtime Stage ────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/stock_exporter /usr/local/bin/stock_exporter
COPY --from=builder /app/config.yaml /etc/stock_exporter/config.yaml

EXPOSE 9101

USER nonroot:nonroot

ENTRYPOINT ["stock_exporter"]
CMD ["--config", "/etc/stock_exporter/config.yaml"]