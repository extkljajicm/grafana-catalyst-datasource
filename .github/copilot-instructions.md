# AI Instructions: grafana-catalyst-datasource

You are an AI assistant for the **grafana-catalyst-datasource** project.
This document provides development rules, architecture details, and procedures for this repository.
All instructions comply with Simple Technical English (STE) and Cisco Catalyst Center REST API guidelines.
Follow these instructions for all code changes and reviews.

---

## 1. Project Overview

- **Project Name:** `grafana-catalyst-datasource`
- **Target Platform:** Grafana v12.1.0 or newer.
- **Purpose:** Query Cisco Catalyst Center (formerly DNA Center) assurance issues and alerts with the Catalyst REST API.
- **Plugin Type:** Backend data source plugin (Go backend + React/TypeScript frontend).
- **License:** Apache-2.0.

---

## 2. Architecture and Directory Structure

The project uses a standard Grafana plugin architecture.

### Directory Layout

| Path | Purpose | Technologies |
| :--- | :--- | :--- |
| `cmd/grafana-catalyst-datasource/` | Backend application entry point | Go >= 1.21 |
| `pkg/backend/` | Backend logic, authentication, parameter building, and models | Go >= 1.21, Grafana Plugin SDK for Go |
| `src/` | Frontend plugin source code | TypeScript 5.5, React 18 |
| `src/components/` | UI components for configuration, panels, and variables | `@grafana/ui` |
| `tests/` | End-to-end tests | Playwright |
| `Magefile.go` | Backend build automation targets | Mage |
| `create_release.sh` | Local packaging script for testing | Bash |
| `docker-compose.yaml` | Local Grafana environment for development | Docker Compose |

### Data Contract

The frontend and the backend share the query schema:
- **Frontend definition:** `src/types.ts` (`CatalystQuery` interface).
- **Backend definition:** `pkg/backend/model.go` (`QueryModel` struct).

**Rule:** You must keep `src/types.ts` and `pkg/backend/model.go` in sync.
When you add or update a query parameter in one file, update the other file immediately.

---

## 3. Cisco Catalyst Center API Compliance Guidelines

Always align plugin communication with official Cisco Catalyst Center API standards.

### 3.1 API Path Hierarchy and Base URLs

Catalyst Center organizes endpoints by domain:
- **Authentication API:** `/dna/system/api/v1/auth/token`
- **Assurance Issues API:** `/dna/data/api/v1/assuranceIssues`
- **Intent Site API:** `/dna/intent/api/v2/site`
- **Site Health API:** `/dna/intent/api/v1/site-health`

**Base URL Rules:**
- The configured `baseUrl` must specify the root host (for example: `https://catalyst.example.com`).
- Do not append `/dna` or `/dna/intent/api` to `baseUrl`.
- Preserve any reverse-proxy path prefix that precedes `/dna` (for example: `https://proxy.corp/catalyst`).

### 3.2 Authentication and Header Standards

- **Token Endpoint:** Send `POST /dna/system/api/v1/auth/token` using HTTP Basic Authentication (`Authorization: Basic <base64(user:password)>`).
- **Token Delivery:** Catalyst Center returns the security token in the `X-Auth-Token` response header and in the JSON response body (`{"Token": "<token>"}`).
- **Subsequent Requests:** Send the header `X-Auth-Token: <token>` with every API request.
- **Accept Header:** Always send `Accept: application/json` on all requests.
- **Content-Type Header:** Always send `Content-Type: application/json` when sending a request body.
- **Token Expiry:** Catalyst Center tokens expire after one hour (3600 seconds) by default. Cache the token with its expiration time. Refresh the token when expired.

### 3.3 Pagination and Limit Standards

Catalyst Center Assurance APIs use `limit` and `offset` query parameters.

- **One-Based Indexing:** The `offset` parameter uses one-based indexing.
  - Set `offset=1` to retrieve the first record.
  - Do not use `offset=0`.
- **Page Size:**
  - The API supports a maximum `limit` of 1000 records per request.
  - Clamp user limits between `1` and `1000`.
  - Default to `25` or `100` records per page.
- **Pagination Loop Termination:**
  - Stop the pagination loop when the returned record count is zero.
  - Stop the pagination loop when the returned record count is less than `limit`.
  - Stop the pagination loop when the total fetched records reach the requested maximum limit.

### 3.4 Parameter Format and Normalization

Follow Catalyst Center parameter naming and type rules:

| Field | Catalyst Center Parameter | Expected Format / Values | Description |
| :--- | :--- | :--- | :--- |
| Time range | `startTime`, `endTime` | Integer epoch milliseconds | UTC timestamp in milliseconds |
| Pagination | `limit` | Integer (`1` to `1000`) | Maximum records per request |
| Pagination | `offset` | Integer (`>= 1`) | One-based starting record index |
| Site filter | `siteId` | UUID string or comma-separated UUIDs | Example: `7f6c0f40-...` |
| Device filter | `deviceId` | UUID string | Example: `9b2d3a10-...` |
| Client filter | `macAddress` | Hexadecimal string | Colon-separated or hyphen-separated MAC |
| Priority | `priority` | Uppercase comma-separated: `P1,P2,P3,P4` | API priority values |
| Issue status | `status` | Lowercase string: `active`, `resolved`, `ignored` | Note: Query parameter is `status`, response field is `issueStatus` |
| AI-driven | `aiDriven` | Lowercase string: `"true"` or `"false"` | Boolean filter |

**Normalization Rules:**
- Omit empty or whitespace-only parameters from the request URL.
- Convert priority values to uppercase (`P1`, `P2`, `P3`, `P4`).
- Convert status values to lowercase (`active`, `resolved`, `ignored`).
- Map legacy aliases before parameter generation (`severity` -> `priority`; `status` -> `issueStatus`).

### 3.5 Rate Limiting and Performance Rules

- Catalyst Center enforces rate limits per endpoint (commonly 100 calls per minute; varies from 20 to 500 calls per minute).
- **Avoid N+1 Requests:** Do not make single API calls per issue row.
- **Batch Resolution:** Query `/dna/intent/api/v1/site` with comma-separated IDs (`?siteId=id1,id2,id3`) to resolve multiple site names in one call.
- **Debounce Frontend Inputs:** Debounce query inputs for 400 milliseconds to reduce API load.
- **Limit Variable Lookups:** Restrict template variable issue lookups to 5 pages (`MAX_PAGES = 5`).
- **HTTP 429 Handling:** Respect HTTP 429 (Too Many Requests) responses. Do not immediately retry without backoff.

### 3.6 Response Payload Handling

- Catalyst Center APIs return data inside a `response` attribute: `{"response": [...], "version": "1.0"}`.
- Parse responses defensively:
  - Check for the envelope structure `{"response": [...]}`.
  - Check for raw JSON arrays `[...]`.
- Read field names with fallbacks to support multiple Catalyst Center releases:
  - Timestamp: `timestamp`, `firstOccurredTime`, `startTime`.
  - Issue ID: `issueId`, `id`, `instanceId`.
  - Issue Title: `name`, `title`, `issueTitle`.
  - Priority: `priority`, `severity`.
  - Status: `issueStatus`, `status`.
  - Category: `category`, `type`.
  - Device: `deviceId`, `deviceIp`, `device`.
  - MAC Address: `macAddress`, `clientMac`.
  - Description: `description`, `details`, `issueDescription`.

### 3.7 HTTP Status Code Handling

- `200 OK`: Request succeeded. Parse payload.
- `400 Bad Request`: Query parameters or UUID formats are invalid. Return an error message.
- `401 Unauthorized`: Token expired or credentials invalid. Invalidate token cache, acquire a new token, and retry the request once.
- `403 Forbidden`: User account lacks required permissions or token expired. Invalidate token cache, acquire a new token, and retry the request once.
- `429 Too Many Requests`: Rate limit exceeded. Return a rate-limit error.
- `5xx Server Error`: Catalyst Center service error. Log details and report error to Grafana.

---

## 4. Backend Rules (Go)

All backend code lives in `pkg/backend/` and `cmd/grafana-catalyst-datasource/`.

### 4.1 Authentication and Token Management (`token.go`)

- Do not implement custom authentication routines.
- Always call `d.tm.getToken(ctx, instanceUID, settings, httpClient)` to get an authentication token.
- The token manager handles:
  1. Manual API token override from `InstanceSettings.APIToken`.
  2. Cached tokens with expiry validation.
  3. Automatic token acquisition with HTTP Basic Auth at `/dna/system/api/v1/auth/token`.
  4. Automatic retry: On HTTP 401 or HTTP 403 responses, the datasource clears the cached token, requests a new token, and retries the API call once.

### 4.2 URL Building and Reverse Proxy Support (`model.go`)

- Always use the URL helper functions in `model.go`:
  - `TokenURL(baseURL)`: Returns the token endpoint path (`/dna/system/api/v1/auth/token`).
  - `IssuesURL(baseURL)`: Returns the issues endpoint path (`/dna/data/api/v1/assuranceIssues`).
  - `SiteURL(baseURL)`: Returns the site lookup endpoint path (`/dna/intent/api/v2/site`).
- These helper functions call `dnacPrefix()`. This preserves reverse-proxy prefixes before `/dna`.
- Do not build endpoint URLs by direct string concatenation.

### 4.3 Query Parameter Construction (`params.go`)

- Always use `buildAssuranceParamsFromQuery()` to convert `QueryModel` into `url.Values`.
- Enforce Catalyst Center API parameter standards:
  - Set `offset` to one-based index (`offset=1` for the first page).
  - Convert status values to lowercase (`active`, `resolved`, `ignored`).
  - Convert priority values to uppercase (`P1`, `P2`, `P3`, `P4`).
  - Normalize boolean values to `"true"` or `"false"`.
  - Support legacy field aliases (`severity` maps to `priority`; `status` maps to `issueStatus`).
  - Omit empty or whitespace-only parameters from the request URL.

### 4.4 Query Execution and Data Enrichment (`datasource.go`)

- `QueryData()` executes the main query flow:
  1. Reads `QueryModel` from `backend.DataQuery`.
  2. Paginates requests to `/dna/data/api/v1/assuranceIssues` until it reaches the user limit or receives an empty page.
  3. Uses a default page size of 25.
  4. Enriches site IDs with site names if `qm.Enrich` is `true`.
  5. Converts issues into a `data.Frame`.
- **Enrichment Rule:** Collect all unique site IDs after the query loop. Make one bulk API request to `SiteURL()` with comma-separated IDs. Do not make individual API requests per issue row.
- **Code Organization:** Keep `datasource.go` focused on orchestration. Move URL logic to `model.go` and parameter logic to `params.go`.

### 4.5 Resource Routing (`CallResource`)

- `CallResource()` routes custom frontend resource requests:
  - `sites`: Returns cached site hierarchy objects via `handleSitesRequest()`.
  - `issues`: Forwards query requests to `/dna/data/api/v1/assuranceIssues` via `resourceIssues()` with authentication headers to power dynamic template variable resolution (`sites`, `devices`, `macs`).
  - Any other path: Returns HTTP 404 Not Found.

---

## 5. Frontend Rules (TypeScript and React)

All frontend code lives in `src/`.

### 5.1 Component Library and Styling

- Use Grafana UI components from `@grafana/ui` (`Field`, `Input`, `Select`, `Switch`, `SecretInput`).
- Do not import unauthorized third-party UI libraries.

### 5.2 Type Definitions (`src/types.ts`)

- `src/types.ts` is the single source of truth for frontend models.
- Use explicit union types:
  - `CatalystPriority = 'P1' | 'P2' | 'P3' | 'P4'`
  - `CatalystIssueStatus = 'ACTIVE' | 'RESOLVED' | 'IGNORED'`
  - `QueryType = 'alerts'`

### 5.3 Data Source Configuration (`ConfigEditor.tsx`)

- Users configure:
  - `baseUrl`: Base HTTP endpoint for Catalyst Center (example: `https://catalyst.example.com` or `https://proxy.corp/catalyst`).
  - `insecureSkipVerify`: Switch to disable TLS certificate verification for self-signed certificates.
  - `username` and `password`: Secure fields used to obtain `X-Auth-Token`.
  - `apiToken`: Optional secure field to provide a pre-issued `X-Auth-Token`.
- **Base URL Rule:** Do not append `/dna` or `/dna/intent/api` to `baseUrl`. The backend appends all API paths automatically.

### 5.4 Panel Query Editor (`QueryEditor.tsx`)

- Use `useDebounced` for text and select inputs (400 ms delay). This prevents excessive requests to the backend during typing.
- When calling `onChange(nextQuery)`:
  - Set empty text inputs to `undefined`.
  - Set empty arrays to `undefined`.
  - Set default aliases (`severity`, `status`) to `undefined`.
- Support all query editor fields:
  - `siteId`: Site UUID filter. Supports Grafana template variables.
  - `deviceId`: Device UUID filter.
  - `macAddress`: MAC address filter.
  - `priority`: Multi-select (`P1`, `P2`, `P3`, `P4`).
  - `issueStatus`: Single-select (`ACTIVE`, `RESOLVED`, `IGNORED`).
  - `aiDriven`: Single-select (`Any`, `True`, `False`).
  - `limit`: Maximum row count (default 25).
  - `enrich`: Switch labeled "Fetch full details" to enable backend site name enrichment.

### 5.5 Template Variable Editor (`VariableQueryEditor.tsx` and `datasource.ts`)

- Implement variable query functions:
  - `priorities()`: Returns static list `P1`, `P2`, `P3`, `P4`.
  - `issueStatuses()`: Returns static list `ACTIVE`, `RESOLVED`, `IGNORED`.
  - `sites(search:"<text>")`: Fetches recent issues and extracts unique site IDs.
  - `devices(search:"<text>")`: Fetches recent issues and extracts unique device identifiers.
  - `macs(search:"<text>")`: Fetches recent issues and extracts unique MAC addresses.
- `DataSource.metricFindQuery()` calls `this.getResource('issues?...')` to query the backend resource handler for dynamic variable values.

---

## 6. Development and Build Commands

### Prerequisites

- **Git**
- **Node.js >= 22** and **npm**
- **Go >= 1.21**
- **Docker** with the Compose plugin
- CLI utilities: `zip`, `awk`, `sed`, `jq`

### Command Reference

| Action | Command |
| :--- | :--- |
| Install frontend dependencies | `npm ci` |
| Start frontend watch mode | `npm run dev` |
| Run linter | `npm run lint` |
| Automatically fix lint and style | `npm run lint:fix` |
| Run TypeScript checks | `npm run typecheck` |
| Run frontend unit tests | `npm run test` (interactive) or `npm run test:ci` (CI mode) |
| Run backend unit tests | `go test ./pkg/...` |
| Run backend benchmarks | `go test -bench=. ./pkg/...` |
| Build frontend bundle | `npm run build` |
| Build backend binary | `go run github.com/magefile/mage -v BuildAll` |
| Clean backend binaries | `go run github.com/magefile/mage -v Clean` |
| Run local Grafana container | `docker compose up -d --build` |
| Stop local Grafana container | `docker compose down` |

---

## 7. Git and Release Workflows

### 7.1 Git Commits

- Write commit messages in Conventional Commits format:
  `type(scope): short description in imperative mood`
- Common types:
  - `feat`: New user-facing feature.
  - `fix`: Bug fix.
  - `docs`: Documentation updates.
  - `refactor`: Code change that does not fix a bug or add a feature.
  - `test`: Adding or updating tests.
  - `chore`: Build tasks, configuration, or dependency updates.
- Update `CHANGELOG.md` manually for release notes.

### 7.2 Release Option A — CI Release from Git Tag (Recommended)

1. Create and push an annotated Git tag:
   ```bash
   export V=1.0.6
   git tag -a v$V -m "Release v$V"
   git push origin v$V
   ```
2. The GitHub Actions workflow (`.github/workflows/release.yml`) handles the release:
   - Reads the version from the tag (`v1.0.6` -> `1.0.6`).
   - Runs `npm ci` and `npm run build`.
   - Builds backend binaries for `linux/amd64` using `Magefile.go`.
   - Packages the plugin into `grafana-catalyst-datasource-<version>.zip`.
   - Generates a SHA-1 checksum file.
   - Publishes a GitHub Release with attached assets.

### 7.3 Release Option B — Local Packaging

Use this option to generate a local ZIP package for testing:

1. Create a local Git tag (optional, recommended):
   ```bash
   export V=1.0.6
   git tag -a v$V -m "Release v$V"
   ```
2. Execute the release script:
   ```bash
   ./create_release.sh
   ```
3. The script outputs `grafana-catalyst-datasource-<version>.zip` in the repository root.

---

## 8. Common Step-by-Step Procedures

### How to Add a New Query Parameter

1. **Update Frontend Type:** Add the property to `CatalystQuery` in `src/types.ts`.
2. **Update Backend Model:** Add the property to `QueryModel` in `pkg/backend/model.go`.
3. **Update Parameter Builder:** Update `buildAssuranceParamsFromQuery()` in `pkg/backend/params.go` to validate and map the field according to Catalyst Center parameter specifications.
4. **Update Query UI:** Add the input component to `src/components/QueryEditor.tsx` with debouncing.
5. **Add Backend Tests:** Add unit test cases to `pkg/backend/params_test.go`.
6. **Verify Changes:** Run `npm run typecheck`, `npm run test:ci`, and `go test ./pkg/...`.

### How to Add a New ID Resolution Lookup

1. **Add URL Helper:** Add a new endpoint helper in `pkg/backend/model.go`.
2. **Define Data Structures:** Add response envelope structs in `pkg/backend/model.go`.
3. **Implement Batch Lookup:** Add a batch lookup method in `pkg/backend/datasource.go` that accepts a slice of IDs and returns a `map[string]string`.
4. **Collect IDs in `QueryData`:** After the issues pagination loop, extract unique IDs into a slice.
5. **Resolve Names:** Call the batch lookup method once with comma-separated IDs to avoid rate limits, and map the resolved values into the output data frame.

---

## 9. Verification Checklist

Before you submit code changes, complete every step:

- [ ] `npm run typecheck` succeeds without errors.
- [ ] `npm run lint` reports zero warnings and errors.
- [ ] `npm run test:ci` passes all frontend tests.
- [ ] `go test ./pkg/...` passes all backend tests.
- [ ] Frontend and backend models remain synchronized.
- [ ] Empty query parameters are set to `undefined`.
- [ ] Status values are converted to lowercase (`status=active`) for Catalyst Center API calls.
- [ ] Offsets for pagination remain one-based (`offset=1`).
- [ ] Time parameters (`startTime`, `endTime`) are formatted as epoch milliseconds.
- [ ] Endpoint lookups use batching to respect Catalyst Center rate limits.
- [ ] `CHANGELOG.md` is updated if the change affects users.
