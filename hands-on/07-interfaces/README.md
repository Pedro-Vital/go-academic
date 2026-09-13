# Notes+Todo App

**This is a notes-taking and todos-taking app that uses structs and JSON for storing data.**

**We are using interfaces to create a better code.**

# Doubt

There is only 1:

### > PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP

Can't I add the error handling to the outputData function? Example:

func outputData(data outputtable) {
	data.Display()
	err := saveData(data)

	if err != nil {
		return
	}
}

### > PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP

You **can**, but then you change the contract of `outputData` and lose the ability for `main()` to know that `saveData` failed.

The key difference is **who is responsible for handling the error**.

### Current version

```go
func outputData(data outputtable) error {
    data.Display()
    return saveData(data)
}
```

Here, `outputData` **propagates** the error:

```text
saveData()
    ↓ error
outputData()
    ↓ error
main()
    ↓ handles it
```

So in `main`:

```go
err = outputData(todo)

if err != nil {
    return
}
```

`main` can decide what to do if saving fails.

---

### Your version

```go
func outputData(data outputtable) {
    data.Display()

    err := saveData(data)

    if err != nil {
        return
    }
}
```

This **handles/consumes** the error inside `outputData`.

But there's an important problem:

```go
if err != nil {
    return
}
```

doesn't tell the caller that anything went wrong.

For example:

```go
err := saveData(data)

if err != nil {
    return
}
```

The function simply exits. From `main`'s perspective:

```go
outputData(todo)

fmt.Println("Everything worked!")
```

would be perfectly possible—even though `saveData` failed.

### You could print the error there

For example:

```go
func outputData(data outputtable) {
    data.Display()

    err := saveData(data)

    if err != nil {
        fmt.Println("Error:", err)
        return
    }
}
```

Now the error is handled, but `main` still doesn't know about it.

---

## Why return the error instead?

A common Go pattern is:

> **A function that encounters an error should return it to the caller unless it is the appropriate layer to handle it.**

`saveData` knows **what went wrong**, but `outputData` may not know what the application should do about it.

So:

```go
func outputData(data outputtable) error {
    data.Display()
    return saveData(data)
}
```

means:

> "I attempted the operation. If something went wrong, I'm giving the error to whoever called me."

Then `main` decides:

```go
err = outputData(todo)

if err != nil {
    fmt.Println("Error:", err)
    return
}
```

This separation is useful because `outputData` can be reused:

```go
if err := outputData(todo); err != nil {
    // maybe retry
}

if err := outputData(userNote); err != nil {
    // maybe log differently
}
```

---

### One subtle point

You **could** handle it inside `outputData` and return nothing, but then you're saying:

```go
func outputData(data outputtable)
```

has the semantic contract:

> "I'll try to save this, and I'll deal with any failure myself."

Whereas:

```go
func outputData(data outputtable) error
```

says:

> "I'll try to save this, and I'll report whether it succeeded."

For application/business logic, the second approach is generally much more flexible.

And notice that this isn't about Go's `error` being an exception. **Errors are ordinary return values in Go**, so returning the error is the mechanism for passing failure information up the call stack.

