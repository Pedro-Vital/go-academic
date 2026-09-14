# Supplementary Guide: Generics

**Roadmap position:** Supplementary — not numbered in the main sequence. Best read **after Guide 4 (Interfaces & Polymorphism)**, since Go's generic constraints *are* interfaces, and constraints won't make sense without that foundation. Also useful to revisit before **Guide 12 (Architecture & Dependency Injection)** and **Guide 13 (Database Integration)**, where generic repository patterns show up.
**Prerequisites:** Guide 2 (types), Guide 4 (interfaces) — constraints are interfaces under the hood.

## Why This Guide Exists (and Why It's Supplementary, Not Numbered)

Back at Guide 4, generics were deliberately deferred: interfaces are a prerequisite for understanding constraints, and error handling/concurrency were more directly on the path to the REST API goal. That was the right call for *pacing*, but it left a real gap — a lot of real-world Go code (standard library `slices`/`maps` packages, ORMs, generic repository patterns) leans on generics, and you'll hit unfamiliar `[T any]` syntax reading other people's Go before you'd naturally reach it in the numbered sequence.

This guide fills that gap on its own track. It doesn't block Phase E or F — you can read Guides 10–14 without it — but it *will* make some patterns in Guide 12/13 (a generic `Repository[T]`) click faster, and it'll stop `[T any]` syntax from being a stumbling block the first time you see it in someone else's code.

---

## Table of Contents

1. [The Problem Generics Solve](#1-the-problem-generics-solve)
2. [Mental Model: Go Generics vs. Python Typing](#2-mental-model-go-generics-vs-python-typing)
3. [Type Parameters — Syntax Basics](#3-type-parameters--syntax-basics)
4. [Constraints — `any`, `comparable`, and Custom Interfaces](#4-constraints--any-comparable-and-custom-interfaces)
5. [Generic Functions](#5-generic-functions)
6. [Type Inference — When You Can Drop `[T]`](#6-type-inference--when-you-can-drop-t)
7. [Generic Types (Structs)](#7-generic-types-structs)
8. [Generic Methods — The Receiver Limitation](#8-generic-methods--the-receiver-limitation)
9. [Multiple Type Parameters](#9-multiple-type-parameters)
10. [The `cmp` and `slices`/`maps` Standard Packages](#10-the-cmp-and-slicesmaps-standard-packages)
11. [Generics + Interfaces — How They Compose](#11-generics--interfaces--how-they-compose)
12. [Common Patterns for REST APIs](#12-common-patterns-for-rest-apis)
13. [Gotchas & Footguns](#13-gotchas--footguns)
14. [When *Not* to Use Generics](#14-when-not-to-use-generics)
15. [Exercises](#15-exercises)
16. [Key Takeaways](#16-key-takeaways)

---

## 1. The Problem Generics Solve

Before Go 1.18 (generics landed in March 2022), if you wanted a `Max` function that worked for both `int` and `float64`, you had two bad options: write it twice, or use `interface{}` and lose type safety.

```go
// Option A: duplication
func MaxInt(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func MaxFloat64(a, b float64) float64 {
    if a > b {
        return a
    }
    return b
}
```

```go
// Option B: interface{} — compiles, but unsafe and clunky
func Max(a, b interface{}) interface{} {
    // now you need a type switch just to compare them...
    switch a := a.(type) {
    case int:
        b := b.(int) // panics if b isn't also an int — no compile-time safety at all
        if a > b {
            return a
        }
        return b
    case float64:
        // ... repeat
    }
    return nil
}
```

Generics give you a third option: write the logic **once**, parameterized over a type, with the compiler still checking everything at compile time:

```go
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

Max(3, 5)       // T inferred as int
Max(3.1, 2.9)   // T inferred as float64
Max("a", "b")   // T inferred as string
```

One function body, full type safety, no runtime type assertions, no `interface{}` boxing.

---

## 2. Mental Model: Go Generics vs. Python Typing

This is the section where the Python-comparison approach needs the most care, because **Python's typing system and Go's generics solve different problems.**

Python is dynamically typed at runtime — `def max_val(a, b): return a if a > b else b` already works for `int`, `float`, `str`, anything with `__gt__`, with **zero type annotations**, because Python dispatches on the actual runtime type of `a` and `b`. `TypeVar` and `Generic[T]` exist purely to give **static type checkers** (mypy, pyright) something to check — they have **no effect at runtime** and are erased entirely once the code executes.

Go is statically and strictly typed — there is no dynamic dispatch on arbitrary types the way Python has. Before generics, Go's *only* way to write one function that handles multiple types was `interface{}` (Option B above), which sacrifices compile-time type checking. Generics in Go exist to let the **compiler** (not just an optional external linter) verify type safety while still writing the logic once.

| Concept | Python | Go |
|---|---|---|
| Why it exists | Static type *checking* is optional and external (mypy); the language itself never needed generics to be polymorphic | Static type checking is built into the compiler; generics are the *only* way to get compile-time-checked polymorphism without `interface{}` |
| Runtime effect | None — `TypeVar`/`Generic` are erased; Python dispatches dynamically regardless | Real — the compiler generates/shares code per concrete type (a process closer to C++ templates than to Java's type erasure) |
| Syntax | `T = TypeVar("T")`, `class Stack(Generic[T])` | `[T any]` on functions/types |
| Constraining a type param | `T = TypeVar("T", bound=Comparable)` or a `Protocol` | A constraint interface: `[T cmp.Ordered]`, `[T MyConstraint]` |
| "Any type is fine" | Just don't annotate, or use `Any` | `[T any]` — `any` is just an alias for `interface{}` used as a constraint |
| Do you *need* it for basic polymorphism? | No — duck typing already gives you that | Yes — without generics, `interface{}` + type assertions is your only other option |

The honest takeaway: **generics matter more in Go than in Python** precisely *because* Go doesn't have duck typing at the language level. In Python, "does this object support `+`?" is a runtime question resolved via `__add__`. In Go, "does this type support `>`?" has to be resolved at compile time, either through a constraint (generics) or by giving up type safety (`interface{}`).

---

## 3. Type Parameters — Syntax Basics

A **type parameter list** goes in square brackets, right after the function or type name:

```go
func Identity[T any](v T) T {
    return v
}

x := Identity(42)      // T = int, inferred
y := Identity("hello") // T = string, inferred
```

`T` is a placeholder — a **type parameter** — that stands in for whatever concrete type is used at the call site (called a **type argument**). You can name it anything (`T`, `V`, `K`, `Item` — single uppercase letters are conventional, mirroring `T` in C++/Java/TypeScript, but not mandatory).

You can explicitly instantiate instead of relying on inference:

```go
y := Identity[string]("hello") // explicit type argument — rarely needed, inference usually suffices
```

---

## 4. Constraints — `any`, `comparable`, and Custom Interfaces

A **constraint** is what goes inside the brackets after the type parameter name — it's an interface that restricts which types are allowed, and which operations are legal on values of type `T` inside the function body.

### `any` — no constraint at all

```go
func Identity[T any](v T) T { return v }
```

`any` is a built-in alias for `interface{}` (added in Go 1.18 specifically to read better in this position). It means "literally any type" — but consequently, **you can't do anything type-specific with a `T` constrained only by `any`** — no `+`, no `>`, no field access — because the compiler has no guarantee what operations `T` supports.

### `comparable` — supports `==` and `!=`

```go
func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target { // only legal because T is constrained to comparable
            return true
        }
    }
    return false
}

Contains([]int{1, 2, 3}, 2)          // true
Contains([]string{"a", "b"}, "c")    // false
```

`comparable` is another built-in constraint, satisfied by any type where `==`/`!=` are valid — numbers, strings, booleans, pointers, structs made entirely of comparable fields. Slices, maps, and functions are **not** comparable in Go and don't satisfy this constraint (§13 gotcha).

### Custom constraints — interfaces with a type set

```go
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
        ~float32 | ~float64 | ~string
}

func Max[T Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

This is a constraint expressed as a **union of types** inside an interface — new syntax specific to generics, not something you'd write for an ordinary Guide-4-style interface. The `~` prefix means "this type, **or any type whose underlying type is this**" — so `~int` matches not just `int` but also `type UserID int`. Without `~`, the constraint would only match the exact named type `int`, rejecting your own named types built on top of it.

You don't have to write `Ordered` from scratch — it's exactly what the standard library's `cmp.Ordered` (§10) already provides.

| Concept | Python | Go |
|---|---|---|
| "Any type" | No annotation, or `Any` | `[T any]` |
| "Must support equality" | Implicit — everything supports `==` via `__eq__`, or a `Protocol` requiring `__eq__` | `[T comparable]` |
| "Must be one of these types" | `Union[int, float, str]` (a type hint, checker-only) | A constraint interface with a `|`-separated type set (compiler-enforced) |
| "Must support this behavior" | `Protocol` with the required methods | A constraint interface with a method set (same as Guide 4's ordinary interfaces) |

---

## 5. Generic Functions

Beyond `Max` and `Contains`, generic functions are where most of the immediate, practical value shows up — utility functions that work across your whole codebase regardless of the element type:

```go
func Map[T, U any](items []T, f func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = f(item)
    }
    return result
}

func Filter[T any](items []T, predicate func(T) bool) []T {
    var result []T
    for _, item := range items {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}

func Reduce[T, U any](items []T, initial U, f func(U, T) U) U {
    acc := initial
    for _, item := range items {
        acc = f(acc, item)
    }
    return acc
}
```

Usage:

```go
nums := []int{1, 2, 3, 4, 5}

doubled := Map(nums, func(n int) int { return n * 2 })          // []int{2,4,6,8,10}
evens := Filter(nums, func(n int) bool { return n%2 == 0 })       // []int{2,4}
sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })   // 15

names := Map(nums, func(n int) string { return fmt.Sprintf("#%d", n) }) // []string, T=int, U=string
```

**Python comparison:** `Map`/`Filter`/`Reduce` are Go's answer to Python's built-in `map()`, `filter()`, and `functools.reduce()` — except Python's versions work on *any* iterable with zero type ceremony (duck typing again), while Go's need the `[T, U any]` type parameter list precisely because Go can't just "try it and see if it works at runtime."

---

## 6. Type Inference — When You Can Drop `[T]`

Go can usually infer type arguments from the function's regular arguments, so you rarely write `Max[int](3, 5)` explicitly — `Max(3, 5)` is enough. Inference works when the type parameter appears in a regular parameter's type. It does **not** work when the type parameter *only* appears in the return type:

```go
func Zero[T any]() T {
    var zero T
    return zero
}

x := Zero[int]()    // MUST specify explicitly — nothing to infer from
// x := Zero()      // compile error: cannot infer T
```

Rule of thumb: if `T` shows up in an argument, Go infers it from what you pass in; if `T` only shows up in the return type (or isn't tied to any argument), you must instantiate explicitly with `[Type]`.

---

## 7. Generic Types (Structs)

Type parameters aren't limited to functions — structs (and other type declarations) can carry them too, which is how you build reusable generic containers:

```go
type Stack[T any] struct {
    items []T
}

func NewStack[T any]() *Stack[T] {
    return &Stack[T]{items: make([]T, 0)}
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    var zero T
    if len(s.items) == 0 {
        return zero, false
    }
    n := len(s.items) - 1
    item := s.items[n]
    s.items = s.items[:n]
    return item, true
}

func (s *Stack[T]) Len() int {
    return len(s.items)
}
```

Usage:

```go
intStack := NewStack[int]()
intStack.Push(1)
intStack.Push(2)
v, ok := intStack.Pop() // v = 2, ok = true

userStack := NewStack[User]()
userStack.Push(User{Name: "Ada"})
```

Note `var zero T` in `Pop` — this is Go's idiom for "give me the zero value of whatever T is" (0 for numbers, `""` for strings, `nil` for pointers/slices/maps/interfaces, a zero-valued struct for struct types). You can't write a literal like `0` or `nil` directly because you don't know which of those is valid for an arbitrary `T` — `var zero T` always works, for any constraint.

**Python comparison:** `Stack[T]` corresponds to `class Stack(Generic[T])` with a `list[T]` field internally — but again, Python's version is a hint for mypy only; you could build and use a Python `Stack` without `Generic` at all and it'd behave identically at runtime, just without static checking. In Go, `Stack[T]` genuinely changes what the compiler allows — pushing a `string` onto an `intStack` is a compile error, not a runtime one.

---

## 8. Generic Methods — The Receiver Limitation

This is a rule that trips people up: **a method can use the type parameters already declared on its receiver's type, but it cannot introduce brand-new type parameters of its own.**

```go
type Container[T any] struct {
    value T
}

// OK — reusing T from the receiver
func (c Container[T]) Get() T {
    return c.value
}

// NOT ALLOWED — methods cannot declare their own new type parameters
// func (c Container[T]) Convert[U any]() U { ... }  // compile error
```

If you need a genuinely new type parameter for a single operation (like the `Map` example converting `T` to `U`), it has to be a **free function**, not a method:

```go
// Free function — allowed, because U is fresh, not tied to a method receiver
func ConvertContainer[T, U any](c Container[T], f func(T) U) Container[U] {
    return Container[U]{value: f(c.value)}
}
```

This is a genuine language limitation (not present in, say, C++ templates or Java generics), and it's the single most common "wait, why won't this compile" moment when people start writing generic types with transform-style methods.

---

## 9. Multiple Type Parameters

Functions and types can take more than one type parameter, each with its own (possibly different) constraint:

```go
type Pair[K comparable, V any] struct {
    Key   K
    Value V
}

func NewPair[K comparable, V any](key K, value V) Pair[K, V] {
    return Pair[K, V]{Key: key, Value: value}
}

func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
    groups := make(map[K][]T)
    for _, item := range items {
        k := keyFn(item)
        groups[k] = append(groups[k], item)
    }
    return groups
}
```

```go
type User struct {
    Name string
    Age  int
}

users := []User{{"Ada", 30}, {"Grace", 30}, {"Alan", 25}}
byAge := GroupBy(users, func(u User) int { return u.Age })
// map[int][]User{30: [{Ada 30} {Grace 30}], 25: [{Alan 25}]}
```

**Python comparison:** `Pair[K, V]` mirrors `dict`-like pair typing such as `tuple[K, V]` with `TypeVar` bounds in a type-checked codebase; `GroupBy` mirrors `itertools.groupby` (though Python's version requires pre-sorted input and returns an iterator of groups, whereas this Go version builds a full `map` directly — a meaningfully different tradeoff, not just a syntax difference).

---

## 10. The `cmp` and `slices`/`maps` Standard Packages

Since Go 1.21, the standard library ships generic utilities so you rarely need to write `Max`, `Contains`, or `Map`/`Filter` yourself for common cases:

```go
import (
    "cmp"
    "slices"
    "maps"
)

cmp.Compare(3, 5)         // -1
cmp.Less(3, 5)            // true
max := max(3, 5)          // built-in generic function since Go 1.21! (also `min`)

slices.Contains([]int{1, 2, 3}, 2)   // true
slices.Sort(mySlice)                  // in-place, works for any cmp.Ordered element type
slices.Index([]string{"a", "b"}, "b") // 1
slices.Max([]int{3, 1, 4, 1, 5})      // 5

keys := maps.Keys(myMap)   // iterator over keys (Go 1.23+: range-over-func iterator)
```

`cmp.Ordered` is the standard library's version of the `Ordered` constraint hand-written in §4 — you should use `cmp.Ordered` in your own code rather than redefining it, exactly the way you'd reach for a well-known Pydantic/typing utility instead of hand-rolling your own.

| Concept | Python | Go |
|---|---|---|
| `max()`/`min()` over any comparable pair | Built-in `max()`/`min()`, works via `__gt__`/`__lt__` | Built-in generic `max()`/`min()` (Go 1.21+), works via `cmp.Ordered` |
| Sort a list | `list.sort()` / `sorted()` | `slices.Sort(s)` — generic, replaces the older `sort.Ints`/`sort.Strings`/`sort.Slice` zoo |
| Check membership | `x in some_list` | `slices.Contains(s, x)` |
| Dict keys as a list | `list(d.keys())` | `slices.Collect(maps.Keys(m))` (Go 1.23+) |

---

## 11. Generics + Interfaces — How They Compose

Constraints *are* interfaces (§4), which means everything from Guide 4 about interfaces applies to constraint design too — including "accept interfaces, return structs." A subtlety worth naming explicitly: a **constraint interface** (used only in `[T Constraint]` position, defining what operations `T` supports) is a different *use* of the interface mechanism than an **ordinary interface** (used as a variable/parameter type, for polymorphic dispatch at runtime) — even though both are declared with the same `type X interface { ... }` syntax.

```go
// Ordinary interface — used for runtime polymorphism (Guide 4 style)
type Stringer interface {
    String() string
}

// Constraint interface — used only inside [T ...], never as a variable type
type Numeric interface {
    ~int | ~int64 | ~float64
}

// A generic function CAN combine both roles:
func Describe[T Numeric](values []T) string {
    var sb strings.Builder
    for _, v := range values {
        sb.WriteString(fmt.Sprintf("%v ", v))
    }
    return sb.String()
}

// You can also constrain T by requiring it implement an ordinary interface:
func PrintAll[T fmt.Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}
```

`PrintAll` above shows constraints and ordinary interfaces meeting directly: `T` is constrained to "any type implementing `fmt.Stringer`," which means inside the function body you can call `.String()` on a `T` — something you couldn't do with a bare `[T any]`.

---

## 12. Common Patterns for REST APIs

These are the patterns most likely to actually show up once you reach Guide 12/13 — worth previewing now so they aren't a surprise.

### Generic repository

```go
type Repository[T any] interface {
    GetByID(ctx context.Context, id string) (T, error)
    Create(ctx context.Context, item T) error
    Update(ctx context.Context, item T) error
    Delete(ctx context.Context, id string) error
}

type InMemoryRepository[T any] struct {
    mu    sync.RWMutex
    items map[string]T
}

func NewInMemoryRepository[T any]() *InMemoryRepository[T] {
    return &InMemoryRepository[T]{items: make(map[string]T)}
}

func (r *InMemoryRepository[T]) GetByID(ctx context.Context, id string) (T, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    item, ok := r.items[id]
    if !ok {
        var zero T
        return zero, fmt.Errorf("item %s not found", id)
    }
    return item, nil
}
// ... Create/Update/Delete follow the same shape
```

This gives you one repository implementation shared across `User`, `Product`, `Order`, etc., instead of hand-writing near-identical CRUD boilerplate per entity — directly relevant to Guide 13's repository-pattern coverage.

### Generic `Result`-style wrapper (an `Optional`/`Result` type Go doesn't have built in)

```go
type Optional[T any] struct {
    value T
    valid bool
}

func Some[T any](v T) Optional[T] { return Optional[T]{value: v, valid: true} }
func None[T any]() Optional[T]    { return Optional[T]{} } // valid stays false

func (o Optional[T]) Get() (T, bool) {
    return o.value, o.valid
}
```

**Python comparison:** this is Go's hand-built version of `Optional[T]` from `typing`, though Python's `Optional[T]` is really just `Union[T, None]` — a type-checker hint, with `None` as an actual runtime value already baked into the language. Go has no built-in "nullable" concept for non-pointer/interface/slice/map types (an `int` can never be `nil`), so a wrapper like `Optional[T]` is a genuinely useful pattern, not just a stylistic mirror.

---

## 13. Gotchas & Footguns

**1. `comparable` doesn't mean "has an `Equals` method" — it means the built-in `==` works.** A struct containing a slice or map field is **not** `comparable`, even though you might want to compare it field-by-field:
```go
type Bad struct {
    Tags []string // slices are never comparable
}
// func Contains[T comparable](s []T, v T) bool { ... }
// Contains([]Bad{...}, someBad) // compile error: Bad does not implement comparable
```

**2. `~T` in a constraint means "underlying type T," not "type T."** Forgetting the `~` silently narrows your constraint to reject perfectly reasonable named types:
```go
type Constraint interface {
    int // exact match only — rejects `type UserID int`
}
type Constraint2 interface {
    ~int // matches int AND any named type based on int, like UserID
}
```

**3. Methods cannot introduce new type parameters (§8)** — this isn't a bug, it's a deliberate language restriction, and reaching for a free function instead is the correct fix, not a workaround.

**4. Generics are not a substitute for interfaces when you need runtime polymorphism.** `[T any]` is resolved at *compile time* — the compiler needs to know the concrete type at each call site (or infer it). If you need a slice holding a mix of different concrete types decided at runtime (`[]Shape` holding both `Circle` and `Square` values), that's Guide 4's ordinary interfaces, not generics — the two solve different problems and aren't interchangeable.

**5. Over-parameterizing hurts readability.** A function with three type parameters and a gnarly constraint can be harder to read than the small amount of duplication it saves. Idiomatic Go leans toward writing the concrete version first, and only reaching for generics once you actually have two or more real call sites that would otherwise duplicate logic (see §14).

**6. Type inference failures produce notoriously unhelpful error messages** in older Go versions (pre-1.21) when multiple type parameters interact in a function signature — if inference fails unexpectedly, try specifying type arguments explicitly (`Func[int, string](...)`) to narrow down which parameter Go couldn't infer.

---

## 14. When *Not* to Use Generics

Idiomatic Go culture is notably more cautious about reaching for generics than Python culture is about reaching for `TypeVar`, or than TypeScript culture is about generic utility types. The community norm, echoed directly by the Go team, is roughly: **write the concrete version first; only generalize once duplication actually hurts.**

Prefer a plain interface over introducing a type parameter when:
- You need runtime polymorphism (a slice of mixed concrete types) — that's what interfaces are *for* (Guide 4).
- You only have **one** concrete type today — a generic `Repository[T]` with a single instantiation (`Repository[User]`) is premature; just write `UserRepository` and generalize later if `ProductRepository` turns out to need identical logic.
- The constraint would end up being `any` with no real operations used on `T` inside the function — if you're not actually calling any type-specific method or operator on `T`, you probably don't need `T` to be generic at all; `interface{}`/`any` as a plain (non-generic) parameter type may be simpler.

This is a genuinely different cultural default from Python, where adding `TypeVar` bounds is close to free (pure static-analysis sugar, zero runtime cost, easy to remove). In Go, every additional type parameter is a small tax on readability at every call site, so the bar for "this is worth generalizing" is higher.

---

## 15. Exercises

1. **Write `Reverse[T any](s []T) []T`** that returns a new slice with elements in reverse order, without mutating the input. Test it on `[]int` and `[]string`.

2. **Write a generic `Set[T comparable]` type** backed by a `map[T]struct{}`, with `Add`, `Contains`, and `Remove` methods. (Why `map[T]struct{}` instead of `map[T]bool`? Look up the memory-size argument once you've got it working.)

3. **Hit the method-type-parameter wall on purpose.** Try writing a method on `Stack[T]` called `MapTo[U any]() Stack[U]` and confirm it fails to compile. Then rewrite it as a free function `MapStack[T, U any](s Stack[T], f func(T) U) Stack[U]` and confirm that one works.

4. **Build the generic repository from §12** for two different entity types (`User` and `Product`), and confirm the *same* `InMemoryRepository[T]` implementation serves both without any code duplication.

5. **Break `comparable` on purpose.** Try instantiating `Contains[T comparable]` with a struct type that has a slice field, read the exact compiler error, then fix it by either removing the slice field or writing a non-generic `ContainsFunc` that takes an explicit equality function instead.

6. **Constraint design.** Write a `Summable` constraint covering all numeric types where `+` is meaningful, then write `Sum[T Summable](items []T) T`. Compare your constraint to `cmp.Ordered` — which types does yours include that `cmp.Ordered` doesn't, and vice versa (hint: think about `string` and `+`)?

---

## 16. Key Takeaways

- **Generics exist in Go to give compile-time-checked polymorphism without `interface{}` + type assertions** — Python doesn't need this because it's dynamically typed at the language level; Go's `TypeVar`-equivalent has real runtime/compile-time teeth, unlike Python's checker-only typing.
- **A constraint is an interface** used in `[T Constraint]` position — `any` (no restriction), `comparable` (built-in, `==`/`!=` only), or a custom interface expressing a type set (`~int | ~float64 | ...`) or a method set (`fmt.Stringer`).
- **`~T` means "underlying type T," not "exact type T"** — omitting the tilde silently narrows a constraint more than you probably intend.
- **Type inference works from arguments, not return types** — if `T` only appears in what a function returns, you must instantiate it explicitly.
- **Methods cannot introduce new type parameters beyond their receiver's** — a hard language rule; reach for a free function when you need a fresh type parameter for a single operation.
- **The standard library (`cmp`, `slices`, `maps`, built-in `min`/`max`) already covers most everyday generic utility needs** since Go 1.21 — reach for those before writing your own `Max`/`Contains`/`Sort`.
- **Generics and (ordinary) interfaces solve different problems and aren't interchangeable** — generics for compile-time-shared logic across known types, interfaces for runtime polymorphism across a mix of types decided at runtime.
- **Idiomatic Go defers generics until duplication actually hurts** — write the concrete version first; this is a real cultural difference from how freely Python/TypeScript reach for type parameters.

This closes the generics gap flagged back at Guide 4. From here, Guides 10–14 (HTTP layer, DI, database, capstone) proceed as planned — you'll now recognize `[T any]` on sight when it shows up in router internals, ORM code, or your own repository layer, instead of it being unfamiliar syntax.