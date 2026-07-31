# S3 Store

**가볍고 셀프호스팅 가능한 S3 호환 오브젝트 스토리지 Web 콘솔.**

**Cloudflare R2**, **AWS S3**, **MinIO** 등을 깨끗한 데스크탑스타일 UI로 관리합니다. **단일 바이너리** 또는 **Docker 이미지**로 배포할 수 있습니다.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

**언어:** [English](./README.md) · [简体中文](./README.zh-CN.md) · [繁體中文](./README.zh-TW.md) · [日本語](./README.ja.md) · [한국어](./README.ko.md)

---

## 왝c S3 Store인가요?

- **로컬 우선** 관리 UI — 앱 자체는 SaaS 불필요
- **Cloudflare R2** 일급 지원
- 프론트엔드가 포함된 **실행 파일 하나**, 또는 가볍운 Docker
- 자격 증명은 로컬/볼륨에 남음

---

## 기능

- 다중 연결 프로필 (R2 / S3 / MinIO / 사용자 정의)
- 버킷 목록/생성/삭제, 프로필 내 버킷 전환
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

http://127.0.0.1:17890 — 설정 볼륨: `/data/config.json`

```bash
docker build -t s3store:latest .
docker run -d --name s3store \
  -p 17890:17890 \
  -v s3store-data:/data \
  s3store:latest
```

### 바이너리 (Windows)

```powershell
./build.ps1
./dist/s3store.exe
```

### 바이너리 (Linux / macOS)

```bash
cd web && npm ci && npm run build && cd ..
rm -rf internal/static/dist && cp -R web/dist internal/static/dist
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.1.0" -o dist/s3store ./cmd/s3store
./dist/s3store -no-browser
```

`config.json`은 기본으로 **실행 파일 엺**에 저장됩니다.

---

## Cloudflare R2 설정

1. Dashboard → **R2** → **Manage R2 API Tokens** → **Account API token** 생성
2. **Access Key ID** / **Secret Access Key** 복사 (비밀키는 한 번만 표시)
3. 계정 상세의 **S3 API**: `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`
4. S3 Store → **연결**:

| 항목 | 값 |
|------|-----|
| Provider | Cloudflare R2 |
| Endpoint | 대시보드 S3 API URL |
| Region | `auto` |
| Access Key / Secret | R2 API 토큰 쌍 |
| Force Path Style | 꺼짐 |

**주의**

- `cfat_…` 토큰 값은 **사용하지 마세요** (Cloudflare API용)
- 공개 커스텀 도메인을 S3 Endpoint로 쓰지 마세요
- 한 연결로 **여러 버킷** (드롭다운)
- **다른 계정** 은 여러 프로필

공식: [R2 S3 API](https://developers.cloudflare.com/r2/api/s3/api/) · [API tokens](https://developers.cloudflare.com/r2/api/tokens/)

---

## UI 다국어 (i18n)

기본: **English**

사이드바 **Language** 에서 전환. `localStorage` (`s3store.locale`) 에 저장.

지원: `en` · `zh-CN` · `zh-TW` · `ja` · `ko`

로케일 파일: `web/src/i18n/locales/*.json`

---

## 설정

### CLI

| 플래그 | 설명 | 기본 |
|------|------|------|
| `-addr` | 리슨 주소 | `127.0.0.1:17890` |
| `-config` | `config.json` 경로 | 실행 파일 엺 |
| `-no-browser` | 브라우저 열지 않음 | 데스크탑에서는 염 |
| `-version` | 버전 출력 | |

### 환경 변수

| 변수 | 설명 |
|------|------|
| `S3STORE_ADDR` | `-addr`와 동일 (Docker: `0.0.0.0:17890`) |
| `S3STORE_CONFIG` | `-config`와 동일 (Docker: `/data/config.json`) |
| `S3STORE_NO_BROWSER` | 설정 시 브라우저 비활성화 |

```bash
./s3store -addr 0.0.0.0:8080 -config /etc/s3store/config.json -no-browser
```

---

## 개발

**필요:** Go 1.22+, Node.js 20+

```bash
go run ./cmd/s3store -addr 127.0.0.1:17890 -no-browser
cd web
npm install
npm run dev
```

- UI: http://127.0.0.1:5173
- Vite가 `/api` 를 `http://127.0.0.1:17890` 로 프록시

### 디렉터리

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

### 빌드

1. `npm run build` → `web/dist`
2. `internal/static/dist` 로 복사
3. `go build`가 `//go:embed` 로 포함
4. UI+API 통합 바이너리

---

## HTTP API (개요)

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

## CI / CD (GitHub Actions)

| Workflow | 트리거 | 내용 |
|----------|---------|------|
| `CI` | `main` push/PR | 프론트+Go+Docker 스모크 |
| `Release` | tag `v*` (예: `v0.1.0`) | 다중 플랫폼 바이너리 + GitHub Release |
| `Docker` | `main` / tag `v*` | 다중 아키 이미지 → GHCR |

### 릴리스 공개

```bash
git tag v0.1.0
git push origin v0.1.0
```

예: `s3store_v0.1.0_windows_amd64.exe`, `linux_amd64/arm64`, `darwin_amd64/arm64`, `checksums.txt`

### Docker (GHCR)

```bash
docker pull ghcr.io/hakuzero4/s3_store_gui:latest
docker pull ghcr.io/hakuzero4/s3_store_gui:0.1.0
```

---

## 보안

- **로컬/사설 네트워크** 용
- 비밀키는 `config.json`에 평문 저장. 호스트 보호
- TLS/인증 없이 공개 노출 금지
- 최소 권한 R2/S3 토큰 권장

---

## 기술 스택

| 층 | 기술 |
|-------|--------|
| UI | Vue 3, TypeScript, Vite, Naive UI, Pinia, Vue Router, vue-i18n |
| API | Go `net/http`, AWS SDK for Go v2 (`service/s3`) |
| 배포 | `go:embed`, `CGO_ENABLED=0`, multi-stage Docker (Alpine) |

---

## 로드맵

- [ ] 대용량 파일 멀티파트 업로드
- [ ] prefix/버킷 간 복사
- [ ] 다크 모드
- [ ] Docker Basic Auth (선택)

Issue/PR 환영합니다.

---

## 라이선스

MIT — [LICENSE](./LICENSE)

https://github.com/hakuzero4/S3_store_gui
