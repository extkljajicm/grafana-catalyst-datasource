# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.2.0] - 2025-09-30

### Fixed
- **Site Health Metrics:** Fixed an issue where `clientCount` and `healthScore` were returning zero due to incorrect metric name mapping in the backend.
- **UI Duplication:** Resolved a UI bug where filter fields in the `siteHealth` query editor were duplicated.
- **Build Process:** Corrected the build script to ensure the plugin logo is always included in the release package.
- **Time Range Filtering:** Fixed a bug where the `siteHealth` endpoint was not correctly using the Grafana time range selector.

### Added
- **Expanded Site Health Filtering:** Added support for filtering by `siteId` and `parentSiteId` in the `siteHealth` query.
- **Expanded Site Health Metrics:** The "Metrics" selector for `siteHealth` now includes all relevant metrics from the API documentation.

### Changed
- **Refactored Query Editor:** The `QueryEditor` component was refactored to use a single state object, improving maintainability.
- **Improved Docker Workflow:** The build and deployment process now correctly uses `docker compose down` to ensure a clean environment.

## [1.1.0] - 2025-09-29

### Added
- **Endpoint filter feature:** Users can now select the API endpoint in the data source config, enabling support for multiple endpoints and dynamic query editor filters.
- **Site Health endpoint support:** Added support for `/dna/intent/api/v1/site-health` to fetch overall health metrics for all sites, with filters for site type, limit, offset, timestamp, parent site name, and site name.
- **Time series selection:** Query editor allows users to select which site health metrics to visualize as time series (e.g., accessGoodCount, clientHealthWired, networkHealthAP, etc.).

### Changed
- Backend and frontend logic updated to support dynamic filters and endpoint selection without breaking existing functionality.

### Notes
- Existing issue/alerts endpoint and features remain unchanged.
- This is a minor version bump due to new feature addition.

## [1.0.4] - 2025-09-27

### Changed
- Added and improved code comments for all frontend files:
  - `src/components/ConfigEditor.tsx`
  - `src/components/QueryEditor.tsx`
  - `src/components/VariableQueryEditor.tsx`
  - `src/module.ts`
  - `src/types.ts` (already well-documented, no changes needed)
- Added and improved code comments for backend files:
  - `pkg/backend/datasource.go`
  - `pkg/backend/token.go`
  - `pkg/backend/model.go`
  - `pkg/backend/params.go`

### Notes
- No logic changes; only documentation and comments for clarity and onboarding.
- Ensures codebase is self-documenting for new contributors.

## [1.0.3] - 2025-09-11

### Added
- **Multi-priority filtering**: The Query Editor now supports selecting multiple priorities (e.g., P1 and P2) to filter issues.

### Changed
- **Priority filtering mechanism**: Due to an API limitation, multi-priority filtering is now handled within the plugin. The datasource fetches all issues for the selected time range and then applies the priority filters on the backend before returning results to Grafana.

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