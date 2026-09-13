# Guide 4: Interfaces & Polymorphism

*Phase B — Core Language*

---

## Table of Contents

1. [Introduction & Motivation](#1-introduction--motivation)
2. [Quick Recap: How Python Handles Polymorphism](#2-quick-recap-how-python-handles-polymorphism)
3. [What Is a Go Interface?](#3-what-is-a-go-interface)
4. [Implicit Satisfaction: The Core Idea](#4-implicit-satisfaction-the-core-idea)
5. [Defining and Using Interfaces](#5-defining-and-using-interfaces)
6. [The Empty Interface and `any`](#6-the-empty-interface-and-any)
7. [Type Assertions](#7-type-assertions)
8. [Type Switches](#8-type-switches)
9. [Interface Composition (Embedding)](#9-interface-composition-embedding)
10. [Interfaces from the Standard Library](#10-interfaces-from-the-standard-library)
11. [Design Guidelines: Accept Interfaces, Return Structs](#11-design-guidelines-accept-interfaces-return-structs)
12. [The Nil Interface Gotcha](#12-the-nil-interface-gotcha)
13. [Python vs. Go: Side-by-Side Summary](#13-python-vs-go-side-by-side-summary)
14. [Exercises](#14-exercises)
15. [Key Takeaways](#15-key-takeaways)

---

## 1. Introduction & Motivation

In Guide 3, you learned how Go builds behavior around structs using methods, and how composition replaces inheritance. That covered *concrete* types — types with a fixed, known shape. This guide covers Go's answer to polymorphism: **interfaces**.

If structs are Go's answer to "how do I group data and attach behavior to it," interfaces are Go's answer to "how do I write code that works with *any* type that behaves a certain way, without caring what that type actually is." This is the mechanism that lets you write a function that accepts "anything that can be written to," or "anything that can be converted to a string," without hardcoding a specific struct.

For you, coming from Python, the destination is familiar — Python code relies on polymorphism constantly. The *mechanism* is what's different, and that mechanism is the single most important idea in this guide: **Go interfaces are satisfied implicitly.** There is no `implements` keyword. There is no explicit declaration of intent. A type satisfies an interface simply by having the right methods. This single design decision ripples through how Go code is structured, tested, and extended, so we'll spend real time on it before moving to syntax.

---

## 2. Quick Recap: How Python Handles Polymorphism

Python gives you polymorphism through a spectrum of mechanisms, roughly in order of how "enforced" they are:

**1. Pure duck typing (no declaration at all).** If an object has the method you call, it works. Nothing is checked until runtime, and nothing is declared in advance.

```python
class Duck:
    def speak(self):
        return "Quack"

class Dog:
    def speak(self):
        return "Woof"

def make_it_speak(animal):
    print(animal.speak())  # works for anything with .speak()

make_it_speak(Duck())
make_it_speak(Dog())
```

There is no shared base class here. `make_it_speak` doesn't know or care what type it receives — it just calls `.speak()` and hopes for the best. If the object doesn't have `.speak()`, you get an `AttributeError` *at call time*, not before.

**2. Abstract Base Classes (`abc` module).** When you want to *enforce* a contract, Python gives you `ABC` and `@abstractmethod`. This requires explicit inheritance — the subclass must declare `class Dog(Animal):` — and Python will refuse to instantiate a class that hasn't implemented all abstract methods.

```python
from abc import ABC, abstractmethod

class Animal(ABC):
    @abstractmethod
    def speak(self) -> str:
        ...

class Dog(Animal):
    def speak(self) -> str:
        return "Woof"

# Animal()  # TypeError: Can't instantiate abstract class
```

This is *explicit* and *nominal* — the relationship between `Dog` and `Animal` is declared in the class definition, and Python checks it.

**3. `typing.Protocol` (structural typing, added in Python 3.8).** This is actually the closest Python analogue to Go interfaces. A `Protocol` defines a shape, and any class matching that shape satisfies it — *without inheriting from it*. This is checked by static type checkers (mypy, pyright), not at runtime by default.

```python
from typing import Protocol

class SpeakerProtocol(Protocol):
    def speak(self) -> str: ...

def make_it_speak(animal: SpeakerProtocol) -> None:
    print(animal.speak())

# Dog from above satisfies this without any inheritance —
# mypy will accept make_it_speak(Dog()) with no declared relationship.
```

So Python actually spans the full spectrum: fully dynamic duck typing, fully nominal ABCs, and structural typing via `Protocol`. Go picks *one* point on that spectrum and uses it everywhere: **structural typing, checked at compile time.** It's `Protocol`, but it's not an opt-in feature bolted on for static analysis — it's the only mechanism Go has, and the compiler enforces it before your program ever runs.

---

## 3. What Is a Go Interface?

An interface in Go is a type that specifies a **method set** — a list of method signatures. It says nothing about data, only behavior.

```go
type Speaker interface {
    Speak() string
}
```

This declares that "a `Speaker` is anything with a `Speak() string` method." No struct is mentioned. No relationship to any concrete type is declared. The interface is purely a *description of behavior*.

Compare this to the struct definitions from Guide 3, which described *shape* (fields and their types). An interface describes *capability* (methods and their signatures). This distinction — structs hold data, interfaces describe behavior — is a good mental anchor as we go forward.

---

## 4. Implicit Satisfaction: The Core Idea

Here is the idea the entire guide pivots on:

> **A type satisfies an interface automatically, just by implementing its methods. There is no explicit declaration anywhere.**

```go
package main

import "fmt"

type Speaker interface {
    Speak() string
}

type Dog struct {
    Name string
}

func (d Dog) Speak() string {
    return d.Name + " says Woof"
}

type Cat struct {
    Name string
}

func (c Cat) Speak() string {
    return c.Name + " says Meow"
}

func announce(s Speaker) {
    fmt.Println(s.Speak())
}

func main() {
    announce(Dog{Name: "Rex"})
    announce(Cat{Name: "Whiskers"})
}
```

Notice what is *absent*: nowhere does `Dog` or `Cat` say "I am a `Speaker`." There's no `type Dog struct { ... } implements Speaker`, no inheritance clause, nothing. `Dog` satisfies `Speaker` purely because it happens to have a method named `Speak` that returns a `string`. If `Cat` didn't exist yet when `Speaker` was designed, it doesn't matter — as soon as `Cat` has a matching `Speak()` method, it satisfies the interface too.

This is exactly the mechanical behavior of Python's `Protocol` — but where `Protocol` is an optional annotation checked by an external tool (mypy), Go's version is baked into the language and checked by the compiler itself, on every build.

**Why this matters in practice:**

- **Decoupling.** The package that defines `Speaker` doesn't need to know about `Dog` or `Cat` at all. In fact, the *idiomatic* pattern in Go is for the *consumer* of an interface to define it, not the producer of the concrete type — the opposite of how you'd typically design an ABC hierarchy in Python, where the base class usually lives with (or above) its implementations.
- **Retroactive conformance.** A type written years ago, in a totally different package, with no awareness that your interface exists, can satisfy your interface today — as long as its method signatures line up. You can even satisfy interfaces from the standard library without ever importing the package that defines them.
- **No fragile hierarchies.** You never need to restructure an inheritance tree to make a new type fit an old interface, because there is no tree.

**The trade-off:** the compiler *is* checking this — you don't get away with sloppy method names — but the check happens where a type is *used* as an interface, not where it's *defined*. Read an error like `Dog does not implement Speaker (missing method Speak)` at the call site, and you'll immediately know what's missing.

---

## 5. Defining and Using Interfaces

### 5.1 Syntax

```go
type InterfaceName interface {
    MethodOne(paramType) returnType
    MethodTwo() (returnType1, returnType2)
}
```

By convention, single-method interfaces are often named with an `-er` suffix describing the action: `Speaker`, `Writer`, `Reader`, `Stringer`, `Closer`. This isn't enforced, but it's idiomatic and you'll see it throughout the standard library.

### 5.2 Interfaces as Variable Types

An interface is a type like any other — you can declare variables of interface type, use it as a function parameter or return type, or store it in a slice or map.

```go
var s Speaker      // zero value is nil (see section 12)
s = Dog{Name: "Rex"}
fmt.Println(s.Speak())

animals := []Speaker{
    Dog{Name: "Rex"},
    Cat{Name: "Whiskers"},
}
for _, a := range animals {
    fmt.Println(a.Speak())
}
```

That last example — a slice holding *different concrete types* united only by shared behavior — is the direct equivalent of a Python list like `[Dog(), Cat()]` passed through a loop calling `.speak()` on each. The difference: Go's version is statically type-checked. If you tried to append a type without a `Speak()` method to that slice, it wouldn't compile.

### 5.3 Value Receivers vs. Pointer Receivers and Interface Satisfaction

This interacts directly with Guide 3's discussion of receivers, and it trips up almost everyone the first time:

```go
type Speaker interface {
    Speak() string
}

type Robot struct{ ID int }

func (r *Robot) Speak() string { // pointer receiver
    return fmt.Sprintf("Robot #%d online", r.ID)
}

func main() {
    var s Speaker
    s = &Robot{ID: 1} // OK — *Robot satisfies Speaker
    // s = Robot{ID: 1}  // COMPILE ERROR — Robot (value) does NOT satisfy Speaker
    fmt.Println(s.Speak())
}
```

The rule: **the method set of a pointer type `*T` includes methods with both value and pointer receivers; the method set of a value type `T` includes only methods with value receivers.** If any method in your interface is implemented with a pointer receiver, only the pointer type satisfies the interface — the value type does not. This is a compile-time check, so you find out immediately, but the error message ("`Robot` does not implement `Speaker` (method `Speak` has pointer receiver)") is worth recognizing on sight.

Python has no equivalent friction here, because Python methods don't distinguish between "operates on the object" and "operates on a reference to the object" — everything is a reference. This is one of the few places where Go's value/pointer semantics genuinely surface at the interface level, so keep it in your back pocket.

---

## 6. The Empty Interface and `any`

### 6.1 What It Is

Go has a special interface with *zero* methods: `interface{}`. Since every type has at least zero methods, **every single type in Go satisfies the empty interface.** It's the closest thing Go has to "a variable that can hold anything."

```go
var anything interface{}
anything = 42
anything = "hello"
anything = Dog{Name: "Rex"}
```

Since Go 1.18, the built-in alias `any` means exactly the same thing and is now the idiomatic spelling:

```go
var anything any // identical to interface{}
```

You'll see both forms in the wild — `interface{}` in older codebases and tutorials, `any` in anything written post-2022 — but write `any` in new code.

### 6.2 Python Comparison

This is roughly analogous to a Python variable with no type hint at all, or one hinted as `typing.Any`:

```python
from typing import Any

def process(value: Any) -> None:
    ...
```

In Python, *everything* is already like this by default — Python doesn't require you to declare a type for a variable, so the "can hold anything" behavior is the norm rather than an escape hatch. In Go, the opposite is true: normal variables are strictly typed, and `any` is a deliberate, visible opt-out of that guarantee. Seeing `any` in a Go function signature is a signal — "this function doesn't care about, or doesn't yet know, the concrete type" — in a way that a plain Python parameter is not.

### 6.3 Why `any` Should Be a Last Resort

Using `any` throws away the compiler's ability to check anything about the value at compile time. Once a value is stored as `any`, you can't call any methods on it or use it in any type-specific way without first recovering its concrete type (see the next two sections). Overusing `any` in Go is roughly equivalent to overusing `Any` (or no type hints at all) in a codebase that otherwise uses `mypy` strictly — it works, but you've opted a chunk of your program out of the safety net.

Common legitimate uses: generic containers before Go had generics (rare to write today — see the note below), JSON unmarshaling into unknown structures, and function signatures like `fmt.Println(a ...any)` that genuinely need to accept arbitrary values for printing.

> **Note on generics:** Go added true generics (type parameters) in Go 1.18, which cover most of the cases that used to require `any` plus type assertions. Generics are covered in a later guide — for now, know that `any` is for genuinely unknown/dynamic data (like JSON), not a substitute for generic functions.

---

## 7. Type Assertions

Once a value is behind an interface (whether a narrow one like `Speaker` or the empty interface `any`), you sometimes need to ask: "is this *specifically* a `Dog`?" That's a **type assertion**.

### 7.1 Basic Syntax

```go
var s Speaker = Dog{Name: "Rex"}

d := s.(Dog)          // "panicking" form
fmt.Println(d.Name)   // Rex
```

`s.(Dog)` asserts that the concrete value stored inside `s` is a `Dog`, and if so, gives you back a `Dog`-typed value you can use normally (access fields, call `Dog`-specific methods, etc.). This is the mechanism for recovering type-specific behavior after you've generalized it away.

### 7.2 The Danger: Panicking Form

If the assertion is wrong, the program panics:

```go
var s Speaker = Cat{Name: "Whiskers"}
d := s.(Dog) // panics: interface conversion: main.Cat is not main.Dog
```

This crashes your program if unhandled — similar in spirit to an unguarded Python `isinstance` failure combined with unsafe access, e.g. calling a method that doesn't exist and getting an `AttributeError` you didn't catch. Because of this risk, the panicking form should only be used when you are *certain* of the underlying type (for example, immediately after constructing the value yourself).

### 7.3 The Safe Form: Comma-OK Idiom

The idiomatic, safe way to do a type assertion uses two return values:

```go
d, ok := s.(Dog)
if !ok {
    fmt.Println("s is not a Dog")
    return
}
fmt.Println(d.Name)
```

If the assertion succeeds, `ok` is `true` and `d` holds the properly-typed value. If it fails, `ok` is `false`, `d` holds the *zero value* of `Dog`, and — critically — **the program does not panic.** This is the same "comma-ok" pattern you've already seen for map lookups in Guide 2 (`v, ok := m[key]`) — Go reuses this idiom deliberately across the language so it becomes a recognizable shape rather than a one-off trick.

### 7.4 Python Comparison

The direct equivalent is `isinstance`:

```python
if isinstance(s, Dog):
    print(s.name)
else:
    print("s is not a Dog")
```

Python's `isinstance` is a runtime check against the class hierarchy (or, with `Protocol` and `@runtime_checkable`, a structural check). Go's comma-ok type assertion is also a runtime check — this is one of the few places where Go, despite being statically typed, defers a decision to runtime, because the *static* type at that point genuinely is only "something satisfying `Speaker`," and recovering the concrete type is inherently a runtime question. The difference from Python is mainly ergonomic: Go forces you to explicitly handle the failure case via the second return value, whereas Python's `isinstance` naturally sits inside an `if`, and unguarded attribute access elsewhere in the code could still blow up separately.

---

## 8. Type Switches

When you need to branch across *several* possible concrete types, repeated `if`/type-assertion chains get noisy. Go provides a dedicated construct: the **type switch**.

### 8.1 Syntax

```go
func describe(s Speaker) {
    switch v := s.(type) {
    case Dog:
        fmt.Printf("%s is a dog\n", v.Name)
    case Cat:
        fmt.Printf("%s is a cat\n", v.Name)
    case nil:
        fmt.Println("s is nil")
    default:
        fmt.Printf("unknown speaker type: %T\n", v)
    }
}
```

The special syntax `s.(type)` is only legal inside a `switch` statement. In each `case`, the variable `v` is automatically re-typed to match that case — inside `case Dog:`, `v` has type `Dog` and you can access `v.Name` directly, with no further assertion needed. This is the same underlying comma-ok mechanism from section 7, just extended to check against a list of types instead of one.

### 8.2 Python Comparison

The nearest Python equivalent is a chain of `isinstance` checks, or, since Python 3.10, structural pattern matching with `match`:

```python
match s:
    case Dog():
        print(f"{s.name} is a dog")
    case Cat():
        print(f"{s.name} is a cat")
    case None:
        print("s is None")
    case _:
        print(f"unknown speaker type: {type(s)}")
```

This is a close visual and semantic parallel, and if you've used Python's `match`/`case`, Go's type switch should feel immediately familiar. The main difference is scope of purpose: Python's `match` is a general pattern-matching construct (it can match on values, sequences, and structure, not just types), while Go's type switch is exclusively for dispatching on the concrete type behind an interface value.

### 8.3 When to Reach for a Type Switch

Needing frequent type switches over the same interface is often a *design smell* — in many cases it means you should have put another method on the interface itself, so each concrete type handles its own behavior via polymorphism instead of a central function inspecting types. This is the same instinct that in Python nudges you away from long `isinstance` chains and toward polymorphic methods or the visitor pattern. That said, type switches are entirely legitimate for things like generic JSON handling, error inspection, or writing genuinely generic utility functions (e.g., a debug-printer that needs to special-case a few known types).

---

## 9. Interface Composition (Embedding)

Just as structs can embed other structs (Guide 3), interfaces can embed other interfaces. This builds larger contracts out of smaller ones.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// ReadWriter embeds both — a type must satisfy BOTH to satisfy ReadWriter.
type ReadWriter interface {
    Reader
    Writer
}
```

Any concrete type that implements both a `Read` method and a `Write` method with the correct signatures automatically satisfies `ReadWriter` — again, with no explicit declaration. This is precisely how the standard library builds up its I/O interfaces (see the next section) — small, single-method interfaces composed into larger ones as needed, rather than one large monolithic interface.

The Python analogue is composing multiple `Protocol`s, or historically, using multiple inheritance with ABCs (`class Thing(Reader, Writer):`). Go's version has the same compositional spirit as multiple inheritance, but without any of multiple inheritance's classic problems (no diamond problem, no MRO to reason about) — because interfaces carry *no implementation*, only method signatures. Combining them is pure addition of requirements, never a question of which parent's method wins.

---

## 10. Interfaces from the Standard Library

Three interfaces show up constantly in idiomatic Go, and recognizing them will make standard library code far more readable.

### 10.1 `error`

You've been returning `error` values since Guide 2's discussion of multiple return values. `error` is itself just an interface:

```go
type error interface {
    Error() string
}
```

Any type with an `Error() string` method is an error — this is why you can define your own custom error types (a later guide covers this in depth) and still return them wherever the built-in `error` type is expected.

### 10.2 `fmt.Stringer`

```go
type Stringer interface {
    String() string
}
```

If a type implements `String() string`, functions like `fmt.Println` and `fmt.Printf("%v", ...)` will automatically call it to get a human-readable representation — this is Go's equivalent of Python's `__str__` / `__repr__` dunder methods, but surfaced as an ordinary, implicitly-satisfied interface rather than a special reserved method name.

```go
type Dog struct{ Name string }

func (d Dog) String() string {
    return "Dog(" + d.Name + ")"
}

fmt.Println(Dog{Name: "Rex"}) // prints: Dog(Rex)
```

### 10.3 `io.Reader` and `io.Writer`

These underpin nearly all of Go's I/O — files, network connections, in-memory buffers, HTTP request/response bodies — all of it is built on two tiny interfaces:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Because these are so minimal and so universally implemented, a function written to accept an `io.Writer` can transparently write to a file, a network socket, an in-memory buffer, or `os.Stdout`, without ever being rewritten. This is the payoff of small, structurally-satisfied interfaces: maximum reuse from minimum surface area.

---

## 11. Design Guidelines: Accept Interfaces, Return Structs

This is one of the most quoted Go proverbs, and it's worth understanding *why* it holds:

- **Function parameters should be interfaces** whenever the function only needs specific behavior, not a specific type. This maximizes what can be passed in — any conforming type, present or future, works.
- **Function return values should usually be concrete types (structs)**, not interfaces. Returning a concrete type gives the caller the full, unrestricted method set of that type to work with; returning an interface artificially limits them to only the methods the interface declares, and makes it harder for the compiler to help them.

```go
// Good: accepts an interface (flexible input)
func LogSpeak(w io.Writer, s Speaker) {
    fmt.Fprintln(w, s.Speak())
}

// Good: returns a concrete type (full capability for the caller)
func NewDog(name string) Dog {
    return Dog{Name: name}
}
```

The Python framing: this is close to the "be liberal in what you accept, conservative in what you return" principle, and it maps onto typing-with-`Protocol` guidance — accept the narrowest `Protocol` a function actually needs, but let constructors and factories return concrete, fully-featured classes rather than a `Protocol`-typed value.

---

## 12. The Nil Interface Gotcha

This is a well-known Go trap, and it's worth seeing once deliberately rather than discovering it in production.

```go
type MyError struct{}

func (e *MyError) Error() string { return "something broke" }

func doSomething(fail bool) error {
    var err *MyError // nil pointer
    if fail {
        err = &MyError{}
    }
    return err // BUG: always returns a non-nil error!
}

func main() {
    err := doSomething(false)
    if err != nil {
        fmt.Println("got an error") // this prints, even though nothing failed!
    }
}
```

Why does this happen? An interface value in Go is really a pair: `(type, value)`. When you assign a `nil` `*MyError` to a variable of interface type `error`, the interface's *type* part becomes `*MyError` and the *value* part becomes `nil` — but the interface itself is **not** `nil`, because it has a concrete type recorded. An interface is only `== nil` when **both** its type and value are unset.

```go
var err *MyError = nil
var i error = err
fmt.Println(err == nil) // true  (comparing the concrete pointer)
fmt.Println(i == nil)   // false (interface has a type, even though value is nil)
```

**The fix:** return a *literal* `nil` (of the interface type itself) when there's no error, rather than a typed nil pointer variable:

```go
func doSomething(fail bool) error {
    if fail {
        return &MyError{}
    }
    return nil // explicit, untyped nil — this is genuinely nil as an interface
}
```

There's no direct Python parallel here, because Python's `None` doesn't carry a "recorded type" the way a Go interface value does — `x is None` is a single, unambiguous check regardless of what `x` used to hold. This particular sharp edge is a consequence of the `(type, value)` pair representation, and it's specific enough to Go that it's worth just memorizing the pattern: **never return a typed nil pointer as an error; return a literal `nil` instead.**

---

## 13. Python vs. Go: Side-by-Side Summary

| Concept | Python | Go |
|---|---|---|
| Basic polymorphism | Duck typing (no declaration) | Implicit interface satisfaction (structural, compiler-checked) |
| Enforced contract | `abc.ABC` + `@abstractmethod` (nominal, explicit inheritance) | Interface method set (structural, no declaration) |
| Structural typing | `typing.Protocol` (opt-in, checked by mypy/pyright, not at runtime by default) | The only mechanism; checked by the compiler on every build |
| "Accept anything" type | Untyped variable, or `typing.Any` | `any` (alias for `interface{}`) |
| Recover a specific type | `isinstance(x, T)` | Type assertion: `v, ok := x.(T)` |
| Multi-way type dispatch | `match`/`case` (3.10+) or `isinstance` chain | Type switch: `switch v := x.(type) { ... }` |
| Combine contracts | Multiple inheritance / multiple `Protocol`s | Interface embedding |
| String representation hook | `__str__` / `__repr__` (special dunder methods) | `String() string` (ordinary method, `Stringer` interface) |
| Nil/None comparison | `x is None` — always unambiguous | Typed-nil-in-interface trap: an interface holding a nil pointer is **not** `== nil` |
| Design guidance | Accept the narrowest `Protocol`/ABC needed | "Accept interfaces, return structs" |

---

## 14. Exercises

1. Define an interface `Shape` with methods `Area() float64` and `Perimeter() float64`. Implement it for `Rectangle` and `Circle` structs (from Guide 3's struct patterns). Write a function that accepts a `[]Shape` and prints the total area.
2. Write a function `Describe(x any)` that uses a type switch to print a different message for `int`, `string`, `bool`, and a `default` case, including the value in each message.
3. Deliberately reproduce the nil-interface gotcha from Section 12: write a function that returns a typed-nil `*MyError` as an `error`, and a `main` that shows the interface is non-nil despite the pointer being nil. Then fix it.
4. Define two single-method interfaces (e.g., `Flyer` with `Fly() string` and `Swimmer` with `Swim() string`), embed them into a combined `Amphibious` interface, and implement a struct that satisfies it.
5. Using the comma-ok type assertion (not a type switch), write a function that accepts a `Speaker` (from Section 4) and prints `"Rex the dog says: <speak text>"` only if the concrete type is `Dog`, and does nothing otherwise.

---

## 15. Key Takeaways

- Go interfaces describe **behavior** (method sets), never data — a sharp contrast to structs, which describe data shape.
- **Implicit satisfaction is the core idea of this guide**: a type satisfies an interface automatically by having the right methods, with no `implements` keyword and no declared relationship. This is structurally identical to Python's `typing.Protocol`, except it's the *only* mechanism Go has, and it's enforced by the compiler on every build rather than by an opt-in static analysis tool.
- `any` (the modern spelling of `interface{}`) can hold any value, but using it discards compile-time type safety — use it sparingly, and prefer narrow interfaces or generics where possible.
- Recover a concrete type from an interface using a **type assertion** (`v, ok := x.(T)` — always prefer the comma-ok form over the panicking form) or a **type switch** (`switch v := x.(type) { ... }`) when branching across several possible types.
- Interfaces compose via **embedding**, letting you build larger contracts from small, single-purpose ones — this is the pattern behind the standard library's `io.Reader`/`io.Writer`/`io.ReadWriter` family.
- Watch for the **pointer-receiver-only satisfaction** rule (only `*T` satisfies an interface if any method has a pointer receiver) and the **nil interface gotcha** (an interface holding a typed nil pointer is not itself `nil`) — both are compile-time-safe but runtime-subtle, and both are classic sources of confusion for newcomers.
- Design guidance: **accept interfaces, return structs** — take the narrowest behavioral contract a function needs as input, but return concrete types so callers get full capability.

---

*Next: Guide 5 begins Phase C — Concurrency, starting with goroutines and the fundamentals of Go's concurrency model.*
