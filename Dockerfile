# syntax=docker/dockerfile:1

# ---------- frontend ----------
FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

# ---------- backend ----------
FROM golang:1.25-alpine AS go
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./internal/static/dist
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o /out/s3store ./cmd/s3store

# ---------- runtime ----------
FROM alpine:3.21
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
