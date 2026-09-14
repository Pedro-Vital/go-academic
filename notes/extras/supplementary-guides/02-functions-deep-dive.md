# Supplementary Guide: Go Functions Deep Dive

**Roadmap Position:** Supplementary material, sitting alongside the main track. It expands on the function basics from Guide 2 (which covered plain function declarations, simple closures, and slices/maps) and is worth finishing before or during Phase E, since HTTP handlers, middleware chains, and dependency injection in Chi/Gin all lean heavily on functions as values.

**Goal:** By the end of this guide you will be able to treat functions as ordinary values, write anonymous functions and closures fluently, implement recursive algorithms safely in Go, and design flexible APIs with variadic parameters — including the functional options pattern you'll see constantly in production Go code and open-source libraries.

## Table of Contents

1. Why This Supplement Exists
2. Quick Recap: Function Basics
3. Functions as Values
4. Function Types as a First-Class Citizen
5. Higher-Order Functions: Passing Functions as Arguments
6. Returning Functions from Functions
7. Anonymous Functions
8. Immediately-Invoked Function Expressions (IIFE)
9. Closures, Revisited
10. Recursion in Go
11. Recursion Gotchas: Stack Growth & Missing TCO
12. Mutual Recursion
13. Variadic Functions
14. Spreading Slices into Variadic Functions
15. Variadic Parameters Combined with Regular Parameters
16. Real-World Pattern: Functional Options
17. Python ↔ Go Comparison Summary
18. Exercises
19. Key Takeaways

---

## 1. Why This Supplement Exists

Guide 2 introduced functions as a syntax feature: `func` declarations, parameters, return values, and a first taste of closures. That's enough to write straight-line code. It is **not** enough to read idiomatic Go, because idiomatic Go treats functions as data constantly:

- `http.HandlerFunc` wraps a function so it satisfies an interface.
- Middleware is "a function that takes a handler and returns a handler."
- `sort.Slice` takes a comparison function.
- Dependency injection in Go often means passing constructor functions around, not classes.

This guide closes that gap, and adds two topics Guide 2 didn't touch at all: **recursion** (with Go's real stack behavior, not Python's) and **variadic functions** (Go's `...T`, the analog of Python's `*args`).

---

## 2. Quick Recap: Function Basics

```go
func add(a, b int) int {
    return a + b
}

func divmod(a, b int) (int, int) {
    return a / b, a % b
}
```

| Concept | Go | Python |
|---|---|---|
| Declaration | `func name(params) returnType` | `def name(params):` |
| Multiple returns | Native, comma-separated | Tuple, comma-separated |
| Default params | Not supported | Supported (`def f(x=1)`) |
| Named params at call site | Not supported | Supported (`f(x=1)`) |

Go has no default arguments and no keyword arguments. If you want that ergonomics, you reach for the functional options pattern (Section 16) or a config struct.

---

## 3. Functions as Values

In Go, a function is a value with a type, just like an `int` or a `string`. You can assign it to a variable, store it in a slice or map, pass it as an argument, and return it from another function.

```go
package main

import "fmt"

func add(a, b int) int { return a + b }
func sub(a, b int) int { return a - b }

func main() {
    var op func(int, int) int // a variable whose type is "function from (int,int) to int"

    op = add
    fmt.Println(op(3, 4)) // 7

    op = sub
    fmt.Println(op(3, 4)) // -1

    // Functions in a map -> a simple dispatch table
    ops := map[string]func(int, int) int{
        "+": add,
        "-": sub,
    }
    fmt.Println(ops["+"](10, 5)) // 15
}
```

**Python analogy:** this is exactly `operator.add` / `operator.sub`, or your own `def`-defined functions, stored in a `dict` and looked up by key. Python programmers do this instinctively (`handlers = {"GET": get_handler, "POST": post_handler}`); the Go version reads almost identically, just with explicit types.

```python
def add(a, b): return a + b
def sub(a, b): return a - b

ops = {"+": add, "-": sub}
print(ops["+"](10, 5))  # 15
```

### Gotcha: functions are not comparable (except to `nil`)

```go
var f func()
fmt.Println(f == nil) // OK: true

var g, h func()
// fmt.Println(g == h) // compile error: func can only be compared to nil
```

This mirrors Python, where comparing two functions with `==` checks identity, not behavior — but Go goes further and simply forbids the comparison at compile time (except against `nil`). If you need to check "has this callback been set," compare to `nil`; don't try to compare two funcs for equality.

---

## 4. Function Types as a First-Class Citizen

Because function signatures are types, you can name them, which cleans up long declarations and gives you a vocabulary for documentation.

```go
type BinOp func(int, int) int

func apply(op BinOp, a, b int) int {
    return op(a, b)
}

func main() {
    fmt.Println(apply(add, 3, 4)) // 7
}
```

`BinOp` is a **defined type** whose underlying type is `func(int, int) int`. Any function with that exact signature satisfies it — no explicit "implements" declaration needed, the same structural typing you saw with interfaces in Guide 4.

**Python analogy:** `Callable[[int, int], int]` from the `typing` module. It's a type hint, not enforced at runtime the way Go's function types are enforced at compile time — but the intent (documenting "this parameter expects a two-int-argument callable returning an int") is identical.

```python
from typing import Callable

BinOp = Callable[[int, int], int]

def apply(op: BinOp, a: int, b: int) -> int:
    return op(a, b)
```

---

## 5. Higher-Order Functions: Passing Functions as Arguments

A higher-order function takes another function as a parameter (or returns one — Section 6). The standard library leans on this heavily.

```go
package main

import (
    "fmt"
    "sort"
)

type Person struct {
    Name string
    Age  int
}

func main() {
    people := []Person{
        {"Zara", 30},
        {"Amir", 25},
        {"Mina", 35},
    }

    sort.Slice(people, func(i, j int) bool {
        return people[i].Age < people[j].Age
    })

    fmt.Println(people) // sorted by Age ascending
}
```

`sort.Slice` takes a `less(i, j int) bool` function — you tell it how to compare, it handles the algorithm. This is the same shape as Python's `sorted(people, key=lambda p: p.age)`, except Go's callback compares two indices directly rather than extracting a sort key.

| Task | Go | Python |
|---|---|---|
| Custom sort | `sort.Slice(s, func(i, j int) bool {...})` | `sorted(s, key=lambda x: ...)` |
| Filter | write your own loop, or use `slices` package helpers | `filter(pred, iterable)` / list comprehension |
| Map | write your own loop | `map(fn, iterable)` / list comprehension |

Go's standard library does **not** ship generic `Map`/`Filter`/`Reduce` functions the way Python has built-ins for `map`/`filter`. Since Go 1.21 the `slices` and `maps` packages (built on generics) added some helpers, but the idiomatic style is still often a plain `for` loop — Go favors explicitness over point-free chains.

```go
func filter[T any](items []T, keep func(T) bool) []T {
    var out []T
    for _, item := range items {
        if keep(item) {
            out = append(out, item)
        }
    }
    return out
}

evens := filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
// evens == [2 4 6]
```

(This uses generics syntax briefly — you haven't covered generics formally yet per the roadmap's open item, but the pattern reads naturally: `[T any]` just means "works for any type T," analogous to Python's duck typing.)

---

## 6. Returning Functions from Functions

A function can also produce a function as its result — the core of building configurable behavior without a class hierarchy.

```go
func multiplier(factor int) func(int) int {
    return func(n int) int {
        return n * factor
    }
}

func main() {
    double := multiplier(2)
    triple := multiplier(3)

    fmt.Println(double(5)) // 10
    fmt.Println(triple(5)) // 15
}
```

`multiplier` returns a closure (Section 9) that "remembers" `factor`. This is precisely Python's closure-returning-factory pattern:

```python
def multiplier(factor):
    def inner(n):
        return n * factor
    return inner

double = multiplier(2)
print(double(5))  # 10
```

Same idea, same mental model — Go just requires you to write out the return type (`func(int) int`) instead of letting it stay implicit.

---

## 7. Anonymous Functions

An anonymous function is a function literal with no name attached — you either use it immediately or assign it to a variable on the spot.

```go
// Assigned to a variable
square := func(n int) int {
    return n * n
}
fmt.Println(square(5)) // 25

// Passed inline as an argument
result := apply(func(a, b int) int { return a * b }, 3, 4)
fmt.Println(result) // 12
```

**Python analogy:** this is broader than Python's `lambda`, because Go's anonymous functions can have full bodies with multiple statements, loops, and multiple return values — a `lambda` in Python is restricted to a single expression. The closer Python analog for a multi-line anonymous function is actually a locally-defined `def` inside another function, just without binding it to a name before use.

```python
# Python lambda: single expression only
square = lambda n: n * n

# For anything more complex, Python needs a full def — Go's anonymous func doesn't have that limit
def make_validator():
    def _inner(x):
        if x < 0:
            return False
        return x % 2 == 0
    return _inner
```

```go
// Go's anonymous func handles the same multi-statement logic inline, no name required
validator := func(x int) bool {
    if x < 0 {
        return false
    }
    return x%2 == 0
}
```

---

## 8. Immediately-Invoked Function Expressions (IIFE)

Sometimes you want to run a block of code once, right where it's written, usually to scope some setup logic or to work around the lack of expression-level `if`.

```go
result := func() int {
    total := 0
    for i := 1; i <= 10; i++ {
        total += i
    }
    return total
}() // <-- called immediately

fmt.Println(result) // 55
```

**Python analogy:** Python doesn't really have an idiomatic IIFE — the closest equivalent is defining a throwaway function and calling it once, or more commonly, just using a list comprehension or a walrus operator for the same "compute this once, inline" need. IIFEs show up in Go more often than in Python because Go doesn't have comprehensions or ternary expressions to fall back on.

A very common real Go use of the IIFE pattern is `go func() { ... }()` — launching a goroutine (Guide 7) with an anonymous function body, executed immediately in its own goroutine.

---

## 9. Closures, Revisited

Guide 2 introduced closures briefly. Here's the deeper mental model, plus the one gotcha that trips up almost everyone coming from Python.

A closure is a function value that references variables from outside its own body. The function "closes over" those variables — it keeps a live reference, not a copy.

```go
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

func main() {
    c1 := counter()
    c2 := counter()

    fmt.Println(c1()) // 1
    fmt.Println(c1()) // 2
    fmt.Println(c2()) // 1 — independent state, separate closure
}
```

This matches Python's closures exactly (modulo Python needing `nonlocal` to *mutate* an outer variable from within a nested function; Go needs no such keyword because `count++` mutates the variable Go already captured by reference).

```python
def counter():
    count = 0
    def inner():
        nonlocal count
        count += 1
        return count
    return inner
```

### The loop-variable gotcha (and the Go 1.22 fix)

This is one of the most infamous Go footguns, and it's worth knowing about even though **Go 1.22+ fixed it**, because you will still see pre-1.22 code (and pre-1.22 explanations) in the wild.

```go
// Go 1.22+ behavior (current default): this is safe and does what you'd expect.
funcs := make([]func(), 0, 3)
for i := 0; i < 3; i++ {
    funcs = append(funcs, func() { fmt.Println(i) })
}
for _, f := range funcs {
    f()
}
// Go 1.22+: prints 0, 1, 2
// Go 1.21 and earlier: printed 3, 3, 3 — every closure captured the SAME variable i,
// and by the time the funcs ran, the loop had already finished with i == 3.
```

Before 1.22, the fix was to shadow the loop variable inside the loop body:

```go
for i := 0; i < 3; i++ {
    i := i // shadow: each iteration gets its own copy
    funcs = append(funcs, func() { fmt.Println(i) })
}
```

Check `go.mod`'s `go` directive for your module — if it says `go 1.22` or later, you get the fixed, per-iteration variable semantics automatically. Since your roadmap targets modern Go, you're safe by default, but you should recognize the old pattern immediately when reading legacy code or older tutorials/Stack Overflow answers.

Python has no equivalent gotcha in `for` loops (each loop body shares the same loop variable across iterations too, but Python closures capture by *name lookup at call time* rather than by reference-to-a-fixed-storage-slot, so the practical symptom — and the classic `lambda`-in-a-loop bug — is actually the same failure mode as *pre-1.22* Go):

```python
funcs = [lambda: i for i in range(3)]
print([f() for f in funcs])  # [2, 2, 2] — same "gotcha" as old Go!

# fix: default-argument trick to capture by value
funcs = [lambda i=i: i for i in range(3)]
print([f() for f in funcs])  # [0, 1, 2]
```

So: modern Go (1.22+) is actually *safer* here than Python's `lambda` default behavior. Worth remembering, since it inverts the usual "Python is more forgiving" expectation.

---

## 10. Recursion in Go

Recursion works the same way it does everywhere: a function calls itself with a smaller version of the problem until it hits a base case.

```go
func factorial(n int) int {
    if n <= 1 {
        return 1 // base case
    }
    return n * factorial(n-1) // recursive case
}

func main() {
    fmt.Println(factorial(5)) // 120
}
```

```python
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)
```

Structurally identical. The differences that matter are about *runtime behavior*, not syntax — covered next.

A classic recursive tree-walk (binary search tree lookup), since this is the shape you'll actually use recursion for in real Go code more than toy factorials:

```go
type TreeNode struct {
    Val         int
    Left, Right *TreeNode
}

func contains(node *TreeNode, target int) bool {
    if node == nil {
        return false // base case: fell off the tree
    }
    if node.Val == target {
        return true
    }
    if target < node.Val {
        return contains(node.Left, target)
    }
    return contains(node.Right, target)
}
```

---

## 11. Recursion Gotchas: Stack Growth & Missing TCO

This is the section Python doesn't really prepare you for, because Python's recursion limit and Go's are governed by completely different mechanisms.

### Python's recursion limit is artificial and shallow

Python enforces `sys.getrecursionlimit()` (default 1000) as a hard, configurable ceiling, specifically to protect the C call stack from overflowing. Hit it and you get a `RecursionError` well before you'd actually crash the process.

### Go has no such artificial limit — goroutine stacks grow dynamically

A goroutine (including `main`'s initial goroutine) starts with a small stack (a few KB) and the Go runtime **automatically grows it** as needed, up to a default maximum of 1 GB on 64-bit systems (configurable via `debug.SetMaxStack`). This means:

- Go will happily recurse tens of thousands of levels deep for many workloads without complaint.
- There is no equivalent of Python's `sys.setrecursionlimit(10000)` — you don't need to raise a ceiling, because there isn't a low one to begin with.
- The failure mode when you *do* blow the stack is a fatal, unrecoverable runtime error (`fatal error: stack overflow`) that **cannot be recovered** with `recover()` (unlike a panic) — it crashes the whole program immediately.

```go
func infiniteRecursion(n int) int {
    return infiniteRecursion(n + 1) // no base case — will eventually crash
}
```

Running this produces `fatal error: stack overflow` and exits — `recover()` in a deferred function will not save you, because this isn't a panic, it's the runtime itself refusing to continue.

### No tail-call optimization (TCO)

Some languages (Scheme, and to a partial extent modern JS engines) optimize tail-recursive calls into loops, so tail recursion costs O(1) stack space. **Go's compiler does not do this**, and neither does CPython. If you write deeply tail-recursive Go expecting loop-like stack behavior, you won't get it — each call is a real stack frame.

```go
// Tail-recursive style — but Go will NOT optimize this into a loop.
func sumTo(n, acc int) int {
    if n == 0 {
        return acc
    }
    return sumTo(n-1, acc+n) // tail call, but still a real stack frame in Go
}
```

**Practical rule of thumb:** for anything that might recurse deeper than a few thousand levels on user-controlled input (parsing untrusted/attacker-controlled nested JSON, deeply nested file trees, etc.), prefer an explicit loop with your own stack (a slice used as a stack) over recursion. Go's dynamic stack growth makes recursion much more forgiving than Python's default, but it is not infinite, and a stack-overflow fatal error is worse than a caught exception — it takes the whole process down.

| Behavior | Go | Python |
|---|---|---|
| Recursion depth limit | None enforced; bounded by actual stack memory (default max 1 GB) | `sys.getrecursionlimit()`, default 1000, raisable |
| Failure mode | `fatal error: stack overflow` — unrecoverable, crashes process | `RecursionError` — a catchable exception |
| Tail-call optimization | No | No |
| Typical practical depth before trouble | Tens of thousands+ (workload dependent) | ~1000 by default |

---

## 12. Mutual Recursion

Two (or more) functions can call each other. Go handles this naturally as long as both functions are declared in the same package — order of declaration doesn't matter at the package level (unlike inside a function body).

```go
func isEven(n int) bool {
    if n == 0 {
        return true
    }
    return isOdd(n - 1)
}

func isOdd(n int) bool {
    if n == 0 {
        return false
    }
    return isEven(n - 1)
}
```

This works even though `isEven` refers to `isOdd` before `isOdd` has been declared in the source file — Go resolves all package-level names in a first pass, so declaration order at package scope is irrelevant (this is unlike Python, where a `def` must appear, or at least be bound, before it's called at runtime — though since Python only checks names when the function actually *executes*, mutual recursion works fine there too as long as both `def`s exist by call time).

```python
def is_even(n):
    if n == 0:
        return True
    return is_odd(n - 1)

def is_odd(n):
    if n == 0:
        return False
    return is_even(n - 1)
```

---

## 13. Variadic Functions

A variadic parameter accepts zero or more arguments of a given type, collected into a slice inside the function. It's Go's answer to Python's `*args`.

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func main() {
    fmt.Println(sum())           // 0
    fmt.Println(sum(1))          // 1
    fmt.Println(sum(1, 2, 3, 4)) // 10
}
```

Inside the function, `nums` is just an ordinary `[]int` — nothing magical about how you use it once you're inside the body.

```python
def total(*nums):
    return sum(nums)  # nums is a plain tuple inside the function

total()          # 0
total(1)         # 1
total(1, 2, 3, 4) # 10
```

### Key difference: Go allows only one variadic parameter, and it must be last

```go
func log(level string, args ...interface{}) { /* ... */ } // OK — level, then variadic

// func bad(args ...int, level string) {} // compile error: variadic must be final param
```

Python is more permissive: `*args` can be followed by keyword-only arguments (`def f(*args, key=1)`), because Python's calling convention separates positional and keyword binding. Go has no keyword arguments, so this flexibility doesn't apply — variadic is strictly "last parameter, absorbs everything remaining."

### `fmt.Println` and `fmt.Printf` are variadic

You've been using variadic functions since Guide 1 without necessarily naming the concept:

```go
func Println(a ...any) (n int, err error)
func Printf(format string, a ...any) (n int, err error)
```

That's why `fmt.Println("x:", x, "y:", y)` accepts any number of arguments — same shape as Python's `print(*values)`.

---

## 14. Spreading Slices into Variadic Functions

If you already have a slice and want to pass its elements as individual variadic arguments (rather than as one slice argument), use the `...` spread operator at the **call site**.

```go
nums := []int{1, 2, 3, 4, 5}
fmt.Println(sum(nums...)) // spreads nums into individual args — 15

// Without the ..., this is a compile error:
// fmt.Println(sum(nums)) // cannot use nums (variable of type []int) as int value
```

This directly parallels Python's `*` unpacking at a call site:

```python
nums = [1, 2, 3, 4, 5]
print(total(*nums))  # 15 — same spread concept
```

One restriction Go has that Python doesn't: you can only spread a slice whose element type exactly matches the variadic parameter's type, and it must be the **only** argument in that position — you can't mix `sum(nums..., 6)` to spread a slice and add one more value in the same call. If you need that, build a combined slice first: `sum(append(nums, 6)...)`.

---

## 15. Variadic Parameters Combined with Regular Parameters

Regular parameters come first, the variadic parameter comes last and absorbs everything else.

```go
func joinWithPrefix(prefix string, parts ...string) string {
    result := prefix
    for _, p := range parts {
        result += " " + p
    }
    return result
}

fmt.Println(joinWithPrefix(">>", "hello", "world")) // ">> hello world"
fmt.Println(joinWithPrefix(">>"))                    // ">>" — parts is an empty slice, not nil-panic
```

### Gotcha: nil vs. empty slice, again

Just like Guide 9's JSON nil-vs-empty-slice trap, calling a variadic function with zero variadic arguments gives you an **empty, non-nil slice** in most Go versions' observable behavior for iteration purposes — but you should never rely on nil-ness here for behavior. Always range over it or check `len()`, never assume you need a nil guard before ranging (`range` over a nil slice is always safe and simply doesn't iterate).

```go
func describe(items ...string) {
    fmt.Println("count:", len(items)) // safe even if items is nil/empty
    for _, item := range items {      // safe even if items is nil — just zero iterations
        fmt.Println(item)
    }
}
```

---

## 16. Real-World Pattern: Functional Options

Because Go has no default arguments and no keyword arguments, constructing objects with many optional settings is awkward with a plain constructor. The idiomatic fix — seen throughout the standard library and major Go projects (e.g. `grpc-go`, `aws-sdk-go-v2`) — combines **variadic parameters** + **functions as values** + **closures** into one pattern: functional options.

```go
type Server struct {
    Host    string
    Port    int
    Timeout int
    TLS     bool
}

type Option func(*Server) // a function type that mutates a *Server

func WithPort(port int) Option {
    return func(s *Server) { s.Port = port }
}

func WithTimeout(seconds int) Option {
    return func(s *Server) { s.Timeout = seconds }
}

func WithTLS() Option {
    return func(s *Server) { s.TLS = true }
}

func NewServer(host string, opts ...Option) *Server {
    s := &Server{
        Host:    host,
        Port:    8080, // sensible defaults
        Timeout: 30,
    }
    for _, opt := range opts {
        opt(s) // each option mutates s
    }
    return s
}

func main() {
    s1 := NewServer("localhost")
    // s1: Port=8080, Timeout=30, TLS=false

    s2 := NewServer("localhost", WithPort(9090), WithTLS())
    // s2: Port=9090, Timeout=30, TLS=true
}
```

This gives you the ergonomics of Python's keyword arguments with defaults, built entirely out of concepts you now have: a named function type (Section 4), closures that capture a configuration value (Section 9), and a variadic parameter to accept "zero or more settings" (Section 13).

```python
# The direct Python equivalent needs none of this ceremony —
# this is precisely the gap the functional-options pattern is compensating for.
class Server:
    def __init__(self, host, port=8080, timeout=30, tls=False):
        self.host = host
        self.port = port
        self.timeout = timeout
        self.tls = tls

s2 = Server("localhost", port=9090, tls=True)
```

Recognizing this pattern is important: once you see `Option func(*T)` plus a variadic constructor parameter, you'll immediately know "this is Go's substitute for keyword defaults" the next time you encounter it in a library's API (and you will encounter it often).

---

## 17. Python ↔ Go Comparison Summary

| Concept | Python | Go |
|---|---|---|
| Functions as values | Native — everything is an object | Native — functions have types, assignable like any value |
| Anonymous function | `lambda` (single expression only) | `func(...) {...}` literal (full statement bodies allowed) |
| Closures | Native; needs `nonlocal` to mutate outer var | Native; no keyword needed, captures by reference |
| Loop-variable capture bug | Present (`lambda`s in loop share the loop var) | Fixed by default in Go 1.22+; present pre-1.22 |
| Recursion depth limit | ~1000, artificial, raisable | Unbounded by policy; bounded by real stack memory (default max 1 GB) |
| Recursion failure mode | `RecursionError` (catchable) | `fatal error: stack overflow` (unrecoverable, crashes process) |
| Tail-call optimization | None | None |
| Variadic args | `*args` (any position before keyword-only args) | `...T` (must be the single, final parameter) |
| Spreading a collection into a call | `f(*my_list)` | `f(mySlice...)` |
| Default / keyword arguments | Native (`def f(x=1)`, `f(x=1)`) | Not supported — functional options pattern instead |

---

## 18. Exercises

1. **Dispatch table:** Write a `calculate(op string, a, b float64) (float64, error)` function backed by a `map[string]func(float64, float64) float64` for `+`, `-`, `*`, `/`. Return an error for unknown ops and for division by zero.
2. **Memoized recursion:** Write a recursive Fibonacci function, then rewrite it using a closure that captures a `map[int]int` cache so repeated calls don't redo work. Compare how many times the un-cached version calls itself for `fib(30)` vs. the cached version (add a counter).
3. **Deliberately overflow the stack:** Write a function with no base case, call it, and observe the `fatal error: stack overflow`. Then wrap the call in a `defer`/`recover()` and confirm (as the guide claims) that `recover()` does **not** catch it.
4. **Variadic validator:** Write `func allPositive(nums ...int) bool` that returns `true` only if every argument is positive (and `true` for zero arguments — think about why that's the sane default, same as Python's `all([])`).
5. **Spread a slice:** Given `scores := []int{88, 92, 74, 100}`, call your `sum(...)` from Section 13 by spreading `scores` into it, and separately by summing it manually with a loop. Confirm both give the same answer.
6. **Functional options:** Extend the `Server` example in Section 16 with a `WithHost(host string) Option` and a `WithMaxConnections(n int) Option`. Then write a version of `NewServer` that returns an error if `opts` produces an invalid configuration (e.g., `Port <= 0`) — this previews the error-handling style from Guide 5 combined with functional options.

---

## 19. Key Takeaways

- **Functions are values** in Go, just like `int` or `string`. You can store them in variables, slices, and maps, and pass or return them freely — no special "first-class function" ceremony required, same as Python.
- **Anonymous functions** are Go's `lambda`-and-more: unlike Python's single-expression `lambda`, Go's anonymous functions can have full multi-statement bodies, loops, and multiple return values.
- **Closures** capture variables by reference. Go 1.22+ fixed the classic "loop variable capture" bug that still exists in Python's `lambda`-in-a-loop pattern — a rare case where modern Go is the safer default.
- **Recursion** works the same logically, but the runtime story is very different: Go has no artificial recursion-depth ceiling (Python enforces one at ~1000), Go's goroutine stacks grow dynamically, and a real Go stack overflow is an unrecoverable fatal error, not a catchable exception like Python's `RecursionError`. Neither language optimizes tail calls.
- **Variadic functions** (`...T`) are Go's `*args`, with two restrictions Python doesn't have: only one variadic parameter is allowed, and it must be the last one. Spread a slice into a call with `slice...`.
- **Functional options** (`type Option func(*T)` + a variadic constructor) is how idiomatic Go recreates the ergonomics of Python's default/keyword arguments, since Go supports neither natively. Recognizing this pattern will help you read most well-designed Go library APIs on sight.

**Where this fits next:** Phase E's web layer will use every concept in this guide immediately — HTTP handlers are functions passed as values, middleware is a function that wraps a handler function and returns a new one (Section 6's "return a function" pattern, applied to `http.Handler`), and router setup in Chi/Gin often uses functional options for configuration.