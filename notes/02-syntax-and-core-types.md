# Guide 2 — Syntax & Core Types

> **Roadmap position:** Phase B, Guide 2 of 14 — *Core language, hands-on*
> **Goal of this guide:** Get real Go syntax under your fingers — variables, types, control flow, functions, and Go's two most important built-in data structures (slices and maps). Everything here is contrasted against Python so you can map your existing instincts onto the new syntax, while flagging the places those instincts will actively mislead you.

---

## Table of Contents

1. [How This Guide Builds on Guide 1](#1-how-this-guide-builds-on-guide-1)
2. [Variable Declarations: `var`, `:=`, and `const`](#2-variable-declarations-var--and-const)
3. [Zero Values — Go's Answer to `None`](#3-zero-values--gos-answer-to-none)
4. [Basic Types in Depth](#4-basic-types-in-depth)
5. [Type Conversion — No Implicit Coercion, Ever](#5-type-conversion--no-implicit-coercion-ever)
6. [Control Flow](#6-control-flow)
7. [Functions in Depth](#7-functions-in-depth)
8. [Arrays vs Slices — The Most Important Section in This Guide](#8-arrays-vs-slices--the-most-important-section-in-this-guide)
9. [Maps](#9-maps)
10. [Packages & Imports: The Mechanics](#10-packages--imports-the-mechanics)
11. [Constants and `iota`](#11-constants-and-iota)
12. [Summary & What's Next](#12-summary--whats-next)

---

## 1. How This Guide Builds on Guide 1

Guide 1 gave you the *why*. This guide gives you the *how*, at the statement-and-expression level. Two things carry over directly and won't be re-explained:

- **Static typing is enforced at compile time** (Guide 1, §4) — every declaration below has a fixed type, inferred or explicit, and it never changes.
- **Capitalization controls visibility** (Guide 1, §10) — every identifier you name in this guide (variable, function, type) is a real decision about whether it's visible outside its package. We won't repeat the mechanics, but keep it in mind as you name things.

One new mechanical concept before we start: **every statement in Go ends implicitly without a semicolon in source** — the compiler inserts them automatically based on line breaks (a process governed by a formal but rarely-thought-about rule set). In practice this just means: one statement per line, no trailing `;` — `gofmt` handles this for you and you'll never think about it again after today.

---

## 2. Variable Declarations: `var`, `:=`, and `const`

### 2.1 The Long Form: `var`

```go
var age int = 30
var name string = "Caio"
var isActive bool = true
```

You can also split the declaration from the type when the compiler can infer it:

```go
var age = 30       // inferred as int
var name = "Caio"  // inferred as string
```

Or declare without any initial value at all — this is where **zero values** come in (fully explained in §3):

```go
var age int      // age is 0, not "undefined" or "uninitialized"
var name string  // name is "", not None
```

You can declare multiple variables in one block, which is idiomatic at the top of a function or for grouping related package-level variables:

```go
var (
    host string = "localhost"
    port int    = 8080
    debug bool  = false
)
```

### 2.2 The Short Form: `:=`

Inside a function body (never at package level — this is a hard rule, not a style choice), you'll use the **short variable declaration** operator almost everywhere:

```go
age := 30
name := "Caio"
host, port := "localhost", 8080  // multiple assignment in one line
```

`:=` **declares and infers the type simultaneously**. This is the idiomatic default for local variables in Go — reach for `var` only when: you need a zero value with no initializer, you're declaring at package scope (where `:=` isn't legal), or you want to be explicit about a type the compiler would otherwise infer differently (e.g., `var ratio float64 = 3` where you specifically want `float64`, not the default-inferred `int`).

**A sharp edge to know about now: shadowing.** `:=` creates a *new* variable in the current scope, even if a variable with the same name exists in an outer scope:

```go
x := 10
if true {
    x := 20        // this is a DIFFERENT x, scoped to this if-block
    fmt.Println(x) // 20
}
fmt.Println(x)     // 10 — outer x was never touched
```

This bites newcomers constantly, especially combined with Go's `if err := foo(); err != nil` idiom (§6.1) inside nested blocks — an inner `err :=` can silently shadow an outer `err` you meant to check. Keep this in your peripheral vision; it'll be relevant again in Guide 5 (Error Handling).

### 2.3 `const`

Constants are declared with `const` and **must be assignable at compile time** — no calling a function, no runtime computation:

```go
const MaxRetries = 3
const AppName = "MyService"
```

Unlike Python (which has no true constants — `MAX_RETRIES = 3` at module level is just a variable you *promise* not to reassign), Go's compiler **enforces immutability**:

```go
const MaxRetries = 3
MaxRetries = 5 // compile error: cannot assign to MaxRetries (neither addressable nor a map index expression)
```

We'll return to constants in more depth in §11 once you've seen enough types to make `iota` (Go's enumeration mechanism) make sense.

---

## 3. Zero Values — Go's Answer to `None`

This is a genuinely new concept with no direct Python equivalent, and it's worth sitting with.

In Python, an unassigned variable simply doesn't exist yet (`NameError`), and `None` is an explicit, distinct value you assign on purpose to mean "nothing here." In Go, **every type has a well-defined default value it takes on automatically when declared without an initializer** — there is no concept of an "unset" variable at all.

| Type | Zero Value |
|---|---|
| `int`, `int8`...`int64`, `uint`... | `0` |
| `float32`, `float64` | `0.0` |
| `bool` | `false` |
| `string` | `""` (empty string, **not** `nil`) |
| pointer (`*T`) | `nil` |
| slice (`[]T`) | `nil` |
| map (`map[K]V`) | `nil` |
| channel, function, interface | `nil` |
| struct | every field set to *its own* zero value, recursively |

```go
var count int
var name string
var enabled bool
var scores []int

fmt.Println(count, name, enabled, scores)
// 0 "" false []
```

**Why this matters practically:** in Python, forgetting to initialize something typically surfaces immediately as a loud `NameError` or `AttributeError`. In Go, a forgotten initialization silently becomes a *usable, well-defined* zero value — which is often exactly what you want (an empty slice you're about to `append` to, a struct you're about to fill in field by field) but can also mask a genuine oversight, since the compiler will never complain about "did you mean to set this?" It's a trade-off: fewer crashes from missing initialization, but you must be deliberate about checking whether a zero value is meaningful in context (this becomes especially important with `nil` maps in §9.2 and `nil` slices in §8, which behave subtly differently from an empty-but-initialized one).

This design is also *directly* why Go's error handling works the way it does (previewed in Guide 1, fully covered in Guide 5): a function that "fails" doesn't throw — it returns a zero value for its normal result *alongside* a non-nil `error`, and the zero value is always safe to reference even though it's meaningless in the failure case.

---

## 4. Basic Types in Depth

### 4.1 Numeric Types

Go is far more explicit about numeric width than Python, where `int` is arbitrary-precision and `float` is always a 64-bit double:

| Go Type | Description |
|---|---|
| `int`, `uint` | Platform-dependent width (32 or 64 bit depending on architecture — effectively always 64-bit on modern systems). **Default choice for general-purpose integers.** |
| `int8`/`int16`/`int32`/`int64` | Explicit signed widths |
| `uint8`/`uint16`/`uint32`/`uint64` | Explicit unsigned widths |
| `float32`, `float64` | IEEE-754 floating point. **`float64` is the default choice**, matching Python's `float` precision. |
| `complex64`, `complex128` | Complex numbers (rare outside scientific computing — mentioned for completeness, not something you'll use in API work) |

Two aliases you'll see constantly and should know by name:

- **`byte`** is an alias for `uint8` — used whenever you're working with raw bytes (file I/O, network data, `[]byte` conversions of strings).
- **`rune`** is an alias for `int32` — represents a single Unicode code point. This exists because Go strings are UTF-8 byte sequences under the hood (§4.2), and iterating "characters" correctly requires a type that isn't just a raw byte.

**Critically: Go never implicitly converts between numeric types, even "safe" ones.** This is the single biggest numeric-type surprise coming from Python:

```go
var i int = 10
var f float64 = 3.5
result := i + f // compile error: mismatched types int and float64
```

You must convert explicitly (§5) — this is by design, not an oversight: Go wants every precision-affecting conversion to be a visible, deliberate line of code, never something the compiler papers over silently.

### 4.2 Strings

Go strings are **immutable sequences of bytes**, conventionally holding UTF-8-encoded text — conceptually close to Python 3's `str`, but implemented very differently under the hood (Python 3 strings are sequences of Unicode code points; Go strings are raw byte sequences that are *usually* valid UTF-8 by convention, not enforcement).

```go
s := "Olá, mundo"
fmt.Println(len(s)) // 11 — this is BYTE length, not character count!
```

That `len(s) == 11` for a 10-character string (á is 2 bytes in UTF-8) is the classic gotcha. If you need to count actual characters (runes), convert first:

```go
fmt.Println(len([]rune(s))) // 10 — actual character count
```

Iterating a string with `for range` (§6.3) automatically decodes UTF-8 and gives you runes, not raw bytes — this is usually what you want and mirrors Python's character-based string iteration:

```go
for i, r := range "olá" {
    fmt.Printf("%d: %c\n", i, r)
}
// 0: o
// 1: l
// 2: á   (note: next index will jump by 2, since á takes 2 bytes)
```

String concatenation uses `+`, same as Python:

```go
greeting := "Hello, " + "world"
```

But there's no `f"{name}"` f-string equivalent — Go uses the `fmt` package's `Sprintf`/`Printf` with `%`-style verbs instead (you'll use this constantly):

```go
message := fmt.Sprintf("Hello, %s! You are %d years old.", name, age)
```

### 4.3 Booleans

`bool` with literal values `true`/`false` (lowercase, unlike Python's capitalized `True`/`False`). There is **no truthy/falsy coercion** in Go — an `int`, `string`, or slice can never be used directly where a `bool` is expected:

```python
# Python — this works
if some_list:
    ...
```

```go
// Go — this does NOT compile
if someSlice { ... } // compile error: non-boolean condition in if statement

// You must be explicit:
if len(someSlice) > 0 { ... }
```

This is another instance of Go's "explicit over implicit" philosophy from Guide 1 — every condition must be a genuine boolean expression, never an object standing in for one.

---

## 5. Type Conversion — No Implicit Coercion, Ever

Since Go never auto-converts (§4.1), you convert explicitly using a **type as a function-like conversion**:

```go
var i int = 42
var f float64 = float64(i)  // explicit
var u uint = uint(f)         // explicit

s := strconv.Itoa(i)         // int -> string, NOT via `string(i)` — see warning below
n, err := strconv.Atoi(s)    // string -> int, returns (int, error) — note the error return!
```

**Important warning:** `string(i)` where `i` is an integer does *not* do what a Python-minded developer expects — it converts the integer as a **Unicode code point** (e.g., `string(65)` produces `"A"`, not `"65"`). For actual number-to-string formatting, always use the `strconv` package (`strconv.Itoa`, `strconv.FormatFloat`, etc.) or `fmt.Sprintf`. This specific footgun catches nearly every Go newcomer at least once.

Note also that `strconv.Atoi` returns **two values**: the parsed integer and an `error`. This is your first real taste of Go's multiple-return-value error handling pattern, formally introduced in §7.2 and covered exhaustively in Guide 5. Unlike Python's `int("abc")` raising a `ValueError` you must `try`/`except`, Go simply hands you back an `error` value to check:

```go
n, err := strconv.Atoi("not a number")
if err != nil {
    fmt.Println("conversion failed:", err)
    return
}
fmt.Println(n)
```

---

## 6. Control Flow

### 6.1 `if` / `else` — With an Optional Init Statement

Basic form looks familiar, minus parentheses around the condition (curly braces are mandatory, unlike C):

```go
if age >= 18 {
    fmt.Println("adult")
} else if age >= 13 {
    fmt.Println("teenager")
} else {
    fmt.Println("child")
}
```

The distinctive Go idiom is the **init statement**, which scopes a variable to just the `if`/`else` chain:

```go
if err := doSomething(); err != nil {
    fmt.Println("failed:", err)
    return
}
// err is NOT visible here — it only exists within the if/else block
```

This pattern is everywhere in idiomatic Go, especially paired with error handling (Guide 5) — you'll type `if err :=` and `if err !=` hundreds of times. It keeps error-check variables tightly scoped instead of leaking into the surrounding function body, which matters more than it sounds given Go has no exceptions to bail you out of forgetting to check something.

### 6.2 `for` — The *Only* Looping Construct

Go has no `while`, no `do-while`, and no separate "for-each" keyword — a single `for` keyword covers every looping need via four forms.

**Classic three-clause form** (like C/Python's `range`-based loop):

```go
for i := 0; i < 10; i++ {
    fmt.Println(i)
}
```

**Condition-only form** (this *is* Go's `while`):

```go
count := 0
for count < 5 {
    count++
}
```

**Infinite loop** (this *is* Go's `while True:`):

```go
for {
    // loop forever until an explicit break
    if shouldStop {
        break
    }
}
```

**`for range`** — iterating collections, Go's closest equivalent to Python's `for x in iterable:`:

```go
nums := []int{10, 20, 30}
for index, value := range nums {
    fmt.Println(index, value)
}

// if you only need the value, discard the index with the blank identifier:
for _, value := range nums {
    fmt.Println(value)
}
```

The **blank identifier `_`** is important and recurring: Go requires you to use every declared variable (an unused variable is a compile error — Guide 1, §2.2), so `_` exists specifically as a way to explicitly discard a value you're not interested in, satisfying the compiler while signaling intent clearly.

`for range` also works over maps (§9), strings (§4.2, giving runes), channels (Guide 7), and integers directly in modern Go (`for i := range 5` iterates `0..4`, a newer convenience form similar to Python's `range(5)`).

**No list/dict comprehensions.** Where Python has `[x*2 for x in nums]`, Go requires an explicit loop:

```go
doubled := make([]int, 0, len(nums))
for _, n := range nums {
    doubled = append(doubled, n*2)
}
```

This will feel verbose at first. It's consistent with Guide 1's philosophy: fewer syntactic forms means anyone reading Go code recognizes the pattern instantly, at the cost of typing a few more lines than an equivalent Python comprehension.

### 6.3 `switch` — More Flexible Than You'd Expect

Go's `switch` **does not fall through by default** (the opposite of C, and unlike anything in Python, which has no native switch statement until the very different `match` in 3.10+):

```go
switch day {
case "Mon", "Tue", "Wed", "Thu", "Fri":
    fmt.Println("weekday")
case "Sat", "Sun":
    fmt.Println("weekend")
default:
    fmt.Println("unknown")
}
```

Each `case` automatically breaks after executing — you'd need the explicit `fallthrough` keyword to get C-style behavior, and it's rarely used.

A very idiomatic pattern: **`switch` with no condition at all**, used as a cleaner alternative to a long `if/else if` chain:

```go
switch {
case age < 13:
    fmt.Println("child")
case age < 20:
    fmt.Println("teenager")
default:
    fmt.Println("adult")
}
```

A preview (fully explored in Guide 4): Go also has a **type switch**, used to branch on the concrete type stored in an interface value — syntactically similar but conceptually distinct, so just note the name for now:

```go
switch v := someInterfaceValue.(type) {
case int:
    fmt.Println("it's an int:", v)
case string:
    fmt.Println("it's a string:", v)
}
```

### 6.4 No Ternary Operator

Python's `x if condition else y` has **no equivalent** in Go — there is deliberately no ternary operator. You write a full `if/else` (often assigning inside each branch), or extract a tiny named function if you find yourself wanting one repeatedly. This is a conscious simplicity choice from the language designers, not a missing feature.

---

## 7. Functions in Depth

### 7.1 Basic Syntax

```go
func Add(a int, b int) int {
    return a + b
}

// consecutive parameters of the same type can share a single type annotation:
func Add(a, b int) int {
    return a + b
}
```

### 7.2 Multiple Return Values — The Foundation of Go's Error Handling

This is the single most important functional concept in the entire language, and everything in Guide 5 builds directly on it. Go functions can return **more than one value**, and by convention, a final `error` return value signals "did this succeed":

```go
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := Divide(10, 0)
if err != nil {
    fmt.Println("error:", err)
    return
}
fmt.Println("result:", result)
```

Contrast directly with Python: where you'd `raise ValueError(...)` and rely on a `try`/`except` block somewhere up the call stack to catch it, Go simply **returns the error as an ordinary value**, and the caller is responsible for checking it immediately, every single time, right at the call site. Nothing propagates automatically. We'll go far deeper on *why* this is preferred and how to compose/wrap errors properly in Guide 5 — for now, just internalize the mechanical shape: `(value, error)` return tuples, checked with `if err != nil`.

### 7.3 Named Return Values

Go allows you to name your return values directly in the function signature, which both documents intent and lets you use a bare `return` (returning whatever the named variables currently hold):

```go
func Divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = errors.New("division by zero")
        return // returns (0, err) — result stays at its zero value
    }
    result = a / b
    return // returns (result, nil)
}
```

This is a stylistic tool, not a requirement — you'll see both styles in real code. Named returns tend to be favored for functions with several returns of the same type, or in short functions where the names add real clarity; unnamed returns are preferred when the function body is long enough that a bare `return` would obscure what's actually being sent back.

### 7.4 Variadic Functions

Go's equivalent of Python's `*args` (there's no direct `**kwargs` equivalent — Go has no native keyword-argument concept at all, a consequence of static typing and no default parameter values):

```go
func Sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

Sum(1, 2, 3)       // 6
Sum()               // 0

existing := []int{4, 5, 6}
Sum(existing...)   // spread a slice into variadic args, similar to Python's Sum(*existing)
```

Note: Go has **no default parameter values** and **no keyword arguments** at all — every call must supply every non-variadic parameter positionally. The idiomatic workaround for "optional configuration" (something you're used to leaning on heavily in Python function signatures) is the **functional options pattern**, which you'll meet properly once you're building real services in Guide 12.

### 7.5 Functions as First-Class Values, and Closures

Functions are values in Go, just like in Python — you can assign them to variables, pass them as arguments, and return them from other functions:

```go
func makeMultiplier(factor int) func(int) int {
    return func(n int) int {
        return n * factor
    }
}

double := makeMultiplier(2)
fmt.Println(double(5)) // 10
```

This is a **closure** — the returned anonymous function captures `factor` from its enclosing scope, exactly like Python closures behave. The type of `double` here, `func(int) int`, is itself a legitimate Go type — function types are written out explicitly, another instance of Go's static typing reaching into places Python leaves implicit.

---

## 8. Arrays vs Slices — The Most Important Section in This Guide

Python has one general-purpose ordered, mutable sequence type: `list`. Go has **two**, and conflating them is the single most common source of confusion for newcomers.

### 8.1 Arrays — Fixed Size, Rarely Used Directly

```go
var scores [5]int              // an array of exactly 5 ints, all zero-valued
scores[0] = 90

primes := [3]int{2, 3, 5}       // array literal, size fixed at 3 forever
```

The size **is part of the type** — `[5]int` and `[3]int` are different, incompatible types, not "arrays of different lengths." Arrays are **value types**: assigning one array to another, or passing one to a function, **copies the entire array**:

```go
a := [3]int{1, 2, 3}
b := a          // b is a completely independent COPY of a
b[0] = 999
fmt.Println(a)  // [1 2 3] — a is untouched
```

In practice, **you will almost never use a raw array directly** in idiomatic Go application code — they exist mostly as the underlying mechanism that slices are built on top of. Nearly everywhere you'd reach for a Python `list`, you'll reach for a **slice** instead.

### 8.2 Slices — Your Actual Go "List"

A slice is a small struct-like header containing three things: a **pointer** to an underlying array, a **length**, and a **capacity**. This is the data structure you'll use constantly:

```go
scores := []int{90, 85, 77}   // note: no size in the brackets — this is a slice, not an array
scores = append(scores, 100)   // append returns a (possibly new) slice — always reassign!
```

**Critical idiom:** `append` **must always be reassigned** to a variable (usually the same one), because `append` may or may not return the same underlying array:

```go
s := []int{1, 2, 3}
s = append(s, 4)  // correct
append(s, 4)      // WRONG — return value discarded, s is unchanged; also likely a `go vet` warning
```

### 8.3 Length vs Capacity — Why `append` Sometimes Reallocates

Every slice has:

- **`len(s)`** — how many elements it currently holds.
- **`cap(s)`** — how many elements the underlying array *could* hold before a new, larger array must be allocated.

```go
s := make([]int, 3, 5)  // length 3, capacity 5
fmt.Println(len(s), cap(s)) // 3 5

s = append(s, 10) // still fits in existing capacity — no reallocation, len becomes 4
s = append(s, 20) // still fits — len becomes 5, now at capacity
s = append(s, 30) // capacity exceeded — Go allocates a NEW, larger underlying array and copies everything over
```

This matters practically: if you know roughly how many elements you'll end up with, pre-allocating capacity via `make([]T, 0, expectedSize)` avoids repeated reallocation-and-copy cycles — a small but real performance idiom you'll see in real Go code, roughly analogous to knowing you *could* pre-size a Python list but rarely bother because CPython's list growth is cheap enough that nobody thinks about it.

### 8.4 Slicing Syntax

```go
nums := []int{0, 1, 2, 3, 4, 5}
sub := nums[1:4]   // [1 2 3] — same half-open range semantics as Python's nums[1:4]
sub2 := nums[:3]    // [0 1 2]
sub3 := nums[3:]    // [3 4 5]
```

This part will feel comfortably familiar — Go's slice syntax deliberately mirrors Python's.

### 8.5 The Sharp Edge: Slices Share Underlying Memory

This is the single biggest slice gotcha, and it has no equivalent warning needed in Python (where slicing a list always produces a genuinely independent copy):

```go
original := []int{1, 2, 3, 4, 5}
sub := original[1:3]   // sub is [2 3], but shares the SAME underlying array as original

sub[0] = 999
fmt.Println(original)  // [1 999 3 4 5] — original was mutated through sub!
```

Because a slice is just a *view* (pointer + length + capacity) into a shared underlying array, mutating elements through one slice can silently affect another slice that overlaps the same backing array. This becomes especially important once you pass slices into functions (they're passed "by reference" in effect, since the slice header is copied but it still points at the same underlying array) — a function that mutates elements of a slice parameter mutates the caller's data too, similar in *effect* to Python passing a mutable list into a function, but arising from a different mechanism (slice headers vs. Python's object-reference model).

If you need a genuinely independent copy, use the built-in `copy` function:

```go
dst := make([]int, len(original))
copy(dst, original)
```

### 8.6 Nil Slices vs Empty Slices

Recall §3 (zero values): the zero value of a slice is `nil`, not an empty-but-allocated slice:

```go
var s []int          // s is nil
fmt.Println(s == nil)     // true
fmt.Println(len(s))       // 0 — len() is always safe to call, even on nil

s2 := []int{}         // s2 is NOT nil — it's an empty, allocated slice
fmt.Println(s2 == nil)    // false
```

In practice, both behave identically for reading (`len`, `range`) and for `append` (appending to a nil slice works fine and allocates on first use) — the distinction mostly matters for equality checks (`== nil`) and certain JSON marshaling edge cases you'll hit directly in Guide 9.

---

## 9. Maps

Go's `map[K]V` is conceptually close to Python's `dict`, with a few important behavioral differences.

### 9.1 Declaration and Basic Use

```go
ages := map[string]int{
    "Alice": 30,
    "Bob":   25,
}

ages["Carol"] = 40         // add/update
fmt.Println(ages["Alice"]) // 30
delete(ages, "Bob")         // remove a key
```

### 9.2 The Nil Map Trap

Just like slices, a map's zero value is `nil` — but unlike slices, **writing to a nil map panics at runtime**:

```go
var ages map[string]int  // ages is nil
ages["Alice"] = 30        // PANIC: assignment to entry in nil map
```

Reading from a nil map is perfectly safe and returns the value type's zero value — only *writing* panics. Always initialize maps you intend to write to, either via a literal (`map[string]int{}`) or `make`:

```go
ages := make(map[string]int)
ages["Alice"] = 30 // fine
```

This is a genuinely common beginner bug, precisely because *reading* from a nil map works fine, so the bug often doesn't surface until the first write happens on some rarely-hit code path.

### 9.3 The "Comma-Ok" Idiom

Python distinguishes "key missing" from "key present with a falsy value" via `dict.get(key, default)` or a `KeyError` on plain indexing. Go uses a distinctive **two-value return from a map index expression**:

```go
value, ok := ages["Dave"]
if !ok {
    fmt.Println("Dave not found")
}
```

If the key doesn't exist, `value` is the zero value for the map's value type (`0` for `int` here) and `ok` is `false`. If you index a map with only one return value (`value := ages["Dave"]`) and the key is missing, you silently get the zero value with **no way to distinguish "missing" from "present but zero"** — so the comma-ok form is the idiomatic, safe default whenever that distinction matters. You'll see this exact `value, ok :=` shape again in Guide 4 for type assertions, and in Guide 7 for reading from channels — it's a recurring Go pattern, not a map-specific quirk.

### 9.4 Iteration Order Is Randomized — On Purpose

Since Go 1.0, iterating a map with `for range` deliberately **randomizes the order** on every single run, specifically to prevent developers from ever accidentally relying on an iteration order that was never guaranteed in the first place:

```go
for key, value := range ages {
    fmt.Println(key, value) // order is different every time you run this
}
```

This is stricter than Python, where dicts have guaranteed insertion-order iteration since 3.7 — if you need ordered map iteration in Go, you must sort the keys yourself (typically: extract keys into a slice, sort with `sort.Strings`/`sort.Slice`, then iterate that sorted slice).

---

## 10. Packages & Imports: The Mechanics

Guide 1 (§9–10) covered *what* packages are and *why* capitalization controls visibility. Here's the practical mechanics of working with them day to day.

### 10.1 Multiple Files, One Package

Unlike Python (where every `.py` file is automatically its own importable module), **a Go package is an entire directory**, and every `.go` file in that directory must declare the same `package` name and shares the same scope — meaning a function defined in `handler_users.go` can call an unexported function defined in `handler_orders.go` directly, no import needed, as long as both files declare `package handler`.

### 10.2 Import Syntax

```go
import "fmt"

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"    // third-party import — blank line groups them by convention
)
```

`gofmt`/`goimports` conventionally groups standard-library imports separately from third-party ones, separated by a blank line — this is a formatting convention the tooling enforces automatically, not something you need to manage by hand.

### 10.3 Import Aliasing

```go
import (
    m "math"
)

m.Sqrt(16)
```

Rare in normal code, but useful when two imported packages would otherwise share the same default name.

### 10.4 The Blank Import: `_`

```go
import (
    _ "github.com/lib/pq" // imported purely for its side effects (registers a database driver), never referenced by name
)
```

Recall the blank identifier from §6.2 — here it's used at the import level to satisfy the compiler's "every import must be used" rule (Guide 1, §2.2) when you genuinely only need a package's `init()` side effects (§10.5), not any of its exported names directly. You'll see this exact pattern in Guide 6 when registering SQL database drivers.

### 10.5 `init()` — Automatic Setup Before `main`

Any package (not just `main`) can define one or more `init()` functions, which run automatically — in dependency order across all imported packages — before `main()` executes, with no explicit call needed:

```go
func init() {
    fmt.Println("this runs before main(), automatically")
}
```

There's no direct Python equivalent (top-level module code runs on import, which is *similar* in spirit, but Python has no dedicated, explicitly-named hook function for it). `init()` is used sparingly in idiomatic modern Go — mostly for package-level setup like registering a driver or validating a configuration invariant — since overuse makes program startup behavior harder to trace than explicit initialization in `main()`.

---

## 11. Constants and `iota`

### 11.1 Typed vs Untyped Constants

Recall `const` from §2.3. Go constants can be **untyped** (flexible, adapting to context) or explicitly typed:

```go
const Pi = 3.14159        // untyped constant — can be used as float32, float64, etc. as needed
const MaxUsers int = 100  // explicitly typed as int
```

### 11.2 `iota` — Go's Enum Mechanism

Go has no native `enum` keyword (unlike Python's `enum.Enum`). Instead, the idiomatic pattern combines `const` blocks with the special `iota` identifier, which **auto-increments starting at 0** within a `const` block:

```go
type Weekday int

const (
    Sunday Weekday = iota // 0
    Monday                 // 1 (implicitly repeats "= iota" from the line above)
    Tuesday                 // 2
    Wednesday               // 3
    Thursday                // 4
    Friday                   // 5
    Saturday                 // 6
)
```

This is meaningfully different from Python's `enum.Enum`, which is a genuine class with named members you access as `Weekday.MONDAY`. Go's `iota`-based constants are just typed integers with names — there's no built-in `.name`/`.value` introspection unless you write it yourself (commonly, a `String() string` method on the type, which you'll be well-equipped to write once Guide 3 covers methods on custom types).

`iota` also supports more advanced patterns (skipping values, bit-shifting for flag-style constants), but the simple enumeration form above is what you'll reach for constantly, including later when modeling things like HTTP-adjacent status categories or application-specific state machines in your REST API work.

---

## 12. Summary & What's Next

**What you can now actually write:**

- Declare variables with `var` and `:=`, understand when each is idiomatic, and recognize the shadowing trap.
- Reason correctly about zero values instead of expecting `None`/`NameError` behavior.
- Work confidently with Go's explicit numeric types and mandatory type conversions, including the `string(int)` footgun.
- Write every form of Go's control flow: `if` with init statements, all four `for` forms, `switch` (including the no-condition and type-switch variants).
- Write functions with multiple return values and named returns — the exact mechanism Guide 5's error handling is built on.
- Distinguish arrays from slices, understand `len`/`cap` and why `append` sometimes reallocates, and recognize when two slices silently share memory.
- Use maps correctly, including the comma-ok idiom and the nil-map-write panic trap.
- Understand how multi-file packages, imports, blank imports, and `init()` fit together mechanically.
- Build simple enumerations with `const` + `iota`.

**What comes next — Guide 3: Structs, Methods & Composition.** You now have Go's *data* (variables, slices, maps) and *control flow* fully in hand, but everything so far has been built-in types only. Guide 3 introduces **structs** — Go's answer to a Python class's data — along with **methods** (functions attached to a type), **value vs. pointer receivers** (a distinction with real consequences given what you just learned about copying in §8.1), and **struct embedding**, which is how Go achieves code reuse and "is-a" style relationships entirely without inheritance. This is where Go starts looking less like "C with slices" and more like the object-modeling tool you'll actually build your REST API's domain layer with.
