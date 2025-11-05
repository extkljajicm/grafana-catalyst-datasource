# Comprehensive Audit Report: grafana-catalyst-datasource
**Date:** 2025-11-05  
**Auditor:** GitHub Copilot  
**SSOT Reference:** `.github/copilot-instructions.md`

---

## Executive Summary

This audit compares the **grafana-catalyst-datasource** project against its Single Source of Truth (SSOT) defined in `.github/copilot-instructions.md`. The audit identifies discrepancies in:
- Backend code implementation vs. SSOT mandates
- Test coverage and test-code alignment
- Documentation accuracy and completeness
- Architectural patterns and best practices

### Overall Status
- ✅ **Frontend Tests:** All passing (82/82 tests)
- ❌ **Backend Tests:** Build failures due to test-code misalignment
- ⚠️ **Documentation:** Some inconsistencies with implementation
- ⚠️ **Code Patterns:** Several deviations from SSOT mandates

---

## Discrepancy Categories

### 1. Backend Test Failures

#### 1.1 Function Signature Misalignment
**Location:** `pkg/backend/params_test.go`  
**Severity:** HIGH  
**Issue:** Test functions expect different signatures than implemented functions

**Details:**
- `normalizePriority()` test expects 2 parameters `(priority string, severity string)` but implementation has 1 parameter `(priority string)`
- `normalizeIssueStatus()` test expects 2 parameters `(issueStatus string, status string)` but implementation has 1 parameter `(status string)`

**SSOT Reference:** Section 2 (Backend Development) states:
> "Parameter building is centralized in `params.go`"
> "Normalization: Handles legacy field aliases (e.g., `severity` → `priority`)"

**Evidence:**
```
params_test.go:23:45: too many arguments in call to normalizePriority
	have (string, string)
	want (string)
params_test.go:46:51: too many arguments in call to normalizeIssueStatus
	have (string, string)
	want (string)
```

**Proposed Task:**
Update `params.go` to include legacy field handling as tests expect, OR update tests to match current implementation. The SSOT explicitly mentions handling "legacy field aliases (e.g., `severity` → `priority`)", suggesting the two-parameter version is correct.

---

#### 1.2 Data Structure Misalignment in Tests
**Location:** `pkg/backend/params_test.go:55`  
**Severity:** HIGH  
**Issue:** Test uses `SiteID` as string, but `QueryModel` struct defines it as `[]string`

**Details:**
```go
// Test expects:
q := QueryModel{
    SiteID: "site-123",  // string
    ...
}

// But QueryModel defines:
type QueryModel struct {
    ...
    SiteID []string `json:"siteId,omitempty"`
    ...
}
```

**SSOT Reference:** Section 1 states:
> "These must be kept in sync." (referring to frontend and backend models)

**Proposed Task:**
Update test to use `SiteID: []string{"site-123"}` to match the current data model, or document if the model changed and tests were not updated.

---

#### 1.3 Expected Test Behavior Misalignment
**Location:** `pkg/backend/params_test.go:69`  
**Severity:** MEDIUM  
**Issue:** Test expects `status` field to remain lowercase `"resolved"`, but SSOT mandates uppercase normalization

**Details:**
Test expects:
```go
want := url.Values{
    "status": []string{"resolved"},  // lowercase
    ...
}
```

But SSOT Section 7 states:
> "Normalize status values to lowercase before sending to the API"

However, the implementation in `params.go` normalizes to uppercase (`ACTIVE`, `RESOLVED`, `IGNORED`).

**Proposed Task:**
Clarify SSOT intent - if API expects lowercase, update `normalizeIssueStatus()`. If API expects uppercase, update test expectations and SSOT documentation.

---

#### 1.4 Test Logic for Default Values
**Location:** `pkg/backend/params_test.go:82-105`  
**Severity:** MEDIUM  
**Issue:** Test expects default limit of 100, but implementation doesn't show this default in `buildAssuranceParamsFromQuery()`

**Details:**
Test expects:
```go
if params.Get("limit") != "100" { // default page size
    t.Fatalf("limit = %q, want 100", params.Get("limit"))
}
```

But `buildAssuranceParamsFromQuery()` directly uses the `pageSize` parameter passed to it:
```go
p.Set("limit", strconv.Itoa(pageSize))
```

No default value is applied if `pageSize <= 0`.

**Proposed Task:**
Either add default value handling in `buildAssuranceParamsFromQuery()` using the `clampLimit()` helper (which exists but isn't used), or update test expectations.

---

### 2. Code Organization Issues

#### 2.1 Unused Helper Function
**Location:** `pkg/backend/params.go:65-76`  
**Severity:** LOW  
**Issue:** `clampLimit()` function defined but never used

**SSOT Reference:** Section 4 states:
> "Code Organization: Avoid direct API manipulation in `datasource.go`"

**Evidence:**
```go
// Function exists:
func clampLimit(n, def, min, max int) int { ... }

// But is never called in the codebase
```

**Proposed Task:**
Either use `clampLimit()` in `buildAssuranceParamsFromQuery()` or remove it as dead code. The SSOT emphasizes clean, organized code.

---

#### 2.2 Direct API Manipulation in datasource.go
**Location:** `pkg/backend/datasource.go` (multiple locations)  
**Severity:** MEDIUM  
**Issue:** `datasource.go` contains low-level API handling logic

**SSOT Reference:** Section 4 states:
> "Avoid direct API manipulation in `datasource.go`"
> "All logic for building URLs and parameters should be delegated to helpers in `params.go` and `model.go`"
> "The main `datasource.go` file should orchestrate the process, not handle the low-level details"

**Evidence:**
- Lines 161-311: Direct HTTP request construction, pagination logic, error handling
- Lines 232-280: Token refresh retry logic embedded directly
- Lines 287-294: JSON unmarshaling logic with fallback

**Proposed Task:**
Refactor to extract:
1. Pagination logic into a helper function (e.g., `fetchAllIssues()` in a new `api_client.go`)
2. Token-refresh-retry pattern into `token.go`
3. Response parsing into `model.go`

Leave only orchestration logic in `datasource.go`.

---

### 3. Documentation Discrepancies

#### 3.1 Missing MANUAL.md
**Location:** N/A (file doesn't exist)  
**Severity:** LOW  
**Issue:** Comment references MANUAL.md which doesn't exist

**Note:** The original problem statement mentioned auditing `MANUAL.md`, but this file doesn't exist in the grafana-catalyst-datasource repository. This appears to be a copy-paste error from a different project (likely "netscan"). The actual user-facing documentation is in `src/README.md` and `README.md`.

**Proposed Task:**
N/A - This is not a real discrepancy. The grafana-catalyst-datasource project uses different documentation files.

---

#### 3.2 Missing config.yml.example
**Location:** N/A (file doesn't exist)  
**Severity:** LOW  
**Issue:** Problem statement references config.yml.example which doesn't exist

**Note:** Similar to above, this appears to be from a different project. The grafana-catalyst-datasource uses JSON configuration via Grafana's datasource settings, not YAML files.

**Proposed Task:**
N/A - This is not a real discrepancy.

---

#### 3.3 Documentation: Incorrect Repository Owner
**Location:** `src/README.md:133`, `DEVELOPER_GUIDE.md:98`  
**Severity:** LOW  
**Issue:** Documentation references incorrect GitHub username

**Details:**
- `src/README.md` line 133: `Apache-2.0 © extkljajicm`
- `DEVELOPER_GUIDE.md` line 98: `git clone https://github.com/extkljajicm/grafana-catalyst-datasource.git`
- `DEVELOPER_GUIDE.md` line 380: `Apache-2.0 © extkljajicm`

But repository is `kljama/grafana-catalyst-datasource` (visible in git remote and package.json).

**Proposed Task:**
Update all documentation to use correct owner: `kljama` instead of `extkljajicm`.

---

#### 3.4 Documentation: Enrich Toggle Description Mismatch
**Location:** `src/README.md` vs `copilot-instructions.md`  
**Severity:** LOW  
**Issue:** User documentation doesn't mention the `Enrich` toggle, but SSOT describes it

**SSOT Reference:** Section 2 states:
> "Important: The `Enrich` toggle in the query is for future, performance-intensive lookups (like fetching full device details per issue). The **Site Name resolution**, however, should always be active as it is efficient."

**Current State:**
- `src/README.md` doesn't mention an "Enrich" toggle
- `src/types.ts` defines `enrich?: boolean` in CatalystQuery
- `datasource.go` always performs site name enrichment (no toggle check)

**Proposed Task:**
Either:
1. Add Enrich toggle UI control to QueryEditor.tsx if it's meant to be user-facing, OR
2. Remove the `enrich` field from types if it's not yet implemented, OR
3. Update SSOT to clarify this is a future feature, not current

---

#### 3.5 Documentation: README Build Commands vs SSOT
**Location:** `README.md` vs `copilot-instructions.md`  
**Severity:** LOW  
**Issue:** README and SSOT show different command examples

**SSOT Commands:**
```bash
npm run dev
npm run lint
npm run test
go test ./pkg/...
go run github.com/magefile/mage -v BuildAll
docker compose up -d --build
```

**README.md Commands:**
```bash
npm ci
docker compose down
docker compose up -d --build
docker compose up -d --force-recreate
git commit -m "feat: concise summary"
export V=1.0.5
git tag -a v$V -m "Release v$V"
./create_release.sh
```

**Observation:**
README focuses on deployment/release workflow. SSOT focuses on development workflow. Both are correct for their contexts, but there's some overlap and potential confusion.

**Proposed Task:**
Clarify in SSOT that it's for AI agents (development focus) while README is for human developers (includes release procedures). No code changes needed.

---

### 4. Frontend-Backend Model Synchronization

#### 4.1 Case Sensitivity in Status Values
**Location:** `src/types.ts` vs `pkg/backend/params.go`  
**Severity:** MEDIUM  
**Issue:** Frontend defines uppercase status values, backend normalizes to uppercase, but SSOT says "lowercase"

**Frontend:** `src/types.ts:11`
```typescript
export type CatalystIssueStatus = 'ACTIVE' | 'RESOLVED' | 'IGNORED';
```

**Backend:** `pkg/backend/params.go:54-61`
```go
func normalizeIssueStatus(status string) (string, bool) {
    s := strings.ToUpper(strings.TrimSpace(status))
    if _, ok := allowedIssueStatus[s]; ok {
        return s, true  // Returns UPPERCASE
    }
    return "", false
}
```

**SSOT:** Section 7 states:
> "Normalize status values to lowercase before sending to the API"

**Proposed Task:**
1. Verify what the actual Catalyst Center API expects (likely UPPERCASE based on working code)
2. Update SSOT to say "uppercase" instead of "lowercase", OR
3. Change backend to send lowercase if that's what API actually needs

---

#### 4.2 Field Name Consistency
**Location:** `src/types.ts` vs `pkg/backend/model.go`  
**Severity:** LOW  
**Issue:** Field names are camelCase in frontend, but JSON tags vary in backend

**Example:**
```typescript
// Frontend: src/types.ts
networkDeviceId?: string;
```

```go
// Backend: pkg/backend/model.go
NetworkDeviceID string `json:"networkDeviceId,omitempty"`
```

**Observation:**
This is actually correct - Go uses PascalCase for exported fields, JSON tags provide camelCase mapping. The synchronization is working as expected.

**Proposed Task:**
N/A - No action needed. This is best practice.

---

### 5. Testing Coverage Issues

#### 5.1 Missing Backend Tests
**Location:** `pkg/backend/`  
**Severity:** MEDIUM  
**Issue:** Some files lack test coverage

**SSOT Reference:** Section 4 states:
> "Backend changes should be accompanied by updates to the tests in `*_test.go` files"

**Files without tests:**
- `site_translator.go` - No `site_translator_test.go`
- `datasource.go` - No `datasource_test.go` (complex orchestration logic)
- `token.go` - No `token_test.go` (critical authentication logic)

**Proposed Task:**
Add comprehensive unit tests for:
1. `SiteTranslator` (caching, refresh, lookups)
2. Token management (caching, expiry, refresh)
3. At least integration tests for `datasource.go` QueryData flow

---

#### 5.2 Frontend Test Coverage Gaps
**Location:** `src/components/`  
**Severity:** LOW  
**Issue:** React components lack unit tests

**Current Coverage:** 82 tests covering:
- `errors.test.ts` ✅
- `validation.test.ts` ✅
- `cache.test.ts` ✅
- `datasource.test.ts` ✅

**Missing Tests:**
- `QueryEditor.tsx` - No tests
- `VariableQueryEditor.tsx` - No tests
- `ConfigEditor.tsx` - No tests
- `HelpTooltip.tsx` - No tests

**SSOT Reference:** Section 3 mentions these files but doesn't mandate tests for components.

**Proposed Task:**
Add React component tests using @testing-library/react for:
1. QueryEditor interactions and debouncing
2. VariableQueryEditor type selection
3. ConfigEditor validation

---

### 6. Architectural Pattern Deviations

#### 6.1 Data Enrichment Always Active
**Location:** `pkg/backend/datasource.go:346-359`  
**Severity:** LOW  
**Issue:** Site name enrichment always runs, regardless of `enrich` flag

**SSOT Reference:** Section 2 states:
> "Important: The `Enrich` toggle in the query is for future, performance-intensive lookups (like fetching full device details per issue). The **Site Name resolution**, however, should always be active as it is efficient."

**Current Implementation:**
Site name resolution runs unconditionally (lines 346-359), which matches SSOT.

**Observation:**
This is actually CORRECT per SSOT. However, the `enrich` field exists in the query model but isn't used anywhere in the backend.

**Proposed Task:**
Either:
1. Remove unused `enrich` field from QueryModel and types, OR
2. Document in code comments that it's reserved for future use

---

#### 6.2 Pagination Implementation
**Location:** `pkg/backend/datasource.go:169-311`  
**Severity:** LOW  
**Issue:** Pagination uses correct offset (1-based) but implementation could be cleaner

**SSOT Reference:** Section 7 states:
> "Always paginate through API responses to collect all results"
> "Use one-based offset for pagination (the API expects `offset=1` for the first page)"

**Current Implementation:**
Correctly implements 1-based offset (line 230: `offsetForSite+1`), but the pagination logic is deeply nested in `datasource.go`.

**Proposed Task:**
Extract pagination logic to a helper function per SSOT's code organization mandate.

---

### 7. Git Commit Convention

#### 7.1 Recent Commits
**Location:** Git history  
**Severity:** LOW  
**Issue:** Most recent commit follows convention correctly

**SSOT Reference:** Section 4 states:
> "Use concise, **imperative** commit messages in the format: `type(scope): message`"

**Evidence:**
```
0e1b705 Initial plan
a7ee5b4 Enhance agent instructions for grafana-catalyst-datasource
```

**Observation:**
- "Initial plan" - Missing type prefix
- "Enhance agent instructions..." - Missing type prefix

**Proposed Task:**
Future commits should follow convention: `docs: enhance agent instructions for grafana-catalyst-datasource`

---

## Summary of Proposed Tasks

### High Priority (Blocking Issues)

1. **Fix Backend Test Failures:**
   - Update `normalizePriority()` and `normalizeIssueStatus()` to accept legacy field aliases
   - Update test expectations for `QueryModel.SiteID` to use array syntax
   - Align test expectations with actual API status value requirements (uppercase vs lowercase)

2. **Add Missing Backend Tests:**
   - Create `site_translator_test.go` with comprehensive coverage
   - Create `token_test.go` for authentication logic
   - Add integration tests for `datasource.go` main flows

### Medium Priority (Architectural Improvements)

3. **Refactor datasource.go:**
   - Extract pagination logic to helper functions
   - Move HTTP request/retry logic to separate module
   - Keep only orchestration logic in main file

4. **Clarify Enrich Field Status:**
   - Document whether `enrich` is current or future feature
   - Remove if unused, or implement if intended

5. **Update SSOT Documentation:**
   - Correct "lowercase" to "uppercase" for status normalization
   - Add any missing architectural patterns discovered during audit

### Low Priority (Polish)

6. **Fix Documentation Issues:**
   - Update GitHub username from `extkljajicm` to `kljama`
   - Add React component tests
   - Improve commit message convention adherence

7. **Code Cleanup:**
   - Remove or use `clampLimit()` function
   - Document unused fields in data models

---

## Methodology

This audit was conducted by:
1. Reading the complete SSOT (`.github/copilot-instructions.md`)
2. Examining all backend files (`pkg/backend/*.go`)
3. Examining all frontend files (`src/**/*.ts`, `src/**/*.tsx`)
4. Running test suites (Frontend: PASS 82/82, Backend: FAIL 3 build errors)
5. Reviewing documentation files
6. Cross-referencing implementation against SSOT mandates

---

## Conclusion

The grafana-catalyst-datasource project is generally well-structured and follows most SSOT principles. The main issues are:
1. **Test-code synchronization** - Tests expect older API signatures
2. **Code organization** - `datasource.go` contains too much low-level logic
3. **Documentation consistency** - Minor discrepancies in ownership and command examples

**Recommended Next Steps:**
1. Fix backend tests to match current implementation OR update implementation to match test expectations (verify with SSOT intent)
2. Add comprehensive backend test coverage
3. Refactor `datasource.go` to delegate low-level logic
4. Update documentation for accuracy

The frontend is in excellent shape with 82 passing tests and clean architecture. The backend needs attention primarily in testing and organization.
