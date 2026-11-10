# syntax=docker/dockerfile:1
FROM node:24.22.1-bookworm-slim AS frontend
WORKDIR /src
# Optional local proxy CA is a build secret; it never enters the runtime image.
RUN --mount=type=secret,id=build-ca \
    if [ -f /run/secrets/build-ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/build-ca; fi; \
    npm install --global pnpm@12.18.0
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY backend/package.json ./backend/package.json
COPY api/ ./api/
COPY frontend/ ./frontend/
COPY scripts/prepare-miniapp-assets.mjs ./scripts/prepare-miniapp-assets.mjs
RUN --mount=type=secret,id=build-ca \
    if [ -f /run/secrets/build-ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/build-ca; fi; \
    pnpm install --frozen-lockfile
RUN pnpm --filter @repo/miniapp build && chmod -R a+rX frontend/miniapp/out

FROM golang:1.27.3 AS build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
COPY backend/cmd/ ./cmd/
COPY backend/internal/ ./internal/
RUN --mount=type=secret,id=build-ca \
    if [ -f /run/secrets/build-ca ]; then export SSL_CERT_FILE=/run/secrets/build-ca; fi; \
    CGO_ENABLED=0 go build -trimpath -o /server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o /richmenu ./cmd/richmenu

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /server /server
COPY --from=build /richmenu /richmenu
COPY --from=frontend /src/frontend/miniapp/out/ /app/frontend/
COPY --chmod=644 assets/richmenu/* /assets/richmenu/
ENV PORT=8080 FRONTEND_DIR=/app/frontend
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
