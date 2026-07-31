# syntax=docker/dockerfile:1

ARG VERSION=dev

# ---------- frontend ----------
FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ ./
RUN npm run build

# ---------- backend ----------
FROM golang:1.25-alpine AS go
ARG VERSION=dev
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./internal/static/dist
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/s3store ./cmd/s3store

# ---------- runtime ----------
FROM alpine:3.21
ARG VERSION=dev
LABEL org.opencontainers.image.title="S3 Store" \
      org.opencontainers.image.description="Self-hosted S3 / Cloudflare R2 web console" \
      org.opencontainers.image.source="https://github.com/hakuzero4/S3_store_gui" \
      org.opencontainers.image.version="${VERSION}"
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 10001 s3store
WORKDIR /data
COPY --from=go /out/s3store /usr/local/bin/s3store
USER s3store
EXPOSE 17890
ENV S3STORE_ADDR=0.0.0.0:17890 \
    S3STORE_CONFIG=/data/config.json \
    S3STORE_NO_BROWSER=1 \
    TZ=Asia/Shanghai
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/s3store"]
CMD ["-addr", "0.0.0.0:17890", "-config", "/data/config.json", "-no-browser"]
