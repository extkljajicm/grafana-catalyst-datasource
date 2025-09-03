# Catalyst Datasource (Plugin Docs)

![Logo](https://raw.githubusercontent.com/extkljajicm/grafana-catalyst-datasource/main/src/img/logo.svg)

Query **Cisco Catalyst Center (formerly DNA Center)** issues/alerts directly from Grafana via the Catalyst REST API.

![Screenshot](https://raw.githubusercontent.com/extkljajicm/grafana-catalyst-datasource/main/src/img/screenshot-1.png)

---

## Features

- Fetch issues from `/dna/data/api/v1/assuranceIssues`
- Pagination uses one-based offset
- Filters: **Site**, **Device**, **MAC**, **Priority**, **Issue Status**, **AI-driven**, **Limit**
- Variable support: **priorities**, **statuses**, **sites**, **devices**, **macs**
- Secure credentials via Grafana `secureJsonData`
- Go backend + React/TypeScript frontend
- Token handling:
  - Auto-fetch via `/dna/system/api/v1/auth/token` using Basic Auth
  - Cache with expiry (reads headers/body when available)
  - Manual override with pre-issued API token

---

## Requirements

- Grafana **v12.1.0+**
- Cisco Catalyst Center with API access and network reachability from Grafana

---

## Configuration (Data Source)

- **Base URL** — `https://<host>` **or** `https://<host>/dna/intent/api/v1`  
  (Any reverse-proxy prefix **before** `/dna` is preserved.)
- **Skip TLS verification** — only for self-signed certs (use with care)
- **Username / Password** — used by backend to obtain a short-lived `X-Auth-Token`
- **API Token (override)** — optional; paste an existing token to bypass login

Click **Save & test** to verify connectivity.

---

## Query Editor (Panels)

Fields:
- **Query Type** — `alerts` (issues API)
- **Site ID** — filter by site (UUID)
- **Device ID** — filter by device (UUID)
- **MAC Address** — optional MAC filter (`aa:bb:cc:dd:ee:ff`)
- **Priority** — CSV: `P1,P2,P3,P4`
- **Issue Status** — CSV: `ACTIVE,IGNORED,RESOLVED`
- **AI Driven** — `YES`/`NO` (or blank for any)
- **Limit** — maximum rows returned (default 100)

Variables are supported in text inputs.

Returned columns (for **Table** panels):
- Time, Issue ID, Title
- Priority/Severity, Status, Category
- Device ID, MAC, Site ID, Rule, Details

---

## Variable Query Editor

Available functions:
- `priorities()`
- `issueStatuses()`
- `sites(search:"<text>")`
- `devices(search:"<text>")`
- `macs(search:"<text>")`

The optional `search` parameter narrows results (supports Grafana variables).

---

## Notes & Troubleshooting

- Ensure Grafana can reach your Catalyst Center (VPN/proxy/firewall).
- Prefer enabling TLS verification unless you have a valid reason not to.
- 401/403 responses: the backend will refresh the token and retry once.
- If you use a reverse proxy, include its prefix in the **Base URL**; the plugin preserves it for both `/dna/system/api/v1/auth/token` and `/dna/intent/api/v1/issues`.

---

## License

Apache-2.0 © extkljajicm
