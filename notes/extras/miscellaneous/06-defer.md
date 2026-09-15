In Go, **`defer` schedules a function call to run later**, when the surrounding function is about to return.

It is mainly used for **cleanup**: closing files, unlocking mutexes, closing connections, etc.

### 1. Basic example

```go
func main() {
	fmt.Println("start")

	defer fmt.Println("cleanup")

	fmt.Println("end")
}
```

Output:

```text
start
end
cleanup
```

The important idea is:

```go
defer fmt.Println("cleanup")
```

means roughly:

> "Don't execute this function call now. Execute it when `main()` is about to finish."

---

### 2. `defer` executes even when returning early

```go
func example() {
	defer fmt.Println("cleanup")

	fmt.Println("doing something")

	return

	fmt.Println("this never runs")
}
```

Output:

```text
doing something
cleanup
```

So `defer` is useful when you have multiple possible ways to leave a function.

---

### 3. Common use: closing a file

```go
func readFile() {
	file, err := os.Open("data.txt")
	if err != nil {
		return
	}

	defer file.Close()

	// work with file...
}
```

The sequence is:

```text
os.Open()
    ↓
defer file.Close()    ← schedule cleanup
    ↓
work with file
    ↓
function returns
    ↓
file.Close()          ← executes automatically
```

This is much safer than manually remembering to close the file at every return point.

---

### 4. Multiple `defer`s: LIFO order

This is particularly important.

```go
func main() {
	defer fmt.Println("first")
	defer fmt.Println("second")
	defer fmt.Println("third")
}
```

Output:

```text
third
second
first
```

`defer` follows **LIFO**:

> **Last In, First Out**

You can think of it as a stack:

```text
defer first
defer second
defer third

       ↓

┌─────────────┐
│ third       │ ← executed first
├─────────────┤
│ second      │
├─────────────┤
│ first       │ ← executed last
└─────────────┘
```

This is useful for nested resources:

```go
file1, _ := os.Open("file1.txt")
defer file1.Close()

file2, _ := os.Open("file2.txt")
defer file2.Close()
```

When the function ends:

```text
file2.Close()
file1.Close()
```

---

### 5. Arguments are evaluated immediately

This is a subtle but important property.

```go
func main() {
	x := 10

	defer fmt.Println(x)

	x = 20
}
```

Output:

```text
10
```

Why?

When Go encounters:

```go
defer fmt.Println(x)
```

it evaluates the argument `x` **right then**.

Conceptually:

```go
defer fmt.Println(10)
```

is scheduled.

It does **not** wait and then evaluate `x`.

---

### 6. But closures behave differently

Compare this:

```go
func main() {
	x := 10

	defer func() {
		fmt.Println(x)
	}()

	x = 20
}
```

Output:

```text
20
```

Here, the deferred function is a **closure** that refers to `x`.

The closure captures the variable itself, so when it eventually runs, `x` is `20`.

Compare:

```go
defer fmt.Println(x)
```

with:

```go
defer func() {
	fmt.Println(x)
}()
```

The first evaluates `x` immediately; the second evaluates `x` when the deferred function executes.

---

### 7. `defer` and return

`defer` happens **after the return value has been determined but before the function actually returns to its caller**.

For example:

```go
func example() int {
	defer fmt.Println("deferred")

	return 42
}
```

Conceptually:

```text
determine return value → 42
        ↓
run deferred functions
        ↓
return 42 to caller
```

This becomes especially interesting with **named return values**:

```go
func example() (result int) {
	defer func() {
		result++
	}()

	return 42
}
```

The result is:

```text
43
```

because `result` is a named return variable. The `return 42` assigns `42` to `result`, then the deferred function changes it to `43`, and only then does the function return.

---

### 8. `defer` is not a timer

This is important given your recent questions about `select`, `time.After`, and `time.Sleep`.

`defer` does **not** mean:

> "run this after some amount of time."

It means:

> **"run this when the current function is leaving."**

For example:

```go
func main() {
	defer fmt.Println("done")

	time.Sleep(5 * time.Second)

	fmt.Println("working")
}
```

Output after 5 seconds:

```text
working
done
```

The `defer` waited because **the function was still running**, not because `defer` itself waits.

---

### The mental model

When you see:

```go
defer someFunction()
```

think:

```text
Execute someFunction later,
when the current function is about to return.
```

And remember these four rules:

1. **Deferred calls don't execute immediately.**
2. **They execute when the surrounding function returns.**
3. **Multiple defers execute in reverse order (LIFO).**
4. **Function arguments are evaluated when the `defer` statement is reached.**

For Go, `defer` is one of the key mechanisms for writing reliable **resource cleanup and concurrency code**.
