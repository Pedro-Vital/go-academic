# Packages in Go

Packages are one of the core organizational units in Go. They define how code is grouped, reused, imported, tested, documented, and compiled.

A Go program is not written as one large file. It is split into **packages**, and packages live inside a **module**.

---

## 1. What is a package?

A **package** is a collection of Go source files that are compiled together.

In practice:

- A package usually corresponds to one directory.
- All `.go` files in the same directory must declare the same package name.
- A package groups related functions, types, variables, and constants.
- Other packages can import and use its exported identifiers.

Example:

```go
package calculator

func Add(a int, b int) int {
	return a + b
}
````

This file belongs to the `calculator` package.

---

## 2. Package declaration

Every Go source file starts with a package declaration.

```go
package main
```

or:

```go
package calculator
```

The package declaration tells Go which package the file belongs to.

It must appear before imports, functions, types, variables, and constants.

Example:

```go
package greetings

import "fmt"

func SayHello(name string) {
	fmt.Println("Hello,", name)
}
```

---

## 3. Packages and directories

In Go, packages are strongly connected to directories.

Example project:

```text
myapp/
├── go.mod
├── main.go
└── mathutils/
    └── operations.go
```

`main.go`:

```go
package main

import (
	"fmt"

	"myapp/mathutils"
)

func main() {
	result := mathutils.Add(2, 3)
	fmt.Println(result)
}
```

`mathutils/operations.go`:

```go
package mathutils

func Add(a int, b int) int {
	return a + b
}
```

Here:

* `main.go` belongs to package `main`.
* `operations.go` belongs to package `mathutils`.
* The `main` package imports `mathutils`.

---

## 4. Package `main`

The package named `main` is special.

A package named `main` is used to build an executable program.

Example:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

To create a runnable program, you need:

```go
package main
```

and:

```go
func main() {
	// program entry point
}
```

The `main` function is where execution starts.

A package that is not named `main` is usually a reusable library package.

Example:

```go
package calculator

func Add(a int, b int) int {
	return a + b
}
```

This package cannot be executed directly. It must be imported by another package.

---

## 5. Package name vs import path

This is one of the most important distinctions in Go.

The **package name** is written inside the source file:

```go
package mathutils
```

The **import path** is how another package imports it:

```go
import "myapp/mathutils"
```

The import path depends on the module name and directory structure.

Example `go.mod`:

```go
module github.com/pedro/myapp

go 1.22
```

Project:

```text
myapp/
├── go.mod
├── main.go
└── mathutils/
    └── operations.go
```

Import:

```go
import "github.com/pedro/myapp/mathutils"
```

The package name and import path are related, but they are not the same thing.

Usually, the final part of the import path matches the package name:

```go
import "github.com/pedro/myapp/mathutils"
```

```go
package mathutils
```

But Go technically allows them to be different. Even so, it is best to keep them consistent unless you have a strong reason not to.

---

## 6. Multiple files in the same package

A package can be split across multiple files.

Example:

```text
calculator/
├── add.go
├── subtract.go
└── multiply.go
```

`add.go`:

```go
package calculator

func Add(a int, b int) int {
	return a + b
}
```

`subtract.go`:

```go
package calculator

func Subtract(a int, b int) int {
	return a - b
}
```

`multiply.go`:

```go
package calculator

func Multiply(a int, b int) int {
	return a * b
}
```

All files belong to the same package because they all declare:

```go
package calculator
```

Code inside the same package can use identifiers from other files in that package without importing them.

Example:

```go
package calculator

func DoubleSum(a int, b int) int {
	return Add(a, b) * 2
}
```

No import is needed because `Add` belongs to the same package.

---

## 7. Imports

The `import` keyword is used to bring another package into the current file.

Example:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello")
}
```

For multiple imports, use an import block:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	message := strings.ToUpper("hello")
	fmt.Println(message)
}
```

After importing a package, you access its exported identifiers using the package name:

```go
fmt.Println("Hello")
strings.ToUpper("hello")
```

---

## 8. Importing your own packages

Suppose your module is:

```go
module github.com/pedro/go-academic
```

And your project structure is:

```text
go-academic/
├── go.mod
├── main.go
└── greetings/
    └── greetings.go
```

`greetings/greetings.go`:

```go
package greetings

func Hello(name string) string {
	return "Hello, " + name
}
```

`main.go`:

```go
package main

import (
	"fmt"

	"github.com/pedro/go-academic/greetings"
)

func main() {
	message := greetings.Hello("Pedro")
	fmt.Println(message)
}
```

The import path is:

```go
github.com/pedro/go-academic/greetings
```

because:

* `github.com/pedro/go-academic` is the module path.
* `greetings` is the package directory.

---

## 9. Exported and unexported identifiers

Go does not use `public`, `private`, or `protected`.

Instead, visibility is controlled by capitalization.

An identifier is **exported** if it starts with an uppercase letter.

Examples:

```go
func Add(a int, b int) int
type User struct
const MaxRetries = 3
var DefaultTimeout = 30
```

These can be used from other packages.

An identifier is **unexported** if it starts with a lowercase letter.

Examples:

```go
func add(a int, b int) int
type user struct
const maxRetries = 3
var defaultTimeout = 30
```

These can only be used inside the same package.

Example:

```go
package calculator

func Add(a int, b int) int {
	return add(a, b)
}

func add(a int, b int) int {
	return a + b
}
```

Here:

* `Add` is exported.
* `add` is unexported.
* Other packages can call `calculator.Add`.
* Other packages cannot call `calculator.add`.

---

## 10. Exported types and fields

The same capitalization rule applies to struct fields.

Example:

```go
package users

type User struct {
	Name  string
	email string
}
```

Another package can access:

```go
user.Name
```

But cannot access:

```go
user.email
```

because `email` is unexported.

This allows you to expose only the fields that should be part of the public API.

---

## 11. Package-level variables and constants

Variables and constants can be declared at package level.

Example:

```go
package config

const AppName = "Go Academic"

var Debug = true
```

They can be accessed from another package if exported:

```go
fmt.Println(config.AppName)
```

Package-level declarations should be used carefully. Too many global variables make code harder to test and reason about.

Prefer constants for fixed values:

```go
const MaxRetries = 3
```

Avoid mutable global state unless it is really necessary.

---

## 12. Package initialization

When a package is imported, Go initializes it before it is used.

Initialization happens in this general order:

1. Imported packages are initialized first.
2. Package-level variables are initialized.
3. `init` functions are executed.
4. The importing package continues initialization.

Example:

```go
package config

import "fmt"

var AppName = "Go Academic"

func init() {
	fmt.Println("config package initialized")
}
```

The `init` function runs automatically when the package is initialized.

You do not call `init` manually.

Use `init` sparingly. It can make code harder to understand because it runs implicitly.

Good uses for `init` include:

* registering plugins
* setting up package-level defaults
* validation that must happen before the package is used

Avoid using `init` for complex application setup. Prefer explicit initialization functions.

---

## 13. Import aliases

You can rename a package when importing it.

Example:

```go
import f "fmt"

func main() {
	f.Println("Hello")
}
```

This is called an import alias.

A more realistic example:

```go
import (
	cryptoRand "crypto/rand"
	mathRand "math/rand"
)
```

This is useful when two packages have the same name or when the default package name would be unclear.

---

## 14. Blank imports

A blank import uses `_` as the package name.

Example:

```go
import _ "github.com/lib/pq"
```

This imports a package only for its side effects.

The package is initialized, but you do not directly use its exported identifiers.

Blank imports are commonly used when a package registers itself during initialization.

Example use case:

```go
import (
	"database/sql"

	_ "github.com/lib/pq"
)
```

Here, the PostgreSQL driver is imported so it can register itself with `database/sql`.

Use blank imports only when the side effect is intentional.

---

## 15. Dot imports

A dot import brings exported identifiers directly into the current file’s namespace.

Example:

```go
import . "fmt"

func main() {
	Println("Hello")
}
```

This allows `Println` instead of `fmt.Println`.

Dot imports are usually discouraged because they make code less clear. It becomes harder to see where identifiers come from.

Prefer normal imports:

```go
import "fmt"

func main() {
	fmt.Println("Hello")
}
```

---

## 16. Unused imports

Go does not allow unused imports.

This will not compile:

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Hello")
}
```

The `strings` package is imported but not used.

The compiler will produce an error.

This strictness keeps Go code clean.

Use `go fmt` and `goimports` to automatically format and manage imports.

---

## 17. Import cycles are not allowed

Go does not allow circular imports.

Example of an invalid dependency:

```text
package a imports package b
package b imports package a
```

This creates an import cycle.

Go will reject it.

Example:

```go
import cycle not allowed
```

To fix import cycles, you usually need to redesign the package structure.

Common solutions:

* Move shared code to a third package.
* Define interfaces in the package that consumes behavior.
* Reduce unnecessary dependencies between packages.
* Avoid overly granular packages.

Bad structure:

```text
user imports service
service imports user
```

Better structure:

```text
models/
service/
repository/
```

or:

```text
user/
userstore/
```

The correct structure depends on the domain.

---

## 18. Package naming conventions

Go package names should be simple, short, lowercase, and descriptive.

Good package names:

```text
users
config
auth
parser
server
storage
```

Avoid names like:

```text
utils
helpers
common
misc
```

These names are too vague.

Instead of:

```text
utils
```

prefer something more specific:

```text
stringsx
mathx
validation
```

Package names should usually be:

* lowercase
* one word when possible
* not plural unless the domain naturally suggests it
* not too generic
* not redundant with the import path

Example:

```go
import "github.com/pedro/project/config"
```

Then usage:

```go
config.Load()
```

This reads well.

Bad example:

```go
import "github.com/pedro/project/configpackage"
```

Usage:

```go
configpackage.Load()
```

This is verbose and unnatural.

---

## 19. Package API design

A package exposes an API through its exported identifiers.

Example:

```go
package calculator

func Add(a int, b int) int {
	return a + b
}

func Subtract(a int, b int) int {
	return a - b
}
```

The exported API is:

```go
calculator.Add
calculator.Subtract
```

When designing a package, ask:

* What should other packages be allowed to use?
* What should remain internal?
* Is the package name clear?
* Are exported function names readable with the package name?
* Is the package too broad?
* Is the package too small?

Good API:

```go
cache.Get(key)
cache.Set(key, value)
cache.Delete(key)
```

Bad API:

```go
cache.CacheGet(key)
cache.CacheSet(key, value)
cache.CacheDelete(key)
```

The package name already provides context, so repeating it in function names is usually unnecessary.

Prefer:

```go
http.Get()
json.Marshal()
strings.Contains()
```

Not:

```go
http.HTTPGet()
json.JSONMarshal()
strings.StringContains()
```

---

## 20. Standard library packages

Go comes with a rich standard library.

Common packages include:

```go
fmt
strings
strconv
errors
os
io
net/http
encoding/json
time
math
context
testing
```

Example:

```go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	user := User{Name: "Pedro", Age: 25}

	data, err := json.Marshal(user)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(string(data))
}
```

Here:

* `encoding/json` handles JSON encoding.
* `fmt` handles formatted output.

---

## 21. Packages inside a module

A module can contain many packages.

Example:

```text
go-academic/
├── go.mod
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   └── validator/
│       └── validator.go
├── calculator/
│   └── calculator.go
└── greetings/
    └── greetings.go
```

Possible packages:

```text
github.com/pedro/go-academic/cmd/app
github.com/pedro/go-academic/internal/validator
github.com/pedro/go-academic/calculator
github.com/pedro/go-academic/greetings
```

However, only packages that are meant to be imported should expose reusable APIs.

The `cmd/app` package usually contains the executable entry point.

---

## 22. The `cmd` directory convention

The `cmd` directory is commonly used for executable applications.

Example:

```text
myproject/
├── go.mod
├── cmd/
│   └── api/
│       └── main.go
└── internal/
    └── server/
        └── server.go
```

`cmd/api/main.go`:

```go
package main

import "github.com/pedro/myproject/internal/server"

func main() {
	server.Start()
}
```

The `cmd/api` directory contains the `main` package for the API executable.

This is a convention, not a compiler rule.

It is useful when a repository may contain multiple executables.

Example:

```text
cmd/
├── api/
│   └── main.go
├── worker/
│   └── main.go
└── cli/
    └── main.go
```

Each subdirectory is a separate executable package.

---

## 23. The `internal` directory

The `internal` directory has special meaning in Go.

Packages inside an `internal` directory can only be imported by code inside the parent tree.

Example:

```text
myproject/
├── go.mod
├── main.go
└── internal/
    └── auth/
        └── auth.go
```

Package:

```go
package auth
```

Import from inside the same project:

```go
import "github.com/pedro/myproject/internal/auth"
```

This is allowed.

But another module cannot import it:

```go
import "github.com/pedro/myproject/internal/auth"
```

This is rejected by Go.

Use `internal` for implementation details that should not be part of your public API.

---

## 24. The `pkg` directory convention

Some Go projects use a `pkg` directory for packages intended to be reused externally.

Example:

```text
myproject/
├── go.mod
├── cmd/
├── internal/
└── pkg/
    └── client/
        └── client.go
```

Unlike `internal`, `pkg` has no special compiler behavior.

It is only a convention.

Use it only if it makes the project clearer.

For small and medium projects, you often do not need `pkg`.

This is perfectly fine:

```text
myproject/
├── go.mod
├── cmd/
├── internal/
├── client/
└── config/
```

---

## 25. Tests and packages

Tests in Go are written in files ending with `_test.go`.

Example:

```text
calculator/
├── calculator.go
└── calculator_test.go
```

`calculator.go`:

```go
package calculator

func Add(a int, b int) int {
	return a + b
}
```

`calculator_test.go`:

```go
package calculator

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)

	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}
}
```

The test file belongs to the same package:

```go
package calculator
```

This means the test can access both exported and unexported identifiers in the package.

---

## 26. External test packages

A test file can also use a separate package ending in `_test`.

Example:

```go
package calculator_test

import (
	"testing"

	"github.com/pedro/go-academic/calculator"
)

func TestAdd(t *testing.T) {
	result := calculator.Add(2, 3)

	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}
}
```

This style tests the package from the perspective of an external user.

It can only access exported identifiers.

Use this style when you want to test the public API.

Comparison:

```text
package calculator       -> internal test style
package calculator_test  -> external test style
```

Internal test style:

* Can access unexported identifiers.
* Useful for implementation-level tests.

External test style:

* Can only access exported identifiers.
* Useful for API-level tests.
* Encourages better package design.

---

## 27. Documentation for packages

Go documentation is generated from comments.

A package comment should describe what the package does.

Example:

```go
// Package calculator provides basic arithmetic operations.
package calculator
```

Exported identifiers should also have comments.

Example:

```go
// Add returns the sum of a and b.
func Add(a int, b int) int {
	return a + b
}
```

The convention is that the comment starts with the name of the identifier:

```go
// Add returns...
func Add(...)
```

```go
// User represents...
type User struct {}
```

For larger package documentation, you can create a `doc.go` file:

```text
calculator/
├── calculator.go
└── doc.go
```

`doc.go`:

```go
// Package calculator provides simple arithmetic operations.
//
// It includes functions for addition, subtraction, multiplication,
// and division.
package calculator
```

This keeps package-level documentation separate from implementation code.

---

## 28. Building packages

You can build a package with:

```bash
go build
```

If you are inside a directory with a `main` package, Go builds an executable.

If you are inside a library package, Go checks and compiles the package but does not produce a standalone executable.

To build all packages in the module:

```bash
go build ./...
```

The pattern `./...` means all packages recursively from the current directory.

---

## 29. Testing packages

Run tests in the current package:

```bash
go test
```

Run tests in all packages:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run a specific test:

```bash
go test -run TestAdd
```

---

## 30. Listing packages

List the current package:

```bash
go list
```

List all packages in the module:

```bash
go list ./...
```

Example output:

```text
github.com/pedro/go-academic
github.com/pedro/go-academic/calculator
github.com/pedro/go-academic/greetings
```

This is useful to understand how Go sees your project.

---

## 31. Formatting imports

Go code should be formatted with:

```bash
go fmt ./...
```

For automatic import organization, use `goimports`.

Typical import grouping:

```go
import (
	"fmt"
	"net/http"

	"github.com/pedro/myproject/internal/server"
)
```

Common grouping:

1. Standard library imports
2. Third-party imports
3. Local module imports

Example:

```go
import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pedro/myproject/internal/handlers"
)
```

---

## 32. Good package structure

A simple project can start like this:

```text
go-academic/
├── go.mod
├── main.go
├── calculator/
│   ├── calculator.go
│   └── calculator_test.go
└── greetings/
    ├── greetings.go
    └── greetings_test.go
```

A larger application may look like this:

```text
myapp/
├── go.mod
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── server/
│   │   └── server.go
│   ├── handlers/
│   │   └── users.go
│   └── storage/
│       └── postgres.go
└── README.md
```

In this structure:

* `cmd/api` contains the executable entry point.
* `internal/config` handles configuration.
* `internal/server` handles server setup.
* `internal/handlers` handles HTTP handlers.
* `internal/storage` handles persistence.

The goal is not to create many packages early. The goal is to split code when there is a clear responsibility.

---

## 33. When to create a new package

Create a new package when a part of the code has a clear, independent responsibility.

Good reasons:

* The code has a distinct domain concept.
* The code is reused in multiple places.
* The code has a clear API.
* The code can be tested independently.
* The current package is becoming too large or unfocused.

Bad reasons:

* Creating a package for every file.
* Creating generic `utils` packages.
* Splitting code before understanding the domain.
* Trying to copy folder structures from other languages without adapting to Go.

A package should provide cohesion.

Good package:

```text
validation
```

Contains validation logic.

Bad package:

```text
helpers
```

Contains unrelated random functions.

---

## 34. Avoid package names based only on architecture layers

In some languages, it is common to organize code like this:

```text
controllers/
services/
repositories/
models/
```

This can work in Go, but it is not always ideal.

Go often benefits from grouping code by domain or responsibility rather than by generic layers.

For example, instead of:

```text
controllers/
services/
repositories/
```

you might have:

```text
users/
orders/
billing/
```

or:

```text
internal/
├── user/
├── order/
└── billing/
```

Each package can contain the code needed for that domain.

The best structure depends on the project.

For small applications, avoid overengineering. Start simple.

---

## 35. Common package mistakes

### Mistake 1: Using `package main` everywhere

Only executable entry points should use:

```go
package main
```

Reusable code should live in normal packages:

```go
package calculator
package config
package server
```

---

### Mistake 2: Importing local packages with relative paths

Do not do this:

```go
import "./calculator"
```

Use the module path instead:

```go
import "github.com/pedro/go-academic/calculator"
```

---

### Mistake 3: Creating vague packages

Avoid:

```text
utils
helpers
common
shared
```

Prefer specific packages:

```text
validation
config
storage
parser
formatter
```

---

### Mistake 4: Exporting too much

Avoid making everything public.

Bad:

```go
func ValidateEmail(email string) bool
func NormalizeEmail(email string) string
func RemoveWhitespace(value string) string
func InternalParsingStep(value string) string
```

If a function is only used inside the package, keep it unexported:

```go
func internalParsingStep(value string) string
```

Export only what other packages should use.

---

### Mistake 5: Creating import cycles

Avoid packages that depend on each other in both directions.

Bad:

```text
auth imports users
users imports auth
```

Better:

```text
auth imports userstore
users does not import auth
```

or move shared types/interfaces to a better location.

---

### Mistake 6: Making packages too small

Do not create a new package for every tiny concept.

Bad:

```text
add/
subtract/
multiply/
divide/
```

Better:

```text
calculator/
```

A package should represent a cohesive unit of functionality.

---

## 36. Practical example

Project:

```text
go-academic/
├── go.mod
├── cmd/
│   └── app/
│       └── main.go
└── internal/
    └── greetings/
        └── greetings.go
```

`go.mod`:

```go
module github.com/pedro/go-academic

go 1.22
```

`internal/greetings/greetings.go`:

```go
package greetings

func Hello(name string) string {
	return "Hello, " + name
}
```

`cmd/app/main.go`:

```go
package main

import (
	"fmt"

	"github.com/pedro/go-academic/internal/greetings"
)

func main() {
	message := greetings.Hello("Pedro")
	fmt.Println(message)
}
```

Run:

```bash
go run ./cmd/app
```

Expected output:

```text
Hello, Pedro
```

This example shows:

* The executable package is inside `cmd/app`.
* The reusable logic is inside `internal/greetings`.
* The `main` package imports the internal package.
* The exported function `Hello` is accessible from another package.
* The package path follows the module path plus directory path.

---

## 37. Mental model

Think of packages as boundaries.

A package should answer:

```text
What responsibility does this code have?
What does it expose?
What does it hide?
Who should be allowed to import it?
```

A good package has:

* a clear name
* a focused responsibility
* a small public API
* meaningful tests
* minimal dependencies
* no import cycles
* documentation for exported identifiers

---

## 38. Using external packages

External packages are packages created outside your project, usually hosted on GitHub or another repository.

You import them using their module path:

```go
import "github.com/google/uuid"
````

Example:

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

Before using an external package, add it to your module with:

```bash
go get github.com/google/uuid
```

This updates your `go.mod` file:

```go
require github.com/google/uuid v1.6.0
```

It may also update `go.sum`, which stores checksums used to verify dependencies.

After that, you can run the program normally:

```bash
go run .
```

Useful commands:

```bash
go get github.com/google/uuid      # add or update a dependency
go mod tidy                        # remove unused dependencies and add missing ones
go list -m all                     # list all module dependencies
```

Important points:

* External packages are managed at the **module** level, not per file.
* You still import them normally inside Go files.
* If you import a package but do not use it, Go will raise an error.
* `go.mod` records direct dependencies.
* `go.sum` helps verify dependency integrity.
* Run `go mod tidy` often to keep dependencies clean.

Typical workflow:

```bash
go get github.com/google/uuid
```

```go
import "github.com/google/uuid"
```

```bash
go run .
```


---

## 39. Summary

Packages are how Go organizes code.

The essential ideas are:

* Every Go file belongs to a package.
* A package is usually one directory.
* All `.go` files in the same directory must use the same package name.
* `package main` creates an executable program.
* Non-main packages are reusable libraries.
* Packages are imported using import paths.
* Exported identifiers start with uppercase letters.
* Unexported identifiers start with lowercase letters.
* Import cycles are not allowed.
* `internal` restricts imports.
* Tests can be written in the same package or in an external `_test` package.
* Package names should be short, lowercase, and meaningful.
* Good package design keeps code cohesive and APIs small.

The most important habit is to start simple.

Do not create many packages too early. Let the code grow, identify responsibilities, and then split packages when the boundaries become clear.


