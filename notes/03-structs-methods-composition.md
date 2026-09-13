# Guide 3 — Structs, Methods & Composition

> **Roadmap position:** Phase B, Guide 3 of 14 — *Core language, hands-on*
> **Goal of this guide:** Learn how Go models data and behavior without classes. Structs give you Python's class-level data modeling; methods attach behavior to types without `self`-based class bodies; and embedding gives you code reuse without inheritance. By the end, you'll be able to design a domain layer (`User`, `Order`, `Product`-style types) the idiomatic Go way — which will look and feel different from a Python class hierarchy, on purpose.

---

## Table of Contents

1. [How This Guide Builds on Guides 1–2](#1-how-this-guide-builds-on-guides-12)
2. [Structs: Declaring Go's Data Containers](#2-structs-declaring-gos-data-containers)
3. [Struct Literals and Zero-Valued Structs](#3-struct-literals-and-zero-valued-structs)
4. [Methods: Attaching Behavior to Types](#4-methods-attaching-behavior-to-types)
5. [Value Receivers vs Pointer Receivers](#5-value-receivers-vs-pointer-receivers)
6. [When to Use Pointer vs Value Receivers](#6-when-to-use-pointer-vs-value-receivers)
7. [Struct Embedding — Composition Mechanics](#7-struct-embedding--composition-mechanics)
8. [Method and Field Promotion](#8-method-and-field-promotion)
9. [Composition Over Inheritance, Contrasted With Python](#9-composition-over-inheritance-contrasted-with-python)
10. [Struct Tags and Anonymous Structs](#10-struct-tags-and-anonymous-structs)
11. [Struct Comparison and Copying Semantics](#11-struct-comparison-and-copying-semantics)
12. [Constructor Patterns — Go Has No `__init__`](#12-constructor-patterns--go-has-no-init)
13. [Summary & What's Next](#13-summary--whats-next)

---

## 1. How This Guide Builds on Guides 1–2

You now know Go's built-in types (`int`, `string`, slices, maps) and every control-flow/function form the language has. Everything so far, though, has been about *values* — nothing yet resembles how you'd model a `User` or an `Order` the way you would with a Pydantic model or a plain Python class.

Two things from Guide 2 are directly load-bearing here:

- **Arrays are value types, copied on assignment** (Guide 2, §8.1). Structs behave the same way — this is the single most important carry-over concept in this guide, and it directly motivates the pointer-receiver discussion in §5.
- **Zero values apply to every type, including custom ones** (Guide 2, §3). A struct's zero value is every field set to *its own* zero value — this will matter constantly once you start writing domain types.

---

## 2. Structs: Declaring Go's Data Containers

A `struct` is a typed collection of named fields — Go's equivalent of the *data* half of a Python class (with none of the behavior baked in; behavior is attached separately via methods, §4).

```go
type User struct {
    Name  string
    Email string
    Age   int
}
```

Compare to the Python you're used to:

```python
class User:
    def __init__(self, name: str, email: str, age: int):
        self.name = name
        self.email = email
        self.age = age
```

The Go version has **no `__init__`, no `self`, and no method bodies at all** — `type User struct {...}` declares *only* the shape of the data. This is a deliberate separation: in Go, a type's data definition and its behavior (methods) are always visually and structurally distinct, even though they're conceptually linked. You'll get used to scanning a file for `type X struct` to understand *what data* a type holds, and separately scanning for `func (x X) MethodName()` to understand *what it can do* — these often live in the same file, but never in the same block.

Recall Guide 1 (§10): field name capitalization controls visibility exactly like any other identifier. `Name`, `Email`, `Age` above are all exported (accessible from other packages); a field like `passwordHash string` (lowercase) would be invisible outside the package that declares `User` — this is Go's built-in equivalent of Python's "leading underscore" convention, except compiler-enforced.

---

## 3. Struct Literals and Zero-Valued Structs

### 3.1 Keyed Struct Literals (Always Prefer This Form)

```go
u := User{
    Name:  "Ana Silva",
    Email: "ana@example.com",
    Age:   28,
}
```

Fields you omit simply take their zero value:

```go
u := User{Name: "Ana Silva"} // Email is "", Age is 0
```

### 3.2 Positional Struct Literals (Avoid These)

```go
u := User{"Ana Silva", "ana@example.com", 28} // works, but fragile
```

This compiles, but it's considered poor style outside of very small, stable structs (or test helper code) — if someone reorders the struct's fields later, every positional literal silently breaks in a way the compiler *won't* catch if the types happen to still line up (e.g., two `string` fields swapped). `go vet` and linters will flag this pattern in most real codebases. **Always use keyed literals in application code.**

### 3.3 The Zero-Valued Struct

Just like any other type (Guide 2, §3), a struct declared with `var` and no literal gets every field set to its zero value:

```go
var u User
fmt.Println(u) // { 0}  -- Name="", Email="", Age=0
```

This is genuinely useful — you'll often declare a zero-valued struct and fill it in field by field, or as the target of a database scan or JSON decode (Guide 9), relying on the fact that a struct is *always* in a valid, safe-to-reference state even before you've populated it — there's no equivalent to Python's `AttributeError: 'User' object has no attribute 'age'` from a half-constructed object.

---

## 4. Methods: Attaching Behavior to Types

A method in Go is just a function with an extra piece of syntax before its name: a **receiver**, which is how Go attaches a function to a specific type without needing a `class` block at all.

```go
type User struct {
    Name string
    Age  int
}

func (u User) Greeting() string {
    return fmt.Sprintf("Hi, I'm %s", u.Name)
}

user := User{Name: "Ana Silva", Age: 28}
fmt.Println(user.Greeting()) // Hi, I'm Ana Silva
```

Breaking down `func (u User) Greeting() string`:

- `(u User)` is the **receiver** — it declares that `Greeting` is a method on type `User`, and that inside the method body, `u` refers to the specific `User` instance the method was called on. This is Go's `self` — except it's not a hidden implicit first parameter like Python's `self`; it's an explicit, named part of the function signature that you choose the name for (convention: a short, often one- or two-letter abbreviation of the type name — `u` for `User`, `s` for `Server`, etc., not `self` or `this`).
- Unlike Python, **methods are not defined inside the type's body** — `Greeting` can be declared anywhere in the same package, even in a completely different file, as long as its receiver type is `User`. This is a direct consequence of Go's "a package is a directory" model from Guide 1 (§9–10) and Guide 2 (§10.1): methods on a type can be spread across multiple files in the same package, commonly organized by *behavior* (e.g., `user_validation.go`, `user_auth.go`) rather than forced into one big class body.

---

## 5. Value Receivers vs Pointer Receivers

This is the most important — and most consequential — concept in this entire guide, and it flows directly from "structs are value types, copied on assignment" (§1, recalling Guide 2 §8.1).

### 5.1 Value Receivers: You Get a Copy

```go
func (u User) Birthday() {
    u.Age++ // this mutates the COPY, not the original
}

user := User{Name: "Ana", Age: 28}
user.Birthday()
fmt.Println(user.Age) // 28 — unchanged! Birthday() modified a throwaway copy.
```

This is the single most common early Go bug for developers coming from Python or any reference-semantics-by-default language: `Birthday()` silently does nothing observable, and the compiler gives you **no warning at all**, because this is perfectly legal Go — you just didn't want it.

### 5.2 Pointer Receivers: You Get the Real Thing

```go
func (u *User) Birthday() {
    u.Age++ // mutates the ACTUAL struct the pointer points to
}

user := User{Name: "Ana", Age: 28}
user.Birthday()
fmt.Println(user.Age) // 29 — correctly mutated
```

Notice something important: you called `user.Birthday()`, not `(&user).Birthday()`, even though `Birthday` has a pointer receiver (`*User`) and `user` is a plain value. **Go automatically takes the address of `user` for you** when calling a pointer-receiver method on an *addressable* value (a local variable, a struct field, etc.) — this auto-addressing is a genuine convenience that removes a lot of the manual `&`/`*` ceremony you might expect from a language with explicit pointers. You'll only need to write `&` and `*` explicitly in a handful of recurring situations (constructing a pointer to a new struct, dereferencing in certain expressions) — covered as they come up in later guides.

### 5.3 Why the Compiler Doesn't Warn You About §5.1

This is worth sitting with because it genuinely surprises people: `func (u User) Birthday()` mutating `u.Age` is **not a bug from the compiler's perspective** — you explicitly asked for a copy (via the value receiver), and you mutated the copy. The compiler did exactly what you told it to do. The "bug" is purely a mismatch between what you *intended* (mutate the real object) and what you *wrote* (a value receiver, which can never do that). This is precisely why the guideline in §6 exists — and why experienced Go developers apply it almost mechanically rather than deciding case-by-case.

---

## 6. When to Use Pointer vs Value Receivers

The Go community has converged on a small set of clear guidelines — memorize these, since they'll govern nearly every type you write from here through your REST API work in Guides 10–14.

**Use a pointer receiver (`*T`) when:**

- The method needs to **mutate the receiver's state** (as in §5.2) — this is the most common reason.
- The struct is **large**, and copying it on every method call would be wasteful (structs with many fields, or fields like large arrays — recall arrays are copied by value, Guide 2 §8.1).
- You need **consistency**: if *any* method on a type needs a pointer receiver, idiomatic Go convention is to make **all** methods on that type use pointer receivers, even ones that don't strictly need to mutate anything. This avoids a confusing mix where some calls on the same variable behave differently regarding mutation.

**Use a value receiver (`T`) when:**

- The struct is **small and naturally immutable in spirit** (e.g., a `Point{X, Y int}` or a `Money{Amount int, Currency string}` value you never mutate in place, only replace).
- You genuinely want copy semantics — e.g., value types that should behave like Go's built-in `int` or `string` (copied freely, no shared mutable state to worry about).

**A critical downstream consequence — interface satisfaction (previewed here, covered fully in Guide 4):** a method set defined with pointer receivers is only satisfied by a `*T`, **not** by a plain `T` — meaning if `Greeting()` has a pointer receiver, a bare `User{}` value (not `&User{}`) will **not** satisfy an interface requiring `Greeting()`. This single rule is one of the most common early Go compiler errors you'll hit once Guide 4 introduces interfaces, so it's worth previewing now: **default to pointer receivers for any struct that represents a "thing with identity and behavior"** (users, services, repositories, database connections) — which, not coincidentally, describes almost every type you'll build in your REST API.

### 6.1 A Practical Rule of Thumb for This Roadmap

Given where you're headed (Guides 10–14: handlers, services, repositories — all inherently stateful, behavior-carrying types), the practical rule you'll apply almost every time is:

> **Domain/service/repository-style structs → pointer receivers, always.**
> **Small, literal "value object" style structs (rare in typical REST API code, more common in specialized domains) → value receivers.**

When genuinely unsure, default to pointer receivers — it's the safer, more common default in real-world Go codebases and avoids the entire class of bug shown in §5.1.

---

## 7. Struct Embedding — Composition Mechanics

Go has **no `extends` keyword and no class hierarchy**. Instead, it gives you **embedding**: placing one struct type inside another, without naming a field for it.

```go
type Person struct {
    Name string
    Age  int
}

type Employee struct {
    Person       // embedded — no field name given, just the type
    CompanyName string
    Salary      float64
}
```

Construct it like this:

```go
emp := Employee{
    Person:      Person{Name: "Ana Silva", Age: 28},
    CompanyName: "Acme Corp",
    Salary:      75000,
}
```

The embedded field is still addressable by its type name if needed (`emp.Person.Name`), but embedding's real value is **promotion**, covered next.

---

## 8. Method and Field Promotion

Because `Person` is embedded (not just referenced as a named field), Go **promotes** its fields and methods up to `Employee` automatically — meaning you can access them as if they were declared directly on `Employee`:

```go
fmt.Println(emp.Name)  // "Ana Silva" — promoted from Person, no need for emp.Person.Name
fmt.Println(emp.Age)   // 28 — same
```

This works for methods too:

```go
func (p Person) Greeting() string {
    return fmt.Sprintf("Hi, I'm %s", p.Name)
}

fmt.Println(emp.Greeting()) // "Hi, I'm Ana Silva" — Employee didn't define Greeting(), it was promoted
```

### 8.1 Overriding a Promoted Method

`Employee` can define its own method with the same name, which simply **shadows** the promoted one (no `super()`/`override` keyword needed or available — it's purely a name-resolution rule):

```go
func (e Employee) Greeting() string {
    return fmt.Sprintf("Hi, I'm %s from %s", e.Name, e.CompanyName)
}

fmt.Println(emp.Greeting()) // uses Employee's version now, not Person's
```

If you specifically need the embedded type's original version after overriding, you call it explicitly through the field name: `emp.Person.Greeting()`.

### 8.2 What Embedding Does *Not* Give You

This is the critical conceptual boundary, and it's where "composition over inheritance" (§9) stops being an abstract slogan and becomes a concrete rule you need to internalize:

- **There is no polymorphism through embedding alone.** A function that accepts a `Person` parameter cannot accept an `Employee` in its place, even though `Employee` embeds `Person` — Go does **not** consider `Employee` to be "a kind of" `Person` for type-checking purposes, unlike Python's `isinstance(employee, Person)` returning `True` under real inheritance. Embedding is purely a mechanism for **code reuse and field/method promotion**, not a type-hierarchy or "is-a" relationship. If you need actual polymorphic behavior (treating different concrete types uniformly through a shared contract), that's what **interfaces** are for — the entire subject of Guide 4, and the real mechanism Go uses in place of inheritance-based polymorphism.
- **You can embed multiple types**, and Go has clear (if occasionally surprising) rules for resolving ambiguous promoted names — if two embedded types both have a `Name` field, you must disambiguate explicitly (`emp.Person.Name`) since Go won't guess which one you meant.

---

## 9. Composition Over Inheritance, Contrasted With Python

You've likely built Python class hierarchies like this at some point:

```python
class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        raise NotImplementedError

class Dog(Animal):
    def speak(self):
        return f"{self.name} says Woof"

class Cat(Animal):
    def speak(self):
        return f"{self.name} says Meow"
```

This relies on **inheritance**: `Dog` and `Cat` genuinely *are* `Animal`s, `isinstance()` confirms it, and you can pass either wherever an `Animal` is expected.

Go deliberately has none of this: no base classes, no `super()`, no method resolution order, no multiple inheritance diamond problem to reason about. Instead, Go's philosophy (recall Guide 1, §2.3) is:

> **Build behavior by composing small, focused pieces together (embedding, §7–8), and describe shared *capabilities* — not shared *ancestry* — through interfaces (Guide 4).**

The equivalent idiomatic Go design for the example above wouldn't use embedding at all (since `Dog` and `Cat` don't share reusable *data*, just a shared *capability* to "speak") — it would use an interface:

```go
type Speaker interface {
    Speak() string
}

type Dog struct{ Name string }
func (d Dog) Speak() string { return d.Name + " says Woof" }

type Cat struct{ Name string }
func (c Cat) Speak() string { return c.Name + " says Meow" }

func Announce(s Speaker) {
    fmt.Println(s.Speak())
}

Announce(Dog{Name: "Rex"})  // works
Announce(Cat{Name: "Tom"})  // works — no shared base type required at all
```

Neither `Dog` nor `Cat` extends anything or embeds anything here — they simply each independently implement a `Speak() string` method, and that alone is enough to satisfy `Speaker`. This is the essence of Go's approach, formally called **structural typing** (a type satisfies an interface automatically, just by having the right methods — no explicit "implements" declaration anywhere), and it's the entire subject of Guide 4.

**The practical takeaway for now:** reach for **embedding** (§7–8) when you have genuinely shared *data and default behavior* you want to reuse across types (e.g., every domain model embedding a common `BaseModel{ID string; CreatedAt time.Time}` for shared fields) — this is a very real, very common pattern you'll use in Guide 6 (databases). Reach for **interfaces** (Guide 4) when you need multiple, otherwise-unrelated types to be treated uniformly through a shared *contract* — the mechanism your REST API's handlers, services, and repositories will lean on constantly starting in Guide 12.

---

## 10. Struct Tags and Anonymous Structs

### 10.1 Struct Tags — A Quick Preview

You'll see struct fields written with a trailing backtick-quoted string like this constantly once you reach Guide 9 (JSON) and Guide 6 (databases):

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age,omitempty"`
}
```

These are **struct tags** — plain string metadata attached to a field, read via reflection by libraries like `encoding/json` to control how a field is serialized (here: use lowercase JSON key names instead of Go's capitalized field names, and omit `Age` from the JSON output entirely if it's zero-valued). Nothing about tags is enforced by the compiler itself — they're just conventionally-formatted strings that specific libraries know how to parse. We're flagging their existence now so the syntax doesn't feel unfamiliar when Guide 9 covers them properly; no need to memorize tag syntax yet.

### 10.2 Anonymous Structs

Occasionally useful for throwaway, one-off data shapes that don't deserve a named type — most commonly in tests or small internal helper functions:

```go
point := struct {
    X int
    Y int
}{X: 3, Y: 4}

fmt.Println(point.X, point.Y) // 3 4
```

You won't reach for this often in application code (a named `type Point struct {...}` is almost always clearer and reusable), but you'll encounter it in table-driven tests in Guide 8, where an anonymous struct is the idiomatic way to define a slice of test cases inline.

---

## 11. Struct Comparison and Copying Semantics

### 11.1 Structs Are Comparable With `==` (Usually)

Unlike Python, where `==` on a plain class instance checks identity by default unless you implement `__eq__`, Go structs support `==`/`!=` **automatically**, comparing every field, **as long as every field's type is itself comparable**:

```go
type Point struct{ X, Y int }

p1 := Point{1, 2}
p2 := Point{1, 2}
fmt.Println(p1 == p2) // true — field-by-field comparison, no method needed
```

**But** if a struct contains a slice, map, or function field, it becomes **non-comparable**, and `==` becomes a **compile error**, not a runtime false:

```go
type Group struct {
    Name    string
    Members []string // slice field
}

g1 := Group{Name: "A", Members: []string{"x"}}
g2 := Group{Name: "A", Members: []string{"x"}}
g1 == g2 // compile error: struct containing []string cannot be compared
```

This is worth remembering now because it will directly affect how you write assertions in tests (Guide 8) — comparing structs with slice/map fields requires `reflect.DeepEqual` or a testing library helper instead of plain `==`.

### 11.2 Copying Semantics Recap

Recall §5: passing a struct **by value** (to a function, or via a value-receiver method) copies the entire struct, field by field — including, importantly, copying a slice/map field's **header**, not deep-copying its underlying data (since slices and maps are themselves reference-like, per Guide 2 §8.5 and §9). This means: copying a struct with a slice field gives you two structs whose slice fields still point at the *same* underlying array — mutating elements through one is visible through the other, exactly like the slice-aliasing gotcha from Guide 2. Struct value semantics and slice/map reference semantics compose together, and this combination is worth re-reading Guide 2 §8.5 to make sure it's solid before moving on.

---

## 12. Constructor Patterns — Go Has No `__init__`

Go has no constructor syntax at all — no special method the compiler calls automatically when you create an instance. The idiomatic replacement is a plain function, conventionally named `New` or `NewX`, that returns either a value or (far more commonly, per §6) a pointer:

```go
type User struct {
    Name  string
    Email string
    Age   int
}

func NewUser(name, email string, age int) *User {
    return &User{
        Name:  name,
        Email: email,
        Age:   age,
    }
}

user := NewUser("Ana Silva", "ana@example.com", 28)
```

`&User{...}` here constructs a `User` value and immediately takes its address, producing a `*User` — this is the standard idiom for "construct and return a pointer" you'll see in essentially every Go codebase, and it directly explains why so many of the function signatures you'll encounter from here on return pointer types.

A `New`-style constructor function also gives you a natural place to enforce invariants Python would normally validate inside `__init__` (e.g., raising `ValueError` for an invalid age) — in Go, this becomes a function that returns `(*User, error)` instead, following exactly the multiple-return-value pattern from Guide 2 (§7.2):

```go
func NewUser(name, email string, age int) (*User, error) {
    if age < 0 {
        return nil, errors.New("age cannot be negative")
    }
    return &User{Name: name, Email: email, Age: age}, nil
}
```

This shape — a `New` function returning `(*T, error)` — is exactly what you'll be writing constantly once Guide 5 covers error handling in full depth, and again in Guide 12 when constructing services and repositories with their dependencies wired in.

---

## 13. Summary & What's Next

**What you can now actually do:**

- Declare structs to model domain data, understanding that this is purely a data declaration with no attached behavior until you write methods separately.
- Write methods with the correct receiver syntax, and — critically — reason correctly about **value vs. pointer receivers**, including *why* a value-receiver mutation silently does nothing observable.
- Apply the community's pointer-vs-value-receiver guidelines confidently, defaulting to pointer receivers for anything resembling a service/domain/repository type.
- Use struct embedding for genuine code/data reuse, and understand method/field promotion — while clearly recognizing that embedding is **not** inheritance and gives you no polymorphism on its own.
- Explain, in concrete terms, why Go's "composition over interfaces, not inheritance" philosophy replaces the Python class-hierarchy patterns you're used to.
- Recognize struct tags on sight (fully useful starting in Guide 9) and know when an anonymous struct is appropriate.
- Understand struct comparison rules (`==` works field-by-field, except when a slice/map field makes the whole struct non-comparable) and how struct value semantics interact with slice/map reference semantics.
- Write idiomatic `New`/`NewX` constructor functions, including the `(*T, error)` returning form you'll use throughout the rest of this roadmap.

**What comes next — Guide 4: Interfaces & Polymorphism.** This guide deliberately left one thread open: §9 showed that Go achieves polymorphism through interfaces, not inheritance, but didn't yet show you how to *declare* one. Guide 4 covers Go's **implicit interface satisfaction** (a type satisfies an interface automatically just by having the right methods — no `implements` keyword anywhere), the `any` type (Go's `object`-like catch-all, formerly `interface{}`), type assertions and type switches (previewed briefly in Guide 2, §6.3), and — critically, tying directly back to §6 of this guide — exactly how pointer vs. value receivers determine which types satisfy which interfaces. This is the guide where Go's data modeling (structs, this guide) and Go's behavior-contracts (interfaces) finally connect into the full picture you'll use to design testable, swappable services starting in Guide 12.
