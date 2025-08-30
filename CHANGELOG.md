# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.3] - 2025-08-30
### Added
- Backend: optional token expiry parsing. The token manager now derives expiry from **headers** (e.g., `X-Auth-Token-Expires-In`, `X-Auth-Token-Expiry`, `Cache-Control: max-age=…`, `Expires`) and also from common **JSON** fields (`expiresIn`, `expires_in`, `expiresAt`, `expiry`, `expiration`, `expireTime`) when available.

### Changed
- Backend: normalized DNAC base URL handling. `TokenURL` and `IssuesURL` now anchor to `/dna/...` and preserve any reverse-proxy prefix preceding `/dna`, ensuring consistent calls to:
  - `<prefix>/dna/system/api/v1/auth/token`
  - `<prefix>/dna/intent/api/v1/issues`
- UI: Config Editor hint updated to accept either `https://<host>` **or** `https://<host>/dna/intent/api/v1` (proxy prefixes are preserved).

### Notes
- Manual `apiToken` override still bypasses dynamic fetching (unchanged behavior).
- Default TTL remains 55 minutes when no expiry hints are present.

## [1.0.2] - 2025-08-30
- Tag created. If you published a release for `v1.0.2` already, consider `v1.0.3` for the above backend + UI adjustments.
