# S3 Store

**A lightweight, self-hosted web console for S3-compatible object storage.**

Manage **Cloudflare R2**, **AWS S3**, **MinIO**, and other S3 API endpoints from a clean desktop-style UI — packaged as a **single binary** or a **Docker image**.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

---

## Why S3 Store?

Cloud provider consoles are powerful but heavy. S3 Store is for people who want:

- A **local** tool that does not depend on a SaaS admin UI
- First-class **Cloudflare R2** support (not just “S3-ish”)
- **One executable** (frontend embedded) — copy, run, done
- Or a **tiny Docker** service with persistent connection profiles

No account system. No cloud dependency for the app itself. Your credentials stay on your machine (or in your volume).

---

## Features

| Area | What you get |
|------|----------------|
| **Connections** | Multiple profiles (R2 / S3 / MinIO / custom endpoint) |
| **Buckets** | List, create, delete; switch buckets under one profile |
| **Objects** | Browse prefixes like folders, search, breadcrumb navigation |
| **Transfer** | Upload (button + drag-and-drop), download, multi-select delete |
| **Organize** | Create folder placeholders, rename (copy + delete) |
| **Share** | Presigned download URLs |
| **Preview** | In-app image preview (jpg / png / webp / gif / …) |
| **UX** | Light, Finder-inspired UI; works great as a personal ops tool |

---

## Screenshots

> Run the app and open `http://127.0.0.1:17890` — file browser + connection manager.

| File browser | Connections |
|--------------|-------------|
| Browse buckets & objects | Save and switch R2 / S3 profiles |

---

## Quick start

### Option A — Docker (recommended for servers)

```bash
git clone https://github.com/<you>/s3store.git
cd s3store

docker compose up -d --build
```

Open **http://127.0.0.1:17890**

Config is stored in the `s3store-data` volume at `/data/config.json`.

```bash
# plain docker
docker build -t s3store:latest .
docker run -d --name s3store \
  -p 17890:17890 \
  -v s3store-data:/data \
  s3store:latest
```

### Option B — Single binary

**Windows**

```powershell
# requires Go 1.22+ and Node.js 20+
.\build.ps1
.\dist\s3store.exe
```

**Linux / macOS**

```bash
cd web && npm ci && npm run build && cd ..
rm -rf internal/static/dist && cp -R web/dist internal/static/dist
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.1.0" \
  -o dist/s3store ./cmd/s3store
./dist/s3store -no-browser
```

`config.json` is written **next to the executable** by default.

---

## Cloudflare R2 setup

1. Cloudflare Dashboard → **R2** → **Manage R2 API Tokens** → create an **Account API token**
2. Copy **Access Key ID** and **Secret Access Key** (secret is shown once)
3. Copy **S3 API** endpoint from Account details  
   `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`
4. In S3 Store → **Connections**:

| Field | Value |
|-------|--------|
| Provider | Cloudflare R2 |
| Endpoint | S3 API URL from dashboard |
| Region | `auto` |
| Access Key / Secret | R2 API token pair |
| Force Path Style | Off |

**Notes**

- Do **not** use the `cfat_…` “token value” — that is for Cloudflare’s API, not S3.
- Do **not** use a public custom domain as the S3 API endpoint.
- One connection can manage **many buckets** (dropdown in the file browser).
- Use multiple connections when you have **different accounts / keys**.

Official docs: [R2 S3 API](https://developers.cloudflare.com/r2/api/s3/api/) · [API tokens](https://developers.cloudflare.com/r2/api/tokens/)

---

## Configuration

### CLI flags

| Flag | Description | Default |
|------|-------------|---------|
| `-addr` | Listen address | `127.0.0.1:17890` |
| `-config` | Path to `config.json` | Next to the binary |
| `-no-browser` | Do not open a browser | Opens on desktop by default |
| `-version` | Print version | |

### Environment variables

| Variable | Description |
|----------|-------------|
| `S3STORE_ADDR` | Same as `-addr` (Docker default `0.0.0.0:17890`) |
| `S3STORE_CONFIG` | Same as `-config` (Docker default `/data/config.json`) |
| `S3STORE_NO_BROWSER` | Set to disable auto-open browser |

Example:

```bash
./s3store -addr 0.0.0.0:8080 -config /etc/s3store/config.json -no-browser
```

---

## Development

**Requirements:** Go 1.22+, Node.js 20+

```bash
# API server
go run ./cmd/s3store -addr 127.0.0.1:17890 -no-browser

# Frontend (hot reload)
cd web
npm install
npm run dev
```

- UI: http://127.0.0.1:5173  
- Vite proxies `/api` → `http://127.0.0.1:17890`

### Project layout

```text
cmd/s3store/          # main entry
internal/api/         # HTTP API
internal/config/      # config.json persistence
internal/s3client/    # AWS SDK v2 S3 client (R2-compatible)
internal/static/      # go:embed of web/dist
web/                  # Vue 3 + Naive UI + Vite + Pinia
build.ps1             # Windows release script
Dockerfile            # multi-stage image
docker-compose.yml
```

### Build pipeline

1. `npm run build` → `web/dist`
2. Copy into `internal/static/dist`
3. `go build` embeds static assets via `//go:embed`
4. Result: one binary that serves UI + API

---

## HTTP API (overview)

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/health` | Health check |
| `GET`/`POST` | `/api/profiles` | List / create connection profiles |
| `PUT`/`DELETE` | `/api/profiles/{id}` | Update / delete profile |
| `POST` | `/api/profiles/{id}/activate` | Activate profile |
| `POST` | `/api/profiles/test` | Test credentials |
| `GET`/`POST` | `/api/buckets` | List / create buckets |
| `DELETE` | `/api/buckets/{name}` | Delete bucket |
| `GET` | `/api/objects` | List objects (`bucket`, `prefix`) |
| `POST` | `/api/objects/upload` | Multipart upload |
| `GET` | `/api/objects/download` | Download (`inline=1` for preview) |
| `POST` | `/api/objects/delete` | Delete keys / prefixes |
| `POST` | `/api/objects/folder` | Create folder key |
| `POST` | `/api/objects/rename` | Rename object |
| `POST` | `/api/objects/presign` | Presigned GET URL |
| `GET` | `/api/objects/detail` | Head object metadata |

---

## Security

- Designed for **local** or **private network** use.
- Connection secrets are stored in plain JSON on disk (`config.json`). Protect the file and the host.
- Do not expose the port publicly without a reverse proxy, TLS, and access control.
- Prefer least-privilege R2/S3 tokens (bucket-scoped when possible).

---

## Tech stack

| Layer | Stack |
|-------|--------|
| UI | Vue 3, TypeScript, Vite, Naive UI, Pinia, Vue Router |
| API | Go `net/http`, AWS SDK for Go v2 (`service/s3`) |
| Ship | `go:embed`, `CGO_ENABLED=0`, multi-stage Docker (Alpine) |

---

## Roadmap / ideas

- [ ] Multipart upload for large files  
- [ ] Object copy across prefixes / buckets  
- [ ] Dark mode toggle  
- [ ] Optional basic auth for Docker deployments  
- [ ] i18n (EN / 中文)

Contributions and issues are welcome.

---

## License

MIT — see [LICENSE](./LICENSE).

---

## Acknowledgements

- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- [Vue](https://vuejs.org/) & [Naive UI](https://www.naiveui.com/)
- Cloudflare [R2](https://developers.cloudflare.com/r2/) S3 compatibility
