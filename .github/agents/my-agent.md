# AI Agent Instructions: grafana-catalyst-datasource

You are an AI assistant for the **grafana-catalyst-datasource** project. Your primary goal is to help develop and maintain this Grafana plugin, which fetches "Assurance Issues" from the Cisco Catalyst Center API.

## 1. Core Project Architecture 🏗️

The project is a standard Grafana plugin with a Go backend and a React/TypeScript frontend.

### Backend (Go)
Located in the `pkg/backend/` directory. It handles all API communication with Catalyst Center, including authentication, data fetching, and transforming data into Grafana data frames.

### Frontend (TypeScript/React)
Located in the `src/` directory. It provides the user interface for the Query Editor and Variable Editor within Grafana.

### Data Models
The contract between the frontend and backend is defined in:
- **Frontend:** `src/types.ts` (TypeScript interfaces)
- **Backend:** `pkg/backend/model.go` (Go structs)

**These must be kept in sync.**

## 2. Backend Development (Go)

When working on the backend, adhere to these critical patterns:

### API Interaction

#### Authentication
- Authentication is handled by `token.go`
- All API requests must acquire a token via `tm.getToken()`
- This function handles caching, credentials, and manual token overrides
- **Do not implement your own authentication logic**

#### Parameter Building
- Parameter building is centralized in `params.go`
- **ALWAYS use the `buildAssuranceParamsFromQuery()` helper** to construct API request parameters
- This function correctly handles:
  - **Case Sensitivity:** Converts status values to lowercase (active, resolved)
  - **Pagination:** Uses a one-based offset
  - **Normalization:** Handles legacy field aliases (e.g., `severity` → `priority`)

#### Endpoint URLs
- Endpoint URLs are managed in `model.go`
- Use the `IssuesURL()` and `SiteURL()` helpers to get the correct API endpoints

### Data Handling

#### Primary Data Source
The primary data source is the `/dna/data/api/v1/assuranceIssues` endpoint.

#### Data Enrichment Pattern
Data enrichment is a key feature. To resolve IDs (like `siteId`) to human-readable names, follow this pattern:

1. Fetch the list of issues
2. Collect all unique IDs from the response
3. Make a **single, bulk API call** to the relevant endpoint (e.g., `SiteURL()`) to resolve the IDs
4. Map the names back to the issues before creating the Grafana data frame

**Important:** The `Enrich` toggle in the query is for future, performance-intensive lookups (like fetching full device details per issue). The **Site Name resolution**, however, should always be active as it is efficient.

## 3. Frontend Development (TypeScript/React)

When working on the frontend, follow these guidelines:

### Component Library
- **Use Grafana UI Components:** Import components like `Field`, `Input`, `Select`, and `Switch` from `@grafana/ui`

### Type Definitions
- Type definitions are in `src/types.ts`
- This is the **single source of truth** for the `CatalystQuery` object and other frontend types
- Use the specific types defined there (e.g., `CatalystPriority`)

### Query Editor (`QueryEditor.tsx`)
- User input should be **debounced** using the `useDebounced` hook to avoid excessive API calls while the user is typing
- When updating the query object via `onChange`, ensure that empty or default values are set to `undefined` to keep the query payload clean

### Variable Editor (`VariableQueryEditor.tsx`)
- This component is for creating Grafana template variables
- The logic is self-contained

## 4. General Rules & Constraints 📜

### Git Commits
Use concise, **imperative** commit messages in the format: `type(scope): message`

Examples:
- `feat(api): Add device name resolution`
- `fix(ui): Correct priority filter options`
- `docs: Update API documentation`
- `refactor(params): Simplify normalization logic`

### Dependencies
- **Go dependencies:** Add with `go get` and tidy with `go mod tidy`
- **Frontend dependencies:** Managed with `npm`

### Testing
- Backend changes should be accompanied by updates to the tests in `*_test.go` files
- Run tests with: `go test ./pkg/...`
- Frontend tests can be run with: `npm run test`

### Code Organization
- **Avoid direct API manipulation in `datasource.go`**
- All logic for building URLs and parameters should be delegated to helpers in `params.go` and `model.go`
- The main `datasource.go` file should orchestrate the process, not handle the low-level details

### Build & Development Commands

```bash
# Install dependencies
npm ci

# Development (hot reload)
npm run dev

# Linting
npm run lint

# Frontend tests
npm run test

# Go tests
go test ./pkg/...

# Build backend
go run github.com/magefile/mage -v BuildAll

# Local Grafana with Docker
docker compose up -d --build
```

## 5. Key Files Reference

| File | Purpose |
|------|---------|
| `pkg/backend/datasource.go` | Main datasource orchestration |
| `pkg/backend/token.go` | Authentication token management |
| `pkg/backend/params.go` | API parameter construction |
| `pkg/backend/model.go` | Data models and endpoint helpers |
| `src/types.ts` | Frontend type definitions |
| `src/components/QueryEditor.tsx` | Query editor UI component |
| `src/components/VariableQueryEditor.tsx` | Variable editor UI component |

## 6. Common Patterns

### Adding a New Query Parameter

1. **Frontend:** Add the field to `CatalystQuery` interface in `src/types.ts`
2. **Backend:** Add the field to `QueryModel` struct in `pkg/backend/model.go`
3. **Backend:** Update `buildAssuranceParamsFromQuery()` in `pkg/backend/params.go` to handle the new parameter
4. **Frontend:** Update `QueryEditor.tsx` to include the UI control for the new parameter
5. **Testing:** Add test cases in `pkg/backend/params_test.go`

### Resolving a New ID Type (e.g., Device ID to Name)

1. Add a new endpoint helper in `model.go` (similar to `SiteURL()`)
2. Create a new lookup function in `datasource.go` (similar to `getSiteNamesByID()`)
3. In `QueryData()`, collect unique IDs after fetching issues
4. Make a bulk API call to resolve all IDs at once
5. Map the resolved names back to the issues before creating the data frame

## 7. Important Notes

- **Always paginate through API responses** to collect all results
- **Handle 401/403 errors** by invalidating the cached token and retrying
- **Use one-based offset** for pagination (the API expects `offset=1` for the first page)
- **Normalize status values to lowercase** before sending to the API
- **Keep frontend and backend models in sync** to avoid runtime errors
- **Test both successful and error scenarios** when making changes
