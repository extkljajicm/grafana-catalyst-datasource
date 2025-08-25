# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-08-25
### Added
- Initial release of Grafana Catalyst Datasource plugin
- Query Cisco Catalyst Center issues via REST API (/dna/intent/api/v1/issues)
- Secure API token storage in Grafana
- Query editor with filters (siteId, deviceId, macAddress, priority, issueStatus, aiDriven, limit)
- Config editor for base URL and API token
- Variable query support (priorities, issueStatuses, sites, devices, macs)
