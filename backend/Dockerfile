# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/call-it-cure-it-api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/create-admin ./cmd/create-admin


FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates sqlite3 curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/call-it-cure-it-api /app/call-it-cure-it-api
COPY --from=builder /out/create-admin /app/create-admin
COPY migrations /app/migrations

RUN mkdir -p /app/data

ENV PORT=8080
ENV DATABASE_PATH=/app/data/app.db
ENV LLM_COACHING_ENABLED=false

EXPOSE 8080

CMD ["/app/call-it-cure-it-api"]