# Grafana Catalyst Datasource

[![Build](https://github.com/extkljajicm/grafana-catalyst-datasource/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/extkljajicm/grafana-catalyst-datasource/actions/workflows/ci.yml)

The **Grafana Catalyst Datasource** plugin lets you query **Cisco Catalyst Center (formerly DNA Center)** issues/alerts directly from Grafana dashboards via the Catalyst REST API.

---

## Features

- Fetch Catalyst Center issues via the REST API (`/dna/intent/api/v1/issues`)
- Filter by site, device, MAC address, severity/priority, status, AI-driven flag, and limit
- Display results in table panels
- Populate Grafana variables dynamically (priorities, statuses, sites, devices, MACs)
- Secure credential and token storage using Grafana's `secureJsonData`
- Written with a **Go backend** and **React/TypeScript frontend**

---

## Requirements

- Grafana **v12.1.0+**
- Cisco Catalyst Center with API access
- User credentials or an API token (`X-Auth-Token`) with permission to query issues

---

## Getting Started

### Development environment

Clone and install dependencies:

```bash
git clone https://github.com/extkljajicm/grafana-catalyst-datasource.git
cd grafana-catalyst-datasource
npm install
```

Run in development mode (frontend hot reload):

```bash
npm run dev
```

Start Grafana test container:

```bash
docker compose up -d
```

### Building a release

The plugin includes a build script that compiles both the frontend and the Go backend and packages them for distribution.

```bash
./create_release.sh
```

This produces a `.zip` file in the project root (`grafana-catalyst-datasource-<version>.zip`) ready to be installed into Grafana.

---

## Configuration

In **Data source settings**:
- **Catalyst Base URL** – e.g. `https://dnac.example.com/dna/intent/api/v1`
- **Skip TLS verification** – toggle only for self-signed certs
- **Username / Password** – Catalyst API credentials (backend fetches a short-lived X-Auth-Token)
- **API Token (override)** – paste an existing token to bypass login

---

## Usage

- In panel queries, use the **Query Editor** to filter issues by:
  - Site ID, Device ID, MAC Address
  - Priority (P1–P4), Issue Status (ACTIVE, IGNORED, RESOLVED)
  - AI Driven flag (YES/NO)
  - Limit (max rows)

- In dashboard variables, use the **Variable Query Editor** to dynamically fetch:
  - Priorities, Issue Statuses, Sites, Devices, MACs

---

## Development Notes

The backend service is written in Go and located in:

- `cmd/main.go`
- `pkg/backend/`

The frontend React/TypeScript code is in:

- `src/`

---

## License

Apache-2.0 © extkljajicm
