# Grafana Catalyst Datasource

![Build](https://github.com/extkljajicm/grafana-catalyst-datasource/actions/workflows/ci.yml/badge.svg?branch=main)

The **Grafana Catalyst Datasource** lets you query **Cisco Catalyst Center (DNA Center)** alerts directly from Grafana dashboards.

---

## Features

- Fetch Catalyst Center alerts via the REST API
- Filter by severity, status, and free-text
- Display results in tables or time series panels
- Use in Grafana variables (e.g., severities, statuses)
- Supports secure token storage (`X-Auth-Token`)

---

## Requirements

- Grafana **v10.4.0+** (v12 recommended)
- Cisco Catalyst Center (DNA Center) with API access
- An API token (`X-Auth-Token`) with permission to query alerts

---

## Getting Started

1. Install the plugin into Grafana’s plugins directory or build from source.
2. In Grafana, go to **Connections → Data sources → Add data source**.
3. Choose **Catalyst-Datasource**.
4. Configure:
   - **Base URL**: e.g. `https://dnac.example.com/dna/intent/api/v1`
   - **API Token**: Catalyst Center `X-Auth-Token`
5. Save & Test. You should see a success message.
6. Add a panel, select this datasource, and run a query.

---

## Example Query

Filter alerts in the last 24h:

- **Severity**: `P1,P2`
- **Status**: `ACTIVE`
- **Text**: `Switch`

Results show as a table with columns:
`Time`, `ID`, `Title`, `Severity`, `Status`, `Category`, `Device`, `Site`, `Rule`, `Details`.

---

## Development

```bash
npm install
npm run dev   # watch mode
npm run build # production build
```

Start Grafana with:
```bash
docker compose up --build
```

In `grafana.ini`, allow unsigned plugins during development:
```ini
[plugins]
allow_loading_unsigned_plugins = grafana-catalyst-datasource
```

---

## License

Apache-2.0 © extkljajicm
