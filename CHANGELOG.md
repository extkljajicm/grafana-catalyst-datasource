# Changelog

All notable changes to this project will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [Unreleased]
- Add support for additional Catalyst Center API endpoints.
- Expand query editor with more filters and metrics.
- Improved error handling and logging.

---

## [1.0.0] - 2025-08-30

### Added
- Initial release of **Grafana Catalyst Datasource** plugin.
- Go backend (`cmd/main.go`, `pkg/backend/`) built with Grafana Plugin SDK.
- React/TypeScript frontend (`src/`) with clean Config, Query, and Variable editors.
- Config editor with:
  - Base URL field
  - Skip TLS verification toggle
  - Username / Password authentication (backend fetches X-Auth-Token)
  - Optional API token override
- Query editor with filters:
  - `siteId`, `deviceId`, `macAddress`
  - `priority` (P1..P4), `issueStatus` (ACTIVE/IGNORED/RESOLVED)
  - `aiDriven` (YES/NO), `limit`
- Variable query support:
  - `priorities`, `issueStatuses`, `sites`, `devices`, `macs`
- Table panel integration: returns issue details (time, ID, title, severity, status, category, device, site, rule, details).
