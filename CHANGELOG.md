# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.0.2] - 2025-09-03

### Fixed
- **Reverse-proxy URL handling**: URL builder now preserves any prefix segments before `/dna` (e.g. `/proxy/dnac`) so requests hit the correct path for both **token** and **assuranceIssues** endpoints.
- **Status filter application**: Frontend reliably forwards the selected status (ACTIVE/RESOLVED/IGNORED) and the backend normalizes it to the API’s expected format, preventing “ACTIVE” queries from returning resolved/ignored rows.
- **Query loop in Explore/Editor**: Reworked the Query Editor to debounce input and only re-run when the effective filter signature actually changes (no more constant re-query/reset behavior).

### Changed
- **Default `limit`** is now **25** to match the assuranceIssues endpoint’s per-request maximum (1–25). The backend still pages internally, so user-selected limits >25 are honored by multiple 25-sized requests.
- **AIDriven handling**: normalized across UI → query → backend so `aiDriven` is consistently shaped and accepted by the API.

### Added
- **User-facing “no data” notice** when the result set is empty for the chosen time range/filters (improves UX feedback instead of a blank table).

## [1.0.1] - 2025-09-01

### Changed
- Switched backend issue queries from **/dna/intent/api/v1/issues** to **/dna/data/api/v1/assuranceIssues**.
- Implemented **one-based** pagination for `offset` to match the assuranceIssues API.
- Tightened query parameter validation:
  - `aiDriven` accepted values: `YES` / `NO` (case-insensitive).
  - `priority` accepts only CSV of `P1..P4`.
  - `issueStatus` accepts only CSV of `ACTIVE`, `IGNORED`, `RESOLVED`.
- Health check (Save & Test) now probes with `?limit=1` only

### Notes
- Token handling is unchanged and continues to use `/dna/system/api/v1/auth/token` with Basic auth or an override token.

## [1.0.0] - 2025-08-31
Initial Release