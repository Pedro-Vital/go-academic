In Go, **generics let you write functions, types, and data structures that work with multiple types while still being statically type-safe**.

They were introduced in **Go 1.18**.

The core idea is:

> **Write the logic once, parameterize the type.**

---

## 1. The problem generics solve

Imagine you want a function that returns the larger of two values.

Without generics, you might need separate functions:

```go
func MaxInt(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func MaxFloat(a, b float64) float64 {
    if a > b {
        return a
    }
    return b
}
```

The logic is identical. Only the type changes.

Generics allow you to write it once:

```go
func Max[T int | float64](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

Now:

```go
Max(10, 20)       // int
Max(3.14, 2.71)   // float64
```

`T` is a **type parameter**.

---

# 2. Type parameters

This syntax:

```go
func Max[T int | float64](a, b T) T
```

can be read as:

> Define a function `Max` with a type parameter `T`, where `T` can be `int` or `float64`.

Compare:

```go
func Max(a, b int) int
```

with:

```go
func Max[T int | float64](a, b T) T
```

The first has a concrete type:

```text
a → int
b → int
return → int
```

The second has a **type parameter**:

```text
a → T
b → T
return → T
```

and `T` is constrained to:

```text
int | float64
```

---

# 3. Type inference

Usually you don't need to explicitly tell Go what `T` is.

```go
Max(10, 20)
```

Go sees that both arguments are `int`, so it effectively determines:

```go
T = int
```

Similarly:

```go
Max(1.5, 2.5)
```

means:

```go
T = float64
```

You *can* explicitly specify it:

```go
Max[int](10, 20)
```

but normally inference makes that unnecessary.

---

# 4. Constraints

This is one of the most important concepts.

You can't simply write:

```go
func Max[T any](a, b T) T {
    if a > b { // ❌
        return a
    }
    return b
}
```

Why?

Because `any` means essentially:

> T can be **anything**.

Go cannot guarantee that arbitrary types support `>`.

For example:

```go
type User struct {
    Name string
}
```

What would this mean?

```go
user1 > user2
```

There's no general definition of `>` for structs.

So we need a **constraint** describing which types are allowed.

```go
func Max[T int | float64](a, b T) T
```

The constraint says:

```text
T ∈ { int, float64 }
```

Therefore Go knows that `>` is valid.

---

# 5. `any`

You'll frequently see:

```go
[T any]
```

`any` is an alias for:

```go
interface{}
```

So:

```go
func Print[T any](value T) {
    fmt.Println(value)
}
```

can accept any type:

```go
Print(10)
Print("hello")
Print(3.14)
Print(User{})
```

But remember:

**`any` provides almost no restrictions on T.**

Consequently, you can only perform operations that are valid without knowing anything specific about `T`.

For example:

```go
func Print[T any](value T) {
    fmt.Println(value)
}
```

works because `fmt.Println` accepts arbitrary values.

But:

```go
func Add[T any](a, b T) T {
    return a + b // ❌
}
```

doesn't work because arbitrary `T` values aren't necessarily addable.

---

# 6. Interfaces can be constraints

This becomes particularly interesting given what you've been studying about Go interfaces.

Suppose:

```go
type Number interface {
    int | float64
}
```

Then:

```go
func Max[T Number](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

Here `Number` isn't being used as a traditional runtime interface.

It's a **type constraint**.

This distinction is important:

### Normal interface

```go
type Animal interface {
    Speak()
}
```

Used to describe **behavior**:

```go
func MakeSound(a Animal) {
    a.Speak()
}
```

### Constraint interface

```go
type Number interface {
    int | float64
}
```

Used to describe **which types are allowed as type arguments**.

So generics introduced a new role for interfaces in Go.

---

# 7. The `~` operator

You'll eventually encounter:

```go
type Number interface {
    ~int | ~float64
}
```

The `~` means:

> The underlying type is `int` or `float64`.

For example:

```go
type Age int
type Price float64
```

These are distinct named types:

```go
Age
Price
```

Without `~`:

```go
type Number interface {
    int | float64
}
```

`Age` would **not** satisfy `Number`.

With:

```go
type Number interface {
    ~int | ~float64
}
```

both satisfy it:

```go
Age(20)
Price(10.5)
```

because:

```text
Age     → underlying type int
Price   → underlying type float64
```

---

# 8. Generic data structures

Generics aren't only for functions.

You can create generic types.

For example, a stack:

```go
type Stack[T any] struct {
    items []T
}
```

Now you can create:

```go
intStack := Stack[int]{}
stringStack := Stack[string]{}
```

And methods:

```go
func (s *Stack[T]) Push(value T) {
    s.items = append(s.items, value)
}

func (s *Stack[T]) Pop() T {
    value := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return value
}
```

Usage:

```go
numbers := Stack[int]{}

numbers.Push(10)
numbers.Push(20)

x := numbers.Pop()
```

Here `T` is part of the type itself.

So:

```text
Stack[int]
```

and:

```text
Stack[string]
```

are different instantiated types.

---

# 9. Generics vs `interface{}`

Before generics, Go often used `interface{}` / `any` to create "generic" structures.

For example:

```go
func PrintAll(values []any) {
    for _, value := range values {
        fmt.Println(value)
    }
}
```

You can put anything in it:

```go
values := []any{10, "hello", 3.14}
```

But you lose type information.

Generics preserve it.

```go
func First[T any](values []T) T {
    return values[0]
}
```

Now:

```go
numbers := []int{10, 20, 30}

x := First(numbers)
```

Go knows:

```text
x → int
```

With `[]any`, you'd often need type assertions:

```go
x := values[0].(int)
```

Generics avoid that.

---

# 10. Generics vs interfaces

This distinction is especially useful given your recent questions about Go interfaces.

### Interface polymorphism

You have different **types that implement behavior**:

```go
type Animal interface {
    Speak()
}

func MakeSound(a Animal) {
    a.Speak()
}
```

The focus is:

> **What can this value do?**

---

### Generics

You have code that works with different **types**:

```go
func Reverse[T any](items []T) []T {
    // ...
}
```

The focus is:

> **What type is this value?**

The compiler preserves the relationship between the input and output types.

For example:

```go
func First[T any](items []T) T
```

means:

```text
[]int    → int
[]string → string
[]User   → User
```

That's something you'd lose with a plain `any`-based API.

---

# 11. A useful mental model

Think of ordinary functions as:

```text
Function
   ↓
fixed types
```

For example:

```go
func Add(a int, b int) int
```

Generics turn this into:

```text
Generic Function
       ↓
   type parameter T
       ↓
 ┌─────┼─────┐
 int string User
```

The **algorithm stays the same**, while the type varies.

---

# 12. When should you use generics?

Generics are particularly useful when:

* the same algorithm works for many types
* you are implementing reusable data structures
* you need type-safe collections/utilities
* the relationship between input and output types matters

Examples:

```go
func Map[T, U any](...)
func Filter[T any](...)
func First[T any](...)
type Stack[T any] struct {...}
type Queue[T any] struct {...}
```

But don't use generics merely because you *can*.

If your function only needs behavior, an interface is often cleaner:

```go
func Save(s Storage)
```

rather than trying to parameterize everything.

---

## The big picture

You can think of the three concepts like this:

| Mechanism         | Main question                          |
| ----------------- | -------------------------------------- |
| **Concrete type** | "What exact type is this?"             |
| **Interface**     | "What can this type do?"               |
| **Generic**       | "What type is this code operating on?" |

And the key syntax to recognize is:

```go
func Function[T Constraint](value T) T
```

where:

* `T` → **type parameter**
* `Constraint` → restricts what `T` can be
* `value T` → value has type `T`
* `) T` → function returns the same type `T`

So:

```go
func Max[T int | float64](a, b T) T
```

essentially means:

> **"Give me two values of some type `T`, provided `T` is an `int` or `float64`, and I'll return a value of that same type."**
