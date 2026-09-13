### Meu resumo:

Um tipo tem o seu conjunto de métodos (method set), os value receivers. O ponteiro desse tipo tem o conjunto de métodos do tipo base + os métodos que são pointer receivers.

T method set -> value receivers
*T method set -> value receivers + pointer receivers

Se um método é implementado como pointer receiver, quem satisfaz a interface que tem esse método é o ponteiro do tipo e não o tipo base.

# > PPPPPPPPPPPPPPPPPPPPPPP

Explain:

**the method set of a pointer type** **`*T`** **includes methods with both value and pointer receivers; the method set of a value type** **`T`** **includes only methods with value receivers.**

# > PPPPPPPPPPPPPPPPPPPPPPP

This is about **which methods belong to the method set of `T` versus `*T`**, and it matters especially when you use interfaces.

The easiest way to understand it is with receivers.

## 1. Two kinds of methods

Suppose we have:

```go
type Dog struct {
    Name string
}
```

A method can have a **value receiver**:

```go
func (d Dog) Speak() {
    fmt.Println(d.Name, "says woof")
}
```

Or a **pointer receiver**:

```go
func (d *Dog) Rename(name string) {
    d.Name = name
}
```

So we have:

```text
Dog method set:
    Speak()       ← value receiver

*Dog method set:
    Speak()       ← value receiver
    Rename(...)   ← pointer receiver
```

That's exactly what the statement means.

---

# 2. Why does `*T` get both?

Think about what a pointer to `T` can do.

If you have:

```go
d := Dog{Name: "Rex"}
p := &d
```

Then:

```text
d  → Dog value
p  → *Dog pointer
```

A `*Dog` can access methods defined on `Dog`:

```go
p.Speak()
```

and methods defined on `*Dog`:

```go
p.Rename("Max")
```

So:

```go
*Dog
├── Speak()        // value receiver
└── Rename(...)    // pointer receiver
```

Conceptually, a pointer can use **everything a value can use, plus pointer-receiver methods**.

---

# 3. But `T` cannot get pointer-receiver methods

Now consider:

```go
d := Dog{Name: "Rex"}

d.Speak()        // OK
d.Rename("Max")   // also OK!
```

This is where Go can be confusing.

You might think:

> "Wait! If `d.Rename()` works, doesn't that mean `Rename` is in the method set of `Dog`?"

**No.**

Go performs some automatic address-taking when you call methods.

Because `d` is an addressable variable, Go effectively allows:

```go
d.Rename("Max")
```

to behave like:

```go
(&d).Rename("Max")
```

But that **doesn't change the method set of `Dog`**.

The method set is still:

```text
Dog:
    Speak()

*Dog:
    Speak()
    Rename()
```

This distinction becomes extremely important with interfaces.

---

> Ou seja, acho que essa justificativa toda nem tinha motivo. Só aceita que o method set de *T tem os value receivers também.

# 4. Interfaces reveal the difference

Suppose:

```go
type Speaker interface {
    Speak()
}
```

And:

```go
func (d Dog) Speak() {
    fmt.Println("Woof")
}
```

Both `Dog` and `*Dog` satisfy `Speaker`:

```go
var s1 Speaker = Dog{}
var s2 Speaker = &Dog{}
```

Because:

```text
Dog  → Speak()
*Dog → Speak()
```

Now suppose the interface requires the pointer-receiver method:

```go
type Renamer interface {
    Rename(string)
}
```

And:

```go
func (d *Dog) Rename(name string) {
    d.Name = name
}
```

Then:

```go
var r1 Renamer = &Dog{} // OK
```

> O pointer type (*T) satisfaz a interface, pois o pointer type tem os métodos da interface no seu method set

but:

```go
var r2 Renamer = Dog{}  // ERROR
```

> O type (T) não satisfaz a interface, pois o type não tem os métodos da interface no seu method set

Why?

Because:

```text
Dog method set:
    Speak()

*Dog method set:
    Speak()
    Rename()
```

Therefore only `*Dog` satisfies `Renamer`.

---

# 5. This is the important rule

Memorize this table:

| Receiver               | Method set of `T` | Method set of `*T` |
| ---------------------- | ----------------: | -----------------: |
| `func (t T) Method()`  |                 ✅ |                  ✅ |
| `func (t *T) Method()` |                 ❌ |                  ✅ |

Or simply:

> **`T` gets value-receiver methods. `*T` gets value-receiver + pointer-receiver methods.**

---

# 6. Why does Go design it this way?

Because a pointer and a value have different capabilities.

A value:

```go
Dog{}
```

doesn't inherently represent an object whose underlying storage you're modifying through that value.

A pointer:

```go
&Dog{}
```

explicitly refers to the underlying `Dog`, so it can use methods that operate on that object.

For example:

```go
func (d *Dog) Rename(name string) {
    d.Name = name
}
```

The pointer receiver is necessary because the method intends to modify the original `Dog`.

Compare:

```go
func (d Dog) Rename(name string) {
    d.Name = name
}
```

Here `d` is a **copy**. Changing `d.Name` doesn't modify the original.

---

# 7. The confusing part: method calls vs method sets

This is probably the most important distinction to understand.

### Method call

Go can automatically take the address of an addressable value:

```go
d.Rename("Max")
```

roughly becomes:

```go
(&d).Rename("Max")
```

### Method set

Go **doesn't** add pointer-receiver methods to `Dog`'s method set just because it can automatically take its address.

Therefore:

```go
d.Rename("Max")       // potentially OK
```

but:

```go
var r Renamer = d     // NOT OK
```

while:

```go
var r Renamer = &d    // OK
```

This apparent contradiction disappears once you separate:

**"Can I call this method on this expression?"**

from

**"Is this method part of this type's method set?"**

---

## 8. Why this matters for interfaces

This is one of the most common Go interface gotchas.

Consider:

```go
type Speaker interface {
    Speak()
}

type Dog struct{}

func (d *Dog) Speak() {
    fmt.Println("Woof")
}
```

Because `Speak` has a pointer receiver:

```text
Dog:
    (nothing relevant)

*Dog:
    Speak()
```

Therefore:

```go
var s Speaker = &Dog{} // ✅
```

but:

```go
var s Speaker = Dog{}  // ❌
```

Even though you might be able to write:

```go
dog := Dog{}
dog.Speak()
```

because Go can automatically take `&dog` for the method call.

---

### A good mental model

Think of the relationship as:

```text
             value receiver
                   │
                   ▼
              ┌─────────┐
              │    T    │
              └────┬────┘
                   │
                   │ pointer adds capabilities
                   ▼
              ┌─────────┐
              │   *T    │
              └─────────┘
              value + pointer
                receivers
```

So **`*T` has the larger method set**.

And because **interface satisfaction is based on method sets**, choosing a pointer receiver can determine whether `T` or only `*T` satisfies an interface.

