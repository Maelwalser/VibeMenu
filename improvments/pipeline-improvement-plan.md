# Realize Pipeline Failure Analysis & Improvement Plan

## Failure Taxonomy

From the logs, there are **5 distinct failure modes**, ranked by frequency and retry cost:

### 1. Library API Hallucination (causes ~40% of retries)

The agent invents types that don't exist in third-party libraries:

```
pgxmock.PgxPoolMock       ← doesn't exist
pgxmock.PgxMock            ← doesn't exist
fiber.As(...)              ← doesn't exist (fiber/v2 has no As function)
```

The correct pgxmock usage is `pgxmock.NewPool()` which returns an unexported
type satisfying an interface. The agent has no way to know this without
documentation.

**Root cause:** LLM training data contains outdated or incorrect API usage.
No library API docs are injected into the prompt.

### 2. Dependency Version Hallucination (causes ~25% of retries)

```
pgservicefile@v0.0.0-20231201235250-de7065d787b0: invalid version
```

The agent guesses pseudo-versions for transitive dependencies. These often
don't exist. `go mod tidy` fails because it can't resolve them.

**Root cause:** The agent generates go.mod with invented version strings.
Later tasks re-generate go.mod, overwriting the validated one from the deps
phase.

### 3. Unknown Escape Sequences (causes ~15% of retries)

```
blog_repository_test.go:277:90: unknown escape sequence
```

The agent writes regex patterns in double-quoted Go strings:
```go
regexp.MustCompile("\\d{4}-\\d{2}-\\d{2}")  // works
regexp.MustCompile("\d{4}-\d{2}-\d{2}")      // FAILS: unknown escape \d
```

**Root cause:** LLMs frequently confuse raw strings and interpreted strings
in Go. This is a deterministic fix — always mechanically convertible.

### 4. gofmt Formatting (causes ~10% of retries)

```
files not gofmt-clean (run gofmt -w): internal/repository/postgres/blog_repository.go
```

**Root cause:** The agent generates code that isn't gofmt-formatted.
The deterministic fix exists but runs too late or doesn't cover all files.

### 5. Type Redeclaration on Retry (causes ~10% of retries)

```
internal/domain/models.go:12:6: Blog redeclared in this block
        internal/domain/blog.go:4:6: other declaration of Blog
```

On retry (especially with model escalation to Opus), the agent creates a
new `models.go` that duplicates types already in existing files.

**Root cause:** When retrying, the agent sees the full error but doesn't
see (or ignores) the existing file layout. It creates new files that
conflict.

---

## Implementation Changes

### Change 1: Inject Library API Docs into Agent Prompts

**File:** `internal/realize/agent/prompt.go`

Add a new section to `SystemPrompt()` that injects API reference docs for
commonly-misused libraries based on the task's technology stack.

```go
// In SystemPrompt(), after skill docs:
if len(libraryDocs) > 0 {
    b.WriteString("\n\n## Library API Reference (AUTHORITATIVE)\n\n")
    b.WriteString("The following are the CORRECT APIs for libraries in this project.\n")
    b.WriteString("Do NOT deviate from these — the types listed below are the only ones that exist.\n\n")
    for _, doc := range libraryDocs {
        b.WriteString(doc)
        b.WriteString("\n")
    }
}
```

The library docs live in `internal/realize/deps/deps.go` (see `LibraryAPIDocs`
map above). This prevents the agent from inventing `pgxmock.PgxPoolMock`.

### Change 2: Lock go.mod After Deps Phase

**File:** `internal/realize/agent/prompt.go`

When a task has dependency outputs from a deps phase, inject a hard rule:

```go
// In UserMessage(), when DependencyOutputs contain a go.mod:
b.WriteString("\n## LOCKED DEPENDENCIES\n\n")
b.WriteString("A validated go.mod already exists from the dependency resolution phase.\n")
b.WriteString("You MUST NOT regenerate go.mod or go.sum.\n")
b.WriteString("Only generate .go source files. The build system will use the locked go.mod.\n\n")
```

**File:** `internal/realize/orchestrator/runner.go`

After writing agent output to tmpDir, if a locked go.mod exists from the
deps phase, copy it over any go.mod the agent may have generated:

```go
// In TaskRunner.Run(), after writing files to tmpDir:
if lockedMod := r.findLockedGoMod(); lockedMod != "" {
    // Overwrite any agent-generated go.mod with the locked version.
    destMod := filepath.Join(tmpDir, "go.mod")
    os.WriteFile(destMod, []byte(lockedMod), 0644)
}
```

### Change 3: Expand Deterministic Fixes

**File:** `internal/realize/verify/deterministic_fixes.go` (new file)

Run these fixes BEFORE every verification attempt:

1. **Escape sequence fix:** Regex-detect double-quoted strings with `\d`,
   `\s`, `\w` etc. and rewrite as raw strings.

2. **gofmt:** Run `gofmt -w` on every `.go` file unconditionally.

3. **Duplicate type removal:** Parse Go files in the same package, detect
   duplicate type declarations, and remove the file with fewer declarations.

**Integration point in runner.go:**

```go
// Before calling verifier:
if fixes := verify.ApplyDeterministicFixes(tmpDir, verify.FilePaths(result.Files)); fixes != "" {
    r.log("[%s] applying deterministic fixes before retry", r.task.ID)
}
```

This should run on EVERY attempt (not just retries), since first-attempt
code also has these issues.

### Change 4: Well-Known Dependency Registry

**File:** `internal/realize/deps/deps.go` (new file, created above)

A static registry of well-known Go modules with their correct import paths
and latest stable versions. Used by:

1. The deps task to generate a correct initial go.mod
2. The prompt builder to inject version constraints
3. The deterministic fix layer to validate go.mod on retry

### Change 5: Smarter Retry with Targeted Context

**File:** `internal/realize/orchestrator/runner.go`

Current retry behavior: send full error output to agent and hope it fixes it.

Improved retry behavior:

```go
// On retry, classify the error and provide targeted instructions:
switch classifyError(lastVerifyOutput) {
case errTypeDeps:
    // Don't retry with LLM — run go mod tidy as deterministic fix
    fixGoModTidy(tmpDir)
case errTypeGofmt:
    // Don't retry with LLM — run gofmt -w as deterministic fix
    fixGofmt(tmpDir, files)
case errTypeUndefined:
    // Retry with LLM but inject the specific missing symbols
    ac.PreviousErrors = extractUndefinedSymbols(lastVerifyOutput)
case errTypeTestFailure:
    // Retry with LLM, include test output
    ac.PreviousErrors = lastVerifyOutput
}
```

This avoids wasting an LLM call (and retry budget) on problems that can
be fixed mechanically.

### Change 6: Model Escalation with Preserved Context

**File:** `internal/realize/orchestrator/runner.go`

Current behavior escalates to Sonnet/Opus but starts from scratch.

Improved: When escalating, include the PREVIOUS attempt's generated files
(not just the error) so the stronger model can see what was already produced
and fix it rather than regenerating from scratch. This prevents the type
redeclaration issue.

```go
// On retry with model escalation:
ac.PreviousAttemptFiles = lastAttemptFiles  // new field on agent.Context
```

And in the prompt:
```
## Previous Attempt (for reference)
The following files were generated by the previous attempt but failed verification.
Fix the specific issues listed in the error section. Do NOT create new files that
duplicate types already declared — modify the existing files instead.
```

---

## Architecture Diagram of Improved Pipeline

```
manifest.json
    │
    ▼
┌─────────────────────┐
│  DAG Builder         │ ← unchanged
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Wave 0: Data       │ ← schemas + migrations (no deps issue)
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Wave 1: Plan       │ ← skeleton + interfaces
│  (LLM generates     │
│   go.mod with        │
│   well-known deps)   │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────────────────┐
│  Wave 2: Deps Resolution        │ ← NO LLM CALL
│  1. Copy go.mod from plan       │
│  2. Run `go mod tidy`           │
│  3. If tidy fails, fix go.mod   │
│     using WellKnownGoModules    │
│  4. Save locked go.mod to       │
│     shared memory               │
└─────────┬───────────────────────┘
          │
          ▼
┌─────────────────────────────────┐
│  Wave 3+: Implementation        │
│  • Locked go.mod from deps      │
│    is injected via shared memory │
│  • Library API docs injected    │
│    into system prompt            │
│  • Agent told "DO NOT generate  │
│    go.mod"                       │
│                                  │
│  After each agent call:          │
│  1. Apply deterministic fixes    │
│     (gofmt, escape seqs)         │
│  2. Copy locked go.mod over      │
│     any agent-generated one      │
│  3. Run go mod tidy              │
│  4. Run verifier                 │
│                                  │
│  On failure:                     │
│  • Classify error type           │
│  • If deterministic → fix & skip │
│    LLM retry                     │
│  • If LLM-needed → retry with   │
│    previous files + targeted     │
│    guidance                      │
└─────────────────────────────────┘
```

---

## Quick Wins (implement first, biggest impact)

1. **Add `LibraryAPIDocs` to prompts** — eliminates pgxmock hallucination
   (~40% of retries). 1-2 hours to implement.

2. **Run gofmt + escape sequence fix before EVERY verification** — eliminates
   ~25% of retries. 30 minutes to implement.

3. **Lock go.mod from deps phase** — copy it over agent output before
   verification. Eliminates transitive dep failures. 1 hour.

4. **Add `WellKnownGoModules` registry** — agents use exact versions
   instead of guessing. 1 hour.

## Medium-Term Improvements

5. **Error classification in runner** — deterministic fixes skip LLM retry.
   Saves tokens and wall-clock time. 2-3 hours.

6. **Previous attempt files on retry** — prevents type redeclaration when
   escalating models. 2 hours.

7. **Pre-flight dependency validation** — before any agent runs, create a
   minimal Go project, import all expected dependencies, run `go mod tidy`,
   and cache the result. 3-4 hours.

## For Non-Go Languages

The same pattern applies:

| Language | Equivalent of go.mod | Lockfile | Verification |
|----------|---------------------|----------|--------------|
| Go | go.mod | go.sum | go build + go vet |
| TypeScript | package.json | package-lock.json | npm install + tsc |
| Python | pyproject.toml | requirements.txt / uv.lock | pip install + ruff |
| Rust | Cargo.toml | Cargo.lock | cargo check |
| Java | pom.xml / build.gradle | — | mvn compile / gradle build |

For each language, add:
1. A `WellKnownModules` map with correct package names and versions
2. A deps resolution step that runs the package manager (npm, pip, cargo)
3. Library API docs for commonly-misused packages
4. Deterministic fixes for that language's common LLM mistakes

### TypeScript-specific
- LLMs often use wrong import paths for ESM vs CJS
- Fix: inject correct import syntax for framework in use
- Common: `import { ... } from 'next/router'` vs `'next/navigation'` (App Router)

### Python-specific  
- LLMs mix up Flask/FastAPI/Django patterns
- Fix: inject framework-specific decorator and type patterns
- Common: wrong async/sync handler signatures

### Rust-specific
- LLMs hallucinate trait implementations
- Fix: inject actual trait signatures for axum/actix handlers
