# Guide 6: Pointers & Reference Semantics

*Phase B — Core Language*

---

## Table of Contents

1. [Introduction & Motivation](#1-introduction--motivation)
2. [Quick Recap: Python's "Everything Is a Reference" Model](#2-quick-recap-pythons-everything-is-a-reference-model)
3. [Go's Baseline: Everything Is Passed By Value](#3-gos-baseline-everything-is-passed-by-value)
4. [Pointers: `&` and `*`, the Basics](#4-pointers--and--the-basics)
5. [When to Use `*T`: A Decision Framework](#5-when-to-use-t-a-decision-framework)
6. [Pointers to Structs and Field Access](#6-pointers-to-structs-and-field-access)
7. [The Range Loop Copy Gotcha](#7-the-range-loop-copy-gotcha)
8. [Slices: Shared Backing Array vs. Copied Header](#8-slices-shared-backing-array-vs-copied-header)
9. [The Append Gotcha](#9-the-append-gotcha)
10. [Maps: True Reference-Like Semantics](#10-maps-true-reference-like-semantics)
11. [Nil Pointers, Nil Slices, Nil Maps](#11-nil-pointers-nil-slices-nil-maps)
12. [Pointer Equality vs. Value Equality](#12-pointer-equality-vs-value-equality)
13. [Python vs. Go: Side-by-Side Summary](#13-python-vs-go-side-by-side-summary)
14. [Exercises](#14-exercises)
15. [Key Takeaways](#15-key-takeaways)

---

## 1. Introduction & Motivation

You've brushed up against pieces of this topic already: Guide 2 flagged that slices share memory in a way arrays don't, and Guide 3 spent real time on value vs. pointer receivers. This guide pulls all of that together and completes the picture, because reference semantics is one of those Go topics where the *individual* rules are each simple, but the *combination* of them produces genuinely surprising bugs — especially for someone coming from Python, where the question "is this shared or copied?" almost never has to be asked explicitly.

The uncomfortable truth for a Python developer is this: Python never made you think about pointers because Python doesn't have visible ones — everything is implicitly a reference, uniformly, all the time. Go makes the underlying distinction *explicit* and, worse, applies it *inconsistently* across its own built-in types: structs and arrays are copied by value; slices and maps look like they behave like Python objects, but only partially. Knowing exactly where that line falls is the entire subject of this guide, and it's where nearly every Python-to-Go beginner gets bitten at least once.

---

## 2. Quick Recap: Python's "Everything Is a Reference" Model

In Python, variables don't hold values directly — they hold references to objects, and assignment binds a name to an object rather than copying it.

```python
a = [1, 2, 3]
b = a          # b refers to the SAME list object as a
b.append(4)
print(a)       # [1, 2, 3, 4] — a sees the mutation too
```

Passing an argument to a function works the same way — Python passes the reference itself (often called "call by object reference" or "call by sharing"):

```python
def add_item(lst: list) -> None:
    lst.append("new")

items = ["a", "b"]
add_item(items)
print(items)   # ["a", "b", "new"] — mutated in place, visible to the caller
```

Crucially, this is uniform: lists, dicts, sets, and custom objects all behave this way, because they're all just objects accessed via references. The only reason immutable types (`int`, `str`, `tuple`, frozen dataclasses) *appear* to behave differently is that you can't mutate them at all — reassigning a name to a new value (`x = x + 1`) never mutates the original object, it just rebinds the local name to a different object.

This uniformity is exactly what Go does *not* have. Go's rules depend on the specific type, and that's the crux of this guide.

---

## 3. Go's Baseline: Everything Is Passed By Value

Go has exactly one parameter-passing rule, with no exceptions: **every argument passed to a function is copied.** There is no "pass by reference" mode in the language. What varies is *what gets copied* — and for some types, what gets copied is itself a small reference-like value, which is where the appearance of "reference semantics" comes from.

```go
type Point struct{ X, Y int }

func move(p Point) {
    p.X += 10 // mutates the LOCAL COPY only
}

func main() {
    pt := Point{X: 1, Y: 1}
    move(pt)
    fmt.Println(pt) // {1 1} — unchanged; move() only had a copy
}
```

This is the direct opposite of the Python list example above. `pt` is a struct — a value type — so passing it to `move` copies the entire thing. Any mutation inside `move` is invisible to the caller. This is the baseline you should assume for **every Go type** unless you know otherwise: structs, arrays, ints, strings, bools — all copied, all mutations local, full stop.

Slices, maps, channels, and function values are the exceptions — not because Go secretly passes them "by reference," but because **the value being copied is itself a small struct containing a pointer.** Copying the pointer is cheap, and both the original and the copy end up pointing at the same underlying data. This is the mechanical explanation for everything that follows in this guide.

---

## 4. Pointers: `&` and `*`, the Basics

A pointer is a value that holds the memory address of another value. Two operators are involved:

- `&x` — the **address-of** operator, producing a pointer to `x`.
- `*p` — the **dereference** operator, accessing the value that pointer `p` points to.

```go
package main

import "fmt"

func main() {
    x := 10
    p := &x        // p is a *int, holding the address of x
    fmt.Println(p)  // e.g., 0xc0000140a0 — a memory address
    fmt.Println(*p) // 10 — the value at that address

    *p = 20        // dereference and assign — mutates x through the pointer
    fmt.Println(x) // 20 — x itself changed
}
```

`p`'s type is `*int` — "pointer to int" — read the `*` in a type position as "pointer to." This is the mechanism that lets a function actually mutate a caller's variable:

```go
func increment(n *int) {
    *n = *n + 1 // dereference to read AND write the original value
}

func main() {
    count := 5
    increment(&count) // pass the ADDRESS of count
    fmt.Println(count) // 6 — mutation is visible!
}
```

Compare this to the `move(pt)` example in Section 3: the only difference is that `increment` takes a `*int` and dereferences it, while `move` took a plain `Point` and never had access to the original at all. **This is the single lever Go gives you to opt in to "mutate the caller's data": pass a pointer, and dereference it to write through.**

There's no Python equivalent to `&`/`*` because Python never needs one — every variable is already reference-like under the hood, invisibly. Go's pointers aren't a new *capability* so much as they're a visible, explicit spelling of something Python was always doing silently for mutable objects.

---

## 5. When to Use `*T`: A Decision Framework

Guide 3 covered this specifically for method receivers; here's the generalized version for any function parameter, variable, or struct field.

**Use a pointer (`*T`) when:**

- **You need the function to mutate the caller's value.** This is the most common reason, and mirrors exactly why Python code relies on mutable objects (lists, dicts) when a function needs to modify shared state.
- **The value is large and copying it is wasteful.** A struct with a dozen fields, or one containing large arrays, is expensive to copy on every function call; a pointer copies only a machine word (the address) regardless of how large the pointed-to data is.
- **You need to represent "no value" for something that isn't naturally nil-able.** A `Point` value can't be `nil` — but a `*Point` can be. This is Go's way of expressing what Python expresses with `None`: `var p *Point = nil` reads as "optional Point, currently absent," directly analogous to `p: Optional[Point] = None`.
- **You need identity, not just equal values.** If two callers should observe changes to the *same* underlying object (like two references to the same Python list), you need a shared pointer — a copied struct gives you two independent, unlinked values instead.
- **The type has a pointer-receiver method** (per Guide 3) that you need to call, or an interface (per Guide 4) that only the pointer type satisfies.

**Prefer a plain value (`T`) when:**

- The struct is small (a handful of primitive fields) — copying is cheap and often *faster* than the indirection of following a pointer.
- You want to guarantee the function *cannot* mutate the caller's data — passing by value is a strong, compiler-enforced guarantee of immutability from the caller's perspective, something Python cannot offer for mutable objects at all.
- The value has natural value semantics (a coordinate, a timestamp, a money amount) — things that are conceptually "just data," where two `Point{1, 2}` values being unrelated-but-equal is exactly what you want, not a bug.

The underlying theme: Go forces you to decide, explicitly, per type and per function signature, whether you want Python-list-like sharing or Python-tuple-like independence. Python never makes you choose for custom objects — they're always shared. Go makes you choose every time, which is more code to write but eliminates an entire category of "wait, why did that mutate?" surprises, *provided* you understand the rules — which is exactly what the rest of this guide is for.

---

## 6. Pointers to Structs and Field Access

Go provides a convenience that removes most of the syntactic pain of working with struct pointers: you can access fields on a pointer *without* manually dereferencing first.

```go
type Point struct{ X, Y int }

func main() {
    p := &Point{X: 1, Y: 2}

    fmt.Println(p.X)   // 1 — no need to write (*p).X
    p.X = 100          // no need to write (*p).X = 100
    fmt.Println(p.X)   // 100
}
```

Under the hood, `p.X` is shorthand for `(*p).X` — Go inserts the dereference automatically. This is why pointer-receiver methods (Guide 3) can read `d.rate` inside a method with receiver `(d *Discount)` without any explicit `*d.rate` noise — the compiler handles it transparently, and idiomatic Go code essentially never writes `(*p).Field` by hand.

---

## 7. The Range Loop Copy Gotcha

This is a classic trap, and it follows directly from Section 3's "everything is copied" rule — it just shows up somewhere non-obvious.

```go
type Point struct{ X, Y int }

func main() {
    points := []Point{{X: 1, Y: 1}, {X: 2, Y: 2}}

    for _, p := range points {
        p.X = 100 // mutates the LOOP VARIABLE'S COPY, not the slice element!
    }
    fmt.Println(points) // [{1 1} {2 2}] — UNCHANGED
}
```

`range` copies each element into the loop variable `p` on every iteration — exactly like passing that element to a function by value (Section 3). Mutating `p` mutates a temporary copy that's discarded at the end of the loop body. This surprises Python developers specifically *because* Python's `for p in points:` binds `p` to the *actual* object each time — `p.x = 100` inside a Python loop over a list of mutable objects genuinely does mutate them, because `p` is a reference to the real object, not a copy.

**Two idiomatic fixes:**

```go
// Fix 1: index into the slice directly
for i := range points {
    points[i].X = 100
}
fmt.Println(points) // [{100 1} {100 2}]

// Fix 2: use a slice of pointers, so the loop variable is a copied
// POINTER (cheap), but it still refers to the same underlying struct
pointPtrs := []*Point{{X: 1, Y: 1}, {X: 2, Y: 2}}
for _, p := range pointPtrs {
    p.X = 100 // mutates through the pointer — visible!
}
fmt.Println(*pointPtrs[0], *pointPtrs[1]) // {100 1} {100 2}
```

The rule of thumb: if you need to mutate elements while ranging over a collection of structs, either index directly (`points[i]`) or range over a collection of pointers — never rely on mutating the range loop's value variable directly.

---

## 8. Slices: Shared Backing Array vs. Copied Header

Guide 2 introduced this briefly; here is the full mechanical picture, because it's essential for what follows.

A slice value is actually a small struct with three fields: a pointer to an underlying array, a length, and a capacity. When you pass a slice to a function, **the three-field header is copied** — but the pointer inside that header still points at the *same* underlying array as the original. This is exactly the "copy a pointer, share the data" pattern from Section 3.

```go
func double(nums []int) {
    for i := range nums {
        nums[i] *= 2 // writes through the shared pointer — visible!
    }
}

func main() {
    nums := []int{1, 2, 3}
    double(nums)
    fmt.Println(nums) // [2 4 6] — mutation IS visible
}
```

This much matches Python list intuition perfectly: mutating elements in place, through indexing, behaves the same as Python's `lst[i] = x`. Where things diverge sharply is anything involving a change in **length** — and that's the subject of the next section.

---

## 9. The Append Gotcha

This is, without exaggeration, **the single most common slice-related bug for developers coming from Python**, because Python's `list.append` and Go's `append` look identical on the surface but behave fundamentally differently underneath.

In Python, a list is always the *same object* — appending mutates it in place, and every reference to that list sees the new element, no matter how deeply nested inside function calls the append happened.

```python
def add_item(lst: list) -> None:
    lst.append(99)

items = [1, 2, 3]
add_item(items)
print(items) # [1, 2, 3, 99] — ALWAYS visible, guaranteed
```

In Go, `append` may or may not return a slice header that shares the same backing array as its input, **depending on whether there's spare capacity.** If the underlying array is full, `append` allocates a brand-new, larger array, copies the existing elements into it, and returns a slice header pointing at the *new* array — completely disconnected from the original.

```go
func addItem(nums []int) {
    nums = append(nums, 99) // reassigns the LOCAL copy of the header
    fmt.Println("inside:", nums)
}

func main() {
    nums := []int{1, 2, 3} // len 3, cap 3 — no spare room
    addItem(nums)
    fmt.Println("outside:", nums) // [1 2 3] — 99 is GONE from the caller's view
}
```

Here, `append` had no spare capacity, so it allocated a new array inside `addItem`. The local variable `nums` inside `addItem` now points at that new array, but the caller's `nums` still points at the *original* three-element array — and even if it didn't reallocate, the caller's slice header still has `len == 3`, since headers are copied by value (Section 3) and `addItem` never communicated a new length back.

**The idiomatic fix is always the same: capture and use the return value of `append`.**

```go
func addItem(nums []int) []int {
    return append(nums, 99)
}

func main() {
    nums := []int{1, 2, 3}
    nums = addItem(nums) // reassign the caller's variable explicitly
    fmt.Println(nums)    // [1 2 3 99] — correct
}
```

This is why you'll see `s = append(s, x)` everywhere in Go, even though it looks redundant to a Python developer's eye ("why reassign to the same variable?") — it's not optional. `append` cannot reliably mutate its input in place from the caller's perspective; the returned value is the only trustworthy result.

### The Even Sneakier Variant: Partial Sharing

There's a subtler version of this bug that goes the *other* direction — where capacity *is* available, so no reallocation happens, and a write inside a function silently leaks into a slice the caller didn't even know was related:

```go
func main() {
    base := make([]int, 3, 10) // len 3, cap 10 — plenty of spare room
    base[0], base[1], base[2] = 1, 2, 3

    func(nums []int) {
        nums = append(nums, 99) // fits within capacity — SAME backing array
    }(base)

    fmt.Println(base) // [1 2 3] — len is still 3, so 99 isn't visible here...

    extended := base[:4] // ...but re-slicing to reveal index 3 exposes it
    fmt.Println(extended) // [1 2 3 99] — surprise!
}
```

The append inside the anonymous function wrote `99` directly into the shared backing array at index 3 (because there was capacity for it), but the caller's `base` slice header still reports `len == 3`, so `base` itself doesn't show the new element — until something creates a new slice header that extends into that same capacity, at which point the "leaked" write becomes visible. This exact pattern is a classic source of subtle, hard-to-reproduce bugs when slices are sliced from a shared, pre-allocated buffer (a common pattern for performance) and passed around independently.

**Practical guidance:** treat any slice passed into a function as **potentially aliased** with the caller's data for existing elements, but **never assume length changes propagate** unless the function returns the slice and you reassign it. When you deliberately want an independent copy, use `copy()`:

```go
original := []int{1, 2, 3}
duplicate := make([]int, len(original))
copy(duplicate, original) // independent backing array — safe to mutate freely
```

---

## 10. Maps: True Reference-Like Semantics

Maps behave much closer to Python dict intuition than slices do, and it's worth being precise about why. A Go map value is, under the hood, a pointer to an internal runtime hash-table structure. Copying a map (assigning it, or passing it to a function) copies that pointer — and unlike slices, there's no separate "length" or "capacity" field to get out of sync, and no user-visible reallocation event to worry about.

```go
func addEntry(m map[string]int) {
    m["new"] = 1 // mutates the SAME underlying table — always visible
}

func main() {
    scores := map[string]int{"a": 1}
    addEntry(scores)
    fmt.Println(scores) // map[a:1 new:1] — visible, no gotcha
}
```

This matches Python dict behavior directly:

```python
def add_entry(d: dict) -> None:
    d["new"] = 1

scores = {"a": 1}
add_entry(scores)
print(scores) # {'a': 1, 'new': 1}
```

**The one thing you cannot do** is have a function *replace* the caller's map entirely (reassign it to a new map, or set it to `nil`) and have that visible to the caller — because that requires rebinding the caller's variable itself, and Go's "everything is copied" rule (Section 3) means the function only has a copy of the map reference, not access to the original variable.

```go
func replace(m map[string]int) {
    m = map[string]int{"fresh": 1} // only rebinds the LOCAL variable m
}

func main() {
    scores := map[string]int{"a": 1}
    replace(scores)
    fmt.Println(scores) // map[a:1] — unchanged; replace() never touched the original variable
}
```

This mirrors the Python equivalent exactly:

```python
def replace(d: dict) -> None:
    d = {"fresh": 1}  # rebinds the LOCAL name d, doesn't touch the caller's dict

scores = {"a": 1}
replace(scores)
print(scores) # {'a': 1} — unchanged, for the identical reason
```

If you truly need a function to reassign the caller's map variable itself (rare), you need a pointer to the map: `*map[string]int`. In practice, idiomatic Go almost never does this — the common pattern is to return the new map and let the caller reassign, exactly like the `append` idiom in Section 9.

---

## 11. Nil Pointers, Nil Slices, Nil Maps

Go's `nil` plays the role of Python's `None`, but its behavior is type-dependent in ways `None` never is.

**Nil pointer:** dereferencing one panics.

```go
var p *int
fmt.Println(*p) // panic: runtime error: invalid memory address or nil pointer dereference
```

**Nil slice:** perfectly safe to read from (behaves like an empty slice) and safe to append to (allocates on first use).

```go
var s []int         // nil slice
fmt.Println(len(s)) // 0 — no panic
s = append(s, 1)     // works fine — allocates a new backing array
fmt.Println(s)       // [1]
```

**Nil map:** safe to *read* from (returns the zero value, with `ok == false` in the comma-ok form), but **panics on write**.

```go
var m map[string]int // nil map
fmt.Println(m["x"])   // 0 — safe read, no panic
v, ok := m["x"]
fmt.Println(v, ok)    // 0 false — safe

m["x"] = 1 // panic: assignment to entry in nil map
```

This read-safe-but-write-unsafe asymmetry has no Python analogue at all — `None["x"]` raises `TypeError` immediately, for both reads and writes, with no special case. In Go, always initialize maps you intend to write to, with `make(map[K]V)` or a map literal, rather than relying on a zero-value `var m map[K]V` declaration.

---

## 12. Pointer Equality vs. Value Equality

Two pointers compared with `==` check whether they hold the **same address** (identity) — not whether the values they point to are equal.

```go
type Point struct{ X, Y int }

a := &Point{X: 1, Y: 1}
b := &Point{X: 1, Y: 1}
c := a

fmt.Println(a == b) // false — different addresses, even though *a == *b
fmt.Println(a == c) // true  — same address
fmt.Println(*a == *b) // true — dereferenced VALUES are equal
```

This maps directly onto Python's distinction between `is` (identity) and `==` (equality, which for many built-in types compares values):

```python
class Point:
    def __init__(self, x, y):
        self.x, self.y = x, y

a = Point(1, 1)
b = Point(1, 1)
c = a

print(a is b) # False — different objects
print(a is c) # True  — same object
```

(Python's default `==` for custom classes actually falls back to identity too, unless you define `__eq__` — so the Go pointer-`==` behavior maps most cleanly onto Python's `is`, while Go's struct-`==`, comparing all fields, maps onto a Python class with a proper `__eq__` defined.)

---

## 13. Python vs. Go: Side-by-Side Summary

| Concept | Python | Go |
|---|---|---|
| Default variable/parameter behavior | Always a reference to an object (uniform across all types) | Always copied by value (uniform rule); some types *contain* a pointer, creating the appearance of sharing |
| Mutating a passed-in mutable collection | Always visible to the caller (list, dict, etc. are the same object) | Slices: element mutation visible; length changes often **not** visible. Maps: mutation of entries always visible |
| Explicit address-of / dereference | Not applicable — no visible pointers | `&x` (address-of), `*p` (dereference) |
| "Optional" value for a non-nullable type | `Optional[T] = None` | `*T = nil` |
| Looping and mutating elements | `for x in items: x.attr = ...` mutates the real object | `for _, x := range items { x.Field = ... }` mutates a **copy** — must index (`items[i].Field = ...`) instead |
| Growing a collection inside a function | `lst.append(x)` — always visible to caller | `s = append(s, x)` — visible only if the caller also reassigns from the return value |
| Replacing a collection variable inside a function | Rebinding a local name — never visible to caller | Reassigning a map/slice parameter — never visible to caller (same underlying reason) |
| Identity vs. equality | `is` (identity) vs. `==` (value equality, if `__eq__` defined) | `==` on pointers (identity/address) vs. `==` on dereferenced values or structs (field equality) |
| `None`/`nil` safety | Any operation on `None` raises immediately | Nil slice: safe to read/append. Nil map: safe to read, **panics on write**. Nil pointer: panics on dereference |

---

## 14. Exercises

1. Write a function `Reset(p *Point)` that sets both fields of a `Point` struct to zero through the pointer. Show that calling it on `&pt` mutates `pt`, while an equivalent function taking `Point` (not `*Point`) does not.
2. Reproduce the range-loop copy gotcha from Section 7 with a slice of a custom struct, then fix it using both approaches shown (indexing, and a slice of pointers).
3. Write a function `AppendAndReturn(nums []int, x int) []int` that correctly appends and returns the result. Then write a *broken* version that appends without returning, and demonstrate with a small `main` that the caller's slice is unaffected by the broken version when the backing array is full.
4. Demonstrate the "partial sharing" gotcha from Section 9: create a slice with `make([]int, 2, 5)`, pass it to a function that appends one element, and show that re-slicing the original to a larger length reveals the appended value.
5. Write a function `SafeGet(m map[string]int, key string) (int, bool)` that works correctly even when `m` is `nil`. Then write a `main` that demonstrates writing to a nil map panics, and show the fix (initializing with `make`).

---

## 15. Key Takeaways

- Go has exactly one parameter-passing rule: **everything is copied.** There is no true "pass by reference" — the *appearance* of shared mutation only happens for types (slices, maps, pointers, channels, function values) whose copied value itself contains a pointer to shared data.
- **Use a pointer (`*T`)** when you need to mutate the caller's data, avoid copying a large struct, represent an optional/absent value, need shared identity, or must satisfy a pointer-receiver interface.
- **The range loop copy gotcha**: `for _, x := range items { x.Field = ... }` mutates a throwaway copy. Index directly (`items[i].Field = ...`) or range over pointers when mutation is needed.
- **The append gotcha is the single biggest slice trap for Python developers**: unlike Python's `list.append`, Go's `append` may or may not share the original backing array, and length changes are never visible to the caller unless the returned slice is captured and reassigned. Always write `s = append(s, x)`, and always return an appended slice from a function rather than mutating a parameter in place.
- A subtler variant — writes landing in shared *capacity* that becomes visible only when another slice re-slices into that region — is a classic source of hard-to-find bugs; use `copy()` when you need a truly independent slice.
- **Maps behave much more like Python dicts**: entry mutations are always visible to the caller, with no append-style gotcha, because there's no separate length/capacity header to fall out of sync — but a function still cannot reassign or nil out the caller's map variable itself, for the same "everything is copied" reason.
- **Nil has type-dependent behavior** in Go, unlike Python's uniform `None`: nil pointers panic on dereference, nil maps panic only on write (reads are safe), and nil slices are safe for both reading and appending.
- Pointer `==` compares addresses (identity, like Python's `is`); dereferenced-value or struct `==` compares contents (like a Python `__eq__`).

---

*Next: Guide 7 begins Phase C — Concurrency, starting with goroutines and the fundamentals of Go's concurrency model.*
