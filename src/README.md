# Catalyst Datasource for Grafana

![Logo](https://raw.githubusercontent.com/kljama/grafana-catalyst-datasource/main/src/img/logo.svg)

Query **Cisco Catalyst Center** (formerly DNA Center) assurance data directly from Grafana. This plugin connects to the Catalyst REST API to fetch network health, alerts, and site information, enabling you to build comprehensive monitoring dashboards.

![Screenshot](https://raw.githubusercontent.com/kljama/grafana-catalyst-datasource/feature/endpoint-filter/src/img/screenshot-1.png)

---

## Features

- **Dual Endpoint Support**: Query both `Alerts` (assurance issues) and `Site Health` endpoints.
- **Dynamic Query Editor**: The UI adapts to the selected endpoint, showing relevant filters.
  - **For Alerts**: Filter by Site, Device, MAC, Priority, and Status.
  - **For Site Health**:
    - Build time series visualizations with a dynamic interval (1/5th of the selected time range).
    - Filter by Site Type, Parent Site Name, and Site Name.
    - Select specific metrics to visualize (e.g., Client Count, Health Score, AP Count).
- **Template Variable Support**: Dynamically populate dashboard variables with `priorities`, `issue statuses`, `sites`, `devices`, and `MAC addresses`.
- **Secure Credential Handling**: Uses Grafana's `secureJsonData` to encrypt credentials.
- **Automatic Token Management**: The Go backend handles API token acquisition and refresh automatically.

---

## Requirements

- Grafana **v12.1.0+**
- Cisco Catalyst Center with API access and network reachability from the Grafana instance.

---

## Configuration

Set up the datasource by providing the following:

- **Base URL**: The root URL of your Catalyst Center instance (e.g., `https://catalyst.example.com`). The plugin handles API paths automatically.
- **Credentials**:
  - **Username/Password**: For token-based authentication.
  - **API Token (Optional)**: Use a pre-issued token to bypass username/password login.
- **Skip TLS Verification**: Enable this only for development environments with self-signed certificates.

Click **Save & test** to confirm connectivity.

---

## Query Editor

### Alerts Endpoint

- **Site ID**: Filter by site UUID.
- **Device ID**: Filter by device UUID.
- **MAC Address**: Filter by client MAC address.
- **Priority**: Comma-separated list (e.g., `P1,P2`).
- **Status**: Comma-separated list (`ACTIVE,RESOLVED,IGNORED`).
- **Limit**: Maximum number of issues to return.

### Site Health Endpoint

- **Site Type**: Filter by `AREA` or `BUILDING`.
- **Parent Site Name / Site Name**: Filter by site hierarchy.
- **Metrics**: Select one or more metrics to visualize as a time series (e.g., `clientCount`, `healthScore`, `accessPointCount`).

---

## Template Variables

Use the following functions in the Variable Query Editor to create dynamic filters:

- `priorities()`: Returns `P1`, `P2`, `P3`, `P4`.
- `issueStatuses()`: Returns `ACTIVE`, `RESOLVED`, `IGNORED`.
- `sites(search:"<text>")`: Fetches unique site names from recent issues.
- `devices(search:"<text>")`: Fetches unique device IDs from recent issues.
- `macs(search:"<text>")`: Fetches unique MAC addresses from recent issues.
