# Guide 8: Modules, Testing & Tooling Deep Dive

> Guide 1 introduced Go Modules just enough to get a project running. This guide goes back for the depth you'll actually need once you're managing real dependencies and writing real tests — `go.mod`/`go.sum` internals, semantic versioning, the `testing` package (Go's built-in, no-install-required answer to `pytest`), table-driven tests, benchmarks, and the static-analysis tools (`go vet`, linters) that catch mistakes before a human reviewer has to. By the end, `go test ./...` should feel as automatic to reach for as `pytest` currently is.

## Table of Contents

1. [Go Modules Revisited: What Problem They Solve](#1-go-modules-revisited-what-problem-they-solve)
2. [`go.mod` in Depth](#2-gomod-in-depth)
3. [`go.sum` and Supply-Chain Integrity](#3-gosum-and-supply-chain-integrity)
4. [Semantic Versioning and Version Selection](#4-semantic-versioning-and-version-selection)
5. [Adding, Upgrading, and Removing Dependencies](#5-adding-upgrading-and-removing-dependencies)
6. [Semantic Import Versioning: the `v2+` Rule](#6-semantic-import-versioning-the-v2-rule)
7. [Workspaces: `go.work` for Multi-Module Development](#7-workspaces-gowork-for-multi-module-development)
8. [The `testing` Package: Go's Built-in `pytest`](#8-the-testing-package-gos-built-in-pytest)
9. [Table-Driven Tests](#9-table-driven-tests)
10. [Subtests with `t.Run`](#10-subtests-with-trun)
11. [Setup, Teardown, and `TestMain`](#11-setup-teardown-and-testmain)
12. [Why Go Has No Assertions (and When to Reach for `testify`)](#12-why-go-has-no-assertions-and-when-to-reach-for-testify)
13. [Test Helpers and `t.Helper()`](#13-test-helpers-and-thelper)
14. [Benchmarks with `testing.B`](#14-benchmarks-with-testingb)
15. [Test Coverage](#15-test-coverage)
16. [`go vet`: Static Analysis Built In](#16-go-vet-static-analysis-built-in)
17. [Linting with `golangci-lint`](#17-linting-with-golangci-lint)
18. [Exercises](#18-exercises)
19. [Key Takeaways](#19-key-takeaways)

---

## 1. Go Modules Revisited: What Problem They Solve

Guide 1 showed you `go mod init` and told you `go.mod` declares your module and its dependencies. That's the "how"; here's the "why," because the reasoning maps directly onto a problem you've already solved in Python, just with different tools.

Python's dependency story has historically been fragmented: `requirements.txt` pins versions but doesn't resolve a dependency graph on its own, `pip` resolves but historically resolved *poorly* (the old resolver could produce inconsistent installs), and tools like Poetry or `uv` layer proper resolution and lockfiles on top. You've likely used a `pyproject.toml` + lockfile combination in your FastAPI work — that pairing (a human-edited manifest of direct dependencies, plus a machine-generated lockfile pinning the entire resolved graph) is exactly the shape Go Modules takes:

| Python (Poetry/uv-style) | Go |
|---|---|
| `pyproject.toml` (direct deps, human-edited) | `go.mod` (direct deps, human/tool-edited) |
| `poetry.lock` / `uv.lock` (full resolved graph, machine-generated) | `go.sum` (cryptographic hashes of every module version touched, machine-generated) |
| `poetry add requests` | `go get github.com/some/package` |
| `poetry install` | `go mod download` (usually implicit via `go build`/`go test`) |
| `pip show`, dependency tree tools | `go mod graph`, `go mod why` |
| Virtual environments (isolate per-project deps) | Not needed — Go's module cache is content-addressed and global, but each module's `go.mod` pins its own versions, so there's no environment to "activate" |

That last row is worth pausing on: Go doesn't have an equivalent of `venv`/`virtualenv` because it doesn't need one. Downloaded module versions live in a shared, global, read-only cache (`$GOPATH/pkg/mod` by default), keyed by module path *and* version — so `github.com/some/package@v1.2.0` and `@v1.3.0` coexist on disk without conflict, and which one your build actually uses is determined entirely by your project's `go.mod`, not by which "environment" happens to be active in your shell. There's nothing to `source activate` — you just `cd` into a directory with a `go.mod` and the toolchain knows exactly what to build with.

---

## 2. `go.mod` in Depth

A representative `go.mod` for a small web service:

```
module github.com/pedro/orderapi

go 1.23

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/google/uuid v1.6.0
)

require (
	github.com/stretchr/testify v1.9.0 // indirect
)
```

Line by line:

- **`module github.com/pedro/orderapi`** — declares this module's *import path*. Every file inside this module imports its own sibling packages using this prefix (e.g., `import "github.com/pedro/orderapi/internal/store"`), the way an installed Python package's internal modules import each other via the package's own name. Unlike Python, the module path is conventionally a real, resolvable location (a GitHub repo path, typically) even for private or unpublished code — it's how `go get` knows where to fetch it from if someone else depends on it, and how your own imports are unambiguous even if two different modules on your machine happen to have similarly named internal packages.

- **`go 1.23`** — the minimum Go language version this module requires, not the version installed on your machine. This affects which language features (like the per-iteration loop variable semantics from Guide 7, introduced in 1.22) the compiler assumes are safe to use, and can trigger a "go.mod requires go >= 1.23" error if a teammate is on an older toolchain. It's the rough analogue of `python_requires = ">=3.11"` in a `pyproject.toml`.

- **`require` blocks** — each line is `module/path version`. The first block here holds **direct** dependencies (things your code actually imports); the second, marked `// indirect`, holds dependencies of your dependencies that Go still needs to record precisely for reproducible builds, similar to how a lockfile records transitive dependencies even though you never `import` them yourself. You'll see `// indirect` comments appear and disappear automatically as you add/remove imports — you don't maintain this by hand.

You essentially never hand-edit version numbers in `go.mod` directly; `go get`, `go mod tidy`, and friends do it for you (Section 5), the same way you wouldn't hand-edit a lockfile's resolved versions in a Poetry project.

---

## 3. `go.sum` and Supply-Chain Integrity

`go.sum` is the piece with no direct visual analogue in a `requirements.txt`-only Python project, but a close conceptual match to a Poetry/`uv` lockfile's hash-pinning behavior. A representative excerpt:

```
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb59gIkOpc=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/stretchr/testify v1.9.0 h1:HtqpZUEo0GezhkE6ImkgTybt15j5+E7VVfHkMOKbcC4=
github.com/stretchr/testify v1.9.0/go.mod h1:1iN4jSvsBcwHzOaLdxEXhy...
```

Every dependency (direct *and* transitive) gets two hash lines: one for the module's full source content, one for just its `go.mod` file. When you build or test, Go verifies downloaded module content against these hashes and refuses to proceed if they don't match — this is a supply-chain integrity guarantee, not just a version pin: even if `v1.6.0` of some dependency were somehow re-published with different content at the same version tag (which registries are supposed to prevent, but `go.sum` doesn't have to trust that promise), your build would fail loudly rather than silently using different code than what you tested against.

**You should commit `go.sum` to version control**, exactly as you'd commit `poetry.lock`/`uv.lock` — it's what makes a `go build` reproducible for every teammate and in CI, not just "approximately the same versions" but byte-for-byte the same module content. Never hand-edit it; it's entirely machine-managed via `go mod tidy`/`go get`/`go build`.

---

## 4. Semantic Versioning and Version Selection

Go Modules assumes every dependency follows [semantic versioning](https://semver.org/): `vMAJOR.MINOR.PATCH`, e.g. `v1.6.0`. The convention (not enforced by the compiler, but treated as a hard contract by tooling and community norms) is:

- **PATCH** (`v1.6.0` → `v1.6.1`) — bug fixes, no API changes. Always safe to upgrade.
- **MINOR** (`v1.6.0` → `v1.7.0`) — new functionality, backward compatible. Safe to upgrade in the overwhelming majority of cases.
- **MAJOR** (`v1.6.0` → `v2.0.0`) — breaking changes. Requires a special mechanism, covered in Section 6, because Go treats a new major version as *effectively a different module*.

This mirrors Python's own semver convention (`^1.6.0` in a `pyproject.toml` meaning "anything compatible up to but not including 2.0.0" is a *caret* range built on the same major/minor/patch assumption), but Go's dependency resolution algorithm — called **Minimum Version Selection (MVS)** — behaves differently from `pip`'s or Poetry's resolver in a way worth understanding:

> **When multiple parts of your dependency graph require different versions of the same module, Go picks the *highest* version requested by anyone in the graph** — not the newest version that exists, and not a range-satisfying "latest compatible" the way `^1.6.0` does in a Python lockfile resolution. If your code requires `v1.2.0` of some library and a dependency of yours requires `v1.4.0`, your build uses `v1.4.0` — the minimum version that satisfies *everyone's* stated minimum, which in practice is just "the maximum of all the minimums."

This is deliberately simpler and more deterministic than the constraint-solving `pip`/Poetry do (which can involve genuine SAT-style backtracking when constraints conflict) — Go's algorithm never needs to backtrack, because "take the max of the requested minimums" always has exactly one answer. The tradeoff is that Go pushes much harder on **library authors never making breaking changes within a major version**, since there's no range syntax (no Go equivalent of `^1.6.0` meaning "anything up to 2.0.0 except what I've explicitly excluded") giving you an escape hatch if a supposedly-compatible minor release actually breaks something.

---

## 5. Adding, Upgrading, and Removing Dependencies

```bash
# Add a dependency (fetches latest compatible version, updates go.mod + go.sum)
go get github.com/go-chi/chi/v5

# Add a specific version
go get github.com/go-chi/chi/v5@v5.1.0

# Upgrade a dependency to its latest minor/patch
go get -u github.com/go-chi/chi/v5

# Upgrade EVERYTHING to latest minor/patch
go get -u ./...

# Remove unused dependencies and add any missing ones your code actually needs —
# reconciles go.mod/go.sum with what your source code actually imports
go mod tidy

# See why a dependency is in your graph at all
go mod why github.com/some/transitive-dep

# Visualize the full dependency graph
go mod graph
```

`go mod tidy` deserves a callout because it's the command you'll run most often, and it has no single-command Python equivalent — it's part `pip-compile`, part manually deleting unused entries from a lockfile. It does two things at once: **adds** any module to `go.mod`/`go.sum` that your source code imports but that isn't yet recorded, and **removes** any module that's recorded but that nothing in your source code actually imports anymore (say, after you delete a feature that used it). Running it after every meaningful set of code changes — especially before committing — is close to a universal Go habit, the same way you might run `poetry lock --no-update` or check `pip check` before a commit.

Deleting an `import` line from your Go source does **not** automatically remove the dependency from `go.mod` — you have to run `go mod tidy` for that cleanup to happen, which is a common point of confusion coming from ecosystems where uninstalling is a separate explicit step (`pip uninstall`) but Go's removal is instead "make the manifest match reality" rather than "explicitly evict this package."

---

## 6. Semantic Import Versioning: the `v2+` Rule

This is one of the more surprising Go-specific conventions, with genuinely no Python parallel. Because Go's Minimum Version Selection assumes a module's public API never breaks within a major version, when a library legitimately needs a breaking change, **the major version becomes part of the import path itself**:

```go
import "github.com/go-chi/chi/v5"    // major version 5, note the /v5
```

Contrast with a hypothetical `v1` (or `v0`) import of the same library, which has no version suffix at all:

```go
import "github.com/go-chi/chi"       // implicitly v0 or v1, no suffix
```

The practical consequence: **`github.com/some/lib` and `github.com/some/lib/v2` are, as far as the Go toolchain is concerned, two entirely separate modules** that happen to share a lineage — you can even import both at once in the same program (useful during a gradual migration), the same way you might in Python temporarily depend on both `some-lib` and `some-lib-v2` as separately-named PyPI packages while migrating, except Go bakes this pattern directly into the import-path convention rather than leaving it to whatever naming scheme a package maintainer improvises.

This matters practically the moment you start depending on any library at `v2.x.x` or higher — you'll see `/v2`, `/v3`, etc. baked right into the import statement, and it's not optional or cosmetic; a library at major version 2+ that *didn't* follow this convention would be violating the module system's assumptions and could produce version-selection inconsistencies for anyone depending on it alongside another module that transitively depends on a different major version.

---

## 7. Workspaces: `go.work` for Multi-Module Development

If you're developing two modules simultaneously — say, a shared internal library and an application that consumes it, both under active local development — you'd otherwise need to publish/tag every small change to the library before the application could pick it up, or resort to manual `replace` directives in `go.mod`. **Workspaces** solve this:

```bash
go work init ./orderapi ./sharedlib
```

This generates a `go.work` file:

```
go 1.23

use (
	./orderapi
	./sharedlib
)
```

With this in place, any command run from within the workspace directory (`go build`, `go test`, your editor's language server) resolves imports of `sharedlib` from your **local, uncommitted, in-progress copy** rather than whatever version is pinned in `orderapi`'s own `go.mod` — without editing `orderapi/go.mod` at all. This is conceptually similar to Python's **editable installs** (`pip install -e .` or Poetry's path dependencies) — "use my local working copy of this dependency instead of fetching a published version" — except scoped at the workspace level rather than requiring each consuming project to individually opt in via its own install step.

`go.work` is a local development convenience and, by strong convention, is **not committed to version control** (it typically goes in `.gitignore`) — it describes *your* local multi-module layout, not something that should affect how CI or teammates build the project, the same way you wouldn't commit a personal editable-install configuration that only makes sense on your machine.

For the roadmap ahead, you likely won't need `go.work` until the capstone if you split it into multiple modules — but recognizing it in the wild (and knowing it's not something to copy into `go.mod` by mistake) is worth having now.

---

## 8. The `testing` Package: Go's Built-in `pytest`

This is the section that earns the rest of the guide's existence. Go ships a testing framework in the standard library — no `pip install pytest` equivalent needed, ever. The conventions are rigid (deliberately so — one less thing to configure or disagree about across a team) and worth learning as a fixed set of rules:

- Test files are named `xxx_test.go`, living alongside the code they test (not in a separate `tests/` directory, though you *can* organize larger integration-style suites separately if you want).
- Test functions are named `TestXxx(t *testing.T)` — must start with a capital `Test`, take exactly one parameter of type `*testing.T`, and return nothing.
- Run tests with `go test` (current package) or `go test ./...` (every package in the module, recursively) — no separate test-runner installation, no `conftest.py`, no plugin ecosystem to configure just to get started.

```go
// math.go
package mathutil

func Add(a, b int) int {
	return a + b
}
```

```go
// math_test.go
package mathutil

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}
}
```

```bash
go test ./...
# ok      github.com/pedro/mathutil    0.002s
```

Note immediately: **there is no `assertEqual`, no `assert result == 5`.** You write a plain `if` statement and call `t.Errorf` (or `t.Fatalf`) yourself when the condition indicates failure. This is such a departure from `pytest`'s bare `assert` statements (or `unittest`'s `self.assertEqual`) that it gets its own section (12) — but first, the two failure-reporting methods that matter:

- **`t.Errorf(format, args...)`** — records the test as failed, logs the formatted message, **and continues running the rest of the test function**. Use this when subsequent assertions in the same test are still meaningful even after one has failed (so you see *all* the failures in one run, not just the first).
- **`t.Fatalf(format, args...)`** — records the test as failed, logs the message, and **immediately stops executing the current test function** (via a call to `runtime.Goexit()` internally). Use this when a failure means continuing is pointless or would panic — e.g., a setup step failed, so there's no point asserting on data that was never created.

```go
func TestFetchUser(t *testing.T) {
	user, err := FetchUser(42)
	if err != nil {
		t.Fatalf("FetchUser(42) returned error: %v", err) // stop — nothing below is safe to check
	}
	if user.Name != "Pedro" {
		t.Errorf("user.Name = %q; want %q", user.Name, "Pedro") // keep going — independent check
	}
	if user.ID != 42 {
		t.Errorf("user.ID = %d; want %d", user.ID, 42) // keep going — independent check
	}
}
```

| Concept | `pytest` | Go `testing` |
|---|---|---|
| Discovery | Files matching `test_*.py` / `*_test.py`, functions `test_*` | Files `*_test.go`, functions `TestXxx` |
| Run all tests | `pytest` | `go test ./...` |
| Assertion | Bare `assert x == y` (rewritten by pytest for rich diffs) | Manual `if` + `t.Errorf`/`t.Fatalf` |
| Stop on failure within a test | `assert` raises immediately, halting the test | `t.Fatalf` halts; `t.Errorf` doesn't |
| Fixtures | `@pytest.fixture`, dependency-injected by parameter name | No direct equivalent — plain functions + `TestMain` (Section 11) |
| Parametrize | `@pytest.mark.parametrize` | Table-driven tests + `t.Run` (Sections 9–10) |
| Setup/teardown | Fixture `yield`, `setup_method`/`teardown_method` | `TestMain`, or manual `defer` inside a test |
| Mocking | `unittest.mock`, `pytest-mock` | Interfaces + hand-written fakes (Guide 4) — rarely a dedicated mocking library |
| Coverage | `pytest-cov` (external plugin) | `go test -cover` (built in) |
| Benchmarks | `pytest-benchmark` (external plugin) | `go test -bench` (built in) |

That last handful of rows is the throughline: functionality that requires installing a `pytest` plugin is, in Go, built directly into the `go test` command and the `testing` package. There's less to configure, but also less flexibility than `pytest`'s fixture/plugin ecosystem — a tradeoff consistent with Go's general "one obvious way, fewer knobs" philosophy you've seen in every guide so far.

---

## 9. Table-Driven Tests

This is the single most important idiom in this guide, and the one that most replaces `@pytest.mark.parametrize` in daily use. Instead of writing one `TestXxx` function per case, you define a slice of structs — each one an input/expected-output pair — and loop over it:

```go
func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{name: "two positives", a: 2, b: 3, want: 5},
		{name: "negative and positive", a: -1, b: 1, want: 0},
		{name: "both negative", a: -2, b: -3, want: -5},
		{name: "zero and zero", a: 0, b: 0, want: 0},
	}

	for _, tt := range tests {
		got := Add(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
```

The anonymous struct slice (`[]struct{ ... }{...}`, an inline type from Guide 3's struct literal syntax) is doing exactly what a `@pytest.mark.parametrize` decorator's list of tuples does in Python:

```python
import pytest

@pytest.mark.parametrize("a, b, want", [
    (2, 3, 5),
    (-1, 1, 0),
    (-2, -3, -5),
    (0, 0, 0),
])
def test_add(a, b, want):
    assert add(a, b) == want
```

The Go version is more verbose (no decorator magic, and you write the loop yourself), but it's plain Go — no framework-specific syntax to learn beyond struct literals and a `for range` loop you already know from Guide 2. The `name` field convention deserves its own section, because on its own this loop doesn't give you per-case failure isolation or readable output — that's what `t.Run` adds next.

---

## 10. Subtests with `t.Run`

Without `t.Run`, a failure in one iteration of the table-driven loop above just reports "TestAdd failed" — you'd have to read the error message to figure out *which* case. `t.Run` turns each table entry into its own named, independently-reportable **subtest**:

```go
func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{name: "two positives", a: 2, b: 3, want: 5},
		{name: "negative and positive", a: -1, b: 1, want: 0},
		{name: "both negative", a: -2, b: -3, want: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
```

```bash
go test -v ./...
# === RUN   TestAdd
# === RUN   TestAdd/two_positives
# === RUN   TestAdd/negative_and_positive
# === RUN   TestAdd/both_negative
# --- PASS: TestAdd (0.00s)
#     --- PASS: TestAdd/two_positives (0.00s)
#     --- PASS: TestAdd/negative_and_positive (0.00s)
#     --- PASS: TestAdd/both_negative (0.00s)
```

This is the direct structural equivalent of `pytest`'s parametrized test IDs (`test_add[2-3-5]`, `test_add[-1-1-0]`) — each case gets its own reportable identity, and critically, **you can run just one subtest by name**:

```bash
go test -run "TestAdd/both_negative" -v ./...
```

which mirrors `pytest -k "both_negative"` or `pytest test_math.py::test_add[params]` for isolating a single case while debugging. Note the closure gotcha from Guide 7, Section 2 applies here too: on Go 1.22+ (your toolchain), `tt` is safely scoped per-iteration inside the closure passed to `t.Run`, so you don't need the older `tt := tt` workaround you might see in pre-1.22 codebases.

`t.Run` isn't limited to table-driven loops — you can use it any time you want to group related assertions under readable sub-names, the same way you might split a `pytest` test file into multiple small `test_` functions for readability, except here it's one `TestXxx` function internally organized into named subtests.

---

## 11. Setup, Teardown, and `TestMain`

Go doesn't have `pytest` fixtures, but it has two mechanisms that cover the same ground for the common cases: `defer` inside individual tests, and `TestMain` for package-wide setup/teardown.

**Per-test setup/teardown** is usually just ordinary Go, using `defer`:

```go
func TestWriteFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // teardown, guaranteed on any exit path

	// ... test logic using tmpFile ...
}
```

This is the same shape as a `pytest` fixture using `yield` for teardown (`yield tmp_path; cleanup()`), just spelled with `defer` instead of a generator-based fixture, and declared inline rather than injected by parameter name.

**Package-wide setup/teardown** — run once before *any* test in the package, and once after *all* of them finish (e.g., spinning up a test database connection, or starting/stopping a test server) — uses a special function:

```go
func TestMain(m *testing.M) {
	fmt.Println("setting up shared resources...")
	setupDatabase()

	code := m.Run() // runs all tests in this package, returns exit code

	fmt.Println("tearing down shared resources...")
	teardownDatabase()

	os.Exit(code) // must forward the exit code, or `go test` won't reflect real pass/fail
}
```

If a package defines `TestMain`, `go test` calls it instead of running tests directly — `m.Run()` is what actually executes `TestXxx` functions, and everything before/after that call is your setup/teardown. This is the closest Go gets to `pytest`'s `conftest.py`-level session-scoped fixtures, but note the scope is coarser: it's all-or-nothing for the whole package, not the fine-grained function/class/module/session scoping `pytest` fixtures give you. In practice, most Go test suites use `TestMain` sparingly (mainly for genuinely expensive, package-wide resources) and lean on ordinary helper functions plus `defer` for everything else.

---

## 12. Why Go Has No Assertions (and When to Reach for `testify`)

This is worth addressing head-on because it's the biggest daily-friction difference from `pytest`. Go's standard library deliberately does not provide `assert.Equal(t, want, got)` — the language's own designers have been explicit that they consider assertion libraries an anti-pattern in Go, on the reasoning that a plain `if` + descriptive `t.Errorf` is clearer to read, easier to customize per-case, and doesn't hide control flow behind a library call whose exact failure behavior (does it stop the test? just log?) isn't visible at the call site.

So the idiomatic baseline really is:

```go
if got != want {
	t.Errorf("got %v, want %v", got, want)
}
```

repeated by hand, every time, for every check. This is more typing than `assert got == want`, and it's a real and common early frustration coming from `pytest`. That said, the ecosystem's answer to "I still want assertion helpers" is the third-party library **`stretchr/testify`** (which you may have already noticed as an `// indirect` dependency in the `go.mod` example back in Section 2, since many other libraries pull it in for their own tests) — specifically its `assert` and `require` subpackages:

```go
import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	assert.Equal(t, 5, result, "Add(2, 3) should equal 5") // like t.Errorf — logs, keeps going
}

func TestFetchUser(t *testing.T) {
	user, err := FetchUser(42)
	require.NoError(t, err) // like t.Fatalf — logs, stops the test immediately if it fails
	assert.Equal(t, "Pedro", user.Name)
}
```

The `assert` package mirrors `t.Errorf`'s "log and keep going" behavior; the `require` package mirrors `t.Fatalf`'s "log and stop immediately." This is a genuinely common, well-regarded library in production Go codebases — using it is not considered a beginner's crutch the way, say, reaching for a heavyweight ORM prematurely might be seen in some Python shops. But it's worth having written the manual `if`/`t.Errorf` version enough times first (this guide's earlier sections, and your exercises) that you understand exactly what `testify` is saving you from, rather than reaching for it as an unexamined `pytest`-habit transplant on day one.

---

## 13. Test Helpers and `t.Helper()`

As test suites grow, you'll factor out repeated assertion logic into helper functions — the equivalent of a small local utility function in a `conftest.py`, short of a full fixture:

```go
func assertNoError(t *testing.T, err error) {
	t.Helper() // marks this function as a helper — see below
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchUser(t *testing.T) {
	user, err := FetchUser(42)
	assertNoError(t, err)
	if user.Name != "Pedro" {
		t.Errorf("user.Name = %q; want %q", user.Name, "Pedro")
	}
}
```

**`t.Helper()`** does one specific, easy-to-miss-the-value-of thing: when a failure is reported *from inside* a function marked with `t.Helper()`, Go's test output attributes the failure's file-and-line-number to the **caller** of the helper, not to the line inside the helper itself. Without it, every failure from `assertNoError` would report the line number *inside* `assertNoError` — technically correct, but useless for finding which actual test call site triggered it. This has no direct `pytest` parallel because `pytest`'s assertion rewriting already gives you accurate tracebacks pointing at your real assertion regardless of helper nesting; in Go, `t.Helper()` is the manual mechanism you opt into, in every helper function you write, to get equivalent traceback quality.

---

## 14. Benchmarks with `testing.B`

Benchmarking is built into the same `testing` package and the same `go test` command — no `pytest-benchmark` plugin install required. Benchmark functions live in the same `*_test.go` files, following a parallel naming convention:

```go
func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(2, 3)
	}
}
```

- Named `BenchmarkXxx`, taking `*testing.B` instead of `*testing.T`.
- **`b.N`** is not something you set — the testing framework runs your loop with increasing values of `N` until it's run long enough to produce a statistically stable measurement, then reports time-per-operation. You just write "do the thing once, `b.N` times" and let the framework calibrate how many iterations that means.

Run benchmarks explicitly (they're skipped by a plain `go test`, since they're slower and not pass/fail checks):

```bash
go test -bench=. ./...
# BenchmarkAdd-8    1000000000    0.25 ns/op
```

The `-8` suffix is `GOMAXPROCS` at benchmark time (Guide 7, Section 3) — useful context if you're comparing benchmark runs across machines with different core counts. For benchmarks involving memory allocation, `-benchmem` adds allocation counts to the report, which is often more actionable than raw timing when you're hunting for unnecessary allocations:

```bash
go test -bench=. -benchmem ./...
# BenchmarkAdd-8    1000000000    0.25 ns/op    0 B/op    0 allocs/op
```

If your benchmark needs setup that shouldn't count toward the timed measurement (constructing test data, opening a file), use `b.ResetTimer()` after the setup:

```go
func BenchmarkProcessLargeSlice(b *testing.B) {
	data := generateLargeTestData() // expensive setup, shouldn't be timed
	b.ResetTimer()                  // clock restarts here

	for i := 0; i < b.N; i++ {
		ProcessSlice(data)
	}
}
```

Python's rough equivalent — `pytest-benchmark`, or hand-rolled `timeit` — requires an external dependency and doesn't integrate as tightly with your existing test-discovery and CI setup; here, benchmarks are first-class citizens of the exact same tool and file layout as your correctness tests, which is part of why performance regressions are relatively cheap to catch continuously in Go projects (e.g., running `go test -bench` in CI and diffing against a baseline).

---

## 15. Test Coverage

Also built directly into `go test` — no `pytest-cov` install required:

```bash
go test -cover ./...
# ok    github.com/pedro/mathutil    0.002s    coverage: 87.5% of statements

# Generate a detailed profile you can inspect
go test -coverprofile=coverage.out ./...

# View it as an interactive, line-by-line HTML report in your browser
go tool cover -html=coverage.out
```

The HTML report highlights exactly which lines were and weren't executed by your test suite (green for covered, red for not), the same visual you'd get from `pytest --cov --cov-report=html`, just generated by the standard toolchain rather than a plugin. Coverage percentage is a useful smell-detector, not a target to game — 100% coverage doesn't mean your tests actually assert anything meaningful (a test that calls a function but checks nothing still "covers" that function's lines), which is the same caveat that applies to coverage metrics in Python or any other language. Use it to find *obviously* untested code paths, not as a quality score to optimize in isolation.

---

## 16. `go vet`: Static Analysis Built In

`go vet` examines your source code for constructs that are syntactically valid Go but almost certainly bugs — the kind of thing a careful code reviewer would flag, automated. It ships with the toolchain (no install step) and is conventionally run alongside every test invocation:

```bash
go vet ./...
```

Examples of what it catches: a `Printf`-style format string whose verbs don't match the argument types or count (`fmt.Printf("%d", "a string")`), a struct copied by value that contains a `sync.Mutex` (Guide 7's "never copy a mutex after use" rule, enforced automatically), an `if` condition that's always true or false due to a comparison mistake, and unreachable code after a `return`. None of these are compile errors — the compiler accepts all of them — but they're overwhelmingly likely to be bugs, which is exactly the niche `go vet` fills.

The closest Python tools are static analyzers like `mypy` (type-focused) or `pylint`/`flake8` (broader style-and-bug-pattern focused), except `go vet` requires zero configuration, zero plugin selection, and ships as part of the same binary as the compiler and test runner — `go build`, `go test`, and several other commands **automatically run a relevant subset of `go vet`'s checks before proceeding**, so you'll sometimes see vet-style errors even if you never typed `go vet` yourself. Running the full `go vet ./...` explicitly (as a pre-commit step or CI gate) catches a superset of what's run implicitly.

---

## 17. Linting with `golangci-lint`

`go vet` is deliberately narrow — it only flags things that are near-certainly bugs. Style conventions, unused variables in some contexts, cyclomatic complexity, import ordering, and dozens of other opinionated checks live in the broader **linter** ecosystem, and the de facto standard tool is [`golangci-lint`](https://golangci-lint.run/) — an aggregator that runs dozens of individual linters (including `go vet` itself) under one configuration and one command:

```bash
# Install (one-time, not part of the base toolchain)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run it
golangci-lint run ./...
```

Configuration lives in a `.golangci.yml` file at your module root, where you enable/disable specific linters and tune their thresholds — structurally similar to a `pyproject.toml`'s `[tool.ruff]` or `[tool.pylint]` section, or a standalone `.flake8`/`setup.cfg`. A minimal example:

```yaml
linters:
  enable:
    - govet
    - errcheck    # flags ignored error return values — a very common real bug
    - staticcheck # a large, well-regarded suite of additional static analysis checks
    - unused      # flags unused variables, functions, imports
```

`errcheck` is worth calling out specifically, because it catches a mistake that's easy to make once you're used to Python's exceptions propagating automatically: in Go, an ignored error return value (Guide 5) doesn't halt anything — the program just silently continues with a zero-value result, which can be a much subtler bug than an unhandled Python exception, since Python at least crashes loudly. `errcheck` flags every place you've called a function returning an `error` without checking it, which is close to mandatory in any serious Go codebase.

The overall picture: `go vet` is free, built-in, and narrow; `golangci-lint` is an install step (like `pip install ruff`) that aggregates broad, configurable, opinionated checks. A typical project's CI runs both — `go vet ./...` and `golangci-lint run ./...` — alongside `go test ./...`, the same three-legged stool (`mypy`, `ruff`/`flake8`, `pytest`) you likely already run in your FastAPI projects, just with different tool names and, notably, two of the three (`go vet`, `go test`) requiring zero extra installation at all.

---

## 18. Exercises

1. **`go mod tidy` in practice.** In a scratch module, add an import to a real third-party package (e.g., `github.com/google/uuid`) in your source code without running `go get` first. Run `go build` and observe the error, then run `go mod tidy` and observe it resolve both `go.mod` and `go.sum` automatically. Then delete the import from your source and run `go mod tidy` again — confirm the now-unused dependency disappears from `go.mod`.

2. **Read a real `go.sum`.** Pick any small open-source Go module on GitHub, open its `go.sum`, and find a dependency that appears with more than one version listed. Investigate (via `go mod graph` if you clone it, or just reasoning about the module graph) why two versions of the same transitive dependency might legitimately coexist in the resolution.

3. **Table-driven test from scratch.** Write a function `IsPalindrome(s string) bool` and a table-driven test with at least 6 cases covering: an ordinary palindrome, an ordinary non-palindrome, an empty string, a single character, mixed case (decide and document whether your function is case-sensitive), and a string with spaces. Use `t.Run` for named subtests.

4. **`t.Errorf` vs `t.Fatalf`, deliberately.** Write a test that calls a function returning `(result int, err error)`. Structure the test so that a nil-`err` check is not fatal (using `t.Errorf`) and observe how the test continues to a subsequent nil-pointer-dereference panic. Then change that check to `t.Fatalf` and observe the test stop cleanly before the panic. Explain in a comment why one choice is clearly correct here.

5. **`TestMain` setup/teardown.** Write a package with a `TestMain` that prints `"suite starting"` before `m.Run()` and `"suite finished"` after, wrapping at least two `TestXxx` functions. Run `go test -v ./...` and confirm the setup/teardown messages each print exactly once, regardless of how many tests run.

6. **Benchmark two implementations.** Write two functions that both reverse a string — one using `[]rune` conversion and index-swapping, one using `strings.Builder` appending in reverse — and a `BenchmarkXxx` for each. Run `go test -bench=. -benchmem ./...` and compare allocations and ns/op between the two. Which one wins, and can you explain why from what you know about slices and strings (Guide 2)?

7. **Coverage gap hunting.** Take any function you've written in a previous guide's exercises with at least two branches (an `if`/`else`, or several `case`s), write tests that deliberately cover only one branch, run `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`, and confirm the uncovered branch shows up highlighted. Then add the missing test case and confirm full coverage.

8. **Catch a `go vet` bug on purpose.** Write a function that calls `fmt.Printf("%d", someString)` (a deliberate format-verb mismatch) and confirm `go vet ./...` flags it. Then write a struct containing a `sync.Mutex` field and a function that accidentally copies an instance of that struct by value (not by pointer) — confirm `go vet` flags that too.

---

## 19. Key Takeaways

- **`go.mod`** declares your module's path and direct dependencies (with `// indirect` entries auto-tracked for transitive ones); **`go.sum`** cryptographically pins every module version touched, direct or transitive, for reproducible, tamper-evident builds — together they play the same role as a `pyproject.toml` + lockfile pair, and both should be committed to version control.
- Go has no virtual environments because it doesn't need them: downloaded modules live in a global, version-keyed, read-only cache, and each project's `go.mod` alone determines what gets built — nothing to "activate."
- Go Modules assumes strict semantic versioning and resolves dependencies via **Minimum Version Selection** (always the maximum of everyone's stated minimum version) — simpler and more deterministic than `pip`/Poetry's constraint solving, at the cost of relying more heavily on library authors never breaking compatibility within a major version.
- **`go get`** adds/upgrades dependencies; **`go mod tidy`** is the workhorse command that reconciles `go.mod`/`go.sum` with what your source code actually imports — run it after any import change, before every commit.
- Breaking (major-version) changes to a Go module are encoded directly in the **import path** (`/v2`, `/v3`, ...) — semantic import versioning — letting two major versions of the same module coexist in one build without conflict, a mechanism with no real Python parallel.
- **`go.work`** lets you develop multiple local modules together without publishing intermediate versions, similar in spirit to Python editable installs, but it's a local dev convenience that's conventionally excluded from version control.
- The **`testing` package** is Go's built-in `pytest`: `*_test.go` files, `TestXxx(t *testing.T)` functions, run via `go test ./...` — zero install required. `t.Errorf` fails and continues; `t.Fatalf` fails and stops the current test immediately.
- **Table-driven tests** (a slice of input/expected-output structs, looped over) are the primary replacement for `@pytest.mark.parametrize`; wrapping each case in **`t.Run(tt.name, ...)`** gives you independently reportable, independently runnable named subtests — the equivalent of `pytest`'s parametrized test IDs and `-k` filtering.
- Go deliberately has **no built-in assertion library** — plain `if` + `t.Errorf`/`t.Fatalf` is the idiomatic default, on the reasoning that it's more explicit and readable than hidden assertion magic. `stretchr/testify`'s `assert`/`require` packages are a well-regarded, widely-used opt-in if you want `pytest`-style assertions once you understand what they're abstracting over.
- **`t.Helper()`** marks a function as a test helper so that failure line numbers are attributed to the caller, not the helper itself — the manual mechanism Go needs in place of `pytest`'s automatic assertion-rewriting tracebacks.
- **`TestMain(m *testing.M)`** gives you package-wide setup/teardown around `m.Run()` — coarser-grained than `pytest` fixture scoping, but the closest built-in equivalent; ordinary per-test setup/teardown just uses `defer`.
- **Benchmarks** (`BenchmarkXxx(b *testing.B)`, run via `go test -bench=.`) and **coverage** (`go test -cover`, `go tool cover -html=...`) are both built directly into the standard toolchain, unlike Python's `pytest-benchmark` and `pytest-cov` plugins.
- **`go vet`** is a free, built-in, narrowly-scoped static analyzer catching near-certain bugs (bad format verbs, mutex copies, unreachable code) — partially run automatically by `go build`/`go test` even if you never invoke it directly.
- **`golangci-lint`** is the de facto community-standard aggregator for broader, configurable style and correctness linting (including `errcheck`, which catches ignored `error` return values — a distinctly Go-flavored bug class, since ignoring an error doesn't crash anything the way an unhandled Python exception would) — an install step, like `ruff` or `pylint`, layered on top of the free built-in tools.

---

**Next up: Guide 9** — let me know what's next in your sequence (continuing Phase D with more on data/serialization, or moving into the web layer) and I'll pick up in the same format.
