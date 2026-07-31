# S3 Store

**輕量、可自架的 S3 相容物件儲存管理介面。**

支援 **Cloudflare R2**、**AWS S3**、**MinIO** 等，可打包成**單一執行檔**或 **Docker 映像**。

**語言：** [English](./README.md) · [????](./README.zh-CN.md) · [????](./README.zh-TW.md) · [???](./README.ja.md) · [???](./README.ko.md)

---

## 功能

- 多組連線設定
- 儲存桶列表 / 建立 / 刪除；同一連線切換多桶
- 物件瀏覽、前綴、搜尋、導覽列
- 上傳（按鈕 + 拖放）、下載、批次刪除、重新命名
- 新建資料夾、詳情、預簽名連結
- 圖片線上預覽
- **介面多語系：** 英語（預設）、簡體中文、繁體中文、日語、韓語

---

## 快速開始

### Docker

```bash
git clone https://github.com/hakuzero4/S3_store_gui.git
cd S3_store_gui
docker compose up -d --build
```

http://127.0.0.1:17890（`/data/config.json`）

### Windows

```powershell
./build.ps1
./dist/s3store.exe
```

---

## Cloudflare R2

| 欄位 | 值 |
|------|----|
| Endpoint | R2 → **S3 API** |
| Region | `auto` |
| Access / Secret | R2 Account API 令牌金鑰對 |
| Force Path Style | 關閉 |

同帳號多桶 → 檔案頁下拉；不同帳號 →「連線」啟用。

---

## 介面語言

預設 **English**。側欄 **Language** 可切換。

---

## 授權

MIT — [LICENSE](./LICENSE)

https://github.com/hakuzero4/S3_store_gui
