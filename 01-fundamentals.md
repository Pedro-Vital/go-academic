# Go Fundamentals

> A comprehensive self-study guide to core Go concepts for ML/AI engineers.

Exercises to follow along with: 
- 01-hello-world
- 02-investment-calculator
- 03-bank

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Modules & Packages](#2-modules--packages)
3. [Values and Types](#3-values-and-types)
4. [Variables, Constants, and Type Inference](#4-variables-constants-and-type-inference)
5. [Basic Types](#5-basic-types)
6. [Input and Output](#6-input-and-output)
7. [Functions](#7-functions)
8. [Control Flow](#8-control-flow)
9. [Range](#9-range)
10. [Error Handling](#10-error-handling)
11. [Nil in Go](#11-nil-in-go)
12. [Working with Files](#12-working-with-files)
13. [Type Conversions](#13-type-conversions)
14. [Scope and Visibility](#14-scope-and-visibility)
15. [Common Go Idioms](#15-common-go-idioms)
16. [Best Practices](#16-best-practices)
17. [What to Learn Next](#17-what-to-learn-next)

---

## 1. Introduction

Go (often called Golang) is an open-source, statically typed, compiled programming language designed at Google in 2007 by Robert Griesemer, Rob Pike, and Ken Thompson — the same lineage as Unix and C. Its core goals: **fast compilation, readable syntax, built-in concurrency, and efficient execution with a small runtime**.

### Why Go for ML/AI Engineers?

Python is the dominant language for training and experimentation. Go excels at everything that surrounds Python ML systems in production:

| Layer | Go's role |
|---|---|
| Feature serving | Low-latency gRPC / REST API servers |
| Data pipelines | High-throughput workers, Kafka consumers |
| Orchestration | Kubernetes operators, custom controllers |
| Infrastructure | Sidecar services, proxies, health checks |
| Tooling | CLI tools, build systems, deployment scripts |

Kubernetes, Docker, Prometheus, Terraform, and Argo Workflows are all written in Go. Understanding Go lets you read, extend, and contribute to the tools that run modern ML infrastructure.

### Go's Design Philosophy

Go was designed to be deliberately minimal. Fewer features means more readable code and fewer ways to go wrong:

- **Simplicity over expressiveness** — there is usually one obvious way to do something
- **Explicit over implicit** — no hidden behavior, no magic
- **Compilation enforces discipline** — unused imports and unused variables are compile errors
- **Composition over inheritance** — no class hierarchies; behavior is composed through interfaces
- **Errors are values** — no exceptions; errors flow through return values and must be handled explicitly

These aren't limitations. They are the reason Go codebases stay clean and comprehensible at scale.

---

## 2. Modules & Packages

### Packages: The Unit of Code Organization

A **package** is the fundamental unit of code organization in Go. Every `.go` file starts with a package declaration:

```go
package main
```

All `.go` files in the same directory must share the same package name. A package is essentially a namespace — a collection of related source files compiled together as a unit.

### The `main` Package and Library Packages

Go distinguishes between two kinds of packages:

**The `main` package** defines an executable program. It must contain a `main()` function, which is the entry point. When you `go build` a `main` package, Go produces a runnable binary.

```go
package main

import "fmt"

func main() {
    fmt.Println("This is an executable.")
}
```

**Library packages** (any package not named `main`) define reusable code that other packages can import. They cannot be executed directly.

```go
// Package fileops provides helpers for reading and writing files.
package fileops

func ReadFloat(path string) (float64, error) {
    // ...
}
```

> **Key rule:** Only one package in a project should be named `main`. All others are libraries.

### The `main` Function

`main()` is the entry point of every Go executable. It takes no parameters and returns nothing:

```go
package main

import "fmt"

func main() {
    fmt.Println("Program starts here.")
    // Program exits when main() returns.
}
```

The program exits automatically when `main()` returns. Multiple `.go` files can belong to `package main` — they are compiled as a single unit.

### Modules and `go.mod`

A **module** is a collection of related packages versioned together. It is the unit of dependency management in Go (introduced in Go 1.11 and now universal).

A module is defined by a `go.mod` file at the root of the project:

```
module example.com/myproject

go 1.22

require (
    github.com/some/dependency v1.4.2
)
```

| Field | Description |
|---|---|
| `module` | The module path — a unique identifier, usually a URL. Used as the base for all internal import paths. |
| `go` | The minimum Go version required. |
| `require` | External dependencies and their pinned versions. |

When external dependencies are added, Go also creates a `go.sum` file containing cryptographic checksums to verify the integrity of every dependency. This file is auto-generated and should be committed to version control.

### Go Project Structure

A typical Go project:

```
myproject/
├── go.mod              # Module definition
├── go.sum              # Dependency checksums (auto-generated)
├── main.go             # Entry point (package main)
├── handlers.go         # Another file in package main
└── fileops/
    └── fileops.go      # Package fileops (a library package)
```

**Important rules:**
- All `.go` files in the same directory must belong to the same package.
- Subdirectories define separate packages.
- By convention, the directory name matches the package name (not enforced by the compiler, but strongly expected).

### Importing Packages

```go
import "fmt"                              // Standard library
import "os"                               // Standard library
import "example.com/myproject/fileops"    // Local package within the same module
import "github.com/some/external/lib"     // External dependency
```

For multiple imports, use a grouped import block (the standard style):

```go
import (
    "fmt"
    "os"
    "strconv"

    "example.com/myproject/fileops"
    "github.com/some/external/lib"
)
```

> **Go enforces that every imported package is used.** An unused import is a compile error. This keeps codebases free of dead dependencies.

To reference exported names from an imported package:

```go
fmt.Println("Hello")
os.Exit(1)
fileops.ReadFloat("data.txt")
```

### Essential Commands

```bash
# Initialize a new module in the current directory
go mod init example.com/myproject

# Run the program directly (compiles in memory, no output file)
go run .

# Run a specific file
go run main.go

# Compile into a binary
go build .
go build -o myapp .    # custom binary name

# Execute the compiled binary
./myapp                # Linux / macOS
.\myapp.exe            # Windows

# Download missing dependencies and remove unused ones
go mod tidy

# Add a specific external dependency
go get github.com/some/package@v1.2.3

# Format all Go source files (writes in place)
go fmt ./...

# Run static analysis
go vet ./...
```

> **`go run .` vs `go build`:** `go run .` compiles and executes in one step — ideal during development. `go build` produces a persistent, deployable binary. In production, always ship a compiled binary.

---

## 3. Values and Types

Go is a **statically typed language**. Every value has a type that is determined at compile time. The compiler uses types to catch mistakes before the program runs — a class of bugs that simply cannot happen at runtime.

```go
var age int = 30
var name string = "Alice"

// age = "thirty"  // COMPILE ERROR: cannot use string as int
```

### Zero Values

Every type in Go has a **zero value** — the default a variable holds when declared without an explicit initializer. There is no concept of an "uninitialized" variable.

| Type | Zero Value |
|---|---|
| `int`, `int32`, `int64`, `uint`, etc. | `0` |
| `float32`, `float64` | `0.0` |
| `string` | `""` (empty string) |
| `bool` | `false` |
| Pointers, functions, interfaces | `nil` |

```go
var count int       // 0
var label string    // ""
var active bool     // false
```

This is a meaningful guarantee: a declared variable always holds a valid, usable value.

### Type Safety

Go does not perform implicit type conversions. Mixing types requires an explicit conversion. This eliminates entire categories of subtle bugs common in dynamically typed languages:

```go
var x int = 10
var y float64 = 3.14
// var z = x + y      // COMPILE ERROR: cannot add int and float64
var z = float64(x) + y  // OK: explicit conversion
```

---

## 4. Variables, Constants, and Type Inference

### Declaring Variables with `var`

```go
var name string = "Alice"
var age int = 30
var temperature float64 = 36.6
var active bool = true
```

Declared without an initializer, the variable holds its zero value:

```go
var counter int     // 0
var message string  // ""
```

Multiple variables in a block:

```go
var (
    host  string = "localhost"
    port  int    = 8080
    debug bool   = false
)
```

### Short Variable Declaration with `:=`

The `:=` operator is the most common way to declare variables inside functions. Go **infers the type** from the assigned value:

```go
func main() {
    name   := "Alice"     // inferred: string
    age    := 30          // inferred: int
    score  := 98.5        // inferred: float64
    active := true        // inferred: bool
}
```

Key constraints:
- `:=` can only be used **inside functions** — not at package level.
- At least one variable on the left side must be new.

```go
x := 10
x, y := 20, 30  // OK: y is new; x is reassigned
// x := 40      // COMPILE ERROR: no new variables
x = 40          // OK: plain assignment (not a declaration)
```

### Constants with `const`

Constants are compile-time values that can never change at runtime:

```go
const Pi         = 3.14159
const AppVersion = "2.1.0"
const MaxRetries = 3
```

Multiple constants in a block:

```go
const (
    StatusOK      = 200
    StatusError   = 500
    DefaultPort   = 8080
)
```

Constants cannot be declared with `:=`, and their values must be determinable at compile time.

> **When to use `const` vs `var`:** Use `const` for values with fixed, semantic meaning — API versions, file paths, retry limits, mathematical constants. It communicates intent and prevents accidental mutation.

### The Blank Identifier `_`

Go requires every declared variable to be used. When you need to discard a value, use `_` — the blank identifier:

```go
value, _ := someFunction()        // discard the second return value
_, err   := os.ReadFile("f.txt")  // discard the file content, keep the error
```

`_` is not a variable. It is a write-only discard slot — you can never read from it.

---

## 5. Basic Types

### Numbers

#### Integer Types

| Type | Size | Typical Use |
|---|---|---|
| `int` | 32 or 64 bit (platform) | Default integer type |
| `int8` | 8 bit | -128 to 127 |
| `int16` | 16 bit | -32,768 to 32,767 |
| `int32` | 32 bit | ~±2.1 billion |
| `int64` | 64 bit | ~±9.2 × 10¹⁸ |
| `uint` | Platform | Non-negative, same width as `int` |
| `uint8` / `byte` | 8 bit | 0–255; often used for raw bytes |

**Default integer type:** Writing `x := 42` infers `int`. Use `int` unless you have a specific reason to use a sized variant (e.g., interfacing with binary protocols or external APIs that specify a width).

```go
var count int = 1_000_000   // underscores for readability (Go 1.13+)
remainder := 17 % 5         // modulo operator → 2
quotient  := 17 / 5         // integer division → 3 (truncates)
```

#### Floating-Point Types

| Type | Size | Precision |
|---|---|---|
| `float32` | 32 bit | ~7 significant digits |
| `float64` | 64 bit | ~15 significant digits |

**Default float type:** Writing `x := 3.14` infers `float64`. Prefer `float64` in almost all cases — it is more precise and is the type expected by the standard `math` package.

```go
var price    float64 = 9.99
var ratio    float64 = 0.75
exact := 17.0 / 5.0         // → 3.4 (float division)
```

**Useful `math` package functions:**

```go
import "math"

math.Abs(-4.5)        // → 4.5
math.Sqrt(16.0)       // → 4.0
math.Pow(2, 10)       // → 1024.0
math.Round(3.7)       // → 4.0
math.Floor(3.9)       // → 3.0
math.Ceil(3.1)        // → 4.0
math.Max(3.0, 7.0)    // → 7.0
math.Log2(1024.0)     // → 10.0
```

### Strings

In Go, strings are **immutable sequences of bytes**, encoded in UTF-8 by default. They use double quotes:

```go
greeting  := "Hello, World!"
escaped   := "Line one\nLine two\tTabbed"
```

**Raw string literals** (backticks) disable all escape processing:

```go
path    := `C:\Users\pedro\projects`
template := `{"model": "gpt-4", "temperature": 0.7}`
```

Raw strings are particularly useful for JSON templates, file paths, and regular expressions.

**Common `strings` package operations:**

```go
import "strings"

s := "Go is great for infrastructure"

len(s)                                    // byte length: 30
strings.ToUpper(s)                        // "GO IS GREAT FOR INFRASTRUCTURE"
strings.ToLower(s)                        // "go is great for infrastructure"
strings.Contains(s, "great")             // true
strings.HasPrefix(s, "Go")               // true
strings.HasSuffix(s, "infrastructure")   // true
strings.Count(s, "r")                    // 3
strings.Replace(s, "great", "ideal", 1) // "Go is ideal for infrastructure"
strings.TrimSpace("  hello  ")           // "hello"
strings.Split("a,b,c", ",")             // ["a", "b", "c"]
strings.Index(s, "great")               // 6 (byte position)
```

**String concatenation:**

```go
// Simple concatenation (fine for a few values)
full := "Hello, " + "World!"

// For building large strings efficiently, use strings.Builder
var b strings.Builder
b.WriteString("model=")
b.WriteString("llama3")
b.WriteString(", temp=0.7")
result := b.String() // "model=llama3, temp=0.7"
```

> **Bytes vs. characters:** Go strings are byte sequences. A single Unicode character (rune) can occupy 1–4 bytes. Use `[]rune(s)` when you need to count or index by character rather than byte. See Section 13 for details.

### Booleans

```go
var active bool = true
var disabled = false

// Comparison operators
10 > 5         // true
10 == 10       // true
"go" != "py"   // true

// Logical operators
true && false  // false (AND)
true || false  // true  (OR)
!true          // false (NOT)
```

---

## 6. Input and Output

The `fmt` package (short for "format") is the primary package for formatted I/O in Go.

### Printing to the Console

```go
fmt.Print("Hello")               // no newline
fmt.Println("Hello, World!")     // with newline
fmt.Println("Value:", 42)        // auto-space between arguments → "Value: 42"
fmt.Printf("Name: %s, Age: %d\n", "Alice", 30)  // formatted output
```

### Reading User Input

`fmt.Scan` reads whitespace-separated values from standard input:

```go
var name string
fmt.Print("Enter your name: ")
fmt.Scan(&name)               // reads one word; stops at whitespace
fmt.Println("Hello,", name)

var a, b int
fmt.Scan(&a, &b)              // reads two integers separated by space or newline
fmt.Println("Sum:", a+b)
```

> The `&` operator passes a pointer to the variable, so `fmt.Scan` can write the scanned value into it. Pointers will be covered thoroughly in a future guide — for now, just know that `&` is required here.

For reading full lines, use `bufio.Scanner`:

```go
import (
    "bufio"
    "fmt"
    "os"
)

scanner := bufio.NewScanner(os.Stdin)
fmt.Print("Enter a sentence: ")
scanner.Scan()
line := scanner.Text()
fmt.Println("You entered:", line)
```

### Format Specifiers (Verbs)

`fmt.Printf`, `fmt.Sprintf`, and `fmt.Fprintf` use **verbs** — placeholders that describe how to format a value.

| Verb | Description | Example output |
|---|---|---|
| `%v` | Default format | `42`, `true`, `3.14` |
| `%T` | Go type of the value | `int`, `string`, `float64` |
| `%d` | Integer (decimal) | `42` |
| `%b` | Integer (binary) | `101010` |
| `%x` | Integer (hexadecimal, lowercase) | `2a` |
| `%o` | Integer (octal) | `52` |
| `%f` | Float (default precision) | `3.140000` |
| `%.2f` | Float with 2 decimal places | `3.14` |
| `%e` | Float (scientific notation) | `3.14e+00` |
| `%s` | String | `hello` |
| `%q` | Quoted string | `"hello"` |
| `%t` | Boolean | `true` |
| `%c` | Character (rune as Unicode) | `A` |
| `%p` | Pointer address | `0xc000014080` |

```go
fmt.Printf("Int:    %d\n", 255)
fmt.Printf("Binary: %b\n", 255)         // → Binary: 11111111
fmt.Printf("Hex:    %x\n", 255)         // → Hex:    ff
fmt.Printf("Float:  %.4f\n", 3.14159)   // → Float:  3.1416
fmt.Printf("Type:   %T\n", 3.14)        // → Type:   float64
fmt.Printf("Bool:   %t\n", true)        // → Bool:   true
fmt.Printf("Char:   %c\n", 65)          // → Char:   A

// Width and alignment
fmt.Printf("[%10d]\n", 42)    // → [        42]  (right-aligned, width 10)
fmt.Printf("[%-10d]\n", 42)   // → [42        ]  (left-aligned)
fmt.Printf("[%010d]\n", 42)   // → [0000000042]  (zero-padded)
```

### String Formatting with `fmt.Sprintf`

`fmt.Sprintf` works like `fmt.Printf` but **returns the formatted string** instead of printing it. This is useful for building strings that mix types:

```go
name     := "alice"
epochs   := 50
accuracy := 0.9234

summary := fmt.Sprintf("Run: %s | Epochs: %d | Accuracy: %.2f%%", name, epochs, accuracy*100)
fmt.Println(summary)
// → Run: alice | Epochs: 50 | Accuracy: 92.34%

// Useful for log messages, file names, API responses
logLine  := fmt.Sprintf("[ERROR] %s: %s", "db connection", "timeout after 30s")
ckptPath := fmt.Sprintf("checkpoints/epoch_%03d.pt", 7)
```

> **`fmt.Sprintf` vs concatenation:** Prefer `fmt.Sprintf` when mixing types. It handles conversions, produces cleaner code, and avoids the multi-step string-building that concatenation requires.

---

## 7. Functions

Functions are the primary building block for organizing logic in Go. Unlike Python, Go functions require explicit type annotations for both parameters and return values.

### Defining Functions

```go
func functionName(param1 type1, param2 type2) returnType {
    // body
    return value
}
```

Examples:

```go
func greet(name string) string {
    return "Hello, " + name + "!"
}

func clamp(value, min, max float64) float64 {
    if value < min {
        return min
    }
    if value > max {
        return max
    }
    return value
}
```

If multiple consecutive parameters share the same type, you can consolidate:

```go
func add(a, b float64) float64 {  // same as (a float64, b float64)
    return a + b
}
```

Functions with no return value (equivalent to `void`):

```go
func logEvent(level, message string) {
    fmt.Printf("[%s] %s\n", level, message)
}
```

### Multiple Return Values

One of Go's most important features is the ability to return multiple values. This is how functions signal both a result and an error simultaneously:

```go
import "errors"

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 3)
if err != nil {
    fmt.Println("Error:", err)
} else {
    fmt.Printf("Result: %.4f\n", result) // → Result: 3.3333
}
```

Multiple return values are not limited to (result, error) pairs:

```go
func minMax(a, b float64) (float64, float64) {
    if a < b {
        return a, b
    }
    return b, a
}

lo, hi := minMax(7.5, 3.2)
fmt.Println(lo, hi) // → 3.2 7.5
```

### Named Return Values

Return values can be named in the signature. They are initialized to their zero values and can be returned with a bare `return`:

```go
func parseCoords(input string) (lat float64, lon float64, err error) {
    // lat, lon, and err are pre-declared and initialized to 0, 0, nil
    // ... parsing logic ...
    if input == "" {
        err = errors.New("empty input")
        return // returns lat=0, lon=0, err=<error>
    }
    lat, lon = 48.85, 2.35
    return // returns lat, lon, nil
}
```

Named return values serve as documentation — they make it clear what each return position means. However, bare `return` in complex functions can hurt readability. Use named returns judiciously.

### Functions as First-Class Values

In Go, functions are first-class values: they can be assigned to variables and passed as arguments.

```go
// Assign a function to a variable
double := func(x float64) float64 {
    return x * 2
}
fmt.Println(double(5.0)) // → 10

// Pass a function as a parameter
func applyTransform(value float64, fn func(float64) float64) float64 {
    return fn(value)
}

result := applyTransform(3.0, double)  // → 6.0
```

Function types describe the signature: `func(float64) float64` means "a function that takes a float64 and returns a float64." This concept is central to callback patterns and functional-style pipelines in Go.

---

## 8. Control Flow

### If-Else

Go's `if` requires no parentheses around the condition, but **always requires braces** — even for single-line bodies:

```go
if condition {
    // ...
} else if anotherCondition {
    // ...
} else {
    // ...
}
```

Example:

```go
confidence := 0.85

if confidence >= 0.90 {
    fmt.Println("High confidence — publish result")
} else if confidence >= 0.70 {
    fmt.Println("Moderate confidence — flag for review")
} else {
    fmt.Println("Low confidence — discard")
}
```

**If with an initialization statement** is a common Go pattern. The variable declared in the init is scoped to the entire `if-else` block:

```go
if val, err := strconv.ParseFloat("3.14", 64); err != nil {
    fmt.Println("Parse error:", err)
} else {
    fmt.Println("Parsed:", val)
}
// val and err are out of scope here
```

This keeps the error-handling variable tightly scoped to where it is relevant — you will see this pattern everywhere in Go codebases.

### For Loops

Go has exactly **one looping construct**: `for`. It covers all loop patterns.

**C-style (init; condition; post):**

```go
for i := 0; i < 5; i++ {
    fmt.Println(i)
}
```

**While-style (condition only):**

```go
attempts := 0
for attempts < 3 {
    err := callAPI()
    if err == nil {
        break
    }
    attempts++
}
```

**Infinite loop:**

```go
for {
    choice := promptUser()
    if choice == "exit" {
        break
    }
    handleChoice(choice)
}
```

An infinite `for {}` is idiomatic in Go for programs that run until an explicit exit condition — interactive CLIs, event loops, and server request loops all follow this pattern.

### `continue` and `break`

**`continue`** skips the rest of the current iteration and jumps to the next:

```go
for i := 0; i < 10; i++ {
    if i%2 == 0 {
        continue        // skip even numbers
    }
    fmt.Println(i)      // prints: 1, 3, 5, 7, 9
}
```

**`break`** exits the loop entirely:

```go
for i := 0; i < 100; i++ {
    if i == 5 {
        break           // stop when i reaches 5
    }
    fmt.Println(i)      // prints: 0, 1, 2, 3, 4
}
```

**Labels** allow `break` and `continue` to target an outer loop:

```go
outer:
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            if i == 1 && j == 1 {
                break outer         // exits the outer loop
            }
            fmt.Printf("(%d, %d)\n", i, j)
        }
    }
```

### Switch

Go's `switch` is cleaner than most languages':
- **No automatic fall-through** — no need for `break` in each case.
- Cases can match **multiple values**.
- The condition is **optional** (switch becomes an if-else chain).

**Basic switch:**

```go
status := "error"

switch status {
case "ok":
    fmt.Println("All systems nominal")
case "warn", "warning":
    fmt.Println("Degraded performance")
case "error", "fatal":
    fmt.Println("Service unavailable")
default:
    fmt.Println("Unknown status:", status)
}
```

**Switch without a condition** (equivalent to an if-else chain, but more readable for multi-branch logic):

```go
score := 78.0

switch {
case score >= 90:
    fmt.Println("A")
case score >= 80:
    fmt.Println("B")
case score >= 70:
    fmt.Println("C")
default:
    fmt.Println("F")
}
```

**`fallthrough`** explicitly continues into the next case (opt-in, unlike C):

```go
n := 1
switch n {
case 1:
    fmt.Println("One")
    fallthrough             // continues into case 2
case 2:
    fmt.Println("Two")     // also prints when n == 1
default:
    fmt.Println("Other")
}
```

> **When to use `switch` vs `if-else`:** Prefer `switch` when branching on a single variable across three or more values. It signals "I'm choosing among discrete options" more clearly than a chain of `else if`.

---

## 9. Range

The `range` keyword produces an iterator over a sequence. It always returns up to two values per iteration: an **index** (or key) and a **value**.

### Range over a String

Iterating over a string with `range` gives you each Unicode code point (rune) with its byte index — not the raw bytes:

```go
word := "hello"

for index, char := range word {
    fmt.Printf("index %d → %c (rune: %d)\n", index, char, char)
}
// index 0 → h (rune: 104)
// index 1 → e (rune: 101)
// index 2 → l (rune: 108)
// index 3 → l (rune: 108)
// index 4 → o (rune: 111)
```

Discard the value when you only need the index:

```go
for i := range "hello" {
    fmt.Println(i)  // 0, 1, 2, 3, 4
}
```

Discard the index when you only need the character:

```go
for _, char := range "hello" {
    fmt.Printf("%c ", char)   // h e l l o
}
```

Range over a string correctly handles multi-byte Unicode characters — it advances by rune, not by byte:

```go
for i, r := range "héllo" {
    fmt.Printf("byte %d → %c\n", i, r)
}
// byte 0 → h
// byte 1 → é      ← occupies bytes 1 and 2
// byte 3 → l
// byte 4 → l
// byte 5 → o
```

### Range over an Integer (Go 1.22+)

Starting in Go 1.22, you can range over an integer directly, iterating from `0` to `n-1`:

```go
for i := range 5 {
    fmt.Println(i)   // 0, 1, 2, 3, 4
}
```

This is a clean, readable replacement for `for i := 0; i < n; i++` in many common cases.

> Range is especially powerful with slices, maps, and channels — collection types we will explore in upcoming guides.

---

## 10. Error Handling

Go takes a fundamentally different approach to error handling than Python, Java, or C++. There are **no exceptions**. Functions that can fail return an `error` value alongside their result, and the caller handles it explicitly.

This is not a compromise — it is a deliberate design choice that makes error paths **visible, local, and impossible to accidentally ignore**.

### The `error` Type

`error` is a built-in interface:

```go
type error interface {
    Error() string
}
```

Any type with an `Error() string` method satisfies it. You will mostly work with errors at this interface level, treating them as opaque values to be checked and propagated.

### Checking Errors

The canonical Go error-handling pattern:

```go
result, err := someFunction()
if err != nil {
    // handle it — log, return, or wrap and propagate
    fmt.Println("Something went wrong:", err)
    return
}
// Safe to use result here
```

This pattern appears in virtually every Go program. Its intentional verbosity is the point: every error path is explicit and accounted for.

```go
import (
    "fmt"
    "os"
)

data, err := os.ReadFile("model_config.json")
if err != nil {
    fmt.Println("Failed to read config:", err)
    return
}

fmt.Printf("Config loaded: %d bytes\n", len(data))
```

### Creating Errors

**`errors.New`** creates a simple error from a string:

```go
import "errors"

func validateThreshold(t float64) error {
    if t < 0 || t > 1 {
        return errors.New("threshold must be between 0 and 1")
    }
    return nil  // nil = no error, success
}

err := validateThreshold(1.5)
if err != nil {
    fmt.Println(err)  // → threshold must be between 0 and 1
}
```

**`fmt.Errorf`** creates errors with formatting and context:

```go
func loadWeights(path string) error {
    _, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("loadWeights: cannot open %q: %w", path, err)
    }
    return nil
}
```

The `%w` verb **wraps** the original error, preserving it so callers can inspect it. Prefer `fmt.Errorf` with `%w` over `errors.New` whenever you have context to add — it produces error messages that are much easier to debug in production.

### Error Wrapping and Inspection

Go 1.13 introduced `errors.Is` and `errors.As` for examining wrapped error chains:

```go
import "errors"

var ErrModelNotFound = errors.New("model not found")

func loadModel(name string) (string, error) {
    if name == "" {
        return "", fmt.Errorf("loadModel: %w", ErrModelNotFound)
    }
    return "model_data", nil
}

_, err := loadModel("")
if errors.Is(err, ErrModelNotFound) {
    fmt.Println("Model does not exist — falling back to default")
}
```

**Sentinel errors** like `ErrModelNotFound` above are package-level error variables that represent well-defined, expected failure conditions. By convention, they are named `ErrSomething` and defined at the package level for callers to compare against.

### Returning `nil` on Success

Every function that can fail should explicitly return `nil` as the error on the success path:

```go
func parseConfidence(s string) (float64, error) {
    val, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0, fmt.Errorf("parseConfidence: invalid input %q: %w", s, err)
    }
    if val < 0 || val > 1 {
        return 0, fmt.Errorf("parseConfidence: value %f out of range [0, 1]", val)
    }
    return val, nil   // ← nil signals success
}
```

`nil` as an error is the Go equivalent of a successful, exception-free return.

---

## 11. Nil in Go

`nil` is the zero value for **pointer, function, interface, slice, map, and channel** types. It represents the absence of a value.

### `nil` in Error Handling

The most frequent use of `nil` in Go:

```go
func compute() (int, error) {
    return 42, nil   // nil error = no error = success
}

result, err := compute()
if err != nil {
    fmt.Println("Error:", err)
    return
}
fmt.Println("Result:", result)
```

The pattern `if err != nil` is ubiquitous in Go. Checking for `nil` is how you distinguish success (`nil`) from failure (non-nil error).

### `nil` for Optional or Unset Values

`nil` also serves as a sentinel for other types:

```go
var err error = nil     // interface with nil value
var p *int  = nil       // pointer to nothing
```

> Calling a method on a `nil` pointer causes a **runtime panic**. Always check for `nil` before using a value that could be nil.

### What Cannot Be `nil`

Basic types — `int`, `float64`, `string`, `bool` — can never be `nil`. They always hold a zero value:

```go
var n int    // 0, not nil
var s string // "", not nil
var b bool   // false, not nil
```

This is why error handling works the way it does: the `error` type is an interface, which can hold `nil`, so returning `nil` from a function that returns `error` is valid and meaningful.

---

## 12. Working with Files

File I/O in Go is handled through the `os`, `bufio`, and `strconv` packages. The standard library provides clean, explicit primitives.

### Reading an Entire File

`os.ReadFile` reads a complete file into memory as a byte slice. Best for small files like configs and data files:

```go
import (
    "fmt"
    "os"
)

data, err := os.ReadFile("config.txt")
if err != nil {
    fmt.Println("Error reading file:", err)
    return
}

content := string(data)    // []byte → string
fmt.Println(content)
```

> `[]byte` is a slice of bytes — Go's raw binary buffer type. Converting it to `string` interprets the bytes as UTF-8 text. Slices will be covered in depth in a future guide.

### Reading a File Line by Line

For large files, `bufio.Scanner` reads one line at a time without loading the entire file into memory:

```go
import (
    "bufio"
    "fmt"
    "os"
)

file, err := os.Open("data.txt")
if err != nil {
    fmt.Println("Error opening file:", err)
    return
}

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Println(line)
}

if err := scanner.Err(); err != nil {
    fmt.Println("Scanner error:", err)
}

file.Close()  // always close the file when done
              // Note: idiomatic Go uses defer for this — covered in a future guide
```

### Writing an Entire File

`os.WriteFile` writes a byte slice to a file, creating it if it doesn't exist and overwriting it if it does:

```go
import "os"

content := []byte("model=llama3\ntemperature=0.7\nmax_tokens=512")
err := os.WriteFile("config.txt", content, 0644)
if err != nil {
    fmt.Println("Error writing file:", err)
    return
}
```

The permission argument `0644` is an octal literal:
- `6` (owner) = read + write
- `4` (group) = read only
- `4` (others) = read only

To write a float or other numeric type, convert to string first:

```go
import (
    "fmt"
    "os"
)

value := 0.9234
text := fmt.Sprintf("%.4f", value)
os.WriteFile("score.txt", []byte(text), 0644)
```

### Writing Incrementally

For appending or writing in multiple steps, use `os.Create` and `bufio.Writer`:

```go
import (
    "bufio"
    "fmt"
    "os"
)

file, err := os.Create("log.txt")
if err != nil {
    fmt.Println("Error creating file:", err)
    return
}

writer := bufio.NewWriter(file)
fmt.Fprintln(writer, "Epoch 1: loss=0.423")
fmt.Fprintln(writer, "Epoch 2: loss=0.381")
fmt.Fprintln(writer, "Epoch 3: loss=0.354")
writer.Flush()    // flush the buffer to disk — don't forget this
file.Close()
```

### Checking if a File Exists

```go
import (
    "errors"
    "os"
)

func fileExists(path string) bool {
    _, err := os.Stat(path)
    return !errors.Is(err, os.ErrNotExist)
}
```

### Parsing a Numeric Value from a File

A pattern common in Go utilities — read a file, parse its content as a number:

```go
import (
    "fmt"
    "os"
    "strconv"
    "strings"
)

func readFloat(path string) (float64, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return 0, fmt.Errorf("readFloat: %w", err)
    }

    text := strings.TrimSpace(string(data))

    value, err := strconv.ParseFloat(text, 64)
    if err != nil {
        return 0, fmt.Errorf("readFloat: invalid float in %q: %w", path, err)
    }

    return value, nil
}
```

`strings.TrimSpace` removes leading and trailing whitespace, which is important when a file's content ends with a newline (as most text editors add one).

---

## 13. Type Conversions

Go has **no implicit type conversions**. Every conversion between types must be explicit. This eliminates a broad class of bugs that arise from accidental coercions in dynamically typed languages.

### Numeric Type Conversions

```go
var i int     = 42
var f float64 = float64(i)    // int → float64
var back int  = int(f)         // float64 → int (truncates toward zero)

var x int32 = 100
var y int64 = int64(x)         // widening: always safe
var z uint8 = uint8(300)       // narrowing: wraps around — 300 % 256 = 44
```

> **Truncation vs rounding:** `int(3.9)` → `3`, not `4`. If you need rounding, use `math.Round` first:
>
> ```go
> rounded := int(math.Round(3.9))   // → 4
> ```

### String Conversions with `strconv`

The `strconv` package is the idiomatic way to convert between strings and primitive types:

```go
import "strconv"

// int ↔ string
n   := 42
s   := strconv.Itoa(n)              // int → string → "42"
n2, err := strconv.Atoi("42")      // string → int (returns an error if it fails)

// float64 ↔ string
f    := 3.14159
fStr := strconv.FormatFloat(f, 'f', 2, 64)   // → "3.14"
f2, err := strconv.ParseFloat("3.14", 64)     // → 3.14

// bool ↔ string
bStr := strconv.FormatBool(true)              // → "true"
b2, err := strconv.ParseBool("true")          // → true

// Always check errors from parsing functions
if err != nil {
    fmt.Println("Parse error:", err)
}
```

**`Itoa` and `Atoi`** are named from the C convention: "integer to ASCII" and "ASCII to integer."

The parsing functions (`Atoi`, `ParseFloat`, `ParseBool`) return an error because any input string might not be a valid representation of the target type.

### String ↔ `[]byte` (Byte Slice)

Strings are immutable in Go, but `[]byte` (a byte slice) is mutable. Converting between them is common in file and network I/O:

```go
s := "Hello, Go!"
b := []byte(s)      // string → []byte (creates a copy)
s2 := string(b)     // []byte → string (creates a copy)
```

These conversions make a copy — the original and the result do not share memory.

### Runes and Strings

A **rune** is an alias for `int32` representing a Unicode code point. Use rune conversions when you need to work with characters rather than bytes:

```go
// string → []rune: correct character count for multi-byte Unicode
s     := "héllo"
runes := []rune(s)
fmt.Println(len(s))      // → 6 (bytes — é is 2 bytes in UTF-8)
fmt.Println(len(runes))  // → 5 (runes / characters)

// rune → string
r := 'A'                // rune literal (single quotes)
fmt.Println(string(r))  // → "A"
fmt.Printf("%T\n", r)   // → int32
```

For ASCII-only text, `len(s)` and character count are the same. For multilingual text, always use `[]rune(s)` when counting or indexing characters.

---

## 14. Scope and Visibility

### Scope Levels in Go

Scope is the region of code where a name is accessible. Go has four levels, from innermost to outermost:

| Scope | Where defined | Accessible from |
|---|---|---|
| **Block** | Inside `{}` braces | That block only |
| **Function** | Inside a function | The whole function |
| **Package** | Outside any function | All files in the package |
| **Universe** | Built-in (`int`, `len`, `true`, etc.) | Everywhere |

Inner scopes can shadow outer ones:

```go
x := "outer"
{
    x := "inner"       // new variable, shadows outer x
    fmt.Println(x)     // → "inner"
}
fmt.Println(x)         // → "outer" (outer x unaffected)
```

### Local Variables

Variables declared inside a function are **local** — they exist only during that function's execution:

```go
func computeLoss() float64 {
    loss := 0.423   // local to computeLoss
    return loss
}

func main() {
    result := computeLoss()
    fmt.Println(result)
    // fmt.Println(loss)  // COMPILE ERROR: undefined: loss
}
```

Variables declared in an `if` initialization statement are scoped to the `if-else` block:

```go
if val, err := strconv.Atoi("42"); err == nil {
    fmt.Println("Parsed:", val)   // val is in scope here
}
// val is NOT in scope here
```

### Package-Level Variables

Variables and constants declared outside any function have **package scope** — they are accessible from any function within the same package:

```go
package main

const defaultEndpoint = "https://api.example.com"
var   requestCount    int = 0

func main() {
    fmt.Println("Endpoint:", defaultEndpoint)
    requestCount++
}

func logStats() {
    fmt.Println("Requests made:", requestCount)  // accessible here too
}
```

Package-level variables must be declared with `var` — the `:=` shorthand is not available outside functions:

```go
var counter int = 0   // ✅ package level
// counter := 0       // ❌ COMPILE ERROR: not inside a function
```

> **Use package-level mutable variables sparingly.** They make behavior harder to reason about and test. Prefer passing values through function parameters.

### Exported vs. Unexported Identifiers

Go's visibility model is defined entirely by **capitalization**:

| First letter | Visibility | Term |
|---|---|---|
| Uppercase (`A–Z`) | Any package that imports this one | **Exported** |
| Lowercase (`a–z`) | Only within the declaring package | **Unexported** |

```go
package fileops

var hitCount int = 0           // unexported — only usable within fileops

var Version = "1.0.0"          // exported — visible to any importer

func parseValue(s string) {}   // unexported helper

func ReadFloat(path string) (float64, error) {  // exported
    parseValue(path)           // OK: same package
    // ...
}
```

From another package:

```go
import "example.com/myproject/fileops"

fileops.ReadFloat("data.txt")   // ✅ exported
fileops.Version                 // ✅ exported
// fileops.parseValue()         // ❌ COMPILE ERROR: unexported
// fileops.hitCount             // ❌ COMPILE ERROR: unexported
```

> This is Go's **entire** visibility model. No `public`, `private`, or `protected` keywords — just capitalization. Elegantly simple, and it works at the package boundary rather than the class boundary.

---

## 15. Common Go Idioms

These patterns appear throughout idiomatic Go code. Learning to recognize and apply them is the fastest way to write Go that feels natural to other engineers.

### Guard Clauses and Early Returns

Rather than nesting logic deep inside `if-else` blocks, idiomatic Go handles error and edge cases at the top of a function and returns immediately. This keeps the "happy path" at the left margin:

```go
// Non-idiomatic: deeply nested, hard to read
func process(data []byte) (string, error) {
    if data != nil {
        result, err := parse(data)
        if err == nil {
            return result, nil
        } else {
            return "", err
        }
    }
    return "", errors.New("data is nil")
}

// Idiomatic Go: guard clauses, flat structure
func process(data []byte) (string, error) {
    if data == nil {
        return "", errors.New("process: data is nil")
    }
    result, err := parse(data)
    if err != nil {
        return "", fmt.Errorf("process: %w", err)
    }
    return result, nil
}
```

### Multiple Assignment and Swap

```go
a, b := 1, 2
a, b = b, a         // swap without a temporary variable
fmt.Println(a, b)   // → 2 1
```

### The `init()` Function

Each package can define one or more `init()` functions that run automatically before `main()` (or before the package is used by another package). They require no explicit call:

```go
package main

import (
    "fmt"
    "os"
)

var configPath string

func init() {
    configPath = os.Getenv("CONFIG_PATH")
    if configPath == "" {
        configPath = "config/default.yaml"
    }
}

func main() {
    fmt.Println("Using config:", configPath)
}
```

Use `init()` for lightweight, one-time setup that cannot fail. Avoid complex logic, I/O errors, or side effects in `init()` — they are difficult to test and hard to reason about.

### The `run()` Pattern

A clean idiom for `main()`: delegate all real logic to a `run()` function that returns an error, keeping `main()` responsible only for exit handling:

```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

func run() error {
    // all real application logic here
    return nil
}
```

Benefits: `run()` is testable, errors are propagated cleanly, and `os.Exit(1)` is called in exactly one place.

### Sentinel Errors

Define expected failure conditions as package-level error variables, following the `ErrSomething` naming convention:

```go
var (
    ErrNotFound   = errors.New("not found")
    ErrTimeout    = errors.New("operation timed out")
    ErrPermission = errors.New("permission denied")
)

func fetchRecord(id int) (string, error) {
    if id <= 0 {
        return "", ErrNotFound
    }
    return "record_data", nil
}

_, err := fetchRecord(-1)
if errors.Is(err, ErrNotFound) {
    fmt.Println("Record does not exist — returning empty state")
}
```

### Adding Context When Propagating Errors

Each layer of your call stack should add context to an error before returning it:

```go
// At the bottom: a descriptive base error
func readConfig(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("readConfig(%q): %w", path, err)
    }
    return data, nil
}

// One level up: more context
func startService(cfgPath string) error {
    _, err := readConfig(cfgPath)
    if err != nil {
        return fmt.Errorf("startService: %w", err)
    }
    return nil
}
```

The final error message assembles into a chain like:
```
startService: readConfig("config.yaml"): open config.yaml: no such file or directory
```

This level of detail is invaluable in production systems.

---

## 16. Best Practices

### Format with `gofmt` / `goimports`

Go has an official, opinionated formatter: `gofmt`. There is no style debate in Go — one formatter, one style, universally applied. Run it before every commit:

```bash
gofmt -w .          # format all .go files in place
goimports -w .      # like gofmt, but also organizes and removes unused imports
```

Most editors (VS Code with the Go extension, GoLand) run `goimports` on save automatically. This is standard practice across every Go team.

### Naming Conventions

| Construct | Convention | Examples |
|---|---|---|
| Local variables & functions | `camelCase` | `userName`, `parseFloat`, `maxRetries` |
| Exported identifiers | `PascalCase` | `GetUser`, `ReadFloat`, `DefaultTimeout` |
| Package names | `lowercase`, short, no underscores | `fileops`, `httputil`, `mathutils` |
| File names | `snake_case.go` | `file_ops.go`, `http_client.go` |
| Error variables | `ErrSomething` | `ErrNotFound`, `ErrTimeout` |

**Keep names short in tight scopes.** Go convention favors `i`, `n`, `b`, `err`, `ok` for short-lived loop and error variables. Longer names belong at the package level and in function signatures.

### Always Handle Errors

The number one rule in Go: never silently discard errors.

```go
// ❌ Never do this
data, _ := os.ReadFile("config.txt")

// ✅ Always do this
data, err := os.ReadFile("config.txt")
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}
```

Use `_` to discard an error only when you have consciously decided the failure mode is genuinely irrelevant — and even then, add a comment explaining why.

### Comment Exported Identifiers

Every exported function, variable, and constant should have a **doc comment** that starts with the identifier's name. These are read by `godoc` and displayed in editor tooling:

```go
// ReadFloat reads a float64 value from the specified file.
// It returns an error if the file cannot be opened or its
// content cannot be parsed as a valid floating-point number.
func ReadFloat(path string) (float64, error) {
    // ...
}

// DefaultPort is the TCP port used when no port is specified.
const DefaultPort = 8080
```

### Keep Functions Small and Focused

A function should do one thing clearly. If a function is handling multiple logical concerns, break it into smaller functions. This also makes testing easier.

### Use `go vet` and `staticcheck`

The compiler catches type errors, but these tools catch a second tier of mistakes:

```bash
go vet ./...        # catches misused format verbs, unreachable code, suspicious constructs

# Install staticcheck:
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...   # broader analysis: deprecated APIs, performance issues, and more
```

Both tools are free, fast, and should be part of your development workflow from the start.

### Reserve `panic` for Programming Errors

`panic` terminates the program immediately. It is appropriate for:
- Invariant violations that indicate a programming bug (e.g., `nil` passed to a function that requires a non-nil value)
- Unrecoverable initialization failures at startup

It is **not** appropriate for expected runtime errors like missing files, network failures, or invalid user input. Those should return `error`.

```go
func mustGetEnv(key string) string {
    val := os.Getenv(key)
    if val == "" {
        panic(fmt.Sprintf("required env var %q is not set", key))
    }
    return val
}
```

The `Must...` or `must...` naming convention signals that a function panics instead of returning an error.

---

## 17. What to Learn Next

This guide covers Go's foundational layer. Here is a structured path forward, ordered by priority for ML/AI engineering.

### Intermediate Go

These are the next essential building blocks, regardless of specialization:

| Topic | Why It Matters |
|---|---|
| **Structs** | Go's primary data modeling tool. Used everywhere. Learn this first. |
| **Methods** | Attach behavior to structs. Required before interfaces. |
| **Interfaces** | Go's polymorphism model. Central to idiomatic design patterns. |
| **Pointers** | Memory efficiency and mutation semantics. Required to understand method receivers. |
| **Slices & Maps** | The fundamental collection types. Used in nearly every program. |
| **Goroutines & Channels** | Go's native concurrency model. The primary reason to choose Go for infrastructure. |
| **Context package** | Cancellation and deadlines. Critical for production service calls and LLM API timeouts. |
| **Defer, closures, recover** | Advanced function mechanics. Essential for resource cleanup and resilience patterns. |
| **Testing** | Table-driven tests with the `testing` package. Mocks, benchmarks, and fuzz tests. |
| **Error types & `errors.As`** | Advanced error inspection beyond `errors.Is`. |
| **Generics** | Type-parameterized code, available since Go 1.18. |
| **Struct embedding** | Go's composition-based alternative to inheritance. |
| **Reflection** | Runtime type inspection via the `reflect` package. Use sparingly. |

### Backend Development

Building production services in Go:

- HTTP servers: `net/http`, and popular routers like Chi, Gin, and Echo
- REST API design and JSON serialization
- Middleware: authentication, logging, rate limiting, request tracing
- Databases: `database/sql`, `sqlx`, `pgx` (PostgreSQL), GORM
- Authentication: JWT, OAuth2, session management

### For ML Infrastructure (Especially Valuable)

Given your work with LangGraph multi-agent systems, Kubernetes, and production ML pipelines, these Go topics carry the highest return on investment:

| Topic | Application to ML/AI Engineering |
|---|---|
| **Goroutines + Worker pools** | Parallel inference requests, concurrent batch processing |
| **Channels** | Coordination between concurrent workers and agent pipelines |
| **Context** | Cancellation of LLM API calls, request timeouts, propagating deadlines |
| **gRPC** | High-performance inter-service communication for model serving |
| **HTTP APIs** | REST endpoints wrapping ML models and data pipelines |
| **Message queues** | Kafka / NATS clients for event-driven ML pipeline orchestration |
| **Kubernetes client libraries** | Build custom controllers and operators for ML workloads (`client-go`, Kubebuilder) |
| **Prometheus metrics** | Expose inference latency, queue depth, error rates |
| **OpenTelemetry** | Distributed tracing across multi-agent pipelines |
| **Docker integration** | Programmatic container lifecycle management |
| **Microservices patterns** | Circuit breakers, retries with backoff, sidecar patterns |
| **Distributed systems** | Leader election, saga pattern, exactly-once processing |

**Recommended learning order for ML infra work:**

```
Structs & Methods → Interfaces → Pointers
       ↓
Goroutines & Channels → Context
       ↓
gRPC → Kubernetes client → Observability (Prometheus + OpenTelemetry)
```

---

*This guide is part of a self-study series on Go for ML/AI Infrastructure Engineering.*
