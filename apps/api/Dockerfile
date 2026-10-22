# syntax=docker/dockerfile:1
# Build context must be the repository root (the API depends on ./plugins).

FROM golang:1.25-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOWORK=off

# Cache module downloads separately from source
COPY plugins/go.mod plugins/go.sum* ./plugins/
COPY apps/api/go.mod apps/api/go.sum ./apps/api/
RUN cd apps/api && go mod download

COPY plugins ./plugins
COPY apps/api ./apps/api
RUN cd apps/api && go build -trimpath -ldflags="-s -w" -o /out/dboke-api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S dboke && adduser -S -G dboke dboke

# The API resolves plugin directories as ../plugins relative to its working
# directory, so run from /app/api and keep installed plugins in /app/plugins.
WORKDIR /app/api
COPY --from=build /out/dboke-api /app/api/dboke-api
RUN mkdir -p /app/plugins && chown -R dboke:dboke /app
VOLUME /app/plugins

USER dboke
ENV PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1
ENTRYPOINT ["/app/api/dboke-api"]
