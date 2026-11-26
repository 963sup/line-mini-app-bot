# syntax=docker/dockerfile:1
# Update pinned digests with the corresponding runtime and package versions.
FROM node:24.22.1-bookworm-slim@sha256:be8204d4bd327363357f532187b752ff62759c31bedf98e9e18fc08c9cbd605c AS frontend
WORKDIR /src
ENV NEXT_TELEMETRY_DISABLED=1
RUN --mount=type=secret,id=build-ca \
    if [ -f /run/secrets/build-ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/build-ca; fi; \
    npm install --global pnpm@12.18.0
# Dependency layers depend on manifests, not application source.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY backend/package.json ./backend/package.json
COPY api/client/package.json ./api/client/package.json
COPY frontend/miniapp/package.json ./frontend/miniapp/package.json
RUN --mount=type=cache,id=miniapp-pnpm,target=/pnpm/store \
    --mount=type=secret,id=build-ca \
    if [ -f /run/secrets/build-ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/build-ca; fi; \
    pnpm install --frozen-lockfile --store-dir=/pnpm/store
COPY api/client/ ./api/client/
COPY frontend/ ./frontend/
COPY scripts/generate/prepare_miniapp_assets.mts ./scripts/generate/prepare_miniapp_assets.mts
RUN --mount=type=cache,id=miniapp-next,target=/src/frontend/miniapp/.next/cache \
    pnpm --filter @repo/miniapp build && chmod -R a+rX frontend/miniapp/out

FROM golang:1.27.4-bookworm@sha256:4fa7c968ba93f1635670522dbb01a2c78150443e1385746924566079f36570ea AS build
WORKDIR /src/backend
ENV CGO_ENABLED=0 GOTOOLCHAIN=local GOFLAGS=-mod=readonly
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,id=backend-modules,target=/go/pkg/mod \
    --mount=type=secret,id=build-ca \
    if [ -f /run/secrets/build-ca ]; then export SSL_CERT_FILE=/run/secrets/build-ca; fi; \
    go mod download
COPY backend/main.go ./main.go
COPY backend/internal/ ./internal/
RUN --mount=type=cache,id=backend-modules,target=/go/pkg/mod \
    --mount=type=cache,id=backend-build,target=/root/.cache/go-build \
    --mount=type=secret,id=build-ca \
    if [ -f /run/secrets/build-ca ]; then export SSL_CERT_FILE=/run/secrets/build-ca; fi; \
    go build -trimpath -o /server .

FROM debian:bookworm-slim@sha256:734a72b3d6142a71993b4bc580ca01d8ca687f77a2c422d96a961d3376b179bc
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates poppler-utils && rm -rf /var/lib/apt/lists/*
ARG VCS_REF=unknown
LABEL org.opencontainers.image.source="https://github.com/963sup/line-mini-app-bot" \
      org.opencontainers.image.revision="$VCS_REF"
COPY --from=build /server /server
COPY --from=frontend /src/frontend/miniapp/out/ /app/frontend/
# Image-only resources must not invalidate Go compilation.
# Pre-create traversable directories: COPY --chmod also affects new parents.
RUN install -d -m 755 /assets/line/richmenu
COPY --chmod=644 assets/line/richmenu/ /assets/line/richmenu/
# Keep the existing fallback until every deployment provides an explicit model.
ENV PORT=8080 FRONTEND_DIR=/app/frontend AI_DOCUMENT_MODEL=gpt-4.1-mini
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/server"]
