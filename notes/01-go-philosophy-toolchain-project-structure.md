# Guide 1 — Go Philosophy, Toolchain & Project Structure

> **Roadmap position:** Phase A, Guide 1 of 14 — *Foundational / conceptual, minimal code*
> **Goal of this guide:** Build the correct mental model for *why* Go looks and behaves the way it does, before you write a single line of "real" Go logic. Every contrast made here to Python exists because you're coming from a dynamically-typed, interpreted, batteries-included ecosystem, and Go will actively fight you if you try to write it like Python with different syntax.

---

## Table of Contents

1. [Why This Guide Exists (and Why It Has Almost No Code)](#1-why-this-guide-exists-and-why-it-has-almost-no-code)
2. [Go's Design Philosophy](#2-gos-design-philosophy)
3. [Compiled vs Interpreted: What Actually Changes for You](#3-compiled-vs-interpreted-what-actually-changes-for-you)
4. [Static Typing vs Dynamic Typing](#4-static-typing-vs-dynamic-typing)
5. [Garbage Collection in Go vs Python](#5-garbage-collection-in-go-vs-python)
6. [Installing Go & Configuring Your Environment](#6-installing-go--configuring-your-environment)
7. [The Go Toolchain, Command by Command](#7-the-go-toolchain-command-by-command)
8. [Go Modules: Dependency Management](#8-go-modules-dependency-management)
9. [Standard Project Layout & Conventions](#9-standard-project-layout--conventions)
10. [Exported vs Unexported: Go's Encapsulation Model](#10-exported-vs-unexported-gos-encapsulation-model)
11. [Your First Go Program, Dissected Line by Line](#11-your-first-go-program-dissected-line-by-line)
12. [Editor & Tooling Setup](#12-editor--tooling-setup)
13. [Summary & What's Next](#13-summary--whats-next)

---

## 1. Why This Guide Exists (and Why It Has Almost No Code)

When you learned FastAPI, you already understood Python. You knew what a function was, what an exception was, how imports worked, how the interpreter behaved. Learning FastAPI was really just learning a *library* on top of a language you already had internalized.

Go is different. You are not just learning a new library — you are learning a language with a genuinely different execution model, type system, error-handling philosophy, and concurrency model. If you skip straight to "how do I write a handler," you will write Go code that *compiles* but is fundamentally non-idiomatic — the Go equivalent of writing Java-style getters/setters in Python. It works, but every experienced Go developer will immediately recognize you learned the syntax without learning the language.

This guide has almost no application-level code on purpose. Its job is to install the correct *assumptions* in your head so that every guide after this one (syntax, structs, interfaces, error handling, concurrency) lands on solid ground instead of being reinterpreted through a Python lens.

**Key mindset shift to hold onto throughout this entire roadmap:**

> Python optimizes for *how fast you can write code*. Go optimizes for *how fast a team can read, maintain, and safely modify code at scale, years later*. Almost every design decision in Go — verbosity, explicit error returns, no exceptions, no generics until 1.18, no operator overloading, gofmt enforcing one style — traces back to this single priority.

---

## 2. Go's Design Philosophy

### 2.1 Origins and Motivation

Go was created at Google in 2007 (publicly released in 2009) by Robert Griesemer, Rob Pike, and Ken Thompson — names with deep roots in systems programming (Pike and Thompson worked on Unix and Plan 9; Thompson co-created Unix and C). Google's motivation wasn't academic. They had a very concrete, very large-scale engineering problem:

- **Massive codebases** with thousands of engineers touching the same repositories.
- **Slow build times** in C++ at Google's scale (single builds taking many minutes to hours).
- **Difficulty onboarding** engineers into large, inconsistent C++/Java codebases.
- **A growing need for native concurrency** in networked, multi-core server software — this was the era when multi-core CPUs became standard and existing languages' concurrency stories (threads + locks in C++/Java) were painful and error-prone.

So Go was explicitly designed to be:

- **Fast to compile** — even huge codebases build in seconds, not minutes.
- **Simple enough that any engineer on the team can read any other engineer's code** without needing to know that engineer's personal style or favorite abstractions.
- **Safe and boring** — few ways to do the same thing, minimal "clever" syntax tricks, no metaprogramming magic.
- **Concurrent by design**, not concurrency bolted on as a library later.

### 2.2 The "One Way To Do It" Enforcement

Python's philosophy (per PEP 20 / the Zen of Python) *says* "there should be one — and preferably only one — obvious way to do it," but in practice Python gives you many idioms: list comprehensions vs loops, `%` formatting vs `.format()` vs f-strings, classes vs dataclasses vs namedtuples, multiple ways to structure a project, and no enforced code formatter until the community adopted `black` voluntarily.

Go takes this philosophy and **enforces it with tooling, not just convention**:

- `gofmt` (and `go fmt`) is a single canonical formatter. There is no configuration. Every Go codebase on Earth is formatted identically. You will never have a tabs-vs-spaces or brace-placement debate in a Go team, ever — the tool decides, full stop.
- There is (traditionally) no ternary operator, no operator overloading, no implicit type conversions, no exceptions, no inheritance. Fewer features means fewer ways to solve the same problem differently.
- The language spec itself is short enough to read in an afternoon (unlike, say, the C++ spec).

**Why this matters for you:** coming from Python, where flexibility and expressiveness are core virtues, you'll initially feel that Go is "fighting" you — unused imports are compile errors, unused variables are compile errors, there's no list comprehension, error handling looks repetitive. This is not an accident or an oversight. It is the entire point. Go trades your typing speed for the team's long-term reading speed and correctness.

### 2.3 Composition Over Inheritance, Explicit Over Implicit

Two philosophical threads you'll see recur constantly in later guides:

- **Composition over inheritance:** Go has no classes and no inheritance. You build behavior by embedding structs and satisfying interfaces, not by extending base classes. (Covered in depth in Guide 3.)
- **Explicit over implicit:** errors are returned values you must explicitly check (`if err != nil`), not exceptions that silently propagate up a call stack until someone remembers to catch them. Nothing happens "automatically" behind your back. (Covered in depth in Guide 5.)

Hold onto both ideas loosely for now — you'll get hands-on with them soon. The point here is just to recognize that these aren't arbitrary syntax quirks; they follow directly from the "readable, boring, safe at scale" philosophy above.

---

## 3. Compiled vs Interpreted: What Actually Changes for You

### 3.1 What Python Actually Does

Python is technically compiled to bytecode (`.pyc` files), but that bytecode is then interpreted by the CPython virtual machine at runtime. Practically, this means:

- You run `python app.py` and the interpreter reads, compiles-to-bytecode, and executes more or less simultaneously.
- Type errors, `NameError`s, and typos in a rarely-executed branch of code can lie dormant for months and only surface when that code path finally runs in production.
- Distribution requires shipping either your source code or bundling an interpreter (e.g., via Docker, PyInstaller, etc.) — you can't just hand someone one executable file and expect it to run without a Python installation (PyInstaller-style bundling aside).

### 3.2 What Go Actually Does

Go is **ahead-of-time (AOT) compiled** directly to native machine code for a specific OS/architecture (e.g., linux/amd64, darwin/arm64). This produces a **single, statically-linked binary** with no external runtime dependency (Go's small runtime — including its garbage collector and scheduler — is compiled *into* that binary, not installed separately).

Concrete consequences for you as a developer:

| Aspect | Python | Go |
|---|---|---|
| Errors caught | Mostly at runtime | Many caught at **compile time** (type mismatches, unused variables/imports, wrong function signatures) |
| Deployment artifact | Source + interpreter (or container) | Single native binary, no runtime needed |
| Startup time | Interpreter startup + import overhead | Near-instant (already-compiled machine code) |
| Cross-compilation | Not really applicable | Trivial — `GOOS=linux GOARCH=amd64 go build` compiles a Linux binary *from your Mac*, no VM/container needed |
| "Does it even run" feedback loop | Run it and see | Compiler often tells you before you run it |

This is the single biggest quality-of-life shift you'll feel in week one: **an entire category of bugs (typos, wrong argument counts, wrong types passed to a function, forgetting to handle a return value) simply cannot compile in Go.** You'll get immediate, precise compiler errors instead of a traceback five minutes into a script's execution.

The trade-off: you must satisfy the compiler *before* you can run anything at all, even to quickly test one line. This feels slower moment-to-moment coming from Python's REPL-like edit-run-repeat loop, but compilation is so fast (seconds, even for large projects) that in practice it's a minor friction, not a real bottleneck.

### 3.3 Why This Matters for the Web/API Track You're Heading Toward

When you get to Guide 10–14 (building REST APIs), this distinction becomes very concrete: a Go binary you build for your API is a **single self-contained executable** you can `scp` to a server or `COPY` into a minimal Docker `scratch`/`distroless` image with no Python interpreter, no `pip install`, no virtualenv, no dependency resolution happening at deploy time. All dependency resolution happens once, at build time, on your machine or CI server.

---

## 4. Static Typing vs Dynamic Typing

### 4.1 Python's Dynamic Typing (With Optional Hints)

Python variables don't have a fixed type — they're references that can point to any object, and the *object* carries the type, not the variable:

```python
x = 5
x = "now I'm a string"  # perfectly legal
```

Type hints (`x: int`) are optional, checked only by external tools like `mypy`, and never enforced by the interpreter itself at runtime unless you add your own validation (which is exactly the gap Pydantic fills in FastAPI).

### 4.2 Go's Static Typing

In Go, every variable has a fixed, known type determined **at compile time**, and that type can never change:

```go
var x int = 5
x = "now I'm a string" // compile error: cannot use "now I'm a string" (untyped string constant) as int value
```

Function signatures declare exact parameter and return types, and the compiler enforces that every call site matches:

```go
func Add(a int, b int) int {
    return a + b
}

Add(2, "three") // compile error: cannot use "three" (untyped string constant) as int value in argument to Add
```

This is roughly the equivalent of what Pydantic + FastAPI give you at your API's *boundary* (validating incoming JSON) — except in Go, this discipline applies to **every function in your entire codebase**, not just at the HTTP boundary. You don't need a validation library to know that a function receiving a `User` struct actually received a `User` struct; the compiler already guaranteed it before the program could even build.

### 4.3 Type Inference — Go Isn't as Verbose as You Might Fear

Go does have type inference via the `:=` short variable declaration operator, so you don't have to spell out every type explicitly:

```go
x := 5          // compiler infers int
name := "Caio"  // compiler infers string
```

This is *not* the same as Python's dynamic typing — `x` is still permanently and immutably typed as `int` after this line; the compiler simply figured out the type for you instead of you writing `var x int = 5`. You'll use `:=` constantly; it's the idiomatic default for local variables.

### 4.4 Why This Matters Given Your Background

You already have a taste of the value of static-ish typing because you use Pydantic models and type hints heavily in FastAPI. Think of Go as: **that discipline, but load-bearing and compiler-enforced everywhere, all the time, not optional and not just at request/response boundaries.** You'll likely find this *reduces* certain categories of bugs you're used to chasing down in Python (wrong type passed three functions deep, `None` where an object was expected, etc.), at the cost of writing a bit more upfront declaration.

---

## 5. Garbage Collection in Go vs Python

Both languages are garbage-collected — meaning you, the developer, never manually `free()` memory like in C — but the *mechanisms* differ, and it's worth understanding at a conceptual level since it affects how you reason about performance later.

### 5.1 Python's Memory Management

CPython uses two mechanisms together:

1. **Reference counting** — every object carries a count of how many references point to it. When the count hits zero, the object is immediately deallocated.
2. **A generational cyclic garbage collector** — reference counting alone can't detect *cycles* (object A references B, B references A, neither is reachable from anywhere else), so Python periodically runs a separate collector to find and clean up these cycles.

Consequence you may have felt already: reference counting means deallocation is deterministic and immediate in the common case, but the GIL (Global Interpreter Lock) — a related but separate mechanism — means only one thread executes Python bytecode at a time, which is *the* central limitation that pushes Python toward multiprocessing or async I/O for concurrency rather than true multi-threaded parallelism.

### 5.2 Go's Garbage Collector

Go uses a **concurrent, tri-color mark-and-sweep garbage collector**. In plain terms:

- It periodically scans your program's object graph, starting from "roots" (global variables, stack variables), marking every object that's still reachable.
- Anything not marked is considered garbage and its memory is reclaimed.
- Critically, this collector runs **concurrently with your program** — your goroutines (Go's lightweight threads, covered in Guide 7) keep running while collection happens, with only very brief "stop-the-world" pauses (typically sub-millisecond in modern Go versions) to coordinate.
- Go has **no GIL**. Multiple goroutines can execute Go code on multiple OS threads truly in parallel across CPU cores, simultaneously.

### 5.3 Why This Matters for You

This is precisely *why* Go is the language of choice for high-concurrency network services (API servers, proxies, infrastructure tools like Docker and Kubernetes, both written in Go): the combination of no GIL + lightweight goroutines + a GC designed not to stall the program means you can genuinely run tens of thousands of concurrent connections/goroutines without the concurrency ceiling Python's GIL imposes. You'll feel the practical implications of this directly in Guide 7 (Goroutines & Channels) and again in Guide 11 when your REST API handles concurrent requests natively, with no need for an ASGI event loop or `async`/`await` ceremony (Go's concurrency model is fundamentally different — no async/await keywords exist at all, because goroutines make blocking calls "cheap" in a way Python's synchronous calls aren't).

---

## 6. Installing Go & Configuring Your Environment

### 6.1 Installation

Go ships as a simple, official installer/binary distribution — no equivalent of pyenv/conda complexity is needed for basic use, though version managers exist (`g`, `gvm`) if you want to juggle multiple Go versions later.

**macOS (via Homebrew):**
```bash
brew install go
```

**Linux (manual, most reliable method):**
```bash
# Download the latest version from https://go.dev/dl/ - example shown for a specific version
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
```
Then add to your shell profile (`~/.bashrc`, `~/.zshrc`):
```bash
export PATH=$PATH:/usr/local/go/bin
```

**Windows:** use the official MSI installer from go.dev/dl.

Verify installation:
```bash
go version
# go version go1.23.0 linux/amd64
```

### 6.2 The Environment Variables You Actually Need to Understand

Unlike Python's `PYTHONPATH`/virtualenv conceptual model, Go has its own set of environment variables — but modern Go (1.16+) has drastically simplified what you need to think about day to day, since Go Modules (Section 8) replaced the old `GOPATH`-centric workflow.

| Variable | Purpose | Do you need to set it? |
|---|---|---|
| `GOROOT` | Where the Go installation itself lives (standard library, compiler, tools) | No — auto-detected from install location |
| `GOPATH` | Legacy: where Go used to require all your code and dependencies to live. Today: still used for the global module cache (`$GOPATH/pkg/mod`) and installed binaries (`$GOPATH/bin`) | No — has a sensible default (`~/go`), but worth knowing it exists |
| `GO111MODULE` | Legacy toggle for enabling Modules vs GOPATH mode | No — Modules are the default and only sane mode since Go 1.16 |
| `GOBIN` | Where `go install` puts compiled binaries | No — defaults to `$GOPATH/bin`; you may want this on your `PATH` |

**Practical takeaway:** you can install Go and start a project today without configuring anything beyond adding Go's `bin` directory to your `PATH`. This is a deliberate contrast to the historical GOPATH-era pain and is much closer to the "it just works" experience you're used to with a fresh Python virtualenv.

### 6.3 Check Your Setup

```bash
go env GOROOT GOPATH
go version
```

If both commands return sensible output, you're ready.

---

## 7. The Go Toolchain, Command by Command

Python's tooling is a patchwork you assembled yourself over time: `pip` for packages, `black`/`ruff` for formatting/linting, `pytest` for testing, maybe `mypy` for type-checking, a `venv` for isolation. Go ships **all of this as one cohesive `go` command** with subcommands. This is one of the toolchain's biggest ergonomic wins and one you'll come to appreciate quickly.

### 7.1 `go run` — Compile and Execute Immediately

```bash
go run main.go
```
Compiles the program to a temporary location and executes it immediately — this is your closest equivalent to `python app.py` for quick iteration, though remember: it's still fully compiling first, just transparently.

### 7.2 `go build` — Produce a Binary

```bash
go build -o myapp .
```
Compiles your program into a standalone executable named `myapp` in the current directory. This binary has no external dependencies — you can copy it to another machine with the same OS/architecture and run it with zero setup.

### 7.3 `go fmt` — The Non-Negotiable Formatter

```bash
go fmt ./...
```
Reformats every `.go` file in your project (the `./...` pattern means "this directory and all subdirectories recursively" — you'll see `./...` constantly in Go tooling commands). There are no configuration options. This is intentional: it removes formatting bikeshedding entirely from Go teams.

### 7.4 `go vet` — Static Analysis for Suspicious Code

```bash
go vet ./...
```
Catches things that compile successfully but are almost certainly bugs — e.g., a `Printf`-style format string that doesn't match its arguments, or a struct copied when it shouldn't be. Think of this as a much narrower, Go-standard-library-provided equivalent of `pylint`/`ruff` catching logic smells rather than style issues.

### 7.5 `go test` — Built-in Test Runner

```bash
go test ./...
```
Runs all tests in the project (files ending in `_test.go`). No separate `pytest` install required — testing is a first-class part of the standard toolchain. Covered in depth in Guide 8.

### 7.6 `go install` — Build and Install a Binary Globally

```bash
go install github.com/some/tool@latest
```
Downloads, builds, and installs a Go program's binary into `$GOBIN`/`$GOPATH/bin`, making it available as a global CLI command — this is how you'll install most Go developer tools (linters, code generators, etc.), roughly analogous to `pipx install some-tool` in the Python world.

### 7.7 `go doc` — Documentation Without Leaving the Terminal

```bash
go doc fmt.Println
```
Go's standard library and any well-documented package expose their documentation directly through this command (and via pkg.go.dev online) — comments directly above a function/type/package declaration *are* the documentation, no separate docstring convention or Sphinx-style build step needed.

### 7.8 A Quick Command Reference Table

| Command | Python rough equivalent | Purpose |
|---|---|---|
| `go run` | `python script.py` | Compile + run immediately, no artifact kept |
| `go build` | N/A (Python doesn't compile to a binary) | Produce a standalone executable |
| `go fmt` | `black .` | Canonical, non-configurable code formatting |
| `go vet` | `ruff`/`pylint` (logic-focused subset) | Catch suspicious-but-compiling code |
| `go test` | `pytest` | Run the test suite |
| `go install` | `pipx install` | Install a Go binary globally |
| `go mod` | `pip` + `requirements.txt`/`poetry` | Dependency management (see Section 8) |
| `go doc` | `help()` / Sphinx | Inline documentation lookup |

---

## 8. Go Modules: Dependency Management

### 8.1 The Problem Modules Solve

Before Go 1.11 (2018), all Go code had to live inside a single global `GOPATH` workspace, with no real per-project dependency versioning — a genuinely painful era that long-time Go developers still wince at. **Go Modules** (stable default since Go 1.16) solved this properly, and today it's the only workflow you need to know.

### 8.2 `go.mod` — Your `pyproject.toml`/`requirements.txt` Equivalent

Initialize a new module:

```bash
go mod init github.com/yourusername/myproject
```

This creates a `go.mod` file:

```
module github.com/yourusername/myproject

go 1.23

require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/go-playground/validator/v10 v10.22.0
)
```

Key things to notice, contrasted with Python packaging:

- The **module name is typically a repository URL path** (e.g., `github.com/yourusername/myproject`), not an arbitrary name like a PyPI package name. This is because Go's dependency resolution fetches source code directly from version control hosts (GitHub, etc.) rather than from a centralized package index like PyPI — there is a central *proxy/cache* (`proxy.golang.org`) for performance and availability, but conceptually, packages are identified by their source location.
- `go 1.23` pins the language version the module was written against (roughly analogous to specifying a Python version constraint).
- `require` lists direct dependencies with **exact semantic versions** — no loose version ranges like `requests>=2.0,<3.0`. Go resolves a single exact version per dependency for the whole build (Minimal Version Selection algorithm), which produces far more reproducible builds by default than pip's resolver historically has.

### 8.3 `go.sum` — Your Lockfile

Alongside `go.mod`, Go generates a `go.sum` file containing **cryptographic checksums** of every dependency (and its transitive dependencies) to guarantee that what you download today is byte-for-byte identical to what a teammate or CI server downloads tomorrow. This is conceptually your `poetry.lock`/`Pipfile.lock` equivalent, but it ships as a standard, mandatory part of every Go project — not an optional add-on tool.

### 8.4 Adding and Managing Dependencies

```bash
go get github.com/go-chi/chi/v5        # add/upgrade a dependency
go mod tidy                             # add missing + remove unused dependencies
go mod download                         # download dependencies into the local cache
go list -m all                          # list the full dependency tree
```

`go mod tidy` is the command you'll run constantly — it's the closest analog to `pip freeze` reconciliation, keeping `go.mod`/`go.sum` in sync with what your code actually imports.

### 8.5 No Virtual Environments — And Why You Won't Miss Them

There is no Go equivalent of a Python virtualenv, and you don't need one. Since dependencies are resolved and versioned per-module via `go.mod`/`go.sum` (with a shared, content-addressed global cache at `$GOPATH/pkg/mod` deduplicating identical dependency versions across all your projects), there's no risk of one project's dependencies polluting another's, and no "activate" step. You `cd` into a project directory and every Go tool automatically knows exactly which dependencies and versions apply, based on the nearest `go.mod` file.

---

## 9. Standard Project Layout & Conventions

### 9.1 There Is No Single Official Layout — But There Is a De Facto Standard

Unlike, say, a Django project (which has an enforced structure from `django-admin startproject`), Go itself imposes almost no structure — a single `main.go` file is a completely valid, complete Go program. However, the community (through the widely-adopted, unofficial but near-universal **["Standard Go Project Layout"](https://github.com/golang-standards/project-layout)**) converged on strong conventions you'll see in nearly every serious Go codebase, including the ones you'll build toward in Guide 14.

### 9.2 The Core Directories You'll Actually Use

```
myproject/
├── cmd/
│   └── api/
│       └── main.go          # entry point(s) — thin, just wiring
├── internal/
│   ├── handler/              # HTTP handlers
│   ├── service/               # business logic
│   ├── repository/            # data access layer
│   └── model/                 # domain types/structs
├── pkg/                       # code intended to be importable by external projects
├── config/                    # configuration loading
├── migrations/                 # SQL migration files
├── go.mod
├── go.sum
└── README.md
```

**What each directory means and why it exists:**

- **`cmd/`** — holds one subdirectory per *entry point binary* your module produces. A project might build both an `api` server and a separate `worker` binary — each gets its own `cmd/<name>/main.go`. Each `main.go` should be thin: parse config, wire dependencies together, start the server. All real logic lives elsewhere.
- **`internal/`** — this is not just a convention; it's **enforced by the Go compiler itself**. Any package under a directory literally named `internal/` cannot be imported by any code outside the module tree rooted at `internal/`'s parent. This is Go's built-in mechanism for "private to this project" at the package level — something Python has no real enforced equivalent for (leading underscore prefixes are a convention only, not enforced).
- **`pkg/`** — (more debated/optional in the community than `internal/`) conventionally holds code you *do* intend to be reusable/importable by other projects. Many smaller projects skip this entirely and just use `internal/` for everything, which is a perfectly reasonable choice for an API-only service like the one you'll build in this roadmap.
- **`config/`, `migrations/`** — self-explanatory, and you'll build these out concretely in Guides 12–14.

### 9.3 Package Naming Conventions

- Package names are **short, lowercase, single words** — no underscores, no camelCase (`handler`, not `Handler` or `http_handler`).
- The package name is *not* the same as the import path — you `import "github.com/you/myproject/internal/handler"` but refer to it in code as `handler.SomeFunction()`.
- Every `.go` file in the same directory must declare the **same package name** at the top — a Go "package" is a directory, not a file (contrast with Python where every `.py` file is implicitly its own module).

### 9.4 How This Compares to a FastAPI Project

If you're used to structuring a FastAPI project as `app/routers/`, `app/models/`, `app/services/`, `app/db/` — good news: the Go layout above is conceptually almost identical (`handler` ≈ `routers`, `service` ≈ `services`, `repository`/`model` ≈ `db`/`models`). The biggest structural difference is `cmd/` (Python has no real equivalent since `python -m app.main` doesn't require a dedicated directory) and the *compiler-enforced* privacy of `internal/`.

---

## 10. Exported vs Unexported: Go's Encapsulation Model

### 10.1 Python's Convention-Only Privacy

Python signals "this is private, don't touch it" purely through naming convention:

```python
class Service:
    def __init__(self):
        self._internal_state = {}   # convention: "private", but fully accessible
        self.__very_private = {}    # name-mangled, but still accessible if you try
```

Nothing stops another developer (or you, six months later) from reaching in and touching `_internal_state` directly. It's a social contract, not a compiler-enforced one.

### 10.2 Go's Capitalization-Based Enforcement

Go has no `private`/`public` keywords. Instead, **visibility is determined entirely by the first letter's capitalization** of any identifier — a function, type, variable, constant, or struct field:

```go
package service

// Public API of this package — importable and usable from anywhere
func CreateUser(name string) *User {
    return &User{Name: name, secret: generateSecret()}
}

type User struct {
    Name   string // exported — accessible as user.Name from other packages
    secret string // unexported — invisible outside package "service", full stop
}

// unexported — cannot be called from any package other than "service"
func generateSecret() string {
    return "..."
}
```

If another package does `import ".../service"` and tries `service.generateSecret()` or accesses `someUser.secret`, **this is a compile error**, not a style violation someone might catch in code review. The compiler itself enforces your API boundary.

### 10.3 Why This Is a Bigger Deal Than It Sounds

This single rule does a lot of work:

- It makes every package's **public API immediately visible** just by scanning for capitalized names — no `__all__` list, no separate "public interface" documentation needed.
- It eliminates an entire category of "someone reached into internals and now I can't safely refactor" problems that plague large Python codebases lacking strict discipline.
- Combined with `internal/` (Section 9.2), Go gives you **two enforced layers of encapsulation**: package-level (capitalization) and module-tree-level (`internal/`) — both enforced by the compiler, neither dependent on developer discipline.

You'll feel the effects of this directly starting in Guide 2, since *every single identifier you write* — down to struct field names — is a real, permanent decision about whether external code can see it.

---

## 11. Your First Go Program, Dissected Line by Line

This is the only "real code" in this guide — kept deliberately minimal since syntax itself is Guide 2's job. The goal here is purely to connect everything above to something you can actually run.

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello from Go")
}
```

Line by line, connecting back to what you just learned:

- **`package main`** — every Go file declares which package it belongs to. `main` is special: a package named `main` containing a `main()` function is what makes this an **executable program** rather than an importable library package (recall Section 9.3 — packages are directories, and `main` is the one reserved package name the toolchain treats as a binary entry point).
- **`import "fmt"`** — imports the standard library's formatting/I/O package. Note: `fmt` is lowercase (Section 10) — it's a normal package exposing exported functions like `Println`, `Printf`, `Sprintf`. Unlike Python, **an unused import is a compile error**, not a linter warning — another instance of the compiler enforcing cleanliness rather than trusting convention/tooling.
- **`func main()`** — the entry point. Just like `package main`, the function name `main` inside package `main` is specially recognized by the compiler/runtime as where execution begins — conceptually similar to Python's `if __name__ == "__main__":` block, except it's a hard requirement of the language for a runnable program, not a convention.
- **`fmt.Println(...)`** — calling an **exported** function (`Println`, capital P) from the `fmt` package. If it were lowercase (`println` — note: this actually exists as an unexported *built-in*, a rare special case, but ignore that nuance for now), you could not call it from outside package `fmt`.

Run it:

```bash
go run main.go
# Hello from Go
```

Build it into a standalone binary:

```bash
go build -o hello main.go
./hello
# Hello from Go
```

Notice: no `import antigravity`-style whimsy, no interpreter startup, no `if __name__`. The binary produced by `go build` is a real, independent executable — try moving it to a totally different empty directory and running it; it'll work with zero dependencies, since everything (including a chunk of the Go runtime itself: GC, scheduler) was statically compiled in.

---

## 12. Editor & Tooling Setup

You don't need to fully configure this today, but it's worth knowing what to install before Guide 2:

- **VS Code + the official Go extension** (`golang.go`) is the most common setup and works well if you're already comfortable in VS Code from your TypeScript/React work.
- The extension will prompt you to install `gopls` (the official Go language server — think "the `pyright`/`pylance` of Go"), plus `dlv` (Delve, Go's debugger — your `pdb`/debugger equivalent), and `staticcheck` (a much more thorough linter than `go vet` alone, widely used as the de facto community standard).
- GoLand (JetBrains) is a strong paid alternative if you prefer a dedicated IDE (similar positioning to PyCharm vs VS Code for Python).

Install the essentials now so Guide 2 onward is friction-free:

```bash
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

(Recall Section 6.2 — make sure `$GOPATH/bin` or `$GOBIN` is on your shell `PATH`, or these installed binaries won't be found when invoked by name.)

---

## 13. Summary & What's Next

**What you now have a working mental model for:**

- *Why* Go looks and feels different from Python: it optimizes for large-team, long-lived, high-concurrency systems software — not for rapid scripting or exploratory data work.
- Compiled + statically typed + GC'd, and concretely what each of those three properties changes about your day-to-day workflow versus Python.
- The full toolchain (`run`, `build`, `fmt`, `vet`, `test`, `install`, `doc`) as one cohesive replacement for the patchwork of `pip`/`black`/`pytest`/`mypy` you're used to assembling yourself.
- Go Modules (`go.mod`/`go.sum`) as a cleaner, more reproducible-by-default analog to `pyproject.toml`/lockfiles.
- The standard project layout (`cmd/`, `internal/`, `pkg/`) and how it maps onto the FastAPI-style structure you already use.
- Go's compiler-enforced encapsulation via capitalization — a stricter, non-optional version of Python's "convention only" privacy.

**What comes next — Guide 2: Syntax & Core Types.** This is where you'll finally write real Go: variables, zero values (a Go-specific concept with no Python equivalent — every type has a well-defined default value, so there's no `NameError`/`AttributeError` from an undeclared variable), control flow, functions with multiple return values (this is *the* mechanism Go's error handling is built on — previewed conceptually in Section 2.3, made concrete in Guide 2, and fully explored in Guide 5), and a deep look at slices and maps versus Python's lists and dicts.

Everything in this guide was deliberately conceptual so that when you hit `if err != nil` for the fortieth time in Guide 5, or write your first `interface` in Guide 4, you understand *why* the language insists on it — not just *how* to type it.
