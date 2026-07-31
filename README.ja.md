# S3 Store

**軽量・セルフホスト向けの S3 互換オブジェクトストレージ管理 UI。**

**Cloudflare R2** / **AWS S3** / **MinIO** 対応。**単一バイナリ** または **Docker** で配布。

**Languages:** [English](./README.md) · [????](./README.zh-CN.md) · [????](./README.zh-TW.md) · [???](./README.ja.md) · [???](./README.ko.md)

---

## 機能

- 複数接続プロファイル
- バケット一覧/作成/削除、接続内切替
- オブジェクト閲覧、プレフィックス、検索
- アップロード（ボタン+D&D）、ダウンロード、一括削除、リネーム
- フォルダ作成、詳細、署名付き URL
- 画像プレビュー
- **UI 多言語:** English（デフォルト） / 簡体中文 / 繁體中文 / 日本語 / 韓国語

---

## クイックスタート

### Docker

```bash
git clone https://github.com/hakuzero4/S3_store_gui.git
cd S3_store_gui
docker compose up -d --build
```

http://127.0.0.1:17890

### Windows

```powershell
./build.ps1
./dist/s3store.exe
```

---

## Cloudflare R2

| 項目 | 値 |
|------|-----|
| Endpoint | R2 ダッシュボード **S3 API** |
| Region | `auto` |
| Access / Secret | R2 Account API トークンのキーペア |
| Force Path Style | OFF |

同一アカウントの複数バケットはファイル画面のドロップダウン、別アカウントは「接続」で切替。

---

## UI 言語

デフォルトは **English**。サイドバーの **Language** で変更。

---

## ライセンス

MIT — [LICENSE](./LICENSE)

https://github.com/hakuzero4/S3_store_gui
