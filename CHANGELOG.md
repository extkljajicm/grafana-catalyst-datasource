# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

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