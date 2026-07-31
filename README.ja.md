# S3 Store

**軽量・セルフホスト向けの S3 互換オブジェクトストレージ Web コンソール。**

**Cloudflare R2**、**AWS S3**、**MinIO** などを、クリーンなデスクトップ風 UI で管理。**単一バイナリ** または **Docker イメージ** で配布できます。

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

**Languages:** [English](./README.md) · [简体中文](./README.zh-CN.md) · [繁體中文](./README.zh-TW.md) · [日本語](./README.ja.md) · [한국어](./README.ko.md)

---

## なぜ S3 Store？

- **ローカル優先** — アプリ本体は SaaS 不要
- **Cloudflare R2** を一級市民としてサポート
- 前端組み込みの **1 実行ファイル**、または軽量 Docker
- 認証情報は自分の機器 / ボリュームに留まる

---

## 機能

- 複数接続プロファイル（R2 / S3 / MinIO / カスタム）、サイドバーで素早く切替
- バケット一覧 / 作成 / 削除、同一プロファイルで複数バケット
- オブジェクト閲覧、プレフィックス、検索、パンくず
- アップロード（ボタン + D&D、リトライ、大容量マルチパート）、ダウンロード、一括削除、リネーム
- プレフィックス / バケット跨ぎの一括コピー・移動
- 選択項目の ZIP ダウンロード、フォルダ作成、詳細、署名付き URL
- 画像プレビュー、テキスト/コードプレビュー
- プレビュー/ダウンロードのローカルディスクキャッシュ（一時ディレクトリ、ETag / 更新日時で検証）
- プロファイルのインポート / エクスポート
- **UI 多言語:** English（デフォルト）、簡体中文、繁體中文、日本語、韓国語

---

## クイックスタート

### Docker

```bash
git clone https://github.com/hakuzero4/S3_store_gui.git
cd S3_store_gui
docker compose up -d --build
```

http://127.0.0.1:17890 — 設定ボリューム: `/data/config.json`

```bash
docker build -t s3store:latest .
docker run -d --name s3store \
  -p 17890:17890 \
  -v s3store-data:/data \
  s3store:latest
```

### バイナリ（Windows）

```powershell
./build.ps1
./dist/s3store.exe
```

### バイナリ（Linux / macOS）

```bash
cd web && npm ci && npm run build && cd ..
rm -rf internal/static/dist && cp -R web/dist internal/static/dist
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.2.0" -o dist/s3store ./cmd/s3store
./dist/s3store -no-browser
```

`config.json` はデフォルトで**実行ファイルの旁**に保存されます。

---

## Cloudflare R2 設定

1. Dashboard → **R2** → **Manage R2 API Tokens** → **Account API token** 作成
2. **Access Key ID** / **Secret Access Key** をコピー（秘密鍵は 1 度だけ表示）
3. アカウント詳細から **S3 API**: `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`
4. S3 Store → **接続**:

| 項目 | 値 |
|------|-----|
| Provider | Cloudflare R2 |
| Endpoint | ダッシュボードの S3 API URL |
| Region | `auto` |
| Access Key / Secret | R2 API トークン |
| Force Path Style | Off |

**注意**

- `cfat_…` の「トークン値」は **使わない**（Cloudflare API 用）
- 公開カスタムドメインを S3 Endpoint にしない
- 1 接続で**複数バケット**（ドロップダウン）
- **別アカウント** は複数プロファイル

公式: [R2 S3 API](https://developers.cloudflare.com/r2/api/s3/api/) · [API tokens](https://developers.cloudflare.com/r2/api/tokens/)

---

## UI 多言語（i18n）

デフォルト: **English**

サイドバーの **Language** で切替。`localStorage`（`s3store.locale`）に保存。

対応: `en` · `zh-CN` · `zh-TW` · `ja` · `ko`

言語ファイル: `web/src/i18n/locales/*.json`

---

## 設定

### CLI

| フラグ | 説明 | デフォルト |
|------|------|--------------|
| `-addr` | リスン位址 | `127.0.0.1:17890` |
| `-config` | `config.json` パス | バイナリ旁 |
| `-no-browser` | ブラウザを開かない | デスクトップでは開く |
| `-version` | バージョン表示 | |

### 環境変数

| 変数 | 説明 |
|------|------|
| `S3STORE_ADDR` | `-addr` と同様（Docker: `0.0.0.0:17890`） |
| `S3STORE_CONFIG` | `-config` と同様（Docker: `/data/config.json`） |
| `S3STORE_NO_BROWSER` | 設定時ブラウザ不開 |

```bash
./s3store -addr 0.0.0.0:8080 -config /etc/s3store/config.json -no-browser
```

---

## 開發

**必要:** Go 1.22+、Node.js 20+

```bash
go run ./cmd/s3store -addr 127.0.0.1:17890 -no-browser
cd web
npm install
npm run dev
```

- UI: http://127.0.0.1:5173
- Vite が `/api` を `http://127.0.0.1:17890` へプロキシ

### ディレクトリ

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

### ビルド

1. `npm run build` → `web/dist`
2. `internal/static/dist` へコピー
3. `go build` が `//go:embed` で埋め込み
4. UI+API 一体バイナリ

---

## HTTP API（概要）

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

| Workflow | トリガ | 内容 |
|----------|---------|------|
| `CI` | `main` への push/PR | 前端+Go+Docker スモーク |
| `Release` | tag `v*`（例: `v0.2.0`） | 多平台バイナリ + GitHub Release |
| `Docker` | `main` / tag `v*` | 多架構イメージを GHCR へ |

### リリース公開

```bash
git tag v0.2.0
git push origin v0.2.0
```

産物例: `s3store_v0.2.0_windows_amd64.exe`, `linux_amd64/arm64`, `darwin_amd64/arm64`, `checksums.txt`

### Docker（GHCR）

```bash
docker pull ghcr.io/hakuzero4/s3_store_gui:latest
docker pull ghcr.io/hakuzero4/s3_store_gui:0.2.0
```

---

## セキュリティ

- **ローカル / プライベートネット** 向け
- 秘密情報は `config.json` に明文保存。ホストを守る
- TLS/認証なしで公開暴露しない
- 最小権限の R2/S3 トークンを推奨

---

## テックスタック

| 層 | 技術 |
|-------|--------|
| UI | Vue 3, TypeScript, Vite, Naive UI, Pinia, Vue Router, vue-i18n |
| API | Go `net/http`, AWS SDK for Go v2 (`service/s3`) |
| 配布 | `go:embed`, `CGO_ENABLED=0`, multi-stage Docker (Alpine) |

---

## ロードマップ

- [x] 大容量ファイルの分割アップロード
- [x] プレフィックス/バケット跨ぎコピー・移動
- [x] ZIP ダウンロード、テキストプレビュー、ローカルオブジェクトキャッシュ
- [ ] ダークモード
- [ ] Docker 向け Basic Auth（オプション）

Issue / PR 歓迎です。

---

## 変更履歴

[CHANGELOG.md](./CHANGELOG.md) を参照。

## ライセンス

MIT — [LICENSE](./LICENSE)

https://github.com/hakuzero4/S3_store_gui
