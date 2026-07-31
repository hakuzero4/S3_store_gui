# S3 Store

**轻量、可自托管的 S3 兼容对象存储管理面板。**

支持 **Cloudflare R2**、**AWS S3**、**MinIO** 等。可打包为**单文件二进制**或 **Docker 镜像**。

**语言：** [English](./README.md) · [简体中文](./README.zh-CN.md) · [繁體中文](./README.zh-TW.md) · [日本語](./README.ja.md) · [한국어](./README.ko.md)

---

## 特性

- 多连接配置（R2 / S3 / MinIO / 自定义 Endpoint）
- Bucket 列表 / 创建 / 删除；同一连接切换多个桶
- 对象浏览、前缀目录、搜索、面包屑
- 上传（按钮 + 拖拽）、下载、批量删除、重命名
- 新建文件夹、对象详情、预签名链接
- 图片在线预览
- **界面多语言：** 英语（默认）、简体中文、繁体中文、日语、韩语

---

## 快速开始

### Docker

```bash
git clone https://github.com/hakuzero4/S3_store_gui.git
cd S3_store_gui
docker compose up -d --build
```

访问 http://127.0.0.1:17890 ，配置保存在卷 `/data/config.json`。

### Windows

```powershell
./build.ps1
./dist/s3store.exe
```

---

## Cloudflare R2 配置

| 字段 | 值 |
|------|----|
| Endpoint | R2 控制台 → **S3 API** |
| Region | `auto` |
| Access / Secret | R2 **帐户 API 令牌** 密钥对 |
| Force Path Style | 关闭 |

不要使用 `cfat_` 开头的 Cloudflare API 令牌值。  
同一账号多桶 → 文件页 Bucket 下拉；不同账号 →「连接」页激活。

---

## 界面语言

默认 **English**。侧栏 **Language** 可切换（`localStorage`: `s3store.locale`）。

---

## 开发

```bash
go run ./cmd/s3store -addr 127.0.0.1:17890 -no-browser
cd web && npm install && npm run dev
```

---

## 许可

MIT — [LICENSE](./LICENSE)

https://github.com/hakuzero4/S3_store_gui
