# S3 Store

**가볍고 셀프호스팅 가능한 S3 호환 오브젝트 스토리지 관리 UI.**

**Cloudflare R2**, **AWS S3**, **MinIO** 지원. **단일 바이너리** 또는 **Docker** 배포.

**언어:** [English](./README.md) · [????](./README.zh-CN.md) · [????](./README.zh-TW.md) · [???](./README.ja.md) · [???](./README.ko.md)

---

## 기능

- 다중 연결 프로필
- 버킷 목록/생성/삭제, 연결 내 버킷 전환
- 객체 탐색, prefix, 검색, 브레드크럼
- 업로드(버튼+드래그), 다운로드, 일괄 삭제, 이름 변경
- 폴더 생성, 상세, 사전 서명 URL
- 이미지 미리보기
- **UI i18n:** English(기본), 간체중국어, 번체중국어, 일본어, 한국어

---

## 빠른 시작

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

| 항목 | 값 |
|------|-----|
| Endpoint | R2 대시보드 **S3 API** |
| Region | `auto` |
| Access / Secret | R2 Account API 토큰 키 쌍 |
| Force Path Style | 끄기 |

같은 계정의 여러 버킷은 파일 화면 드롭다운, 다른 계정은 **연결**에서 프로필 전환.

---

## UI 언어

기본값은 **English**. 사이드바 **Language** 에서 변경.

---

## 라이선스

MIT — [LICENSE](./LICENSE)

https://github.com/hakuzero4/S3_store_gui
