# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2025-08-25

### Added
- **Initial release** of Grafana Catalyst Datasource plugin.
- Datasource backend written in Go using the Grafana Plugin SDK.
- Query Cisco Catalyst Center issues via REST API (`/dna/intent/api/v1/issues`).
- Secure storage of credentials and tokens (Grafana `secureJsonData`).
- Config editor with:
  - Base URL field
  - Username / Password login (backend fetches short-lived X-Auth-Token)
  - Optional API token override
  - Skip TLS verification toggle (for lab / self-signed environments)
- Query editor with filters:
  - `siteId`, `deviceId`, `macAddress`
  - `priority` (P1..P4), `issueStatus` (ACTIVE/IGNORED/RESOLVED)
  - `aiDriven` (YES/NO), `limit`
- Variable query support:
  - `priorities`, `issueStatuses`, `sites`, `devices`, `macs`
- Table panel integration: returns issue details (time, ID, title, severity, status, category, device, site, rule, details).

---
