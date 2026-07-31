# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Local disk cache for object previews/downloads (temp dir, ETag/LastModified validated)
- Sidebar connection source dropdown
- Copy/Move across prefixes and buckets
- ZIP download for selection
- Text/code preview
- Upload retry
- Profile import/export
- Multipart upload for larger files

### Changed
- Compact sidebar footer for connection + language

## [0.1.0] - 2026-07-31

### Added
- Initial public release of **S3 Store** ? self-hosted S3-compatible object storage console
- Vue 3 + Naive UI frontend embedded in a single Go binary (`go:embed`)
- Multi-profile connections for Cloudflare R2, AWS S3, MinIO, and custom S3 endpoints
- Bucket list / create / delete; object browse with prefixes, search, breadcrumbs
- Upload (button + drag-and-drop), download, bulk delete, rename, folder create
- Object details, presigned download URLs, in-app image preview
- Local `config.json` next to the executable (Docker: `/data/config.json`)
- Docker multi-stage image and `docker-compose.yml`
- GitHub Actions:
  - **CI** on `main` push/PR
  - **Release** multi-platform binaries on `v*` tags
  - **Docker** multi-arch push to GHCR (`ghcr.io/hakuzero4/s3_store_gui`)
- UI i18n: English (default), Simplified Chinese, Traditional Chinese, Japanese, Korean
- Multilingual READMEs: `README.md`, `README.zh-CN.md`, `README.zh-TW.md`, `README.ja.md`, `README.ko.md`
- MIT License

### Notes
- Designed for local / private-network use; secrets are stored in plain JSON
- Prefer least-privilege R2/S3 API tokens

[Unreleased]: https://github.com/hakuzero4/S3_store_gui/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/hakuzero4/S3_store_gui/releases/tag/v0.1.0
