# Guide 5: Error Handling

*Phase B — Core Language*

---

## Table of Contents

1. [Introduction & Motivation](#1-introduction--motivation)
2. [Quick Recap: How Python Handles Errors](#2-quick-recap-how-python-handles-errors)
3. [The `error` Interface, Revisited](#3-the-error-interface-revisited)
4. [The Idiomatic Pattern: Check, Don't Catch](#4-the-idiomatic-pattern-check-dont-catch)
5. [Creating Errors: `errors.New` and `fmt.Errorf`](#5-creating-errors-errorsnew-and-fmterrorf)
6. [Sentinel Errors and `errors.Is`](#6-sentinel-errors-and-errorsis)
7. [Custom Error Types and `errors.As`](#7-custom-error-types-and-errorsas)
8. [Wrapping Errors with `%w`](#8-wrapping-errors-with-w)
9. [Combining Multiple Errors: `errors.Join`](#9-combining-multiple-errors-errorsjoin)
10. [`panic` and `recover`](#10-panic-and-recover)
11. [The `defer` + `recover` Pattern](#11-the-defer--recover-pattern)
12. [When to Panic vs. Return an Error](#12-when-to-panic-vs-return-an-error)
13. [Python vs. Go: Side-by-Side Summary](#13-python-vs-go-side-by-side-summary)
14. [Exercises](#14-exercises)
15. [Key Takeaways](#15-key-takeaways)

---

## 1. Introduction & Motivation

Every guide so far has flagged some difference from Python worth noting. This one is different: error handling is, by a wide margin, **the single biggest mindset shift** you'll make moving from Python to Go. It's not a new keyword to learn or a syntax quirk to memorize — it's a fundamentally different philosophy about what an error *is* and how it should move through a program.

In Python, an error is an **exceptional event** that interrupts normal control flow, propagates up the call stack automatically, and must be explicitly caught to be handled. In Go, an error is **just a value** — an ordinary return value, like an `int` or a `string`, that the caller receives and must explicitly check. Nothing propagates automatically. Nothing is "thrown." If you don't check it, the error is silently sitting in a variable you ignored, not an exception waiting to crash your program down the line.

This guide covers the full toolkit: the `error` interface itself (introduced briefly in Guide 4), the idiomatic "check, don't catch" pattern, creating and wrapping errors, sentinel errors, custom error types, and the two functions — `errors.Is` and `errors.As` — that do for errors what `isinstance` and equality checks do in Python. It closes with `panic`/`recover`, Go's *actual* exception-like mechanism, and clarifies why it is reserved for a narrow set of cases that are not "normal" error handling at all.

---

## 2. Quick Recap: How Python Handles Errors

Python's model will be completely familiar to you, but it's worth stating explicitly so the contrast in the rest of this guide lands clearly.

```python
def divide(a: float, b: float) -> float:
    if b == 0:
        raise ValueError("cannot divide by zero")
    return a / b

def process(a: float, b: float) -> None:
    try:
        result = divide(a, b)
        print(f"Result: {result}")
    except ValueError as e:
        print(f"Error: {e}")
    except Exception as e:
        print(f"Unexpected error: {e}")
    finally:
        print("Cleanup runs regardless")
```

Key properties of this model:

- **Errors are raised, not returned.** `raise` interrupts the normal flow of the function immediately — nothing after the `raise` statement executes.
- **Propagation is automatic.** If `divide` raises and `process` didn't have a `try`/`except`, the exception would keep unwinding up the call stack — through `process`'s caller, and its caller, and so on — until either something catches it or the program crashes with a traceback. You don't have to do anything for this to happen; it's the default.
- **Catching is opt-in and type-based.** `except ValueError` catches only that class (and its subclasses); a bare `except Exception` catches broadly.
- **`finally` guarantees cleanup.** Code in `finally` runs whether or not an exception occurred.
- **Nothing forces you to handle anything at the call site.** You can call `divide(a, b)` with no `try` at all, and if it raises, your function also (silently, from the caller's perspective) starts propagating that exception.

That last point is the seed of the whole contrast. In Python, error handling is **invisible in the type signature** — `def divide(a: float, b: float) -> float` gives no hint that it might raise. In Go, as you're about to see, error handling is **part of the function's return signature**, visible at every call site, every time.

---

## 3. The `error` Interface, Revisited

Guide 4 introduced `error` as an example of an implicitly-satisfied interface. It's worth restating here because it's the foundation for everything in this guide:

```go
type error interface {
    Error() string
}
```

That's it. `error` is not a special language construct — it's an ordinary interface with a single method. **Any type** with an `Error() string` method is a valid error. This is why Go's error handling can be built entirely out of interfaces, functions, and normal control flow (`if` statements), with no dedicated exception syntax needed at all.

---

## 4. The Idiomatic Pattern: Check, Don't Catch

Here is the same `divide` function in Go:

```go
package main

import (
    "errors"
    "fmt"
)

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }
    return a / b, nil
}

func process(a, b float64) {
    result, err := divide(a, b)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Result:", result)
}

func main() {
    process(10, 2)
    process(10, 0)
}
```

Walk through what's different:

- **`error` is a normal return value**, occupying the same "multiple return values" mechanism you learned in Guide 2. `divide` returns `(float64, error)` — a result *and* an error, every time.
- **On success, `err` is `nil`.** There's no separate "success path" type — the function always returns both values; the caller is responsible for checking whether `err` is `nil` before trusting `result`.
- **`if err != nil { ... }` is the single most common idiom in all of Go.** You will type this exact shape hundreds of times. It's not boilerplate to be abstracted away — it's the language's entire error-handling mechanism, in plain sight, at every call site.
- **Nothing propagates automatically.** If `process` doesn't check `err`, the zero-value `result` (`0`) is simply used as if it were valid — silently. This is the sharp edge of the model: Go trades Python's "impossible to accidentally ignore an unhandled exception" (it crashes loudly) for "impossible to accidentally *not* have compiler-checked access to the error" (you must name the variable, but nothing forces you to act on it).
- **Manual propagation.** If a function wants to pass an error up to *its* caller instead of handling it, it does so explicitly, by returning it:

```go
func processAndReturn(a, b float64) (float64, error) {
    result, err := divide(a, b)
    if err != nil {
        return 0, err // explicit propagation — nothing happens "automatically"
    }
    return result, nil
}
```

Every single layer of the call stack must explicitly decide: handle this error here, or return it further up. There is no equivalent of an exception silently sailing through five stack frames untouched — each frame's choice to propagate is visible in its source code.

**Why Go designed it this way:** the language designers' stated rationale is that exceptions encourage treating error handling as an afterthought — code reads cleanly along the "happy path," with error handling bolted on in a separate `except` block, often far from where the error actually occurred. Go's philosophy is that errors are just as much a part of a function's normal behavior as its successful return value, and the syntax should reflect that by making error checks unavoidable to *see*, even if not unavoidable to *act on*.

---

## 5. Creating Errors: `errors.New` and `fmt.Errorf`

Two standard-library functions cover the majority of error creation:

```go
import (
    "errors"
    "fmt"
)

// Simple, static message
err1 := errors.New("connection failed")

// Formatted message, built from dynamic values
userID := 42
err2 := fmt.Errorf("user %d not found", userID)
```

`errors.New` takes a plain string. `fmt.Errorf` works like `fmt.Sprintf` (formatting values into a template) but returns an `error` instead of a `string`. In practice, `fmt.Errorf` is used far more often, since real error messages almost always need to include some contextual value (an ID, a filename, a status code).

The Python equivalent of "creating an error" is instantiating an exception without raising it yet: `err = ValueError(f"user {user_id} not found")`. The difference is what happens next — in Python, you'd typically `raise` it immediately; in Go, you `return` it as a value.

---

## 6. Sentinel Errors and `errors.Is`

A **sentinel error** is a specific, pre-declared error value that callers can check for by identity — Go's rough equivalent of defining a specific, well-known exception class that callers check for with `except SpecificError`.

```go
package store

import "errors"

var ErrNotFound = errors.New("item not found")

func GetItem(id int) (string, error) {
    if id != 1 {
        return "", ErrNotFound
    }
    return "widget", nil
}
```

By convention, sentinel errors are named starting with `Err` and are exported (capitalized) so other packages can reference them. Callers check for a specific sentinel using `errors.Is`:

```go
item, err := store.GetItem(99)
if errors.Is(err, store.ErrNotFound) {
    fmt.Println("that item doesn't exist")
} else if err != nil {
    fmt.Println("unexpected error:", err)
}
```

**Why `errors.Is` instead of `==`?** For a plain sentinel like this, `err == store.ErrNotFound` would actually work. But `errors.Is` also correctly unwraps *wrapped* errors (covered in Section 8) to find a match anywhere in the chain — `==` would fail the moment the error gets wrapped even once. Using `errors.Is` from the start means your comparisons keep working even if the error later passes through additional layers of wrapping. It's directly analogous to always preferring `isinstance(e, SpecificError)` over comparing exception objects by identity — the idiomatic tool that keeps working as code evolves around it.

### Python Comparison

```python
class ItemNotFoundError(Exception):
    pass

def get_item(item_id: int) -> str:
    if item_id != 1:
        raise ItemNotFoundError("item not found")
    return "widget"

try:
    item = get_item(99)
except ItemNotFoundError:
    print("that item doesn't exist")
except Exception as e:
    print("unexpected error:", e)
```

A Python custom exception class and a Go sentinel error value are solving the same problem — "let the caller check for *this specific* failure mode" — but notice the mechanism differs: Python distinguishes error *kinds* via the exception's **class** (checked with `except SomeClass` / `isinstance`), while Go's sentinel pattern distinguishes error kinds via a specific **value's identity** (checked with `errors.Is`, which under the hood does an equality-style comparison, unwrapping as needed). Custom error *types* (next section) bring Go a step closer to Python's class-based model.

---

## 7. Custom Error Types and `errors.As`

Sometimes an error needs to carry structured data, not just a message — a validation error might need to carry the field name, an HTTP error might need to carry a status code. For this, define your own type that satisfies the `error` interface.

```go
package main

import (
    "errors"
    "fmt"
)

type ValidationError struct {
    Field string
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Msg)
}

func validateAge(age int) error {
    if age < 0 {
        return &ValidationError{Field: "age", Msg: "must be non-negative"}
    }
    return nil
}
```

Because `ValidationError` implements `Error() string` (recall from Guide 4: implicit satisfaction), `*ValidationError` is usable anywhere `error` is expected. To recover the structured data at the call site, use `errors.As`:

```go
err := validateAge(-5)

var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Println("Field with problem:", valErr.Field)
    fmt.Println("Details:", valErr.Msg)
} else if err != nil {
    fmt.Println("some other error:", err)
}
```

`errors.As` takes a pointer to a variable of the target error type, and — if `err` (or anything it wraps) matches that type — populates the variable and returns `true`. This is functionally very close to Python's:

```python
class ValidationError(Exception):
    def __init__(self, field: str, msg: str):
        self.field = field
        self.msg = msg
        super().__init__(f"validation failed on {field}: {msg}")

def validate_age(age: int) -> None:
    if age < 0:
        raise ValidationError("age", "must be non-negative")

try:
    validate_age(-5)
except ValidationError as e:
    print("Field with problem:", e.field)
    print("Details:", e.msg)
```

**`errors.Is` vs. `errors.As` — the rule of thumb:**

| Use... | When you want to check... | Python analogue |
|---|---|---|
| `errors.Is(err, target)` | "Is this *specific* error (a sentinel value) present anywhere in the chain?" | `isinstance(e, SpecificSentinelLikeClass)` / identity check |
| `errors.As(err, &target)` | "Is there an error of *this type* anywhere in the chain, and if so, give it to me so I can access its fields?" | `except SomeErrorClass as e:` then using `e`'s attributes |

---

## 8. Wrapping Errors with `%w`

Real programs have layers — a database call fails inside a repository function, which is called by a service function, which is called by an HTTP handler. Each layer often wants to add context ("failed to load user") *without losing* the original underlying error, so that callers further up can still check for the root cause.

`fmt.Errorf` supports a special verb, `%w`, that **wraps** an existing error inside a new one, preserving the chain:

```go
package main

import (
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("not found")

func fetchUser(id int) error {
    if id != 1 {
        return ErrNotFound
    }
    return nil
}

func loadUserProfile(id int) error {
    err := fetchUser(id)
    if err != nil {
        return fmt.Errorf("loading profile for user %d: %w", id, err)
    }
    return nil
}

func main() {
    err := loadUserProfile(99)
    fmt.Println(err) // "loading profile for user 99: not found"

    // errors.Is still finds ErrNotFound, even through the wrap:
    if errors.Is(err, ErrNotFound) {
        fmt.Println("root cause was: not found")
    }
}
```

The printed message reads like a breadcrumb trail — each layer's added context, innermost error last — while `errors.Is`/`errors.As` can still see straight through to `ErrNotFound` at the bottom of the chain. This works because wrapping with `%w` creates an error that also implements an (unexported, implicit) `Unwrap() error` method, and `errors.Is`/`errors.As` both call `Unwrap` repeatedly until they find a match or run out of chain.

**Contrast with `%v` or plain string concatenation:** if you instead wrote `fmt.Errorf("loading profile for user %d: %v", id, err)` (note `%v`, not `%w`), you'd get the same printed message, but the chain would be broken — `errors.Is(err, ErrNotFound)` would return `false`, because there's no `Unwrap` path back to the original error. **Always use `%w` when you want the original error to remain checkable**, and `%v` only when you deliberately want to discard that traceability (rare).

### Python Comparison

Python's rough equivalent is *exception chaining* via `raise ... from ...`:

```python
def fetch_user(user_id: int) -> None:
    if user_id != 1:
        raise LookupError("not found")

def load_user_profile(user_id: int) -> None:
    try:
        fetch_user(user_id)
    except LookupError as e:
        raise RuntimeError(f"loading profile for user {user_id}") from e
```

Python's `from e` sets `__cause__` on the new exception, which shows up in the traceback as "The above exception was the direct cause of the following exception." It's a similar instinct — preserve the original failure while adding context — but the ergonomics differ: Python's chained exception is inspected via `e.__cause__` manually, whereas Go's `errors.Is`/`errors.As` walk the whole chain for you automatically, which is why wrapping with `%w` and checking with `errors.Is`/`errors.As` feels like a much more integrated, first-class part of the language than exception chaining tends to feel in typical Python code.

---

## 9. Combining Multiple Errors: `errors.Join`

Sometimes several independent operations can each fail, and you want to report all of them at once rather than stopping at the first. Since Go 1.20, `errors.Join` combines multiple errors into one:

```go
func validateForm(name string, age int) error {
    var errs []error
    if name == "" {
        errs = append(errs, errors.New("name is required"))
    }
    if age < 0 {
        errs = append(errs, errors.New("age must be non-negative"))
    }
    return errors.Join(errs...) // returns nil if errs is empty
}
```

The joined error's message concatenates each sub-error's message on its own line, and — importantly — `errors.Is` and `errors.As` both search *all* of the joined errors, not just the first. This maps loosely to Python's `ExceptionGroup` (added in Python 3.11) for representing multiple concurrent failures, though Python's version is more heavyweight (it's a real exception you raise and catch with `except*`), while `errors.Join` produces an ordinary `error` value handled with the same `if err != nil` idiom as everything else.

---

## 10. `panic` and `recover`

Everything above — `error`, wrapping, `errors.Is`/`errors.As` — is Go's *normal*, expected error-handling path. `panic` is different: it's Go's mechanism for **unrecoverable, exceptional** conditions, and it behaves much more like an uncaught Python exception than anything discussed so far.

```go
func mustPositive(n int) int {
    if n < 0 {
        panic("n must be positive")
    }
    return n
}

func main() {
    fmt.Println(mustPositive(5))
    fmt.Println(mustPositive(-1)) // panics here
    fmt.Println("this line never runs")
}
```

When `panic` is called: execution of the current function stops immediately, deferred calls (see Guide on `defer`, or note: `defer` schedules a function call to run when the surrounding function returns — covered in full in a later guide, but used here for `recover`) still run, and the panic propagates up the call stack, unwinding each frame, until either something calls `recover` (next section) or the program reaches the top and crashes with a stack trace printed to stderr.

This *is* genuinely analogous to an unhandled Python exception unwinding the stack — the difference is that in idiomatic Go, this path is reserved for situations that indicate a **bug**, not situations a well-behaved program should ever routinely encounter. A missing config file is an `error`. An index out of bounds on a slice, or a nil pointer dereference, or a program invariant being violated — those panic, because they represent a programming mistake, not an expected failure mode.

`recover` stops a panic from continuing to propagate, but it has a strict usage rule: **it only has an effect when called directly inside a `defer`red function.** Calling it anywhere else does nothing.

```go
func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()
    result = a / b // panics if b == 0: "integer divide by zero"
    return result, nil
}

func main() {
    result, err := safeDivide(10, 0)
    if err != nil {
        fmt.Println("Error:", err) // Error: recovered from panic: integer divide by zero
    } else {
        fmt.Println("Result:", result)
    }
}
```

Notice the pattern: `safeDivide` uses **named return values** (`result int, err error` — first seen conceptually in Guide 2) specifically so the deferred function can *modify* `err` after a panic, converting what would have been a crash into an ordinary returned `error` that the caller checks the usual way. This is the standard idiom for "panic containment" at a boundary — often used at the top of a goroutine (Phase C) or an HTTP handler, so one panicking request doesn't take down the entire process.

---

## 11. The `defer` + `recover` Pattern

Since `recover` only works inside a deferred function, the pattern is almost always structured the same way, regardless of context:

```go
func riskyOperation() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("riskyOperation panicked: %v", r)
        }
    }()

    // ... code that might panic ...
    return nil
}
```

A few details worth internalizing:

- `defer func() { ... }()` — an anonymous function, deferred, called with `()` immediately (the call is what gets deferred, not the definition).
- `recover()` returns `nil` if there was no panic — meaning this deferred function is completely harmless to include even when nothing goes wrong; the `if r := recover(); r != nil` guard means the body only does anything when a panic actually occurred.
- Only the *nearest* enclosing deferred `recover()` stops the unwinding — a panic will keep propagating past any function that doesn't have a `recover` in a `defer`, exactly like an exception passing through Python stack frames with no matching `except`.

This is the closest Go gets to a `try`/`except`/`finally` block, but the framing matters: this is not how you handle "the user typed an invalid age." It's how you build a **safety net at a boundary**, so a genuine bug doesn't crash something it shouldn't (like an entire web server, when only one request handler misbehaved).

---

## 12. When to Panic vs. Return an Error

This is the single most important judgment call this guide covers, so it deserves a direct rule of thumb:

**Return an `error` when:**
- The failure is *expected* as part of normal operation — a file might not exist, a network call might time out, a user might submit invalid input.
- The caller has a reasonable chance of handling it meaningfully (retry, show a message, fall back to a default).
- You're writing a library function that other code will call — libraries should almost never panic on bad input; they should return an error and let the caller decide what to do.

**Panic when:**
- The condition represents a **programmer error**, not a runtime/environment failure — e.g., a function precondition was violated in a way that should be impossible if the code calling it is correct (an index that should never be negative but is).
- Continuing execution would be unsafe or meaningless — corrupted internal state, a required invariant broken.
- You're in `main()` or an initialization path where there's genuinely no way to proceed (some Go code uses `panic` for unrecoverable startup failures, though even here, many idiomatic codebases still prefer logging the error and calling `os.Exit(1)`).

**A useful gut check, mapped to Python instincts:** if you'd naturally reach for a custom exception class and expect *some* caller, somewhere, to catch it and respond — that's an `error` in Go. If you'd only ever expect this to be caught by a top-level "something is very wrong" handler (or not caught at all, because it indicates a bug that needs fixing, not handling) — that's a `panic`. Go libraries and idiomatic code lean *heavily* toward the `error` side of this line; reaching for `panic` as a general-purpose "something went wrong" tool is considered a mistake carried over from languages with pervasive exceptions, and is one of the fastest ways to write non-idiomatic Go.

---

## 13. Python vs. Go: Side-by-Side Summary

| Concept | Python | Go |
|---|---|---|
| Default error-handling model | Exceptions (`raise`/`except`), automatic propagation | Returned `error` values, explicit propagation |
| Visible in function signature? | No — a function's exceptions are undocumented unless noted in a docstring | Yes — `(T, error)` return type makes fallibility explicit |
| Creating an error | Instantiate an exception class | `errors.New("msg")` or `fmt.Errorf("msg: %v", val)` |
| Checking for a specific known failure | `except SpecificExceptionClass:` | `errors.Is(err, SentinelErr)` |
| Extracting structured error data | `except CustomError as e:` then `e.attribute` | `errors.As(err, &target)` then `target.Field` |
| Adding context while preserving the original | `raise NewError(...) from original` | `fmt.Errorf("context: %w", err)` |
| Combining several failures | `ExceptionGroup` + `except*` (3.11+) | `errors.Join(err1, err2, ...)` |
| Guaranteed cleanup | `finally` block | `defer` (general-purpose, not error-specific — later guide covers it fully) |
| Truly exceptional/unrecoverable failure | Uncaught exception crashes the program | `panic` — unwinds the stack unless `recover`ed inside a `defer` |
| Containing a crash at a boundary | Top-level `try`/`except Exception` | `defer` + `recover()`, converting panic into a returned `error` |
| Idiomatic default for "something failed" | Raise an exception | Return an `error` — panic is the exception (pun intended), not the rule |

---

## 14. Exercises

1. Write a function `ParsePositiveInt(s string) (int, error)` that converts a string to an int (use `strconv.Atoi`) and returns an error if the string isn't a valid number *or* if the resulting number is negative. Wrap the underlying `strconv.Atoi` error with `%w` and additional context.
2. Define a sentinel error `ErrInsufficientFunds` and a function `Withdraw(balance, amount float64) (float64, error)` that returns it when `amount > balance`. Write a caller that uses `errors.Is` to print a friendly message specifically for that case.
3. Define a custom error type `RangeError` with `Min`, `Max`, and `Value` fields, satisfying the `error` interface with a descriptive `Error()` message. Write a function that returns it, and a caller that uses `errors.As` to print the `Min`/`Max`/`Value` fields individually.
4. Write a function `SafeCall(f func())` that calls `f` and uses `defer` + `recover` to catch any panic from `f`, printing "recovered: <value>" instead of crashing. Test it by passing in a function that panics and one that doesn't.
5. Write a function `ValidateUser(name string, age int) error` that checks multiple conditions (name non-empty, age non-negative, age under some max) and uses `errors.Join` to return all violations at once rather than stopping at the first.

---

## 15. Key Takeaways

- **This is the single biggest mindset shift from Python:** Go errors are ordinary return values checked explicitly (`if err != nil`), not exceptions that propagate automatically until caught.
- `error` is just an interface with one method, `Error() string` — anything satisfying it (implicitly, per Guide 4) is a valid error.
- Create errors with `errors.New` (static messages) or `fmt.Errorf` (formatted messages).
- **Sentinel errors** (`errors.Is`) are Go's rough equivalent of checking `isinstance(e, SpecificClass)` for a known failure mode; **custom error types** (`errors.As`) are the equivalent of custom exception classes carrying structured data.
- **Wrap errors with `%w`**, not `%v`, whenever you want the original error to remain discoverable via `errors.Is`/`errors.As` further up the call stack — this is Go's version of Python's `raise ... from ...` exception chaining, but with automatic chain-walking built into the standard library.
- `panic`/`recover` is Go's actual exception-like mechanism, but it is reserved for programmer errors and unrecoverable conditions — not for everyday, expected failures. Reaching for `panic` where an `error` return would do is one of the clearest tells of non-idiomatic Go, and the pull to do so is usually residual Python/exception-language instinct.
- The standard containment pattern — `defer` + `recover()` inside a function with named return values — converts a would-be crash into an ordinary `error` at a chosen boundary, without making panic a routine part of your program's control flow.

---

*Next: Guide 6 begins Phase C — Concurrency, starting with goroutines and the fundamentals of Go's concurrency model.*
