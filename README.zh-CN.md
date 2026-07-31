# S3 Store

**轻量、可自托管的 S3 兼容对象存储 Web 管理控制台。**

管理 **Cloudflare R2**、**AWS S3**、**MinIO** 及其他 S3 API 端点；提供清爽的桌面风格 UI，可打包为 **单一二进制文件** 或 **Docker 镜像**。

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

**语言：** [English](./README.md) · [简体中文](./README.zh-CN.md) · [繁體中文](./README.zh-TW.md) · [日本語](./README.ja.md) · [한국어](./README.ko.md)

---

## 为什么选 S3 Store？

- **本地优先**的管理 UI — 应用本身不依赖 SaaS
- 对 **Cloudflare R2** 一等公民支持
- **单文件**内嵌前端，或轻量 Docker 镜像
- 凭证保留在你的机器 / 数据卷中

---

## 功能

- 多连接配置（R2 / S3 / MinIO / 自定义），侧边栏可快速切换
- Bucket 列表 / 创建 / 删除；同一配置下管理多个桶
- 对象浏览（前缀目录）、搜索、面包屑
- 上传（按钮 + 拖拽、失败重试、大文件分片）、下载、批量删除、重命名
- 跨前缀 / 跨桶批量复制与移动
- 选中项打包 ZIP 下载；新建文件夹；对象详情；预签名链接
- 图片在线预览、文本/代码预览
- 本地磁盘缓存预览/下载（系统临时目录，按 ETag / 修改时间校验）
- 连接配置导入 / 导出
- **界面多语言：** 英语（默认）、简体中文、繁体中文、日语、韩语

---

## 快速开始

### Docker

```bash
git clone https://github.com/hakuzero4/S3_store_gui.git
cd S3_store_gui
docker compose up -d --build
```

打开 http://127.0.0.1:17890 — 配置卷：`/data/config.json`。

```bash
# 纯 docker
docker build -t s3store:latest .
docker run -d --name s3store \
  -p 17890:17890 \
  -v s3store-data:/data \
  s3store:latest
```

### 二进制（Windows）

```powershell
./build.ps1
./dist/s3store.exe
```

### 二进制（Linux / macOS）

```bash
cd web && npm ci && npm run build && cd ..
rm -rf internal/static/dist && cp -R web/dist internal/static/dist
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.2.0" -o dist/s3store ./cmd/s3store
./dist/s3store -no-browser
```

默认在**可执行文件旁**写入 `config.json`。

---

## Cloudflare R2 配置

1. Cloudflare Dashboard → **R2** → **Manage R2 API Tokens** → 创建 **Account API token**
2. 复制 **Access Key ID** 与 **Secret Access Key**（密钥仅显示一次）
3. 在帐户详情中复制 **S3 API**：`https://<ACCOUNT_ID>.r2.cloudflarestorage.com`
4. 在 S3 Store → **连接** 中填写：

| 字段 | 值 |
|------|----|
| Provider | Cloudflare R2 |
| Endpoint | 控制台中的 S3 API URL |
| Region | `auto` |
| Access Key / Secret | R2 API 令牌密钥对 |
| Force Path Style | 关闭 |

**注意**

- **不要**使用 `cfat_…`「令牌值」——那是 Cloudflare API 用的，不是 S3。
- **不要**把公开自定义域名当作 S3 API Endpoint。
- 一个连接可管理**多个 Bucket**（文件浏览器中下拉切换）。
- **不同账号 / 密钥** 请建多个连接配置。

官方文档：[R2 S3 API](https://developers.cloudflare.com/r2/api/s3/api/) · [API tokens](https://developers.cloudflare.com/r2/api/tokens/)

---

## 界面多语言（i18n）

默认语言：**English**。

在侧栏 **Language** 中切换。选择会保存到 `localStorage`（`s3store.locale`）。

支持：`en` · `zh-CN` · `zh-TW` · `ja` · `ko`

语言包：`web/src/i18n/locales/*.json`

---

## 配置

### 命令行参数

| 参数 | 说明 | 默认 |
|------|------|------|
| `-addr` | 监听地址 | `127.0.0.1:17890` |
| `-config` | `config.json` 路径 | 与二进制同目录 |
| `-no-browser` | 不自动打开浏览器 | 桌面默认会打开 |
| `-version` | 打印版本 | |

### 环境变量

| 变量 | 说明 |
|------|------|
| `S3STORE_ADDR` | 同 `-addr`（Docker 默认 `0.0.0.0:17890`） |
| `S3STORE_CONFIG` | 同 `-config`（Docker 默认 `/data/config.json`） |
| `S3STORE_NO_BROWSER` | 设置后不自动打开浏览器 |

示例：

```bash
./s3store -addr 0.0.0.0:8080 -config /etc/s3store/config.json -no-browser
```

---

## 开发

**依赖：** Go 1.22+、Node.js 20+

```bash
# API
go run ./cmd/s3store -addr 127.0.0.1:17890 -no-browser

# 前端（热更新）
cd web
npm install
npm run dev
```

- UI：http://127.0.0.1:5173
- Vite 将 `/api` 代理到 `http://127.0.0.1:17890`

### 目录结构

```text
cmd/s3store/          # 入口
internal/api/         # HTTP API
internal/config/      # config.json
internal/s3client/    # AWS SDK v2 S3（兼容 R2）
internal/static/      # go:embed web/dist
web/                  # Vue 3 + Naive UI + Vite + Pinia + vue-i18n
web/src/i18n/         # 语言包
build.ps1             # Windows 打包
Dockerfile
docker-compose.yml
.github/workflows/    # CI、Release、Docker
```

### 构建流程

1. `npm run build` → `web/dist`
2. 复制到 `internal/static/dist`
3. `go build` 通过 `//go:embed` 内嵌静态资源
4. 得到同时提供 UI + API 的单二进制

---

## HTTP API（概览）

| 方法 | 路径 | 用途 |
|------|------|------|
| `GET` | `/api/health` | 健康检查 |
| `GET`/`POST` | `/api/profiles` | 列出 / 创建连接 |
| `PUT`/`DELETE` | `/api/profiles/{id}` | 更新 / 删除连接 |
| `POST` | `/api/profiles/{id}/activate` | 激活连接 |
| `POST` | `/api/profiles/test` | 测试凭证 |
| `GET`/`POST` | `/api/buckets` | 列桶 / 建桶 |
| `DELETE` | `/api/buckets/{name}` | 删桶 |
| `GET` | `/api/objects` | 列对象（`bucket`、`prefix`） |
| `POST` | `/api/objects/upload` | 上传 |
| `GET` | `/api/objects/download` | 下载（`inline=1` 预览） |
| `POST` | `/api/objects/delete` | 删除 key / 前缀 |
| `POST` | `/api/objects/folder` | 创建文件夹 |
| `POST` | `/api/objects/rename` | 重命名 |
| `POST` | `/api/objects/presign` | 预签名 GET URL |
| `GET` | `/api/objects/detail` | Head 元数据 |

---

## CI / CD（GitHub Actions）

| Workflow | 触发 | 作用 |
|----------|------|------|
| `CI` | push/PR 到 `main` | 构建前端 + Go + Docker 冒烟 |
| `Release` | tag `v*`（如 `v0.2.0`） | 多平台二进制 + GitHub Release |
| `Docker` | push `main` / tag `v*` | 多架构镜像推送到 GHCR |

### 发布 Release

```bash
git tag v0.2.0
git push origin v0.2.0
```

Release 产物示例：

- `s3store_v0.2.0_windows_amd64.exe`
- `s3store_v0.2.0_linux_amd64`
- `s3store_v0.2.0_linux_arm64`
- `s3store_v0.2.0_darwin_amd64`
- `s3store_v0.2.0_darwin_arm64`
- `checksums.txt`

### Docker 镜像（GHCR）

```bash
docker pull ghcr.io/hakuzero4/s3_store_gui:latest
docker pull ghcr.io/hakuzero4/s3_store_gui:0.2.0
```

> 首次使用可能需要将 Package 设为 Public，或 `docker login ghcr.io`。

---

## 安全

- 设计用于 **本地** 或 **私有网络**。
- 连接密钥以明文 JSON 保存在磁盘（`config.json`）。请保护文件与主机。
- 不要在无反向代理、TLS 与访问控制的情况下公网暴露端口。
- 尽量使用最小权限的 R2/S3 令牌（可按桶限制）。

---

## 技术栈

| 层 | 技术 |
|-------|--------|
| UI | Vue 3、TypeScript、Vite、Naive UI、Pinia、Vue Router、vue-i18n |
| API | Go `net/http`、AWS SDK for Go v2（`service/s3`） |
| 发布 | `go:embed`、`CGO_ENABLED=0`、多阶段 Docker（Alpine） |

---

## 路线图 / 想法

- [x] 大文件分片上传
- [x] 跨前缀 / 跨桶复制与移动
- [x] ZIP 下载、文本预览、本地对象缓存
- [ ] 深色模式切换
- [ ] Docker 部署可选 Basic Auth

欢迎 Issue 与 PR。

---

## 更新日志

见 [CHANGELOG.md](./CHANGELOG.md)。

## 许可

MIT — 见 [LICENSE](./LICENSE)。

仓库：https://github.com/hakuzero4/S3_store_gui
