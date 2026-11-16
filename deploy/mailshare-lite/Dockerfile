# syntax=docker/dockerfile:1.7
# Separate image for IMAP aggregation and read-only guest sharing.
# The existing all-in-one image and deployment stay unchanged.
FROM --platform=$BUILDPLATFORM golang:1.27.3-bookworm AS api-build
WORKDIR /src/apps/api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api ./
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/lanqin-api ./cmd/server

FROM --platform=$BUILDPLATFORM node:20-bookworm-slim AS web-build
WORKDIR /src
COPY pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN corepack enable && corepack prepare pnpm@10.28.2 --activate
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --filter lanqin-email-web...
COPY apps/web apps/web
ENV VITE_MAILSHARE_LITE=true
RUN pnpm --dir apps/web run build

FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive \
    LANQIN_MAILSHARE_LITE=true \
    LANQIN_ADDR=127.0.0.1:8080 \
    LANQIN_DATA_DIR=/data \
    LANQIN_DB_PATH=/data/lanqin.db \
    LANQIN_DB_DRIVER=sqlite \
    LANQIN_MAILDIR_ROOT=/var/mail/vhosts
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates tzdata curl nginx sqlite3 && \
    rm -rf /var/lib/apt/lists/*
COPY --from=api-build /out/lanqin-api /usr/local/bin/lanqin-api
COPY --from=web-build /src/apps/web/dist /usr/share/nginx/html
COPY deploy/mailshare-lite/nginx.conf /etc/nginx/sites-enabled/default
COPY deploy/mailshare-lite/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh && mkdir -p /data /var/mail/vhosts /run/nginx
EXPOSE 80
HEALTHCHECK --interval=10s --timeout=5s --start-period=35s --retries=6 \
  CMD curl --fail --silent --show-error http://127.0.0.1/readyz >/dev/null || exit 1
ENTRYPOINT ["/entrypoint.sh"]
