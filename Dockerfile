# syntax=docker/dockerfile:1
#
# Build context is the PARENT directory of this repo, because nextduel-svr
# depends on the sibling nex-server module (go.mod: replace => ../../nex-server):
#
#   Projects/
#     NextDuel/     <- this repo
#     nex-server/   <- github.com/8xNehanSS/NEX-Server-Framework
#
#   cd Projects && docker build -f NextDuel/Dockerfile -t nextduel:latest .
#
# Dockerfile.dockerignore (next to this file) keeps the context small.

FROM node:22-bookworm-slim AS frontend-build
WORKDIR /src/nextduel-frontend
COPY NextDuel/nextduel-frontend/package*.json ./
RUN npm ci
COPY NextDuel/nextduel-frontend/ ./
RUN npm run build


FROM golang:1.25-bookworm AS server-build
ARG VERSION=dev
COPY nex-server/ /nex-server/
WORKDIR /src/nextduel-svr
COPY NextDuel/nextduel-svr/go.mod NextDuel/nextduel-svr/go.sum ./
RUN go mod download
COPY NextDuel/nextduel-svr/ ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/nextduel-svr .


FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --shell /usr/sbin/nologin nextduel

WORKDIR /app/nextduel-svr

COPY --from=server-build /out/nextduel-svr ./nextduel-svr
COPY --from=frontend-build /src/nextduel-frontend/dist ./htmls

# data/ holds the SQLite database; sessions/ holds nex-server's login tokens
# (the path is fixed by the framework). Both must be volumes to survive
# redeploys.
RUN mkdir -p data sessions && chown -R nextduel:nextduel /app

USER nextduel

ENV PORT=8080 \
    HTML_ROOT=./htmls \
    DB_PATH=/app/nextduel-svr/data/nextduel.db

VOLUME ["/app/nextduel-svr/data", "/app/nextduel-svr/sessions"]

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["./nextduel-svr", "healthcheck"]

CMD ["./nextduel-svr"]
