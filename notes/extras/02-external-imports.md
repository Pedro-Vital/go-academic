In Go, **external imports are managed through Go Modules**. The key idea is that you don't manually download libraries into your project. You declare your module, import packages by their module path, and the Go toolchain resolves and downloads the required versions. ([Go.dev][1])

### 1. Your project starts with `go.mod`

Suppose you create:

```bash
mkdir myapp
cd myapp

go mod init github.com/pedrovital/myapp
```

You now have:

```text
myapp/
├── go.mod
└── main.go
```

`go.mod`:

```go
module github.com/pedrovital/myapp

go 1.24
```

The `module` name becomes the **import-path prefix for your own packages**. ([Go.dev][2])

---

### 2. Import an external package

Suppose you want Google's UUID library:

```go
package main

import (
    "fmt"

    "github.com/google/uuid"
)

func main() {
    id := uuid.New()
    fmt.Println(id)
}
```

Notice something important:

```go
"github.com/google/uuid"
```

isn't a filesystem path like:

```text
~/go/packages/github.com/google/uuid
```

It's a **package import path**.

Go determines which **module** provides that package.

---

### 3. Add the dependency

You can explicitly run:

```bash
go get github.com/google/uuid
```

Go downloads the module and adds a requirement to `go.mod`. ([Go.dev][1])

You might then have:

```go
module github.com/pedrovital/myapp

go 1.24

require github.com/google/uuid v1.6.0
```

And typically a `go.sum` appears as well:

```text
myapp/
├── go.mod
├── go.sum
└── main.go
```

`go.sum` contains checksums that Go uses to verify downloaded module contents. ([Go.dev][1])

---

## 4. What actually happens when you build?

When Go encounters:

```go
import "github.com/google/uuid"
```

it essentially needs to answer:

> "Which module provides `github.com/google/uuid`, and which version should I use?"

The module graph might look like:

```text
Your application
│
├── github.com/google/uuid v1.6.0
│
├── github.com/some/library v2.1.0
│   └── golang.org/x/... v...
│
└── ...
```

Go downloads the required module versions into its **module cache**, rather than copying dependency source code into your project. ([Go.dev][3])

The default ecosystem uses the Go module proxy (`proxy.golang.org`) when appropriate, although modules can also be obtained directly from repositories. ([Go.dev][4])

---

# The important distinction: package vs module

This is probably the most important concept to understand.

### Package

A **package** is what you import:

```go
import "github.com/google/uuid"
```

### Module

A **module** is the versioned unit that contains one or more packages.

For example:

```text
github.com/google/uuid       ← module
        │
        └── uuid              ← package
```

A larger module might contain:

```text
github.com/example/project   ← module
│
├── client                   ← package
├── server                   ← package
├── auth                     ← package
└── internal/...             ← packages
```

Your `go.mod` tracks **modules**, while your source code imports **packages**. ([Go.dev][5])

---

# What about your own packages?

Suppose your project is:

```text
myapp/
├── go.mod
├── main.go
└── mathutil/
    └── mathutil.go
```

with:

```go
module github.com/pedrovital/myapp
```

Then you can import your own package:

```go
import "github.com/pedrovital/myapp/mathutil"
```

So the import path is:

```text
[module path] + [relative package directory]
```

```text
github.com/pedrovital/myapp
              │
              └── /mathutil
```

→

```text
github.com/pedrovital/myapp/mathutil
```

This is why the module name matters. ([Go.dev][5])

---

# What does `go mod tidy` do?

This is a command you'll use constantly:

```bash
go mod tidy
```

It examines the packages your code actually imports and makes `go.mod` consistent:

* adds missing dependencies
* removes dependencies no longer needed
* updates `go.sum` accordingly

([Go.dev][1])

So a common workflow is simply:

```bash
go mod init github.com/pedrovital/myapp

# write code
# add imports

go mod tidy

go run .
```

---

# What about versions?

Go dependencies are versioned, usually with semantic versions:

```text
github.com/google/uuid v1.6.0
                           │
                           └── version
```

You can explicitly request one:

```bash
go get github.com/google/uuid@v1.6.0
```

or update to the latest available version:

```bash
go get github.com/google/uuid@latest
```

([Go.dev][1])

This is fundamentally different from Python's typical:

```bash
pip install package
```

where you might have a virtual environment containing installed packages.

Go's model is more like:

```text
                 go.mod
                   │
                   ▼
          dependency graph
                   │
          ┌────────┴────────┐
          ▼                 ▼
      module A           module B
       v1.2.0             v2.1.0
          │                 │
          └───────┬─────────┘
                  ▼
             Go module cache
```

Your repository generally contains **your code + `go.mod` + `go.sum`**, not a copy of every external dependency.

---

## One more thing: Go has no `requirements.txt` equivalent

The closest conceptual comparison is:

| Python                            | Go                                |
| --------------------------------- | --------------------------------- |
| `requirements.txt`                | `go.mod`                          |
| `pip install`                     | `go get`                          |
| virtual environment               | Go module/cache model             |
| PyPI                              | Go module ecosystem / proxies     |
| package                           | package                           |
| distribution/project              | module                            |
| `pip freeze`                      | not really a direct equivalent    |
| `pip install -r requirements.txt` | `go mod download` / build tooling |

But **don't think of Go modules as simply "Python virtualenv for Go."** The dependency/version model is substantially different.

### The mental model I'd recommend

Since you're learning Go for backend development, remember just this:

> **`import` says what package I want; `go.mod` says what module/version provides it; the Go toolchain resolves, downloads, verifies, and caches it.**

And your normal workflow is:

```bash
go mod init <module-path>   # once
go get <dependency>         # when deliberately adding/updating one
go mod tidy                 # synchronize dependencies
go build / go test / go run # use them
```

Go's official documentation describes this module-based workflow as the standard way to manage external dependencies. ([Go.dev][1])

[1]: https://go.dev/doc/modules/managing-dependencies?utm_source=chatgpt.com "Managing dependencies - The Go Programming Language"
[2]: https://go.dev/doc/modules/gomod-ref?utm_source=chatgpt.com "go.mod file reference - The Go Programming Language"
[3]: https://go.dev/ref/mod?utm_source=chatgpt.com "Go Modules Reference - The Go Programming Language"
[4]: https://go.dev/doc/modules/managing-source?utm_source=chatgpt.com "Managing module source - The Go Programming Language"
[5]: https://go.dev/wiki/Modules?utm_source=chatgpt.com "Go Wiki: Go Modules - The Go Programming Language"
