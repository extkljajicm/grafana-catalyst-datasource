# Developer Guide - Grafana Catalyst Datasource

This guide is for developers who want to contribute to or understand the internal workings of the Grafana Catalyst Datasource plugin.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Structure](#project-structure)
3. [Development Setup](#development-setup)
4. [Testing](#testing)
5. [Code Quality](#code-quality)
6. [Key Components](#key-components)
7. [Backend API](#backend-api)
8. [Contributing](#contributing)

---

## Architecture Overview

The plugin follows a **backend-driven architecture** where:

- **Frontend (TypeScript/React)**: Provides UI for configuration and query building
- **Backend (Go)**: Handles all API communication with Catalyst Center, authentication, and data transformation
- **Plugin SDK**: Uses Grafana's official SDKs for both frontend (@grafana/data, @grafana/ui, @grafana/runtime) and backend (grafana-plugin-sdk-go)

```
┌──────────────────┐
│  Grafana UI      │
└────────┬─────────┘
         │
         ├─► Frontend: Query Editor, Config Editor
         │   - TypeScript/React
         │   - @grafana/ui components
         │
         ├─► Backend Plugin (Go)
         │   - Token management
         │   - API communication
         │   - Data transformation
         │
         └─► Cisco Catalyst Center API
             - /dna/system/api/v1/auth/token
             - /dna/data/api/v1/assuranceIssues
             - /dna/intent/api/v1/site-health
```

---

## Project Structure

```
.
├── src/                    # Frontend source code
│   ├── components/         # React components
│   │   ├── ConfigEditor.tsx    # Datasource configuration UI
│   │   ├── QueryEditor.tsx     # Query building UI
│   │   ├── VariableQueryEditor.tsx  # Template variable UI
│   │   └── HelpTooltip.tsx     # Reusable help component
│   ├── datasource.ts       # Main datasource class
│   ├── types.ts           # TypeScript type definitions
│   ├── errors.ts          # Error handling utilities
│   ├── validation.ts      # Configuration & query validation
│   ├── cache.ts           # In-memory caching with TTL
│   ├── logger.ts          # Structured logging
│   └── module.ts          # Plugin registration
│
├── pkg/backend/           # Backend source code (Go)
│   ├── datasource.go      # Main datasource implementation
│   ├── token.go           # Token management & caching
│   ├── model.go           # Data models
│   ├── params.go          # API parameter building
│   └── *_test.go          # Unit tests
│
├── tests/                 # E2E tests
│   ├── configEditor.spec.ts
│   └── queryEditor.spec.ts
│
├── .config/              # Build configuration
│   └── webpack/          # Webpack config
│
└── dist/                 # Build output (generated)
```

---

## Development Setup

### Prerequisites

- Node.js ≥ 22
- Go ≥ 1.21
- Docker with Compose plugin
- npm or yarn

### Installation

```bash
# Clone the repository
git clone https://github.com/kljama/grafana-catalyst-datasource.git
cd grafana-catalyst-datasource

# Install dependencies
npm ci

# Install Go dependencies
go mod download
```

### Running in Development Mode

#### Option 1: Frontend only (hot reload)

```bash
npm run dev
```

This starts webpack in watch mode. Changes to TypeScript/React files will automatically rebuild.

#### Option 2: Full stack with Grafana

```bash
docker compose up --build
```

This starts:
- Grafana on http://localhost:3000 (admin/admin)
- Plugin automatically provisioned
- Backend plugin running in the container

Access Grafana and the datasource will be pre-configured.

---

## Testing

### Unit Tests (Frontend)

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:ci
```

Current test coverage:
- **82+ unit tests** covering:
  - Error handling (errors.test.ts)
  - Validation (validation.test.ts)
  - Caching (cache.test.ts)
  - DataSource class (datasource.test.ts)

### Unit Tests (Backend)

```bash
# Run Go tests
go test ./pkg/backend/... -v

# Run with coverage
go test ./pkg/backend/... -cover
```

### E2E Tests

```bash
# Run Playwright E2E tests
npm run e2e

# Run in UI mode for debugging
npx playwright test --ui
```

---

## Code Quality

### Linting

```bash
# Run ESLint
npm run lint

# Auto-fix linting issues
npm run lint:fix
```

### Type Checking

```bash
# Run TypeScript compiler
npm run typecheck
```

### Formatting

```bash
# Format code with Prettier
npm run lint:fix
```

---

## Key Components

### Frontend

#### DataSource (`src/datasource.ts`)

Main class that extends `DataSourceWithBackend`. Responsibilities:
- Provides default query values
- Implements template variable queries
- Applies template variable substitution
- Handles error parsing

Key methods:
- `getDefaultQuery()`: Returns default query structure
- `filterQuery()`: Validates queries before execution
- `metricFindQuery()`: Implements variable queries
- `applyTemplateVariables()`: Substitutes variables in queries

#### Error Handling (`src/errors.ts`)

Structured error types for better user feedback:
- `AuthenticationError`: Credential/token issues
- `ConnectionError`: Network connectivity problems
- `ConfigurationError`: Invalid configuration
- `QueryError`: Query execution failures
- `APIError`: API-specific errors with status codes

Helper functions:
- `parseError()`: Converts raw errors to CatalystError types
- `formatErrorMessage()`: Formats errors for user display

#### Validation (`src/validation.ts`)

Input validation utilities:
- `validateConfig()`: Validates datasource configuration
- `validateQuery()`: Validates query parameters
- `assertValidConfig()`: Throws if config is invalid
- `assertValidQuery()`: Throws if query is invalid

#### Cache (`src/cache.ts`)

Simple in-memory cache with TTL:
- `SimpleCache<T>`: Generic cache class
- Automatic expiration
- Cache key generation utilities

#### Logger (`src/logger.ts`)

Structured logging with configurable levels:
- Log levels: DEBUG, INFO, WARN, ERROR, NONE
- Contextual logging with metadata
- Child logger support
- Configurable via localStorage

---

### Backend

#### Datasource (`pkg/backend/datasource.go`)

Main backend implementation:
- `QueryData()`: Handles data queries from panels
- `CheckHealth()`: Implements health check endpoint
- `CallResource()`: Handles resource calls (e.g., for variables)

#### Token Manager (`pkg/backend/token.go`)

Token acquisition and caching:
- Automatic token refresh on expiry
- Multiple token expiry detection methods
- Per-datasource instance caching
- Fallback TTL when expiry not provided

#### Models (`pkg/backend/model.go`)

Data structures:
- `QueryModel`: Frontend query representation
- `InstanceSettings`: Datasource configuration
- API response models

---

## Backend API

### Query Data Flow

1. Frontend sends query via `QueryData` RPC
2. Backend parses query model
3. Backend acquires/refreshes token
4. Backend paginates through API (if needed)
5. Backend transforms API response to Grafana DataFrame
6. Response returned to frontend

### Resource Calls

Used for template variables and health checks:

```
GET /resources/issues?limit=100&offset=0&startTime=<ms>&endTime=<ms>
```

Returns raw API response for processing in frontend.

---

## Contributing

### Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make changes and add tests
4. Run linter and tests: `npm run lint && npm test`
5. Build: `npm run build`
6. Commit with conventional commit message: `feat: add feature`
7. Push and create pull request

### Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation only
- `test:` Adding tests
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

### Code Review Checklist

- [ ] All tests pass
- [ ] Code follows existing patterns
- [ ] Proper error handling
- [ ] Logging added for debugging
- [ ] Documentation updated
- [ ] No breaking changes (or properly documented)

---

## Debugging Tips

### Frontend Debugging

1. Enable DEBUG logging:
```javascript
localStorage.setItem('catalyst_log_level', 'DEBUG');
```

2. Open browser DevTools console

3. Look for log messages prefixed with `[CatalystDatasource]`

### Backend Debugging

1. Check Grafana logs:
```bash
docker compose logs -f grafana
```

2. Look for plugin-specific log messages

3. Use `log.DefaultLogger.Debug()` in Go code for detailed logging

---

## Resources

- [Grafana Plugin Development](https://grafana.com/docs/grafana/latest/developers/plugins/)
- [Plugin SDK for Go](https://github.com/grafana/grafana-plugin-sdk-go)
- [Cisco Catalyst Center API Documentation](https://developer.cisco.com/docs/dna-center/)

---

## License

Apache-2.0 © kljama
