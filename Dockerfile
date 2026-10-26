FROM node:22-bookworm-slim AS frontend-build

WORKDIR /src/nextduel-frontend

COPY NextDuel/nextduel-frontend/package*.json ./

RUN npm ci

COPY NextDuel/nextduel-frontend/ ./

RUN npm run build


FROM golang:1.25-bookworm AS server-build

WORKDIR /src

COPY nex-server/ /nex-server/

COPY NextDuel/nextduel-svr/go.mod NextDuel/nextduel-svr/go.sum /src/nextduel-svr/

WORKDIR /src/nextduel-svr

RUN go mod download

COPY NextDuel/nextduel-svr/ ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/nextduel-svr .


FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --shell /bin/bash nextduel

WORKDIR /app/nextduel-svr

COPY --from=server-build /out/nextduel-svr ./nextduel-svr

COPY --from=frontend-build /src/nextduel-frontend/dist ./htmls

RUN chown -R nextduel:nextduel /app

USER nextduel

ENV HTML_ROOT=./htmls

EXPOSE 8080

CMD ["./nextduel-svr"]