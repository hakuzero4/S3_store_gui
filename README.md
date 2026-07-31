# S3 Store

**A lightweight, self-hosted web console for S3-compatible object storage.**

Manage **Cloudflare R2**, **AWS S3**, **MinIO**, and other S3 API endpoints from a clean desktop-style UI ? packaged as a **single binary** or a **Docker image**.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

**Languages:** [English](./README.md) ? [????](./README.zh-CN.md) ? [????](./README.zh-TW.md) ? [???](./README.ja.md) ? [???](./README.ko.md)

---

## Why S3 Store?

- **Local-first** admin UI ? no SaaS dependency for the app itself
- First-class **Cloudflare R2** support
- **One executable** with embedded frontend, or a small Docker image
- Credentials stay on your machine / volume

---

## Features

- Multiple connection profiles (R2 / S3 / MinIO / custom)
- Bucket list / create / delete; switch buckets per profile
- Object browser with prefixes, search, breadcrumbs
- Upload (button + drag-drop), download, bulk delete, rename
- Folder placeholders, object details, presigned URLs
- In-app image preview
- **i18n:** English (default), Simplified Chinese, Traditional Chinese, Japanese, Korean

---

## Quick start

### Docker

```bash
git clone https://github.com/hakuzero4/S3_store_gui.git
cd S3_store_gui
docker compose up -d --build
```

Open http://127.0.0.1:17890 ? config volume: `/data/config.json`.

### Binary (Windows)

```powershell
.\build.ps1
.\dist\s3store.exe
```

### Binary (Linux / macOS)

```bash
cd web && npm ci && npm run build && cd ..
rm -rf internal/static/dist && cp -R web/dist internal/static/dist
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.1.0" -o dist/s3store ./cmd/s3store
./dist/s3store -no-browser
```

---

## Cloudflare R2

| Field | Value |
|-------|--------|
| Endpoint | Dashboard ? R2 ? **S3 API** URL |
| Region | `auto` |
| Access / Secret | R2 **Account API token** key pair |
| Force Path Style | Off |

Do not use the `cfat_?` Cloudflare API token value as S3 keys.
Do not use a public custom domain as the S3 endpoint.

Same connection ? many buckets (dropdown). Different accounts ? multiple profiles under **Connections**.

---

## i18n (UI)

Default language: **English**.

Switch language in the sidebar (**Language**). Choice is saved in `localStorage` (`s3store.locale`).

Supported: `en` ? `zh-CN` ? `zh-TW` ? `ja` ? `ko`

Locale files: `web/src/i18n/locales/*.json`

---

## Configuration

| Flag / Env | Description |
|------------|-------------|
| `-addr` / `S3STORE_ADDR` | Listen address (Docker: `0.0.0.0:17890`) |
| `-config` / `S3STORE_CONFIG` | Path to `config.json` |
| `-no-browser` / `S3STORE_NO_BROWSER` | Do not open browser |
| `-version` | Print version |

---

## Development

```bash
go run ./cmd/s3store -addr 127.0.0.1:17890 -no-browser
cd web && npm install && npm run dev
```

UI: http://127.0.0.1:5173 (API proxied to `:17890`)

---

## Project layout

```text
cmd/s3store/          entry
internal/api/         HTTP API
internal/config/      config.json
internal/s3client/    S3/R2 client
internal/static/      embedded web dist
web/                  Vue 3 + vue-i18n + Naive UI
web/src/i18n/         locale messages
Dockerfile
docker-compose.yml
```

---

## Security

Intended for **local / private network** use. Secrets are stored in plain JSON ? protect the host and avoid exposing the port without TLS and auth.

---

---

## CI / CD (GitHub Actions)

| Workflow | Trigger | What it does |
|----------|---------|----------------|
| CI | push/PR to main | Build frontend + Go binary + Docker smoke build |
| Release | tag * (e.g. 0.1.0) | Multi-platform binaries + GitHub Release assets |
| Docker | push main / tag * | Multi-arch image to GHCR |

### Publish a release

`ash
git tag v0.1.0
git push origin v0.1.0
`

Release assets (examples):

- s3store_v0.1.0_windows_amd64.exe
- s3store_v0.1.0_linux_amd64
- s3store_v0.1.0_linux_arm64
- s3store_v0.1.0_darwin_amd64
- s3store_v0.1.0_darwin_arm64
- checksums.txt

### Docker image (GHCR)

`ash
docker pull ghcr.io/hakuzero4/s3_store_gui:latest
# or a version tag
docker pull ghcr.io/hakuzero4/s3_store_gui:0.1.0
`

> First time: GitHub Packages may require making the package public, or docker login ghcr.io.

## License

MIT ? see [LICENSE](./LICENSE).

Repository: https://github.com/hakuzero4/S3_store_gui
