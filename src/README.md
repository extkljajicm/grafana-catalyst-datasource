# Catalyst Datasource (Plugin Docs)

![Logo](https://raw.githubusercontent.com/extkljajicm/grafana-catalyst-datasource/main/src/img/logo.svg)

Query **Cisco Catalyst Center (formerly DNA Center)** issues/alerts directly from Grafana via the Catalyst REST API.

![Screenshot](https://raw.githubusercontent.com/extkljajicm/grafana-catalyst-datasource/main/src/img/screenshot-1.png)

---


## Features

- **Endpoint selection:** Choose which Catalyst Center API endpoint to query (e.g., issues/alerts or site health) in the data source config.
- **Site Health support:** Fetch overall health metrics for all sites from `/dna/intent/api/v1/site-health`.
- **Dynamic filters:** Query editor adapts filters based on selected endpoint.
  - For `siteHealth`: Filter by Site Type, Parent Site Name, Site Name, Parent Site ID, and Site ID.
  - For `alerts`: Filter by Site, Device, MAC, Priority, Status, and AI-driven.
- **Time series selection:** Select which site health metrics to visualize (e.g., Client Count, Health Score, Wired/Wireless Client Count).
- Fetch issues from `/dna/data/api/v1/assuranceIssues` (existing)
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

- **Base URL** — The base HTTP endpoint for your Cisco Catalyst Center instance. The plugin will automatically append the correct API paths (e.g., `/dna/system/api/v1/auth/token`).
  
  | ✅ Good | ❌ Bad |
  | :--- | :--- |
  | `https://catalyst.example.com` | `https://catalyst.example.com/dna` |
  | `https://proxy.corp/catalyst` | `https://catalyst.example.com/dna/intent/api` |

- **Skip TLS verification** — only for self-signed certs (use with care)
- **Username / Password** — used by backend to obtain a short-lived `X-Auth-Token`
- **API Token (override)** — optional; paste an existing token to bypass login

Click **Save & test** to verify connectivity.

---


## Query Editor (Panels)

Fields (dynamic based on endpoint):
- **Endpoint** — select which API endpoint to query (e.g., issues/alerts, site health)
- For **Site Health**:
  - **Site Type** — filter by site type (`AREA`, `BUILDING`)
  - **Limit** — max rows returned (1–50, default 25)
  - **Offset** — pagination
  - **Timestamp** — optional time filter
  - **Parent Site Name** — filter by parent site name
  - **Site Name** — filter by site name
  - **Metrics** — select which site health metrics to visualize (e.g., accessGoodCount, clientHealthWired, networkHealthAP, etc.)
- For **Issues/Alerts** (existing):
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
