# S3 Store

**輕量、可自架的 S3 相容物件儲存 Web 管理控制台。**

管理 **Cloudflare R2**、**AWS S3**、**MinIO** 及其他 S3 API；提供清爽桌面風格 UI，可打包為 **單一執行檔** 或 **Docker 映像**。

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

**語言：** [English](./README.md) · [简体中文](./README.zh-CN.md) · [繁體中文](./README.zh-TW.md) · [日本語](./README.ja.md) · [한국어](./README.ko.md)

---

## 為什麼選 S3 Store？

- **本機優先**管理 UI — 應用本身不依賴 SaaS
- 對 **Cloudflare R2** 一等公民支援
- **單檔**內嵌前端，或輕量 Docker 映像
- 憑證留在你的機器 / 資料卷

---

## 功能

- 多連線設定（R2 / S3 / MinIO / 自訂）
- Bucket 列表 / 建立 / 刪除；同一設定下切換多桶
- 物件瀏覽、前綴目錄、搜尋、導覽列
- 上傳（按鈕 + 拖放）、下載、批次刪除、重新命名
- 新建資料夾、物件詳情、預簽名連結
- 圖片線上預覽
- **介面多語言：** 英語（預設）、簡體中文、繁體中文、日語、韓語

---

## 快速開始

### Docker

```bash
git clone https://github.com/hakuzero4/S3_store_gui.git
cd S3_store_gui
docker compose up -d --build
```

開啟 http://127.0.0.1:17890 — 設定卷：`/data/config.json`。

```bash
docker build -t s3store:latest .
docker run -d --name s3store \
  -p 17890:17890 \
  -v s3store-data:/data \
  s3store:latest
```

### 執行檔（Windows）

```powershell
./build.ps1
./dist/s3store.exe
```

### 執行檔（Linux / macOS）

```bash
cd web && npm ci && npm run build && cd ..
rm -rf internal/static/dist && cp -R web/dist internal/static/dist
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.2.0" -o dist/s3store ./cmd/s3store
./dist/s3store -no-browser
```

預設在**執行檔旁**寫入 `config.json`。

---

## Cloudflare R2 設定

1. Cloudflare Dashboard → **R2** → **Manage R2 API Tokens** → 建立 **Account API token**
2. 複製 **Access Key ID** 與 **Secret Access Key**（秘鑰僅顯示一次）
3. 從帳戶詳情複製 **S3 API**：`https://<ACCOUNT_ID>.r2.cloudflarestorage.com`
4. 在 S3 Store → **連線**：

| 欄位 | 值 |
|------|----|
| Provider | Cloudflare R2 |
| Endpoint | 控制台 S3 API URL |
| Region | `auto` |
| Access Key / Secret | R2 API 令牌金鑰對 |
| Force Path Style | 關閉 |

**註意**

- **不要**使用 `cfat_…`「令牌值」——那是 Cloudflare API，不是 S3。
- **不要**用公開自訂網域當 S3 API Endpoint。
- 一個連線可管**多個 Bucket**（檔案瀏覽器下拉）。
- **不同帳號 / 金鑰** 請建多個連線設定。

官方文件：[R2 S3 API](https://developers.cloudflare.com/r2/api/s3/api/) · [API tokens](https://developers.cloudflare.com/r2/api/tokens/)

---

## 介面多語言（i18n）

預設語言：**English**。

在側欄 **Language** 切換。會存入 `localStorage`（`s3store.locale`）。

支援：`en` · `zh-CN` · `zh-TW` · `ja` · `ko`

語言包：`web/src/i18n/locales/*.json`

---

## 設定

### CLI

| 參數 | 說明 | 預設 |
|------|------|------|
| `-addr` | 監聽位址 | `127.0.0.1:17890` |
| `-config` | `config.json` 路徑 | 與執行檔同目錄 |
| `-no-browser` | 不自動開瀏覽器 | 桌面預設會開 |
| `-version` | 列印版本 | |

### 環境變數

| 變數 | 說明 |
|------|------|
| `S3STORE_ADDR` | 同 `-addr`（Docker 預設 `0.0.0.0:17890`） |
| `S3STORE_CONFIG` | 同 `-config`（Docker 預設 `/data/config.json`） |
| `S3STORE_NO_BROWSER` | 設定後不自動開瀏覽器 |

```bash
./s3store -addr 0.0.0.0:8080 -config /etc/s3store/config.json -no-browser
```

---

## 開發

**依賴：** Go 1.22+、Node.js 20+

```bash
go run ./cmd/s3store -addr 127.0.0.1:17890 -no-browser
cd web
npm install
npm run dev
```

- UI：http://127.0.0.1:5173
- Vite 將 `/api` 代理到 `http://127.0.0.1:17890`

### 目錄結構

```text
cmd/s3store/          # entry
internal/api/         # HTTP API
internal/config/      # config.json
internal/s3client/    # AWS SDK v2 S3 (R2-compatible)
internal/static/      # go:embed web/dist
web/                  # Vue 3 + Naive UI + Vite + Pinia + vue-i18n
web/src/i18n/         # locales
build.ps1
Dockerfile
docker-compose.yml
.github/workflows/    # CI, Release, Docker
```

### 建置流程

1. `npm run build` → `web/dist`
2. 複製到 `internal/static/dist`
3. `go build` 以 `//go:embed` 內嵌
4. 得到 UI + API 單一執行檔

---

## HTTP API（概覽）

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/health` | Health |
| `GET`/`POST` | `/api/profiles` | List / create profiles |
| `PUT`/`DELETE` | `/api/profiles/{{id}}` | Update / delete |
| `POST` | `/api/profiles/{{id}}/activate` | Activate |
| `POST` | `/api/profiles/test` | Test credentials |
| `GET`/`POST` | `/api/buckets` | List / create buckets |
| `DELETE` | `/api/buckets/{{name}}` | Delete bucket |
| `GET` | `/api/objects` | List objects |
| `POST` | `/api/objects/upload` | Upload |
| `GET` | `/api/objects/download` | Download (`inline=1` preview) |
| `POST` | `/api/objects/delete` | Delete |
| `POST` | `/api/objects/folder` | Create folder |
| `POST` | `/api/objects/rename` | Rename |
| `POST` | `/api/objects/presign` | Presigned URL |
| `GET` | `/api/objects/detail` | Head metadata |

---

## CI / CD（GitHub Actions）

| Workflow | 觸發 | 作用 |
|----------|------|------|
| `CI` | push/PR 到 `main` | 建前端 + Go + Docker 冒煙 |
| `Release` | tag `v*`（如 `v0.2.0`） | 多平台執行檔 + GitHub Release |
| `Docker` | push `main` / tag `v*` | 多架構映像推到 GHCR |

### 發佈 Release

```bash
git tag v0.2.0
git push origin v0.2.0
```

產物示例：`s3store_v0.2.0_windows_amd64.exe`、`linux_amd64/arm64`、`darwin_amd64/arm64`、`checksums.txt`

### Docker 映像（GHCR）

```bash
docker pull ghcr.io/hakuzero4/s3_store_gui:latest
docker pull ghcr.io/hakuzero4/s3_store_gui:0.2.0
```

---

## 安全

- 設計用於 **本機** 或 **私有網路**。
- 密鑰以明文 JSON 存在 `config.json`。請保護主機。
- 勿在無 TLS/存取控的情況下公網暴露埠口。
- 儘量使用最小權限 R2/S3 令牌。

---

## 技術栈

| 層 | 技術 |
|-------|--------|
| UI | Vue 3、TypeScript、Vite、Naive UI、Pinia、Vue Router、vue-i18n |
| API | Go `net/http`、AWS SDK for Go v2（`service/s3`） |
| 發佈 | `go:embed`、`CGO_ENABLED=0`、多階段 Docker（Alpine） |

---

## 路線圖

- [ ] 大檔分片上傳
- [ ] 跨前綴/跨桶複製
- [ ] 深色模式
- [ ] Docker 可選 Basic Auth

歡迎 Issue 與 PR。

---

## 授權

MIT — [LICENSE](./LICENSE)

https://github.com/hakuzero4/S3_store_gui
